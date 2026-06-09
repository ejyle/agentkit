# agentskills.io Spec Compliance

Reference for the agentskills.io skill specification — required fields, optional fields,
body structure conventions, and naming rules.

---

## Frontmatter Specification

Every skill must have a YAML frontmatter block at the top of SKILL.md:

```yaml
---
name: skill-name        # required
description: >          # required
  Use when [task context] — [what it provides].
license: Apache-2.0     # optional but encouraged
version: 1.0.0          # optional
tags:                   # optional
  - cloud
  - aws
---
```

### Required Fields

#### `name`

- **Type:** String
- **Format:** Lowercase letters, numbers, hyphens only. No underscores, no spaces, no uppercase.
- **Length:** Maximum 64 characters
- **Must match folder name:** If the skill lives in `skills/aws/`, the name must be `aws`.
- **Unique constraint:** No two skills in the registry may share the same name.

```yaml
# Valid
name: aws
name: context-mode
name: skill-author
name: github-copilot-tips

# Invalid
name: AWS              # uppercase
name: my_skill         # underscore
name: My Cool Skill    # spaces
```

#### `description`

- **Type:** Multi-line string (`>` block scalar)
- **Purpose:** Tells the AI agent WHEN to activate this skill
- **Must start with:** "Use when" followed by concrete task context
- **Length:** Aim for 2–4 sentences. Enough for an AI to make an activation decision.
- **Must NOT be:** A marketing description, a definition of what the skill is, or a list of features

```yaml
# Valid
description: >
  Use when working with AWS infrastructure — provisioning EC2 instances, managing S3 buckets,
  configuring IAM policies and roles, or running ECS/EKS workloads via the AWS CLI.

# Valid
description: >
  Use when navigating code structure, finding symbol definitions or references, or performing
  safe renames across a codebase. Serena provides LSP-powered code intelligence tools.

# Invalid — describes the skill, not when to use it
description: >
  A comprehensive AWS skill with EC2, S3, IAM, and ECS coverage.

# Invalid — too vague
description: >
  Use for AWS tasks.
```

### Optional Fields

#### `license`

SPDX license identifier. Use this to specify the license for the skill content.

Common values:
- `Apache-2.0`
- `MIT`
- `CC-BY-4.0`
- `CC0-1.0`

#### `version`

SemVer version string. Increment when making breaking changes to the skill content.

```yaml
version: 1.0.0
version: 1.2.3
```

#### `tags`

YAML list of searchable tags. Used by registry search and discovery.

```yaml
tags:
  - cloud
  - aws
  - infrastructure
  - iam
```

Tags should be: lowercase, singular (not plural), descriptive of the domain.

---

## Body Structure Conventions

### Progressive Disclosure Pattern

The SKILL.md body should follow progressive disclosure:
- **Summary in SKILL.md** — quick reference, essential patterns, activation guide
- **Detail in references/** — deep dives, comprehensive command lists, edge cases

This keeps SKILL.md under 500 lines while still providing depth for complex domains.

### Required Sections

| Section | Required | Notes |
|---------|----------|-------|
| `## When to Use` | Yes | Must appear before Quick Reference |
| `## Quick Reference` | Recommended | Code examples and command tables |
| `## Reference Files` | If references/ exists | Markdown table linking to reference files |

### Recommended Sections

| Section | Notes |
|---------|-------|
| `## Common Gotchas` | Known failure modes and workarounds |
| `## Installation` | If the skill requires CLI tools or MCP server setup |
| `## Authentication` | If the domain requires auth configuration |

### Section Ordering

Recommended order:
1. Frontmatter
2. `## When to Use` (or `## Activation Triggers`)
3. `## Quick Reference` (or named sections like `## Authentication`, `## Common Commands`)
4. `## Reference Files` (table)
5. `## Common Gotchas` (optional)

---

## Body Formatting Rules

### Code Blocks

Always use fenced code blocks with a language tag:

```markdown
```bash
aws ec2 describe-instances --filters "Name=instance-state-name,Values=running"
```
```

```markdown
```go
type Installer interface {
    Method() InstallMethod
    Install(spec InstallSpec) error
}
```
```

Avoid inline code for multi-line commands. Single-token names can use backtick inline code.

### Tables

Use Markdown tables for command references and option lists:

```markdown
| Command | When to Use |
|---------|-------------|
| `git status` | Check working tree state |
| `git log --oneline` | Review recent commits |
```

### Line Length

No hard limit on line length within the body. Readability takes priority over strict 80-char limits.
Exception: code blocks should use realistic command lengths that fit in a terminal.

---

## Directory Structure

### Minimal Skill

```
skills/<name>/
└── SKILL.md
```

Acceptable for simple single-domain skills under 300 lines.

### Standard Skill

```
skills/<name>/
├── SKILL.md
└── references/
    ├── topic-a.md
    └── topic-b.md
```

Required when SKILL.md would exceed 500 lines without offloading content.

### Full Skill

```
skills/<name>/
├── SKILL.md
├── references/
│   ├── topic-a.md
│   ├── topic-b.md
│   └── topic-c.md
└── scripts/
    └── setup.sh
```

Scripts are optional. Use for environment detection, prerequisite checks, or setup automation.

---

## Naming Conventions

### Folder Names

```
skills/aws/           # single word — OK
skills/context-mode/  # hyphen-separated — OK
skills/skill-author/  # hyphen-separated — OK
skills/github/        # single word — OK
```

Avoid:
```
skills/AWS/           # uppercase — FAIL
skills/my_skill/      # underscore — FAIL
skills/github-copilot-tips-and-tricks/  # too long (over 64 chars) — WARN
```

### Reference File Names

```
references/ec2.md             # short, descriptive — OK
references/routing-rules.md   # hyphen-separated — OK
references/lsp-usage.md       # descriptive compound — OK
```

Avoid:
```
references/EC2.md             # uppercase — WARN
references/my_reference.md    # underscore — WARN
references/ref.md             # not descriptive — WARN
```

### Script Names

```
scripts/validate-skill.sh     # verb-noun, hyphen-separated — OK
scripts/detect-aws-env.sh     # descriptive — OK
```

---

## Spec Validation

Run the validation script to check compliance:

```bash
bash skills/skill-author/scripts/validate-skill.sh skills/<name>/
```

The script checks:
- SKILL.md exists
- Line count < 500
- references/ exists if SKILL.md mentions reference files
- No injection patterns in any file

Exit code 0 = all checks PASS or WARN.
Exit code 1 = at least one check FAIL.

---

## Differences from anthropics/skills Spec

agentskills.io adapts the Anthropic skills specification with these additions:

| agentskills.io | anthropics/skills | Reason |
|----------------|-------------------|--------|
| `description` must start with "Use when" | No specific format requirement | Enables reliable activation signal |
| `references/` directory structure required for multi-domain skills | No references/ convention | Progressive disclosure pattern |
| `name` must match folder name | No folder-matching rule | Registry lookup consistency |
| Line count limit (500) | No limit | Token budget enforcement |
| Injection safety checks mandatory | Not specified | Security baseline |
