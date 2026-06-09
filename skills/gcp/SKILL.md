---
name: gcp
description: >
  Use when working with Google Cloud Platform infrastructure — provisioning Compute Engine VMs,
  managing GKE clusters, deploying Cloud Run services, configuring IAM service accounts,
  or automating GCP resources via the gcloud CLI.
license: Apache-2.0
---

## When to Use

Activate this skill when the task involves:

- Creating or managing Compute Engine VM instances, instance groups, or images
- Spinning up or scaling GKE (Google Kubernetes Engine) clusters and node pools
- Deploying, routing, or configuring Cloud Run services
- Creating service accounts, binding IAM roles, or setting up Workload Identity Federation
- Writing or debugging Cloud Build pipelines or triggers
- Configuring VPC networks, firewall rules, or Cloud NAT
- Diagnosing gcloud CLI errors (`PERMISSION_DENIED`, `RESOURCE_NOT_FOUND`, quota limits)

## Quick Reference

### Authentication & Project Setup

```bash
# Login and set default project
gcloud auth login
gcloud config set project my-project-id

# Activate a service account key
gcloud auth activate-service-account --key-file sa-key.json

# Application Default Credentials (for SDKs)
gcloud auth application-default login

# List and switch configs
gcloud config configurations list
gcloud config configurations activate my-config
```

### Compute Engine

```bash
# Create VM
gcloud compute instances create my-vm \
  --zone=us-central1-a --machine-type=e2-medium \
  --image-family=debian-12 --image-project=debian-cloud

# List instances
gcloud compute instances list

# SSH into instance
gcloud compute ssh my-vm --zone=us-central1-a

# Start / stop / delete
gcloud compute instances start my-vm --zone=us-central1-a
gcloud compute instances stop my-vm --zone=us-central1-a
gcloud compute instances delete my-vm --zone=us-central1-a
```

### GKE

```bash
# Create cluster
gcloud container clusters create my-cluster \
  --zone=us-central1-a --num-nodes=3 --machine-type=e2-standard-4

# Get kubectl credentials
gcloud container clusters get-credentials my-cluster --zone=us-central1-a

# List clusters
gcloud container clusters list

# Delete cluster
gcloud container clusters delete my-cluster --zone=us-central1-a
```

### Cloud Run

```bash
# Deploy a container image
gcloud run deploy my-service \
  --image=gcr.io/my-project/my-app:latest \
  --region=us-central1 --allow-unauthenticated

# List services
gcloud run services list --region=us-central1

# Describe service (get URL, traffic splits)
gcloud run services describe my-service --region=us-central1

# Delete service
gcloud run services delete my-service --region=us-central1
```

### IAM

```bash
# Create service account
gcloud iam service-accounts create my-sa \
  --description="My service account" --display-name="My SA"

# Bind a role to the service account at project level
gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:my-sa@my-project.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"

# List IAM bindings for the project
gcloud projects get-iam-policy my-project --format=table
```

## Reference Files

Load the appropriate reference file for deep-dive tasks:

| Task | Reference file |
|------|---------------|
| VM lifecycle, managed instance groups, startup scripts, networking | `references/compute.md` |
| GKE cluster management, node pools, Workload Identity, autoscaling | `references/gke.md` |
| Cloud Run deploy, traffic splits, secrets, VPC connector | `references/cloudrun.md` |
| Service accounts, IAM bindings, Workload Identity Federation | `references/iam.md` |

## Common Gotchas

- **Project ID vs Project Number** — most commands accept either; use `gcloud config set project` to avoid repeating `--project` on every call
- **Zone vs Region** — Compute Engine resources are zonal; GKE can be zonal or regional; Cloud Run is regional; always specify `--zone` or `--region` explicitly
- **API enablement** — new projects need APIs enabled before use: `gcloud services enable container.googleapis.com`
- **Application Default Credentials scope** — `gcloud auth login` is for the CLI only; SDKs need `gcloud auth application-default login` to use your user credentials
- **Quota errors (`RESOURCE_EXHAUSTED`)** — increase via Cloud Console > IAM & Admin > Quotas; or request increase via `gcloud alpha quotas update`
