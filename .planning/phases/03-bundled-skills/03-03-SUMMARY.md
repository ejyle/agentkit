---
phase: "03-bundled-skills"
plan: "03"
subsystem: "skills-content"
tags: ["skills", "cloud", "aws", "gcp", "azure", "playwright", "github", "cicd", "bundled"]
dependency_graph:
  requires:
    - "03-01"  # GitHubReleaseInstaller — skills/ directory structure consumed by installer
  provides:
    - "skills/aws"
    - "skills/gcp"
    - "skills/azure"
    - "skills/playwright"
    - "skills/github"
    - "skills/cicd"
  affects: []
tech_stack:
  added: []
  patterns:
    - "agentskills.io SKILL.md frontmatter (name + description)"
    - "Progressive disclosure: SKILL.md index + references/ deep-dive files"
    - "When to Use / Quick Reference / Reference Files SKILL.md structure"
key_files:
  created:
    - "skills/aws/SKILL.md"
    - "skills/aws/references/ec2.md"
    - "skills/aws/references/s3.md"
    - "skills/aws/references/iam.md"
    - "skills/gcp/SKILL.md"
    - "skills/gcp/references/compute.md"
    - "skills/gcp/references/gke.md"
    - "skills/gcp/references/cloudrun.md"
    - "skills/gcp/references/iam.md"
    - "skills/azure/SKILL.md"
    - "skills/azure/references/vms.md"
    - "skills/azure/references/aks.md"
    - "skills/azure/references/appservice.md"
    - "skills/playwright/SKILL.md"
    - "skills/playwright/references/e2e-patterns.md"
    - "skills/github/SKILL.md"
    - "skills/github/references/prs.md"
    - "skills/github/references/issues.md"
    - "skills/github/references/actions.md"
    - "skills/cicd/SKILL.md"
    - "skills/cicd/references/github-actions.md"
    - "skills/cicd/references/pipelines.md"
    - "skills/cicd/references/deployments.md"
  modified: []
decisions:
  - "Playwright MCP server install is separate (agentkit install playwright-mcp); SKILL.md teaches usage patterns only"
  - "All SKILL.md files use When to Use / Quick Reference / Reference Files structure for consistent agent activation UX"
metrics:
  duration: "~15min"
  completed: "2026-06-09"
  tasks_completed: 2
  files_created: 23
---

# Phase 03 Plan 03: Cloud and Dev Bundle Skills Summary

Six cloud and dev bundle skills authored to full reference quality: AWS (3 refs), GCP (4 refs), Azure (3 refs), Playwright (1 ref), GitHub (3 refs), CI/CD (3 refs).

## What Was Built

### Task 1: Cloud bundle — aws, gcp, azure (commit 172e8f2)

- `skills/aws/SKILL.md` — AWS CLI reference skill with EC2/S3/IAM/ECS quick reference; activation triggers cover infra provisioning, cost management, SSO login
- `skills/aws/references/ec2.md` — Full EC2 lifecycle: run-instances, AMI management, security groups, launch templates, Spot instances, SSM Session Manager, userdata
- `skills/aws/references/s3.md` — Bucket ops, presigned URLs, lifecycle policies, versioning, SSE, cross-account access
- `skills/aws/references/iam.md` — Policy authoring (Effect/Action/Resource), role creation/assumption, permission boundaries, instance profiles, AWS SSO, policy simulation
- `skills/gcp/SKILL.md` — gcloud CLI skill covering Compute, GKE, Cloud Run, IAM
- `skills/gcp/references/compute.md` — VM lifecycle, SSH/IAP, images, VPCs/firewall, managed instance groups with autoscaling
- `skills/gcp/references/gke.md` — Cluster/node-pool lifecycle, kubectl integration, HPA, Workload Identity setup
- `skills/gcp/references/cloudrun.md` — Deploy, traffic splits (canary), secrets, VPC connector, service-to-service auth
- `skills/gcp/references/iam.md` — Service accounts, project/resource bindings, custom roles, Workload Identity Federation (GitHub Actions OIDC)
- `skills/azure/SKILL.md` — az CLI skill covering VMs, AKS, App Service
- `skills/azure/references/vms.md` — VM lifecycle, Bastion SSH/RDP, Custom Script Extension, run-command, VMSS with autoscale
- `skills/azure/references/aks.md` — Cluster creation, node pools, Entra ID RBAC, Workload Identity Federation
- `skills/azure/references/appservice.md` — Plan+app creation, ZIP deploy, deployment slots (swap), autoscale, GitHub Actions CI/CD, log streaming

### Task 2: Dev bundle — playwright, github, cicd (commit da90bbd)

- `skills/playwright/SKILL.md` — Playwright E2E skill with Playwright MCP server tool reference; install instruction points to `agentkit install playwright-mcp`
- `skills/playwright/references/e2e-patterns.md` — Page Object Model in TypeScript, semantic locator strategies, auth storageState fixtures, async/await patterns, retry/timeout config
- `skills/github/SKILL.md` — gh CLI skill covering PRs, issues, Actions
- `skills/github/references/prs.md` — Full PR lifecycle: create, draft, checkout, review, merge strategies, conflict resolution, branch protection
- `skills/github/references/issues.md` — Issue creation, filtering/search, bulk JSON+jq ops, issue templates, PR linking, milestone management
- `skills/github/references/actions.md` — Run list/view/watch/rerun, workflow dispatch, secrets/variables, artifact download, matrix builds, environment protection
- `skills/cicd/SKILL.md` — CI/CD pipeline skill covering GitHub Actions workflows, build stages, deployment strategies
- `skills/cicd/references/github-actions.md` — YAML schema, reusable workflows (workflow_call), composite actions, job dependencies (needs:), output variables, permissions hardening
- `skills/cicd/references/pipelines.md` — Four-stage pipeline structure, fan-out/fan-in parallelism, artifact passing, caching strategies, conditional steps, timeout config
- `skills/cicd/references/deployments.md` — Blue/green traffic switching, canary gradual rollout, rollback procedures, environment promotion (dev→staging→prod), health check integration

## Deviations from Plan

None — plan executed exactly as written.

The `grep` for `placeholder` matched a legitimate code comment in `e2e-patterns.md` (`page.getByPlaceholder('Search products') // placeholder attr`) explaining the HTML `placeholder` attribute, not stub content. No stubs are present in any file.

## Known Stubs

None. All skills contain real, actionable CLI commands and code patterns. No placeholder or TODO content.

## Threat Flags

No new security-relevant surface introduced. All files are Markdown content (no network endpoints, no auth paths, no schema changes). Content reviewed for prompt-injection patterns per T-03-09 — no `---` YAML block injections or system-prompt-like constructs found.

## Self-Check: PASSED

Files verified present:
- skills/aws/SKILL.md ✓
- skills/aws/references/ec2.md ✓
- skills/aws/references/s3.md ✓
- skills/aws/references/iam.md ✓
- skills/gcp/SKILL.md ✓
- skills/gcp/references/compute.md ✓
- skills/gcp/references/gke.md ✓
- skills/gcp/references/cloudrun.md ✓
- skills/gcp/references/iam.md ✓
- skills/azure/SKILL.md ✓
- skills/azure/references/vms.md ✓
- skills/azure/references/aks.md ✓
- skills/azure/references/appservice.md ✓
- skills/playwright/SKILL.md ✓
- skills/playwright/references/e2e-patterns.md ✓
- skills/github/SKILL.md ✓
- skills/github/references/prs.md ✓
- skills/github/references/issues.md ✓
- skills/github/references/actions.md ✓
- skills/cicd/SKILL.md ✓
- skills/cicd/references/github-actions.md ✓
- skills/cicd/references/pipelines.md ✓
- skills/cicd/references/deployments.md ✓

Commits verified:
- 172e8f2 (cloud bundle) ✓
- da90bbd (dev bundle) ✓
