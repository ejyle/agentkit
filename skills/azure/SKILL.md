---
name: azure
description: >
  Use when working with Microsoft Azure infrastructure — provisioning VMs, managing AKS clusters,
  deploying App Service web apps, configuring RBAC, or automating Azure resources via the az CLI.
license: Apache-2.0
---

## When to Use

Activate this skill when the task involves:

- Creating or managing Azure Virtual Machines or VM Scale Sets
- Spinning up, scaling, or upgrading AKS (Azure Kubernetes Service) clusters
- Deploying or configuring Azure App Service web apps and Function Apps
- Setting up Azure role assignments (RBAC), Managed Identities, or Entra ID app registrations
- Working with Azure Container Registry, Azure Storage, or Azure Key Vault
- Writing or debugging Azure Pipelines, GitHub Actions targeting Azure, or Bicep/ARM templates
- Diagnosing az CLI errors (`AuthorizationFailed`, `ResourceNotFound`, subscription context issues)

## Quick Reference

### Authentication & Subscription

```bash
# Interactive login
az login

# Login with service principal
az login --service-principal -u <app-id> -p <password> --tenant <tenant-id>

# List subscriptions
az account list --output table

# Set active subscription
az account set --subscription "My Subscription Name"

# Show current context
az account show
```

### Virtual Machines

```bash
# Create VM
az vm create \
  --resource-group my-rg \
  --name my-vm \
  --image Ubuntu2204 \
  --admin-username azureuser \
  --generate-ssh-keys

# List VMs
az vm list --resource-group my-rg --output table

# Start / stop / deallocate
az vm start --resource-group my-rg --name my-vm
az vm stop --resource-group my-rg --name my-vm       # stopped but still billed
az vm deallocate --resource-group my-rg --name my-vm  # stopped, not billed
```

### AKS

```bash
# Create AKS cluster
az aks create \
  --resource-group my-rg \
  --name my-cluster \
  --node-count 3 \
  --generate-ssh-keys

# Get kubectl credentials
az aks get-credentials --resource-group my-rg --name my-cluster

# Scale node count
az aks scale --resource-group my-rg --name my-cluster --node-count 5

# Delete cluster
az aks delete --resource-group my-rg --name my-cluster
```

### App Service

```bash
# Create App Service plan and web app
az appservice plan create \
  --name my-plan --resource-group my-rg \
  --sku S1 --is-linux

az webapp create \
  --resource-group my-rg \
  --plan my-plan \
  --name my-app \
  --runtime "NODE:20-lts"

# Deploy from local ZIP
az webapp deploy --resource-group my-rg --name my-app \
  --src-path ./dist.zip --type zip

# View logs
az webapp log tail --resource-group my-rg --name my-app
```

## Reference Files

Load the appropriate reference file for deep-dive tasks:

| Task | Reference file |
|------|---------------|
| VM lifecycle, extensions, scale sets, Bastion | `references/vms.md` |
| AKS cluster management, node pools, RBAC, AD integration | `references/aks.md` |
| App Service deploy, slots, autoscale, GitHub Actions | `references/appservice.md` |

## Common Gotchas

- **Resource groups required** — almost every Azure resource belongs to a resource group; create one first with `az group create --name my-rg --location eastus`
- **Subscription context** — `az account set` must be run before resource operations; scripts should always set subscription explicitly
- **Deallocate vs stop** — `az vm stop` keeps the VM allocated (you're still charged for compute); `az vm deallocate` releases the compute allocation (cheaper but IP may change unless static)
- **RBAC propagation** — new role assignments take up to 5 minutes to take effect; automation should retry or wait
- **`az login` device flow in CI** — use a service principal or managed identity in CI pipelines; interactive login is blocked in non-interactive shells
