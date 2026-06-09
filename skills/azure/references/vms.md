## Overview

Azure Virtual Machines (VMs) are IaaS compute resources in a resource group and region. Every VM needs a VNet, subnet, network interface, and OS disk. The `az vm` command handles lifecycle; `az vmss` manages scale sets for auto-scaling workloads.

## Common Commands

### VM Lifecycle

```bash
# Create VM with SSH key auth
az vm create \
  --resource-group my-rg \
  --name my-vm \
  --image Ubuntu2204 \
  --size Standard_D2s_v3 \
  --admin-username azureuser \
  --generate-ssh-keys \
  --vnet-name my-vnet \
  --subnet my-subnet

# Create Windows VM
az vm create \
  --resource-group my-rg \
  --name my-win-vm \
  --image Win2022Datacenter \
  --admin-username adminuser \
  --admin-password "SecureP@ssw0rd!"

# List VMs
az vm list --resource-group my-rg --output table
az vm list --output json | jq '[.[] | {name, location, powerState: .powerState}]'

# Get VM details (power state, IPs)
az vm show --resource-group my-rg --name my-vm --show-details

# Start / stop / deallocate / restart
az vm start --resource-group my-rg --name my-vm
az vm stop --resource-group my-rg --name my-vm
az vm deallocate --resource-group my-rg --name my-vm
az vm restart --resource-group my-rg --name my-vm

# Delete VM (does not delete associated NICs/disks/IPs unless --delete-all-resources)
az vm delete --resource-group my-rg --name my-vm --yes
```

### SSH and RDP Access

```bash
# SSH directly (requires public IP or Bastion)
ssh azureuser@<public-ip>

# SSH via Azure Bastion (no public IP needed)
az network bastion ssh \
  --name my-bastion --resource-group my-rg \
  --target-resource-id $(az vm show -g my-rg -n my-vm --query id -o tsv) \
  --auth-type ssh-key \
  --username azureuser \
  --ssh-key ~/.ssh/id_rsa

# RDP tunnel via Bastion
az network bastion rdp \
  --name my-bastion --resource-group my-rg \
  --target-resource-id $(az vm show -g my-rg -n my-vm --query id -o tsv)
```

### VM Extensions

```bash
# Install custom script extension (run script on VM)
az vm extension set \
  --resource-group my-rg --vm-name my-vm \
  --name CustomScript \
  --publisher Microsoft.Azure.Extensions \
  --settings '{"fileUris": ["https://my-storage.blob.core.windows.net/scripts/setup.sh"], "commandToExecute": "bash setup.sh"}'

# Install Azure Monitor Agent
az vm extension set \
  --resource-group my-rg --vm-name my-vm \
  --name AzureMonitorLinuxAgent \
  --publisher Microsoft.Azure.Monitor

# List installed extensions
az vm extension list --resource-group my-rg --vm-name my-vm --output table

# Run a command directly (no extension needed)
az vm run-command invoke \
  --resource-group my-rg --name my-vm \
  --command-id RunShellScript \
  --scripts "sudo systemctl status nginx"
```

### VM Scale Sets (VMSS)

```bash
# Create scale set
az vmss create \
  --resource-group my-rg \
  --name my-vmss \
  --image Ubuntu2204 \
  --upgrade-policy-mode automatic \
  --instance-count 3 \
  --admin-username azureuser \
  --generate-ssh-keys

# Scale manually
az vmss scale --resource-group my-rg --name my-vmss --new-capacity 5

# Configure autoscale
az monitor autoscale create \
  --resource-group my-rg \
  --resource my-vmss \
  --resource-type Microsoft.Compute/virtualMachineScaleSets \
  --name my-autoscale \
  --min-count 2 --max-count 10 --count 3

# Add CPU-based scale-out rule
az monitor autoscale rule create \
  --resource-group my-rg \
  --autoscale-name my-autoscale \
  --scale out 1 \
  --condition "Percentage CPU > 75 avg 5m"

# Update VMSS image (rolling upgrade)
az vmss update \
  --resource-group my-rg --name my-vmss \
  --set virtualMachineProfile.storageProfile.imageReference.version=latest

az vmss rolling-upgrade start --resource-group my-rg --name my-vmss
```

### Managed Disks

```bash
# Create empty data disk
az disk create \
  --resource-group my-rg --name my-disk \
  --size-gb 128 --sku Premium_LRS

# Attach disk to VM
az vm disk attach \
  --resource-group my-rg --vm-name my-vm \
  --name my-disk

# Detach disk
az vm disk detach --resource-group my-rg --vm-name my-vm --name my-disk

# Create snapshot
az snapshot create \
  --resource-group my-rg --name my-snapshot \
  --source $(az disk show -g my-rg -n my-disk --query id -o tsv)
```

## Patterns

### Golden Image Pipeline

1. Deploy base VM with `az vm create`
2. Provision via `az vm run-command invoke` or Custom Script Extension
3. Generalize: `sudo waagent -deprovision+user -force` (Linux)
4. Capture: `az vm capture` or use Azure Image Builder
5. Store in Shared Image Gallery for multi-region distribution

## Gotchas

- **`az vm stop` vs `az vm deallocate`** — `stop` keeps compute allocated and charges apply; `deallocate` releases the host and stops charges, but dynamic public IPs are released
- **OS disk size cannot shrink** — you can increase OS disk size but not decrease; plan capacity before creation
- **Bastion requires a dedicated subnet** — the subnet must be named `AzureBastionSubnet` with at least /27 CIDR
- **VMSS uniform vs flexible orchestration** — flexible supports mixing VM sizes and offers more resilience; use flexible for new deployments
- **Custom Script Extension logs** — check at `/var/log/azure/custom-script/handler.log` (Linux) or `C:\Packages\Plugins\Microsoft.Compute.CustomScriptExtension\` (Windows)
