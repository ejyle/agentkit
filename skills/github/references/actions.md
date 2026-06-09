## Overview

GitHub Actions is a CI/CD platform built into GitHub. Workflows are YAML files in `.github/workflows/`. Each workflow has triggers (`on:`), jobs, and steps. The `gh run` and `gh workflow` CLI commands manage runs and workflows without opening the browser.

## Common Commands

### Viewing Runs

```bash
# List recent workflow runs
gh run list
gh run list --limit 20

# Filter by workflow
gh run list --workflow ci.yml
gh run list --workflow "CI Pipeline"

# View run details and step status
gh run view 12345

# View logs for a specific run
gh run view 12345 --log

# View logs for failed steps only
gh run view 12345 --log-failed

# Watch run in real time (streams status)
gh run watch 12345

# Rerun failed jobs only
gh run rerun 12345 --failed

# Rerun all jobs
gh run rerun 12345
```

### Triggering Workflows

```bash
# Manually trigger a workflow (workflow_dispatch trigger required)
gh workflow run ci.yml

# Trigger with inputs
gh workflow run deploy.yml \
  --field environment=staging \
  --field version=v1.2.3

# Trigger on a specific branch
gh workflow run ci.yml --ref feature/my-branch

# List workflows
gh workflow list

# Enable / disable workflow
gh workflow enable ci.yml
gh workflow disable deploy.yml
```

### Secrets and Variables

```bash
# Set a repository secret (prompts for value)
gh secret set MY_SECRET

# Set from stdin
echo "my-secret-value" | gh secret set MY_SECRET

# Set from file
gh secret set PRIVATE_KEY < private.pem

# Set for a specific environment
gh secret set DB_PASSWORD --env staging

# List secrets (names only, values never shown)
gh secret list
gh secret list --env staging

# Delete a secret
gh secret delete MY_SECRET

# Repository variables (non-secret values)
gh variable set APP_VERSION --body "1.2.3"
gh variable list
```

### Artifacts

```bash
# List artifacts for a run
gh run view 12345 --json artifacts

# Download all artifacts from a run
gh run download 12345

# Download specific artifact by name
gh run download 12345 --name my-artifact --dir ./artifacts/

# List artifacts via API (with retention info)
gh api repos/OWNER/REPO/actions/runs/12345/artifacts \
  --jq '.artifacts[] | {name, size_in_bytes, created_at, expires_at}'
```

### Cache Management

```bash
# List caches
gh api repos/OWNER/REPO/actions/caches \
  --jq '.actions_caches[] | {key, size_in_bytes, created_at}'

# Delete a specific cache by key
gh api repos/OWNER/REPO/actions/caches \
  --jq '.actions_caches[] | select(.key | startswith("npm-")) | .id' |
  xargs -I{} gh api repos/OWNER/REPO/actions/caches/{} --method DELETE
```

### Matrix Builds

```yaml
# .github/workflows/test.yml excerpt
jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        node: ['18', '20']
        exclude:
          - os: windows-latest
            node: '18'
        include:
          - os: ubuntu-latest
            node: '20'
            experimental: true
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}
```

### Environment Secrets vs Repository Secrets

```bash
# Create environment (for deployment gates)
gh api repos/OWNER/REPO/environments/production --method PUT \
  --field wait_timer=10 \
  --field reviewers='[{"type":"User","id":12345}]'

# Set environment-specific secret
gh secret set DB_PASSWORD --env production

# Secret precedence: environment > repository > organization
```

## Patterns

### Workflow with Job Dependencies

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm run build
      - uses: actions/upload-artifact@v4
        with:
          name: dist
          path: dist/

  deploy:
    needs: build
    environment: production
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
        with:
          name: dist
      - run: ./deploy.sh
```

### Reusable Workflows

```yaml
# .github/workflows/deploy-reusable.yml
on:
  workflow_call:
    inputs:
      environment:
        required: true
        type: string
    secrets:
      deploy_key:
        required: true

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: ${{ inputs.environment }}
    steps:
      - run: echo "Deploying to ${{ inputs.environment }}"

# Caller workflow
jobs:
  deploy-staging:
    uses: ./.github/workflows/deploy-reusable.yml
    with:
      environment: staging
    secrets:
      deploy_key: ${{ secrets.STAGING_DEPLOY_KEY }}
```

## Gotchas

- **`GITHUB_TOKEN` permissions are minimal by default** — workflows need `permissions:` in the YAML to elevate; e.g., `contents: write` for creating releases, `pull-requests: write` for commenting
- **Secrets not available in fork PRs** — for security, secrets are not passed to workflows triggered by PRs from forks; use `pull_request_target` carefully (runs in base repo context)
- **Concurrency groups prevent duplicate runs** — add `concurrency: group: ${{ github.workflow }}-${{ github.ref }}` to cancel in-progress runs when a new commit is pushed
- **Artifact retention default is 90 days** — reduce with `retention-days: 7` to avoid storage costs on high-frequency workflows
- **`workflow_dispatch` requires the workflow to exist on the default branch** — `gh workflow run` on a branch only works if the workflow file is on the default branch
