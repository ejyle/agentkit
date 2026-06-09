## Overview

Google Kubernetes Engine (GKE) is a managed Kubernetes service on GCP. Clusters can be zonal (single control plane zone) or regional (replicated control plane across 3 zones). Node pools define the VM configuration for worker nodes. The `gcloud container` CLI manages cluster lifecycle; `kubectl` handles workloads after credentials are fetched.

## Common Commands

### Cluster Lifecycle

```bash
# Create zonal cluster
gcloud container clusters create my-cluster \
  --zone=us-central1-a \
  --num-nodes=3 \
  --machine-type=e2-standard-4 \
  --release-channel=regular

# Create regional cluster (HA control plane)
gcloud container clusters create my-cluster \
  --region=us-central1 \
  --num-nodes=2 \
  --machine-type=e2-standard-4 \
  --release-channel=regular

# List clusters
gcloud container clusters list

# Get kubectl credentials
gcloud container clusters get-credentials my-cluster --zone=us-central1-a

# Upgrade cluster control plane
gcloud container clusters upgrade my-cluster \
  --zone=us-central1-a --master

# Delete cluster
gcloud container clusters delete my-cluster --zone=us-central1-a
```

### Node Pools

```bash
# Add a node pool (e.g., GPU or high-memory)
gcloud container node-pools create gpu-pool \
  --cluster=my-cluster --zone=us-central1-a \
  --machine-type=n1-standard-4 \
  --accelerator=type=nvidia-tesla-t4,count=1 \
  --num-nodes=2

# List node pools
gcloud container node-pools list --cluster=my-cluster --zone=us-central1-a

# Resize a node pool
gcloud container clusters resize my-cluster \
  --node-pool=gpu-pool --num-nodes=4 --zone=us-central1-a

# Delete a node pool
gcloud container node-pools delete gpu-pool \
  --cluster=my-cluster --zone=us-central1-a
```

### kubectl After Credential Setup

```bash
# Verify connected cluster
kubectl config current-context
kubectl cluster-info

# Deploy workload
kubectl apply -f deployment.yaml

# Check pod status
kubectl get pods -n my-namespace
kubectl describe pod <pod-name> -n my-namespace

# View logs
kubectl logs <pod-name> -n my-namespace --follow

# Exec into pod
kubectl exec -it <pod-name> -n my-namespace -- /bin/bash

# Port-forward for local debugging
kubectl port-forward svc/my-service 8080:80 -n my-namespace
```

### Horizontal Pod Autoscaler

```bash
# Create HPA based on CPU
kubectl autoscale deployment my-app \
  --cpu-percent=60 --min=2 --max=10 -n my-namespace

# View HPA status
kubectl get hpa -n my-namespace

# HPA via YAML (recommended for GitOps)
cat <<'EOF' | kubectl apply -f -
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-app-hpa
  namespace: my-namespace
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 60
EOF
```

### Workload Identity

Workload Identity lets Kubernetes service accounts act as GCP service accounts without key files.

```bash
# Enable Workload Identity on cluster
gcloud container clusters update my-cluster \
  --zone=us-central1-a \
  --workload-pool=my-project.svc.id.goog

# Create GCP service account
gcloud iam service-accounts create gke-workload-sa

# Grant the GCP SA access to a resource (e.g., Cloud Storage)
gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:gke-workload-sa@my-project.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"

# Allow Kubernetes SA to impersonate the GCP SA
gcloud iam service-accounts add-iam-policy-binding \
  gke-workload-sa@my-project.iam.gserviceaccount.com \
  --role=roles/iam.workloadIdentityUser \
  --member="serviceAccount:my-project.svc.id.goog[my-namespace/my-ksa]"

# Annotate Kubernetes SA
kubectl annotate serviceaccount my-ksa -n my-namespace \
  iam.gke.io/gcp-service-account=gke-workload-sa@my-project.iam.gserviceaccount.com
```

## Patterns

### Cluster Upgrade Rolling Strategy

1. Check current versions: `gcloud container get-server-config --zone=us-central1-a`
2. Upgrade control plane first: `--master` flag
3. Upgrade node pools one at a time with surge upgrade settings
4. Monitor with: `kubectl get nodes --watch`

### Multi-Namespace Environment Isolation

Use separate namespaces for dev/staging/prod within one cluster; use separate clusters for strict isolation. Apply ResourceQuota and LimitRange per namespace.

## Gotchas

- **Zonal vs regional** — zonal clusters have a single control plane that can go down for upgrades; regional clusters survive single-zone outages at higher cost
- **Workload Identity annotation must match** — the Kubernetes SA annotation and the GCP SA IAM binding must use the exact same namespace/SA names; a typo in either causes authentication failures
- **Node pool version drift** — node pools don't auto-upgrade unless configured; set `--enable-autoupgrade` at pool creation time
- **Default namespace pitfalls** — avoid deploying workloads to `default` namespace in production; use dedicated namespaces per team/app
- **HPA requires metrics-server** — GKE includes it by default but custom metrics need Stackdriver adapter or Prometheus adapter installed
- **`gcloud container clusters get-credentials` overwrites kubeconfig** — use `--internal-ip` when running from inside GCP to avoid internet egress
