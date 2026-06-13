## Overview

Compute Engine provides virtual machines (instances) on Google's infrastructure. Instances live in zones (e.g., `us-central1-a`); zones belong to regions (e.g., `us-central1`). Every instance requires a network interface in a VPC subnet. The `gcloud compute` command covers the full lifecycle.

## Common Commands

### Instance Lifecycle

```bash
# Create with specific image family
gcloud compute instances create my-vm \
  --zone=us-central1-a \
  --machine-type=e2-medium \
  --image-family=debian-12 \
  --image-project=debian-cloud \
  --boot-disk-size=50GB \
  --tags=http-server,https-server \
  --service-account=my-sa@my-project.iam.gserviceaccount.com \
  --scopes=cloud-platform

# Create with startup script
gcloud compute instances create my-vm \
  --zone=us-central1-a \
  --machine-type=e2-medium \
  --image-family=debian-12 --image-project=debian-cloud \
  --metadata-from-file startup-script=startup.sh

# List instances
gcloud compute instances list
gcloud compute instances list --filter="zone:us-central1-a AND status:RUNNING"

# Start / stop / reset
gcloud compute instances start my-vm --zone=us-central1-a
gcloud compute instances stop my-vm --zone=us-central1-a
gcloud compute instances reset my-vm --zone=us-central1-a  # hard reboot

# Delete instance (keeps boot disk by default)
gcloud compute instances delete my-vm --zone=us-central1-a
# Delete instance and boot disk
gcloud compute instances delete my-vm --zone=us-central1-a \
  --delete-disks=boot
```

### SSH Access

```bash
# SSH via IAP tunnel (no public IP needed)
gcloud compute ssh my-vm --zone=us-central1-a --tunnel-through-iap

# SSH with custom command
gcloud compute ssh my-vm --zone=us-central1-a -- "sudo journalctl -u google-startup-scripts"

# Copy files to/from instance
gcloud compute scp localfile.txt my-vm:~/remote/ --zone=us-central1-a
gcloud compute scp my-vm:~/remote/output.txt ./ --zone=us-central1-a
```

### Images

```bash
# List public images for a project
gcloud compute images list --project=debian-cloud --no-standard-images

# Create custom image from disk
gcloud compute images create my-image \
  --source-disk=my-disk --source-disk-zone=us-central1-a

# Create image from running instance (auto-stops instance)
gcloud compute instances stop my-vm --zone=us-central1-a
gcloud compute images create my-vm-snapshot \
  --source-disk=my-vm --source-disk-zone=us-central1-a

# Deprecate an image
gcloud compute images deprecate my-old-image --state=DEPRECATED
```

### Networking — VPCs and Firewall Rules

```bash
# List VPC networks
gcloud compute networks list

# Create VPC
gcloud compute networks create my-vpc --subnet-mode=custom

# Create subnet
gcloud compute networks subnets create my-subnet \
  --network=my-vpc --range=10.0.0.0/24 --region=us-central1

# Create firewall rule — allow SSH from IAP range
gcloud compute firewall-rules create allow-iap-ssh \
  --network=my-vpc \
  --allow=tcp:22 \
  --source-ranges=35.235.240.0/20 \
  --description="Allow SSH via IAP"

# List firewall rules
gcloud compute firewall-rules list --filter="network=my-vpc"

# Delete firewall rule
gcloud compute firewall-rules delete allow-iap-ssh
```

### Managed Instance Groups (MIGs)

```bash
# Create instance template
gcloud compute instance-templates create my-template \
  --machine-type=e2-medium \
  --image-family=debian-12 --image-project=debian-cloud

# Create regional MIG
gcloud compute instance-groups managed create my-mig \
  --template=my-template \
  --size=3 \
  --region=us-central1

# Scale the group
gcloud compute instance-groups managed resize my-mig \
  --size=5 --region=us-central1

# Enable autoscaling
gcloud compute instance-groups managed set-autoscaling my-mig \
  --region=us-central1 \
  --min-num-replicas=2 \
  --max-num-replicas=10 \
  --target-cpu-utilization=0.6

# Rolling update to new template
gcloud compute instance-groups managed rolling-action start-update my-mig \
  --version=template=my-new-template --region=us-central1
```

## Patterns

### Blue/Green with MIG Traffic Shifting

1. Create new MIG with updated instance template
2. Register both MIGs with load balancer backend service
3. Shift traffic using backend service weight: `gcloud compute backend-services update`
4. Monitor error rates and latency
5. Remove old MIG backend once healthy

### Startup Script Debugging

```bash
# View startup script output from inside instance
sudo journalctl -u google-startup-scripts.service

# Check from outside via gcloud
gcloud compute instances get-serial-port-output my-vm --zone=us-central1-a
```

## Gotchas

- **Zones are required** — most `gcloud compute` commands require `--zone`; set a default with `gcloud config set compute/zone us-central1-a` to avoid repetition
- **Service account scopes** — the `--scopes` flag still matters even with a service account; use `--scopes=cloud-platform` for full API access and rely on IAM roles for fine-grained control
- **Default service account vs custom** — the Compute Engine default service account has broad permissions; always create a dedicated service account with least-privilege roles for production
- **External IP charges** — static external IPs have a cost when not attached to a running instance; release unused IPs with `gcloud compute addresses delete`
- **Live migration** — GCP migrates VMs during host maintenance by default; if an application can't tolerate migration, set `--maintenance-policy=TERMINATE` and optionally `--restart-on-failure`
