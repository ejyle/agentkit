## Overview

Deployment patterns control how new code reaches production with minimal risk. Blue/green and canary are the two primary strategies for zero-downtime releases. Environment promotion gates and rollback procedures complete the lifecycle. GitHub Actions environments with required reviewers enforce human gates before production deploys.

## Deployment Strategies

### Blue/Green Deployment

Two identical environments (blue = current prod, green = new version). Traffic switches instantly; rollback is a traffic re-route, not a code revert.

```yaml
# .github/workflows/deploy-blue-green.yml
jobs:
  deploy-green:
    environment: production
    runs-on: ubuntu-latest
    env:
      GREEN_SERVICE: my-service-green
      PROD_LB: my-prod-load-balancer
    steps:
      - uses: actions/checkout@v4

      # Deploy new version to green (receives no prod traffic yet)
      - name: Deploy to green
        run: |
          aws ecs update-service \
            --cluster prod \
            --service $GREEN_SERVICE \
            --force-new-deployment
          aws ecs wait services-stable \
            --cluster prod \
            --services $GREEN_SERVICE

      # Run smoke test against green
      - name: Smoke test green
        run: |
          ./scripts/smoke-test.sh ${{ env.GREEN_ENDPOINT }}

      # Shift all traffic to green
      - name: Shift traffic to green
        run: |
          aws elbv2 modify-listener \
            --listener-arn ${{ secrets.PROD_LISTENER_ARN }} \
            --default-actions Type=forward,TargetGroupArn=${{ secrets.GREEN_TG_ARN }}

      # Monitor for 5 minutes
      - name: Post-deploy health check
        run: |
          for i in $(seq 1 10); do
            STATUS=$(curl -s -o /dev/null -w "%{http_code}" https://app.example.com/health)
            if [ "$STATUS" != "200" ]; then
              echo "Health check failed (HTTP $STATUS) — initiating rollback"
              exit 1
            fi
            sleep 30
          done

      # On success: swap labels so green becomes blue for next deploy
      - name: Promote green to blue
        if: success()
        run: echo "Green is now the production baseline"

      # On failure: roll back to blue
      - name: Rollback to blue
        if: failure()
        run: |
          aws elbv2 modify-listener \
            --listener-arn ${{ secrets.PROD_LISTENER_ARN }} \
            --default-actions Type=forward,TargetGroupArn=${{ secrets.BLUE_TG_ARN }}
```

### Canary Release (Gradual Rollout)

Incrementally shift traffic to the new version, monitor, then promote or rollback.

```yaml
jobs:
  canary-deploy:
    environment: production
    runs-on: ubuntu-latest
    steps:
      - name: Deploy canary (10% traffic)
        run: |
          # Example: AWS ALB weighted target groups
          aws elbv2 modify-listener \
            --listener-arn ${{ secrets.PROD_LISTENER_ARN }} \
            --default-actions '[
              {"Type":"forward","ForwardConfig":{"TargetGroups":[
                {"TargetGroupArn":"${{ secrets.STABLE_TG_ARN }}","Weight":90},
                {"TargetGroupArn":"${{ secrets.CANARY_TG_ARN }}","Weight":10}
              ]}}
            ]'

      - name: Monitor canary (5 min)
        run: |
          sleep 300
          ERROR_RATE=$(./scripts/get-error-rate.sh ${{ secrets.CANARY_TG_ARN }})
          if (( $(echo "$ERROR_RATE > 1.0" | bc -l) )); then
            echo "Canary error rate $ERROR_RATE% — rolling back"
            exit 1
          fi
          echo "Canary healthy (error rate: $ERROR_RATE%)"

      - name: Promote canary to 100%
        if: success()
        run: |
          aws elbv2 modify-listener \
            --listener-arn ${{ secrets.PROD_LISTENER_ARN }} \
            --default-actions Type=forward,TargetGroupArn=${{ secrets.CANARY_TG_ARN }}

      - name: Rollback canary
        if: failure()
        run: |
          aws elbv2 modify-listener \
            --listener-arn ${{ secrets.PROD_LISTENER_ARN }} \
            --default-actions Type=forward,TargetGroupArn=${{ secrets.STABLE_TG_ARN }}
```

### Rollback Procedures

```yaml
# Manual rollback workflow (triggered by user)
on:
  workflow_dispatch:
    inputs:
      target-sha:
        description: 'SHA to roll back to'
        required: true

jobs:
  rollback:
    environment: production
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ inputs.target-sha }}

      - name: Build rollback image
        run: docker build -t app:rollback .

      - name: Deploy rollback
        run: ./deploy.sh app:rollback

      - name: Verify rollback
        run: |
          DEPLOYED_SHA=$(curl -s https://app.example.com/version | jq -r .sha)
          if [ "$DEPLOYED_SHA" != "${{ inputs.target-sha }}" ]; then
            echo "Rollback verification failed"
            exit 1
          fi
```

### Environment Promotion (dev → staging → production)

```yaml
jobs:
  deploy-dev:
    environment: dev
    runs-on: ubuntu-latest
    steps:
      - run: ./deploy.sh dev ${{ github.sha }}

  smoke-test-dev:
    needs: deploy-dev
    runs-on: ubuntu-latest
    steps:
      - run: ./scripts/smoke-test.sh https://dev.app.example.com

  deploy-staging:
    needs: smoke-test-dev
    environment: staging    # requires staging reviewers to approve
    runs-on: ubuntu-latest
    steps:
      - run: ./deploy.sh staging ${{ github.sha }}

  integration-test-staging:
    needs: deploy-staging
    runs-on: ubuntu-latest
    steps:
      - run: ./scripts/integration-test.sh https://staging.app.example.com

  deploy-production:
    needs: integration-test-staging
    environment: production   # requires production reviewers to approve
    runs-on: ubuntu-latest
    steps:
      - run: ./deploy.sh production ${{ github.sha }}
```

### Health Check Integration Post-Deploy

```bash
#!/usr/bin/env bash
# scripts/health-check.sh <url> <max-attempts> <interval>
URL=$1
MAX=$2
INTERVAL=$3

for i in $(seq 1 "$MAX"); do
  HTTP_CODE=$(curl -sf -o /dev/null -w "%{http_code}" "$URL/health" || echo "000")
  if [ "$HTTP_CODE" = "200" ]; then
    echo "Health check passed (attempt $i)"
    exit 0
  fi
  echo "Attempt $i: HTTP $HTTP_CODE — waiting ${INTERVAL}s"
  sleep "$INTERVAL"
done

echo "Health check failed after $MAX attempts"
exit 1
```

```yaml
- name: Post-deploy health check
  run: ./scripts/health-check.sh https://app.example.com 12 10
  # 12 attempts × 10s = 2 minutes
```

## Patterns

### Deployment Approval Workflow

1. Workflow reaches environment job with required reviewers
2. GitHub sends email/Slack notification to reviewers
3. Reviewer visits GitHub Actions run page and clicks "Approve"
4. Job begins execution within the timeout window (configurable, default 30 days)
5. If reviewer clicks "Reject", workflow fails immediately

Configure in Settings > Environments > required reviewers (up to 6 people or teams).

### Automatic Rollback on Error Rate Spike

Post-deploy, the deployment job monitors error rates via CloudWatch/Datadog/Prometheus for a configurable window. If error rate exceeds the threshold, a rollback script runs automatically and the job exits non-zero (flagging the workflow run as failed for visibility).

## Gotchas

- **Environment approval timeout** — by default, approval requests expire after 30 days; if no reviewer approves, the deployment workflow must be re-run from scratch
- **Concurrent deployments** — use `concurrency: group: deploy-production` with `cancel-in-progress: false` to queue instead of cancel simultaneous deployments
- **Blue/green requires consistent schema** — if a database migration is incompatible with the old version, blue/green breaks during the transition window; use expand-contract (backward-compatible migrations) patterns
- **Canary error rate baseline** — measure your pre-canary error rate first; alert only when the canary's rate exceeds the baseline by a meaningful margin (e.g., 2x)
- **GitHub environment URL** — set `environment.url` in the job config to display a clickable URL in the deployment panel; useful for reviewers to verify before approving the next stage
