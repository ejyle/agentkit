# Skill Authoring Guide

Step-by-step guide for creating a new agentkit skill from scratch.

---

## Overview

A skill teaches an AI agent how to work in a specific domain. A good skill:

- Activates at the right moment (the agent recognizes its own need)
- Provides the right information at the right depth (not too much, not too little)
- Stays within the 500-line token budget
- Contains no stubs, placeholders, or injected instructions

This guide walks through the full authoring process in 7 steps.

---

## Step 1: Identify the Domain and Activation Triggers

Before writing a single line, answer these questions:

**What task does the agent need to do when this skill activates?**

Bad framing: "This skill is about Kubernetes."
Good framing: "This skill activates when the agent needs to deploy a workload to a Kubernetes cluster, debug pod failures, manage namespaces, or configure RBAC."

**What signals tell the agent to activate?**

Write a list of specific scenarios. These become your "When to Use" bullets:
- Creating or scaling a Deployment
- Debugging a CrashLoopBackOff pod
- Configuring ServiceAccount permissions
- Writing or applying Helm charts

**What are the sub-domains?** (drives the references/ structure)
Break the domain into major topics that each need 50+ lines of content:
- Pod management → `references/pods.md`
- Networking (services, ingress) → `references/networking.md`
- RBAC / security → `references/rbac.md`

Write these down before authoring — they define your file structure.

---

## Step 2: Research the Domain Content

### Option A: You Know the Domain

Write from expertise. Focus on:
- Real commands that work in practice
- Common failure modes and their fixes (Gotchas section)
- The 20% of features that cover 80% of tasks (Quick Reference)

### Option B: Use the Auto-Researcher Agent

For unfamiliar domains, invoke the auto-researcher:

```
# In your Claude Code session:
"Research the [domain-name] domain for a new agentkit skill.
Output a draft SKILL.md and references/ skeleton to skills/<name>/.
Budget: max 5 web searches. Output to files only."

See: agents/auto-researcher/AGENT.md for full invocation instructions.
```

The auto-researcher produces a draft skeleton. It will NOT be production-ready:
- Review all generated content for accuracy
- Add real command examples you have verified
- Remove or rewrite any content that seems hallucinated

### Research Sources (Priority Order)

1. Official documentation (primary — always cite if using)
2. Official CLI help (`kubectl --help`, `aws help`, `gh --help`)
3. Official quickstart guides
4. Your own experience with the domain

Avoid:
- Blog posts and tutorials (may be outdated or wrong)
- Stack Overflow answers (without verifying against official docs)
- Other skills repos (without verifying content quality)

---

## Step 3: Write the SKILL.md Entrypoint

Create `skills/<name>/SKILL.md`. Target: 150–350 lines.

### Frontmatter (Required)

```yaml
---
name: <name>           # lowercase-hyphens; matches folder name
description: >
  Use when [specific tasks with this domain] — [what the skill provides].
license: Apache-2.0
---
```

### When to Use (Required)

Write a bulleted list of specific activation scenarios. Be concrete:

```markdown
## When to Use

Activate this skill when the task involves:

- Deploying a new workload to a Kubernetes cluster
- Debugging pod failures (CrashLoopBackOff, OOMKilled, ImagePullBackOff)
- Managing namespaces, quotas, or limit ranges
- Configuring RBAC roles and service accounts
- Writing or applying Helm charts
```

### Quick Reference (Recommended)

Add the 5–10 most-used commands with real examples:

```markdown
## Quick Reference

### Pod Management

```bash
# List pods with status
kubectl get pods -n my-namespace -o wide

# Get pod logs
kubectl logs pod-name -c container-name --tail=100

# Describe pod (events, conditions)
kubectl describe pod pod-name -n my-namespace

# Exec into a running pod
kubectl exec -it pod-name -n my-namespace -- /bin/sh
```
```

### Reference Files Table (Required if references/ exists)

```markdown
## Reference Files

| Task | Reference file |
|------|---------------|
| Pod lifecycle, debugging CrashLoopBackOff, resource limits | `references/pods.md` |
| Services, Ingress, NetworkPolicy configuration | `references/networking.md` |
| RBAC roles, service accounts, ClusterRoleBinding | `references/rbac.md` |
```

### Common Gotchas (Optional but Valuable)

```markdown
## Common Gotchas

- **Namespace required for most operations** — always pass `-n <namespace>` or set
  `kubectl config set-context --current --namespace=<namespace>` to avoid "not found" errors
  in the wrong namespace.
- **ImagePullBackOff vs ErrImagePull** — ErrImagePull is transient (retry); ImagePullBackOff
  means the image URL or registry credentials are wrong (fix the config, not retry).
```

---

## Step 4: Identify Sub-Domains for references/

A sub-domain needs its own reference file when it has more than 50 lines of relevant content.

Decision criteria:
- Could an agent need this section independently? (yes → own file)
- Will SKILL.md exceed 500 lines without offloading it? (yes → own file)
- Does it have enough content to be useful? (50+ lines → own file, else keep inline)

Create the directory and placeholder filenames before writing content:

```bash
mkdir -p skills/<name>/references
touch skills/<name>/references/pods.md
touch skills/<name>/references/networking.md
touch skills/<name>/references/rbac.md
```

---

## Step 5: Create Reference Files

Target: 200–400 lines per file.

### Structure for Each Reference File

```markdown
# Domain: Topic Name

Brief intro sentence (1–2 lines).

---

## Sub-Topic A

Content with real commands and examples.

## Sub-Topic B

Content with real commands and examples.

## Common Gotchas

Failure modes specific to this sub-topic.
```

### Quality Bar

Each reference file must have:
- Real, working commands (not pseudocode)
- Explanation of WHEN to use each command (not just what it does)
- At least one common failure mode and its fix
- No content that duplicates SKILL.md Quick Reference (go deeper, not repeat)

---

## Step 6: Add Scripts (Optional)

Scripts go in `skills/<name>/scripts/`. Use for:

- **Environment detection:** `detect-aws-env.sh` — checks `aws sts get-caller-identity`
- **Prerequisite checks:** `check-deps.sh` — verifies CLI tools are installed
- **Validation:** domain-specific validation beyond what validate-skill.sh checks

Script requirements:
- Must be a bash script (`.sh` extension)
- Must be executable (`chmod +x`)
- Must handle missing tools gracefully (check with `command -v <tool>`)
- Must print PASS/WARN/FAIL per check and exit 1 on FAIL

---

## Step 7: Run Validation and Stub Check

### Validation Script

```bash
bash skills/skill-author/scripts/validate-skill.sh skills/<name>/
```

Fix any FAIL before proceeding. WARN items need a PR comment explaining the exception.

### Stub Check

Search for any remaining placeholder content:

```bash
grep -ri "TODO\|FIXME\|coming soon\|placeholder\|stub\|not yet implemented\|to be added\|example.com" skills/<name>/
```

If matches exist, either fill them in or remove the section entirely. Do NOT commit stubs.

### Injection Check

```bash
grep -ri "ignore previous\|disregard.*instructions\|\[INST\]\|<<SYS>>" skills/<name>/
grep -n "^---" skills/<name>/SKILL.md | tail -n +2
```

Both should return nothing.

### Line Count Check

```bash
wc -l skills/<name>/SKILL.md
```

Must be under 500. If over, move content to references/.

---

## Step 8: Submit PR

### PR Checklist

- [ ] All files in `skills/<name>/` staged
- [ ] `bash skills/skill-author/scripts/validate-skill.sh skills/<name>/` exits 0
- [ ] No TODO/stub/placeholder in any file
- [ ] No personal credentials or absolute home paths
- [ ] PR description includes: domain, activation triggers summary, reference files list

### PR Description Template

```markdown
## New Skill: <name>

**Domain:** [e.g., Kubernetes cluster management]
**Bundle:** [e.g., cloud, dev, context — or "standalone"]

**Activates when:** [2-3 sentence summary of when this skill is useful]

**Files added:**
- skills/<name>/SKILL.md
- skills/<name>/references/topic-a.md
- skills/<name>/references/topic-b.md

**Validation:** `validate-skill.sh` passes (attach output)

**Content sources:** [list official docs or other sources used]
```

---

## Examples of Well-Structured Skills

These existing skills in this repo can serve as references:

| Skill | What to Copy |
|-------|-------------|
| `skills/aws/` | Frontmatter format, Quick Reference tables, Common Gotchas section |
| `skills/context-mode/` | Tool hierarchy table format, progressive disclosure pattern |
| `skills/serena/` | Tool group tables, workflow sections in references/ |
| `skills/skill-author/` | Meta-skill structure with references/ + scripts/ |
