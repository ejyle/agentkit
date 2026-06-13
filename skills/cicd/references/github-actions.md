## Overview

GitHub Actions workflows live in `.github/workflows/*.yml`. Each workflow defines event triggers, one or more jobs, and steps within each job. Jobs run in parallel by default; use `needs:` for sequencing. Reusable workflows and composite actions reduce duplication across repos.

## Common Commands

### YAML Workflow Schema

```yaml
name: CI Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read
  pull-requests: write

env:
  NODE_ENV: test

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Lint
        run: npm run lint

  test:
    name: Test
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm test
        env:
          DATABASE_URL: ${{ secrets.TEST_DATABASE_URL }}
```

### Reusable Workflows (workflow_call)

```yaml
# .github/workflows/_build.yml (reusable, prefixed with _ by convention)
on:
  workflow_call:
    inputs:
      node-version:
        type: string
        default: '20'
      environment:
        type: string
        required: true
    outputs:
      artifact-name:
        description: "Name of the uploaded artifact"
        value: ${{ jobs.build.outputs.artifact-name }}
    secrets:
      NPM_TOKEN:
        required: false

jobs:
  build:
    runs-on: ubuntu-latest
    outputs:
      artifact-name: ${{ steps.set-name.outputs.name }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ inputs.node-version }}
      - run: npm ci
        env:
          NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
      - run: npm run build
      - id: set-name
        run: echo "name=build-${{ github.sha }}" >> $GITHUB_OUTPUT
      - uses: actions/upload-artifact@v4
        with:
          name: ${{ steps.set-name.outputs.name }}
          path: dist/

# Caller
jobs:
  build-app:
    uses: ./.github/workflows/_build.yml
    with:
      environment: staging
    secrets:
      NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
```

### Composite Actions

```yaml
# .github/actions/setup-env/action.yml (composite action)
name: Setup Environment
description: Checkout, setup Node, install deps

inputs:
  node-version:
    description: Node version
    default: '20'
  working-directory:
    description: Working directory
    default: '.'

runs:
  using: composite
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: ${{ inputs.node-version }}
        cache: 'npm'
        cache-dependency-path: ${{ inputs.working-directory }}/package-lock.json
    - run: npm ci
      shell: bash
      working-directory: ${{ inputs.working-directory }}

# Usage in a workflow
steps:
  - uses: ./.github/actions/setup-env
    with:
      node-version: '20'
```

### Job Dependencies (needs:)

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps: [...]

  build:
    needs: test
    runs-on: ubuntu-latest
    steps: [...]

  deploy-staging:
    needs: build
    environment: staging
    runs-on: ubuntu-latest
    steps: [...]

  deploy-production:
    needs: deploy-staging
    environment: production
    runs-on: ubuntu-latest
    steps: [...]

  # Cleanup: run even if deploy fails
  notify:
    needs: [deploy-staging, deploy-production]
    if: always()
    runs-on: ubuntu-latest
    steps:
      - run: echo "Pipeline complete: ${{ needs.deploy-production.result }}"
```

### Environment Protection Rules

Configure in GitHub Settings > Environments > (environment name):

- **Required reviewers** — up to 6 users or teams must approve before job runs
- **Wait timer** — delay in minutes before allowing deployment (e.g., for canary observation)
- **Deployment branches** — restrict which branches can deploy to this environment
- **Secrets** — environment-scoped secrets override repository secrets

```yaml
# Reference environment in job
jobs:
  deploy:
    environment:
      name: production
      url: https://app.example.com   # shown in deployment panel
    runs-on: ubuntu-latest
```

### Workflow Dispatch Inputs

```yaml
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to deploy (e.g. v1.2.3)'
        required: true
        type: string
      environment:
        description: 'Target environment'
        type: choice
        options: [dev, staging, production]
        default: staging
      dry-run:
        description: 'Dry run (no actual changes)'
        type: boolean
        default: false

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: |
          echo "Deploying ${{ inputs.version }} to ${{ inputs.environment }}"
          echo "Dry run: ${{ inputs.dry-run }}"
```

### Output Variables

```yaml
steps:
  - id: get-version
    run: |
      VERSION=$(jq -r .version package.json)
      echo "version=$VERSION" >> $GITHUB_OUTPUT

  - run: echo "Version is ${{ steps.get-version.outputs.version }}"

# Pass outputs between jobs
jobs:
  prepare:
    outputs:
      version: ${{ steps.get-version.outputs.version }}
    steps:
      - id: get-version
        run: echo "version=1.2.3" >> $GITHUB_OUTPUT

  deploy:
    needs: prepare
    steps:
      - run: echo "Deploying ${{ needs.prepare.outputs.version }}"
```

## Patterns

### Workflow Permissions Hardening

Always declare minimum permissions:

```yaml
permissions:
  contents: read   # default; allows checkout
  # Add only what's needed:
  # packages: write  — for GHCR push
  # id-token: write  — for OIDC (e.g., AWS/GCP auth)
  # pull-requests: write  — for PR comments
```

## Gotchas

- **`GITHUB_OUTPUT` replaces deprecated `set-output`** — use `echo "key=value" >> $GITHUB_OUTPUT`; the old `::set-output::` command is deprecated and will be removed
- **`if: always()` vs `if: success()`** — by default, steps skip if a previous step failed; `if: always()` runs regardless; `if: failure()` runs only on failure
- **`needs` propagates skip** — if a required job was skipped (e.g., due to a path filter), dependent jobs are also skipped; check with `needs.job.result == 'success' || needs.job.result == 'skipped'`
- **Workflow YAML merge conflicts** — multiple PRs modifying the same workflow file create merge conflicts; keep workflow files small and use reusable workflows to reduce conflict surface
