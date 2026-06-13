## Overview

Azure Kubernetes Service (AKS) is a managed Kubernetes cluster on Azure. The control plane is managed by Azure; you pay only for worker nodes. Clusters integrate with Entra ID (formerly Azure AD) for RBAC, Azure CNI or Kubenet for networking, and can use Managed Identities for Azure resource access without credential files.

## Common Commands

### Cluster Lifecycle

```bash
# Create cluster with system-assigned managed identity
az aks create \
  --resource-group my-rg \
  --name my-cluster \
  --node-count 3 \
  --node-vm-size Standard_D4s_v3 \
  --enable-managed-identity \
  --generate-ssh-keys

# Create cluster with Azure CNI (required for some features)
az aks create \
  --resource-group my-rg \
  --name my-cluster \
  --network-plugin azure \
  --vnet-subnet-id $(az network vnet subnet show -g my-rg --vnet-name my-vnet -n my-subnet --query id -o tsv) \
  --node-count 3 \
  --generate-ssh-keys

# List clusters
az aks list --resource-group my-rg --output table

# Get kubectl credentials (merges into ~/.kube/config)
az aks get-credentials --resource-group my-rg --name my-cluster

# Get credentials (overwrite existing context)
az aks get-credentials --resource-group my-rg --name my-cluster --overwrite-existing

# Upgrade Kubernetes version
az aks get-upgrades --resource-group my-rg --name my-cluster --output table
az aks upgrade --resource-group my-rg --name my-cluster --kubernetes-version 1.32.0

# Delete cluster
az aks delete --resource-group my-rg --name my-cluster --yes --no-wait
```

### Node Pools

```bash
# Add a node pool
az aks nodepool add \
  --resource-group my-rg \
  --cluster-name my-cluster \
  --name gpupool \
  --node-count 2 \
  --node-vm-size Standard_NC6 \
  --node-taints sku=gpu:NoSchedule

# List node pools
az aks nodepool list --resource-group my-rg --cluster-name my-cluster --output table

# Scale a node pool
az aks nodepool scale \
  --resource-group my-rg --cluster-name my-cluster \
  --name agentpool --node-count 5

# Enable cluster autoscaler on a pool
az aks nodepool update \
  --resource-group my-rg --cluster-name my-cluster \
  --name agentpool \
  --enable-cluster-autoscaler \
  --min-count 2 --max-count 10

# Delete a node pool
az aks nodepool delete \
  --resource-group my-rg --cluster-name my-cluster --name gpupool
```

### RBAC and Entra ID Integration

```bash
# Enable Entra ID (Azure AD) RBAC on existing cluster
az aks update \
  --resource-group my-rg --name my-cluster \
  --enable-azure-rbac \
  --enable-aad

# Grant a user admin access to the cluster
az role assignment create \
  --role "Azure Kubernetes Service RBAC Cluster Admin" \
  --assignee user@example.com \
  --scope $(az aks show -g my-rg -n my-cluster --query id -o tsv)

# Grant namespace-scoped access
az role assignment create \
  --role "Azure Kubernetes Service RBAC Writer" \
  --assignee user@example.com \
  --scope "$(az aks show -g my-rg -n my-cluster --query id -o tsv)/namespaces/my-namespace"

# List role assignments on cluster
az role assignment list \
  --scope $(az aks show -g my-rg -n my-cluster --query id -o tsv) \
  --output table
```

### Managed Identity for Azure Resources

```bash
# Enable workload identity (pod-level identity)
az aks update \
  --resource-group my-rg --name my-cluster \
  --enable-workload-identity \
  --enable-oidc-issuer

# Get the OIDC issuer URL
az aks show --resource-group my-rg --name my-cluster \
  --query "oidcIssuerProfile.issuerUrl" -o tsv

# Create managed identity for workload
az identity create --name my-workload-identity --resource-group my-rg

# Federate the identity with a Kubernetes service account
az identity federated-credential create \
  --name my-federated-credential \
  --identity-name my-workload-identity \
  --resource-group my-rg \
  --issuer $(az aks show -g my-rg -n my-cluster --query "oidcIssuerProfile.issuerUrl" -o tsv) \
  --subject system:serviceaccount:my-namespace:my-ksa

# Grant the managed identity access to a resource
az role assignment create \
  --role "Storage Blob Data Reader" \
  --assignee $(az identity show --name my-workload-identity -g my-rg --query principalId -o tsv) \
  --scope $(az storage account show -n my-storage -g my-rg --query id -o tsv)
```

### Diagnostics

```bash
# View cluster status
az aks show --resource-group my-rg --name my-cluster \
  --query "{provisioningState:provisioningState, kubernetesVersion:kubernetesVersion, powerState:powerState}"

# Enable monitoring (Container Insights)
az aks enable-addons \
  --resource-group my-rg --name my-cluster \
  --addons monitoring \
  --workspace-resource-id $(az monitor log-analytics workspace show -g my-rg -n my-workspace --query id -o tsv)
```

## Patterns

### Zero-Downtime Upgrade

1. Cordon and drain nodes before upgrade (AKS does this automatically with `--upgrade-settings MaxSurge=1`)
2. Set `--upgrade-settings MaxSurge=33%` for faster upgrades on large clusters
3. Use PodDisruptionBudgets on critical workloads to prevent eviction during drain

## Gotchas

- **Node pool OS disk size** — default is 128 GB; increase at creation time (`--node-osdisk-size`); cannot shrink later
- **Azure CNI IP exhaustion** — Azure CNI pre-allocates IPs per node; plan subnet size as `(max_nodes * max_pods_per_node) + buffer`; default is 30 pods/node
- **`az aks get-credentials` requires `kubelogin` for Entra ID** — when AAD integration is enabled, `kubectl` needs `kubelogin convert-kubeconfig -l azurecli` to convert the kubeconfig to use az CLI tokens
- **Stop/Start cluster to save costs** — `az aks stop` and `az aks start` stop/start all node VMs; control plane keeps running; use for dev clusters to reduce costs overnight
- **Private cluster** — API server is not publicly accessible; requires VNet peering or Azure Private Link for kubectl access; plan network topology before creation
