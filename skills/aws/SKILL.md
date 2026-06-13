---
name: aws
description: >
  Use when working with AWS infrastructure — provisioning EC2 instances, managing S3 buckets,
  configuring IAM policies and roles, running ECS/EKS workloads, or automating AWS resources
  via the AWS CLI or CloudFormation.
license: Apache-2.0
---

## When to Use

Activate this skill when the task involves:

- Launching, stopping, or connecting to EC2 instances
- Creating or managing S3 buckets, objects, presigned URLs, or lifecycle policies
- Authoring or debugging IAM policies, roles, or permission boundaries
- Setting up ECS tasks, EKS clusters, or Fargate services
- Writing CloudFormation / CDK templates or troubleshooting stacks
- Configuring VPCs, security groups, subnets, or route tables
- Working with AWS SSO / Identity Center
- Diagnosing AWS CLI errors (`InvalidClientTokenId`, `AccessDenied`, region mismatch)

## Quick Reference

### Authentication

```bash
# Configure named profile
aws configure --profile myprofile

# Use a specific profile for any command
AWS_PROFILE=myprofile aws sts get-caller-identity

# Assume a role (outputs temp credentials)
aws sts assume-role --role-arn arn:aws:iam::123456789012:role/MyRole \
  --role-session-name mysession

# SSO login
aws sso login --profile sso-profile
```

### EC2

```bash
# List running instances
aws ec2 describe-instances --filters "Name=instance-state-name,Values=running" \
  --query 'Reservations[*].Instances[*].[InstanceId,PublicIpAddress,Tags[?Key==`Name`].Value|[0]]' \
  --output table

# Start / stop / terminate
aws ec2 start-instances --instance-ids i-0abcd1234
aws ec2 stop-instances --instance-ids i-0abcd1234
aws ec2 terminate-instances --instance-ids i-0abcd1234

# SSM Session Manager (no SSH key or open port 22 required)
aws ssm start-session --target i-0abcd1234
```

### S3

```bash
# List buckets / objects
aws s3 ls
aws s3 ls s3://my-bucket/prefix/

# Copy and sync
aws s3 cp localfile.txt s3://my-bucket/path/
aws s3 sync ./dist s3://my-bucket/dist/ --delete

# Presigned URL (15 min)
aws s3 presign s3://my-bucket/file.zip --expires-in 900
```

### IAM

```bash
# List roles and policies
aws iam list-roles --query 'Roles[*].[RoleName,Arn]' --output table
aws iam list-attached-role-policies --role-name MyRole

# Attach managed policy
aws iam attach-role-policy --role-name MyRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess

# Simulate policy (dry-run permission check)
aws iam simulate-principal-policy \
  --policy-source-arn arn:aws:iam::123456789012:role/MyRole \
  --action-names s3:GetObject \
  --resource-arns arn:aws:s3:::my-bucket/*
```

### ECS / Fargate

```bash
# List clusters and services
aws ecs list-clusters
aws ecs list-services --cluster my-cluster

# Force new deployment (redeploy latest image)
aws ecs update-service --cluster my-cluster --service my-service \
  --force-new-deployment

# View running task logs via CloudWatch
aws logs tail /ecs/my-task --follow
```

## Reference Files

Load the appropriate reference file for deep-dive tasks:

| Task | Reference file |
|------|---------------|
| EC2 instance lifecycle, AMIs, launch templates, Spot | `references/ec2.md` |
| S3 bucket operations, lifecycle policies, encryption, cross-account | `references/s3.md` |
| IAM policy authoring, roles, permission boundaries, SSO | `references/iam.md` |

## Common Gotchas

- **Region matters everywhere** — always pass `--region` or set `AWS_DEFAULT_REGION`; AMI IDs, VPC IDs, and endpoint URLs are region-specific
- **Profile vs environment variables** — `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` env vars take precedence over `~/.aws/credentials` profile; unexpected profile switches often mean stale env vars
- **`iam:PassRole` often forgotten** — services like EC2, ECS, Lambda need `iam:PassRole` in the caller's policy when attaching a role; missing it causes `AccessDenied` on resource creation
- **SSM requires instance profile** — `aws ssm start-session` only works if the instance has an IAM role with `AmazonSSMManagedInstanceCore` attached
