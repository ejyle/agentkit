## Overview

GCP IAM grants access to GCP resources via a binding: `member + role → resource`. Roles are collections of permissions. The three member types are Google accounts, service accounts, and groups. Service accounts are both identities (for workloads) and resources (that humans and other services can impersonate).

## Common Commands

### Service Accounts

```bash
# Create service account
gcloud iam service-accounts create my-sa \
  --description="App backend SA" \
  --display-name="My App Backend"

# List service accounts
gcloud iam service-accounts list

# Create and download a key (avoid key files in production; prefer Workload Identity)
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=my-sa@my-project.iam.gserviceaccount.com

# Activate key for gcloud commands
gcloud auth activate-service-account --key-file=sa-key.json

# Delete service account (does not delete bindings — clean those up first)
gcloud iam service-accounts delete my-sa@my-project.iam.gserviceaccount.com
```

### IAM Bindings — Project Level

```bash
# Add a role binding
gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:my-sa@my-project.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"

# Add binding for a user
gcloud projects add-iam-policy-binding my-project \
  --member="user:alice@example.com" \
  --role="roles/viewer"

# Add binding for a group
gcloud projects add-iam-policy-binding my-project \
  --member="group:devs@example.com" \
  --role="roles/editor"

# Remove a binding
gcloud projects remove-iam-policy-binding my-project \
  --member="serviceAccount:my-sa@my-project.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"

# Get current policy
gcloud projects get-iam-policy my-project --format=json
```

### IAM Bindings — Resource Level

```bash
# Grant access to a specific Cloud Storage bucket
gcloud storage buckets add-iam-policy-binding gs://my-bucket \
  --member="serviceAccount:my-sa@my-project.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"

# Grant access to a specific Cloud Run service
gcloud run services add-iam-policy-binding my-service \
  --region=us-central1 \
  --member="serviceAccount:caller-sa@my-project.iam.gserviceaccount.com" \
  --role="roles/run.invoker"

# Grant access to a Pub/Sub topic
gcloud pubsub topics add-iam-policy-binding my-topic \
  --member="serviceAccount:my-sa@my-project.iam.gserviceaccount.com" \
  --role="roles/pubsub.publisher"
```

### Custom Roles

```bash
# List available permissions for a service
gcloud iam list-testable-permissions //cloudresourcemanager.googleapis.com/projects/my-project \
  --filter="name:storage"

# Create custom role from YAML
cat > custom-role.yaml <<'EOF'
title: "Custom Storage Reader"
description: "Read objects but not list buckets"
stage: GA
includedPermissions:
  - storage.objects.get
  - storage.objects.list
EOF

gcloud iam roles create customStorageReader \
  --project=my-project \
  --file=custom-role.yaml

# List custom roles
gcloud iam roles list --project=my-project

# Update custom role (add permission)
gcloud iam roles update customStorageReader \
  --project=my-project \
  --add-permissions=storage.objects.create
```

### Workload Identity Federation

Allows external identities (GitHub Actions, AWS, Azure) to impersonate GCP service accounts without a service account key.

```bash
# Create Workload Identity Pool
gcloud iam workload-identity-pools create my-pool \
  --location=global \
  --description="My WIF pool"

# Create OIDC provider in the pool (example: GitHub Actions)
gcloud iam workload-identity-pools providers create-oidc github-provider \
  --location=global \
  --workload-identity-pool=my-pool \
  --issuer-uri=https://token.actions.githubusercontent.com \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository == 'myorg/myrepo'"

# Allow the external identity to impersonate a GCP SA
gcloud iam service-accounts add-iam-policy-binding \
  my-sa@my-project.iam.gserviceaccount.com \
  --role=roles/iam.workloadIdentityUser \
  --member="principalSet://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/my-pool/attribute.repository/myorg/myrepo"
```

### Impersonation

```bash
# Test if a service account has a permission
gcloud iam service-accounts get-iam-policy my-sa@my-project.iam.gserviceaccount.com

# Impersonate SA for a gcloud command
gcloud storage ls --impersonate-service-account=my-sa@my-project.iam.gserviceaccount.com
```

## Patterns

### Least-Privilege SA Setup

1. Create a dedicated SA for each workload
2. Start with predefined roles, then narrow to custom roles if needed
3. Bind at the lowest resource level possible (bucket > project > org)
4. Audit with: `gcloud asset search-all-iam-policies --scope=projects/my-project`

## Gotchas

- **`roles/owner` and `roles/editor` are overly broad** — avoid for service accounts; create custom roles or use predefined ones like `roles/storage.objectAdmin`
- **Service account key files are credentials** — treat them like passwords; prefer Workload Identity or Attached Service Account (for GCE/GKE/Cloud Run) to eliminate key files entirely
- **IAM propagation delay** — new bindings take up to 60 seconds to propagate globally; don't assume immediate effect in automation scripts
- **`add-iam-policy-binding` is additive** — it reads the current policy, adds the binding, and writes it back (read-modify-write); in high-concurrency scenarios, use a retry loop to handle conflicts
- **allUsers vs allAuthenticatedUsers** — `allUsers` means truly public (no login); `allAuthenticatedUsers` means any authenticated Google account (not just your org)
- **Principal hierarchy** — a binding at the organization level inherits down to all projects/resources; be careful granting roles at org level
