## Overview

Cloud Run is a serverless container platform that scales from zero to N replicas automatically. Each service runs one container image per revision; traffic can be split across revisions for canary or blue/green releases. Services are regional; no cluster management required.

## Common Commands

### Deploy and Manage Services

```bash
# Deploy from Container Registry (public traffic)
gcloud run deploy my-service \
  --image=gcr.io/my-project/my-app:latest \
  --region=us-central1 \
  --allow-unauthenticated \
  --memory=512Mi --cpu=1 \
  --max-instances=10

# Deploy (authenticated — Cloud IAM controls access)
gcloud run deploy my-service \
  --image=gcr.io/my-project/my-app:latest \
  --region=us-central1 \
  --no-allow-unauthenticated

# List services
gcloud run services list --region=us-central1

# Describe service (URL, latest revision, traffic config)
gcloud run services describe my-service --region=us-central1

# Delete service
gcloud run services delete my-service --region=us-central1
```

### Revisions and Traffic Splits

```bash
# Deploy a new revision without sending traffic
gcloud run deploy my-service \
  --image=gcr.io/my-project/my-app:v2 \
  --region=us-central1 \
  --no-traffic \
  --tag=v2-candidate

# Canary: send 10% to new revision
gcloud run services update-traffic my-service \
  --region=us-central1 \
  --to-revisions=my-service-v2=10,LATEST=90

# Promote new revision to 100%
gcloud run services update-traffic my-service \
  --region=us-central1 \
  --to-latest

# List revisions
gcloud run revisions list --service=my-service --region=us-central1

# Delete old revision
gcloud run revisions delete my-service-v1 --region=us-central1
```

### Secrets Integration

```bash
# Create a secret in Secret Manager
echo -n "my-db-password" | gcloud secrets create db-password --data-file=-

# Mount secret as environment variable in Cloud Run
gcloud run deploy my-service \
  --image=gcr.io/my-project/my-app:latest \
  --region=us-central1 \
  --set-secrets=DB_PASSWORD=db-password:latest

# Mount secret as a file
gcloud run deploy my-service \
  --image=gcr.io/my-project/my-app:latest \
  --region=us-central1 \
  --set-secrets=/secrets/db-password=db-password:latest
```

### VPC Connector (Private Network Access)

```bash
# Create VPC connector
gcloud compute networks vpc-access connectors create my-connector \
  --region=us-central1 \
  --subnet=my-subnet \
  --subnet-project=my-project \
  --min-instances=2 --max-instances=10

# Deploy service using VPC connector (route all traffic through VPC)
gcloud run deploy my-service \
  --image=gcr.io/my-project/my-app:latest \
  --region=us-central1 \
  --vpc-connector=my-connector \
  --vpc-egress=all-traffic
```

### Service-to-Service Authentication

```bash
# Grant a Cloud Run service permission to call another service
gcloud run services add-iam-policy-binding target-service \
  --region=us-central1 \
  --member="serviceAccount:caller-sa@my-project.iam.gserviceaccount.com" \
  --role="roles/run.invoker"

# Caller service obtains ID token (in code)
# Via metadata server (inside Cloud Run):
# curl -H "Metadata-Flavor: Google" \
#   "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=https://TARGET_SERVICE_URL"
```

### Environment Variables and Configuration

```bash
# Set environment variables
gcloud run deploy my-service \
  --image=gcr.io/my-project/my-app:latest \
  --region=us-central1 \
  --set-env-vars=APP_ENV=production,LOG_LEVEL=info

# Update env vars on existing service (creates new revision)
gcloud run services update my-service \
  --region=us-central1 \
  --update-env-vars=FEATURE_FLAG=true

# Set concurrency and timeout
gcloud run deploy my-service \
  --image=gcr.io/my-project/my-app:latest \
  --region=us-central1 \
  --concurrency=80 \
  --timeout=300
```

## Patterns

### Blue/Green Deployment

1. Deploy new image with `--no-traffic` and tag: `--tag=green`
2. Test tagged URL: `https://green---my-service-HASH-uc.a.run.app`
3. Shift traffic: `--to-latest` or explicit revision split
4. Delete old revision after monitoring period

### Scale to Zero + Warm-Up

- `--min-instances=0` — scale to zero when idle (lowest cost)
- `--min-instances=1` — keep one instance warm (eliminate cold start)
- Use Cloud Tasks or Cloud Scheduler to ping the service periodically if cold starts are a concern

## Gotchas

- **Service account must exist before deploy** — use `--service-account` to specify a dedicated SA; the Compute default SA is not available to Cloud Run by default
- **IAM for public services still works** — `--allow-unauthenticated` grants `roles/run.invoker` to `allUsers`; you can remove it later without redeploying
- **VPC egress costs** — `--vpc-egress=all-traffic` routes all outbound (including internet) through VPC connector; use `--vpc-egress=private-ranges-only` if you only need private network access
- **Max concurrency default is 80** — if your container is single-threaded, set `--concurrency=1` to avoid overloading it; Cloud Run will spin up more instances instead
- **Revision immutability** — once deployed, a revision's configuration is fixed; to change env vars or CPU, redeploy (a new revision is created automatically)
- **Cold start memory** — memory is proportional to CPU allocation; container image size directly impacts cold start time; use distroless or alpine base images
