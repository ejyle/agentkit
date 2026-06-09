## Overview

A well-structured CI pipeline validates code quality at every commit, builds artifacts efficiently with caching, and passes them between stages without re-building. This reference covers stage structure, parallelism patterns, artifact management, and caching strategies for GitHub Actions.

## Common Commands

### Build Stage Structure

A standard four-stage pipeline:

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:

jobs:
  # Stage 1: Quality gates (parallel lint + test)
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm run lint

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm test -- --coverage
      - uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage/

  # Stage 2: Build (requires lint + test to pass)
  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm run build
      - uses: actions/upload-artifact@v4
        with:
          name: build-${{ github.sha }}
          path: dist/
          retention-days: 7

  # Stage 3: Package (Docker image)
  package:
    needs: build
    runs-on: ubuntu-latest
    permissions:
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: build-${{ github.sha }}
          path: dist/
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: .
          push: ${{ github.ref == 'refs/heads/main' }}
          tags: ghcr.io/${{ github.repository }}:${{ github.sha }}

  # Stage 4: Deploy (main branch only)
  deploy:
    if: github.ref == 'refs/heads/main'
    needs: package
    environment: staging
    runs-on: ubuntu-latest
    steps:
      - run: echo "Deploy ${{ github.sha }} to staging"
```

### Fan-Out / Fan-In Parallelism

```yaml
jobs:
  # Fan-out: three parallel test suites
  test-unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm test -- --testPathPattern=unit

  test-integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - run: npm test -- --testPathPattern=integration

  test-e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npx playwright test

  # Fan-in: build only when all tests pass
  build:
    needs: [test-unit, test-integration, test-e2e]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm run build
```

### Artifact Passing Between Jobs

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm run build
      - uses: actions/upload-artifact@v4
        with:
          name: dist-${{ github.run_id }}    # unique per run
          path: |
            dist/
            !dist/**/*.map                   # exclude source maps
          retention-days: 1                  # short-lived build artifacts

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
        with:
          name: dist-${{ github.run_id }}
          path: dist/
      - run: ls -la dist/   # verify contents
      - run: ./deploy.sh
```

### Caching Strategies

```yaml
# Simple: via setup action (most common)
- uses: actions/setup-node@v4
  with:
    node-version: '20'
    cache: 'npm'

# Advanced: manual cache with custom key
- uses: actions/cache@v4
  id: npm-cache
  with:
    path: ~/.npm
    key: ${{ runner.os }}-npm-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-npm-

# Run install only if cache missed
- if: steps.npm-cache.outputs.cache-hit != 'true'
  run: npm ci

# Docker layer caching
- uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### Conditional Steps

```yaml
steps:
  # Run only on main branch
  - name: Deploy
    if: github.ref == 'refs/heads/main'
    run: ./deploy.sh

  # Run only on failure
  - name: Notify on failure
    if: failure()
    run: |
      curl -X POST ${{ secrets.SLACK_WEBHOOK }} \
        -H 'Content-type: application/json' \
        -d '{"text":"Build failed: ${{ github.run_url }}"}'

  # Run regardless of prior steps
  - name: Cleanup
    if: always()
    run: rm -rf /tmp/build-*

  # Run only when specific files changed
  - name: Run database migrations
    if: contains(github.event.commits[*].modified, 'migrations/')
    run: npm run migrate
```

### Timeout Configuration

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    timeout-minutes: 20    # job-level timeout
    steps:
      - name: Long test suite
        timeout-minutes: 15    # step-level timeout
        run: npm test
```

## Patterns

### Cached Build with Invalidation on Source Change

```yaml
- uses: actions/cache@v4
  with:
    path: .next/cache
    key: ${{ runner.os }}-nextjs-${{ hashFiles('**/package-lock.json') }}-${{ hashFiles('**/*.ts', '**/*.tsx') }}
    restore-keys: |
      ${{ runner.os }}-nextjs-${{ hashFiles('**/package-lock.json') }}-
      ${{ runner.os }}-nextjs-
```

The key includes a hash of source files, so the cache is invalidated when any TypeScript file changes.

## Gotchas

- **Artifact size limit** — individual artifacts are limited to 2 GB; zip or split large artifacts
- **Cache size limit** — total cache per repository is 10 GB; older entries are evicted when limit is hit; cache `save-always: true` can bypass eviction for critical caches
- **`needs` skips on upstream skip** — if `test` is skipped (e.g., path filter), `build` is also skipped even if it doesn't care about test results; check `needs.test.result != 'failure'` to allow skipped upstream
- **Service containers require `options: --health-*`** — without health check options, the step may start before the service is ready; always add health check for databases
- **`timeout-minutes` default is 360** — a stuck test suite will block your runner for 6 hours; always set a realistic timeout to catch hangs early
