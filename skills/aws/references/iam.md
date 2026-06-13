## Overview

AWS IAM (Identity and Access Management) controls who can do what to which AWS resources. The key building blocks are policies (JSON permission documents), roles (assumable identities for services or cross-account use), and users/groups (for human access). IAM is global — not region-specific.

## Common Commands

### Policies

```bash
# Create an inline policy document
cat > policy.json <<'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:PutObject"],
      "Resource": "arn:aws:s3:::my-bucket/*"
    },
    {
      "Effect": "Deny",
      "Action": "s3:DeleteObject",
      "Resource": "arn:aws:s3:::my-bucket/*"
    }
  ]
}
EOF

# Create managed policy from file
aws iam create-policy \
  --policy-name MyS3Policy \
  --policy-document file://policy.json

# List managed policies (AWS + customer)
aws iam list-policies --scope Local --query 'Policies[*].[PolicyName,Arn]' --output table

# Get policy version document
aws iam get-policy-version \
  --policy-arn arn:aws:iam::123456789012:policy/MyS3Policy \
  --version-id v1

# Delete policy (must detach from all entities first)
aws iam delete-policy --policy-arn arn:aws:iam::123456789012:policy/MyS3Policy
```

### Roles

```bash
# Create role with trust policy (allows EC2 to assume it)
cat > trust.json <<'EOF'
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"Service": "ec2.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}
EOF

aws iam create-role --role-name MyEC2Role \
  --assume-role-policy-document file://trust.json

# Attach AWS managed policy to role
aws iam attach-role-policy --role-name MyEC2Role \
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore

# Attach customer-managed policy
aws iam attach-role-policy --role-name MyEC2Role \
  --policy-arn arn:aws:iam::123456789012:policy/MyS3Policy

# List policies attached to role
aws iam list-attached-role-policies --role-name MyEC2Role

# Detach and delete role
aws iam detach-role-policy --role-name MyEC2Role \
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
aws iam delete-role --role-name MyEC2Role
```

### Assume Role (Cross-Account / Temporary Credentials)

```bash
# Assume a role and export credentials
CREDS=$(aws sts assume-role \
  --role-arn arn:aws:iam::TARGET_ACCOUNT:role/MyRole \
  --role-session-name my-session \
  --query 'Credentials.[AccessKeyId,SecretAccessKey,SessionToken]' \
  --output text)

export AWS_ACCESS_KEY_ID=$(echo $CREDS | awk '{print $1}')
export AWS_SECRET_ACCESS_KEY=$(echo $CREDS | awk '{print $2}')
export AWS_SESSION_TOKEN=$(echo $CREDS | awk '{print $3}')

# Verify assumed identity
aws sts get-caller-identity
```

### Instance Profiles

```bash
# Create instance profile and attach role (for EC2)
aws iam create-instance-profile --instance-profile-name MyProfile
aws iam add-role-to-instance-profile \
  --instance-profile-name MyProfile --role-name MyEC2Role

# Associate with running instance
aws ec2 associate-iam-instance-profile \
  --instance-id i-0abcd \
  --iam-instance-profile Name=MyProfile
```

### Permission Boundaries

```bash
# Create a permission boundary (caps maximum permissions for a role)
aws iam create-policy \
  --policy-name DeveloperBoundary \
  --policy-document file://boundary.json

# Attach boundary when creating a role
aws iam create-role --role-name SandboxRole \
  --assume-role-policy-document file://trust.json \
  --permissions-boundary arn:aws:iam::123456789012:policy/DeveloperBoundary
```

### AWS SSO / Identity Center

```bash
# List SSO accounts and roles
aws sso list-accounts --access-token <token>
aws sso list-account-roles --account-id 123456789012 --access-token <token>

# Configure SSO profile
aws configure sso
# Follow prompts: SSO start URL, region, account, role, output format

# Login and refresh token
aws sso login --profile my-sso-profile

# Use profile for any command
aws s3 ls --profile my-sso-profile
```

### Policy Simulation

```bash
# Dry-run check — does a role have a specific permission?
aws iam simulate-principal-policy \
  --policy-source-arn arn:aws:iam::123456789012:role/MyRole \
  --action-names s3:PutObject ec2:RunInstances \
  --resource-arns arn:aws:s3:::my-bucket/* \
  --query 'EvaluationResults[*].[EvalActionName,EvalDecision]' --output table
```

## Patterns

### Least-Privilege Service Role

1. Start with `Action: "*"` to get code working
2. Enable CloudTrail, run workload, note actual API calls
3. Narrow `Action` list to only what was called
4. Add `Resource` constraints to specific ARNs where possible
5. Attach permission boundary to cap blast radius

### Cross-Account Role Chaining

Source account calls `sts:AssumeRole` on target account role. Target role trust policy lists source account/role. Use `sts:ExternalId` condition for third-party integrations to prevent confused-deputy attacks.

## Gotchas

- **`iam:PassRole` often omitted** — creating an EC2 instance with a role, or a Lambda with an execution role, requires `iam:PassRole` on the role ARN in the caller's policy; missing it causes `AccessDenied` at resource creation, not IAM validation
- **Confused-deputy attack** — when a third-party service assumes a role in your account, require `sts:ExternalId` in the trust policy condition; without it any customer of that service can access your account if they learn your role ARN
- **Permission boundaries do not grant permissions** — they only cap the effective maximum; you still need an identity policy that grants the action
- **IAM eventual consistency** — newly created roles/policies may take a few seconds to propagate; automation should retry on `AccessDenied` after role creation
- **Inline vs managed policies** — inline policies are embedded in the identity and cannot be reused; managed policies can be attached to multiple roles; prefer managed for shared use, inline for single-use guardrails
- **`aws:RequestedRegion` condition** — use this in SCPs or IAM policies to restrict which regions a role can operate in; without it, a role granted `ec2:*` can launch in any region
