## Overview

EC2 (Elastic Compute Cloud) is AWS's virtual machine service. Instances run in a region and availability zone; every instance requires a VPC subnet and at least one security group. The AWS CLI (`aws ec2`) manages the full lifecycle from launch to termination.

## Common Commands

### Instance Lifecycle

```bash
# Launch an instance with a specific AMI and instance type
aws ec2 run-instances \
  --image-id ami-0abcdef1234567890 \
  --instance-type t3.micro \
  --key-name my-keypair \
  --security-group-ids sg-0abc123 \
  --subnet-id subnet-0abc123 \
  --iam-instance-profile Name=MyInstanceProfile \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=my-server}]' \
  --count 1

# Describe instances with useful fields
aws ec2 describe-instances \
  --filters "Name=instance-state-name,Values=running" \
  --query 'Reservations[*].Instances[*].{ID:InstanceId,Type:InstanceType,IP:PublicIpAddress,Name:Tags[?Key==`Name`]|[0].Value}' \
  --output table

# Start / stop / terminate
aws ec2 start-instances --instance-ids i-0abcd
aws ec2 stop-instances --instance-ids i-0abcd
aws ec2 terminate-instances --instance-ids i-0abcd

# Wait for instance running state
aws ec2 wait instance-running --instance-ids i-0abcd
```

### AMI Management

```bash
# List your own AMIs
aws ec2 describe-images --owners self \
  --query 'Images[*].[ImageId,Name,CreationDate]' --output table

# Create AMI from running instance
aws ec2 create-image --instance-id i-0abcd \
  --name "my-server-backup-$(date +%Y%m%d)" --no-reboot

# Copy AMI to another region
aws ec2 copy-image --source-image-id ami-0abcdef1234567890 \
  --source-region us-east-1 --region us-west-2 --name my-copied-ami
```

### Security Groups

```bash
# Create security group
aws ec2 create-security-group \
  --group-name my-sg --description "My security group" \
  --vpc-id vpc-0abc123

# Allow inbound SSH from specific IP
aws ec2 authorize-security-group-ingress \
  --group-id sg-0abc123 \
  --protocol tcp --port 22 --cidr 203.0.113.0/32

# List security group rules
aws ec2 describe-security-groups --group-ids sg-0abc123 \
  --query 'SecurityGroups[*].IpPermissions'
```

### Launch Templates

```bash
# Create launch template
aws ec2 create-launch-template \
  --launch-template-name my-lt \
  --launch-template-data '{
    "ImageId": "ami-0abcdef1234567890",
    "InstanceType": "t3.micro",
    "IamInstanceProfile": {"Name": "MyProfile"},
    "SecurityGroupIds": ["sg-0abc123"]
  }'

# Launch from template
aws ec2 run-instances --launch-template LaunchTemplateName=my-lt,Version='$Latest'
```

### Spot Instances

```bash
# Request spot instance
aws ec2 request-spot-instances \
  --spot-price "0.02" \
  --instance-count 1 \
  --type one-time \
  --launch-specification '{
    "ImageId": "ami-0abcdef1234567890",
    "InstanceType": "t3.micro",
    "SubnetId": "subnet-0abc123"
  }'

# Check spot price history
aws ec2 describe-spot-price-history \
  --instance-types t3.micro \
  --product-descriptions "Linux/UNIX" \
  --start-time $(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --query 'SpotPriceHistory[*].[AvailabilityZone,SpotPrice]' --output table
```

### SSM Session Manager

```bash
# Connect without SSH (requires SSM agent + instance profile)
aws ssm start-session --target i-0abcd

# Run a command remotely
aws ssm send-command \
  --instance-ids i-0abcd \
  --document-name "AWS-RunShellScript" \
  --parameters 'commands=["sudo systemctl status nginx"]' \
  --query 'Command.CommandId'

# Check command output
aws ssm get-command-invocation \
  --command-id <command-id> --instance-id i-0abcd \
  --query '{Status:Status,Output:StandardOutputContent}'
```

### User Data Scripts

```bash
# Pass userdata (base64 on Linux)
aws ec2 run-instances \
  --user-data file://startup.sh \
  --image-id ami-0abcdef1234567890 \
  --instance-type t3.micro

# View userdata on running instance (from inside)
curl http://169.254.169.254/latest/user-data
```

## Patterns

### Blue/Green Instance Swap

1. Launch new instance from updated AMI
2. Register with load balancer target group
3. Wait for health checks to pass: `aws elbv2 wait target-in-service`
4. Deregister old instance from target group
5. Terminate old instance after drain period

### Placement Groups

```bash
# Create cluster placement group (low-latency, same AZ)
aws ec2 create-placement-group --group-name my-hpc-group --strategy cluster

# Launch into placement group
aws ec2 run-instances --placement '{"GroupName":"my-hpc-group"}' ...
```

## Gotchas

- **AMI IDs are region-specific** — an AMI copied from us-east-1 has a different ID in eu-west-1; never hardcode AMI IDs in cross-region scripts
- **SSM requires instance profile** — the IAM role attached to the instance must have `AmazonSSMManagedInstanceCore`; without it `start-session` returns a timeout
- **t3/t4g CPU credits** — burstable instances (t-family) use CPU credits; sustained high CPU will throttle after credits exhaust unless `unlimited` mode is enabled
- **Security group changes are immediate** — changes to inbound/outbound rules apply without instance restart
- **Userdata runs once at launch** — userdata is not re-executed on stop/start; use `cloud-init` modules with `#cloud-config` if needed on every boot
- **Instance metadata v2 (IMDSv2)** — newer instances require session-oriented metadata; use `curl -H "X-aws-ec2-metadata-token: <TOKEN>"` not the plain v1 endpoint
