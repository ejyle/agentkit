---
name: cicd
description: >
  Use when designing, authoring, or debugging CI/CD pipelines — GitHub Actions workflows,
  build stage structure, deployment strategies (blue/green, canary), or environment promotion flows.
license: Apache-2.0
---

## When to Use

Activate this skill when the task involves:

- Authoring or debugging GitHub Actions workflow YAML files
- Structuring multi-stage pipelines (test → build → package → deploy)
- Setting up deployment environments with approval gates
- Implementing blue/green or canary deployment strategies
- Configuring caching, artifact passing, and job parallelism
- Defining rollback procedures triggered by health check failures
- Setting up environment promotion (dev → staging → production)
- Diagnosing workflow failures: stuck jobs, failing matrix steps, secret issues

## Quick Reference

### Workflow Triggers

```yaml
on:
  push:
    branches: [main, release/**]
    paths-ignore: ['**.md', 'docs/**']
  pull_request:
    branches: [main]
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        options: [staging, production]
        required: true
  schedule:
    - cron: '0 2 * * 1'    # Every Monday at 2am UTC
  release:
    types: [published]
```

### Core Job Structure

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm run build
      - uses: actions/upload-artifact@v4
        with:
          name: build-output
          path: dist/
          retention-days: 7
```

### Deployment Gates (Environment Protection)

```yaml
jobs:
  deploy-production:
    needs: deploy-staging
    environment:
      name: production
      url: https://app.example.com
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
        with:
          name: build-output
      - run: ./deploy.sh production
```

Configure required reviewers in GitHub Settings > Environments.

### Concurrency Control

```yaml
# Cancel in-progress runs when new commit pushed
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

# Allow only one deploy per environment at a time
concurrency:
  group: deploy-${{ github.event.inputs.environment }}
  cancel-in-progress: false   # queue instead of cancel
```

### Matrix Builds

```yaml
strategy:
  fail-fast: false
  matrix:
    os: [ubuntu-latest, windows-latest]
    node: ['18', '20', '22']
    exclude:
      - os: windows-latest
        node: '18'
```

## Reference Files

Load the appropriate reference file for deep-dive tasks:

| Task | Reference file |
|------|---------------|
| YAML schema, reusable workflows, composite actions, job dependencies | `references/github-actions.md` |
| Build stages, parallelism, artifact passing, caching, conditional steps | `references/pipelines.md` |
| Blue/green, canary, rollback, environment promotion, approval gates | `references/deployments.md` |

## Common Gotchas

- **`needs:` creates a hard dependency** — if the depended-on job is skipped, the dependent job is also skipped (not failed); use `if: always()` to run cleanup jobs regardless
- **Secrets unavailable in fork PRs** — secrets are not passed to `pull_request` event from forks; use `pull_request_target` with caution or require contributors to use a separate workflow trigger
- **`cache:` in setup actions vs `actions/cache`** — `setup-node@v4` with `cache: 'npm'` is simpler and sufficient for most cases; reach for `actions/cache` directly only for custom cache key strategies
- **Environment approval gates block all workflows** — if a required reviewer doesn't approve within the timeout, the entire workflow times out and must be re-run; inform reviewers promptly
- **`workflow_dispatch` requires file on default branch** — the workflow must be present on the repository's default branch even when triggered from another branch
