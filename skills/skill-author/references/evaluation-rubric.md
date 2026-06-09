# Skill Evaluation Rubric

Full scoring rubric for evaluating agentkit skills before approval and publication.

---

## How to Use This Rubric

Score each section independently. A single FAIL in any section blocks approval.
WARN items require a comment explaining why the exception is intentional.

---

## Section 1: Frontmatter Quality

### 1.1 Required Fields Present

| Check | Pass | Fail |
|-------|------|------|
| `name` field present | `name: aws` | Missing `name:` line |
| `description` field present | Multi-line `description: >` block | Missing or empty |
| `name` matches folder name | folder `aws/`, name `aws` | `name: AWS` (case mismatch) |
| `name` format | `lowercase-hyphens`, max 64 chars | `my_skill`, `MyCoolSkill` |

### 1.2 Description Quality

The `description` field tells agents WHEN to activate this skill. Evaluate:

| Rating | Criterion | Example |
|--------|-----------|---------|
| PASS | Starts with "Use when" + task context | "Use when working with AWS infrastructure — EC2, S3, IAM, ECS" |
| WARN | Describes what it is, not when to use | "AWS CLI and infrastructure management skill" |
| FAIL | Empty, stub, or vague | "A skill for AWS" / "TODO: write description" |

**Key test:** Can an AI agent read the description and decide whether to activate this skill
for a given task? If yes → PASS. If maybe → WARN. If no → FAIL.

### 1.3 Optional Fields

These are optional but encouraged:

| Field | Format | Notes |
|-------|--------|-------|
| `license` | SPDX identifier | `Apache-2.0`, `MIT`, `CC-BY-4.0` |
| `version` | SemVer | `1.0.0` |
| `tags` | YAML list | `["cloud", "aws", "infrastructure"]` |

---

## Section 2: SKILL.md Body Structure

### 2.1 Line Count

| Count | Rating | Action |
|-------|--------|--------|
| < 500 | PASS | No action needed |
| 500–600 | WARN | Review for content that should move to references/ |
| > 600 | FAIL | Must reduce — move content to references/ |

**Measurement:**
```bash
wc -l skills/<name>/SKILL.md
```

### 2.2 When to Use Section

Required. Must appear near the top of the body (before Quick Reference).

| Check | Pass | Fail |
|-------|------|------|
| Section exists | `## When to Use` heading present | Missing entirely |
| Contains activation triggers | Bulleted list of tasks | "Use this for things" (no specifics) |
| Not a marketing pitch | "Activate when working with EC2" | "This skill is great for all AWS needs" |

### 2.3 Quick Reference

Strongly recommended. Should appear after When to Use.

| Check | Pass | Warn |
|-------|------|------|
| Code examples present | Fenced code blocks with real commands | Pseudocode only |
| Commands are runnable | `aws ec2 describe-instances ...` | `aws [command]` |
| Tables used for lookup | Command/description tables | Long prose paragraphs |

### 2.4 Reference Files Table

Required if any reference files exist.

| Check | Pass | Fail |
|-------|------|------|
| Table present | `## Reference Files` with Markdown table | Missing when references/ exists |
| Links are accurate | File paths match actual files | Broken paths |
| Task descriptions | "When to load this file" column present | Filename only |

---

## Section 3: References/ Structure

### 3.1 File Presence

| Check | Pass | Warn | Fail |
|-------|------|------|------|
| references/ exists | Directory present | — | Missing when SKILL.md mentions reference files |
| Files match SKILL.md table | All listed files exist on disk | One file referenced but missing | Multiple broken references |
| No empty files | All files have content | One file < 10 lines | Files contain only "# Title" |

### 3.2 File Length

| Count | Rating |
|-------|--------|
| 200–400 lines | PASS (ideal) |
| 50–200 lines | WARN (may be too thin — could be merged into SKILL.md) |
| 400–600 lines | WARN (consider splitting) |
| > 600 lines | FAIL (split into two files) |
| < 50 lines | FAIL (stub — too thin to be useful) |

### 3.3 Content Quality

| Check | Pass | Fail |
|-------|------|------|
| Domain-specific content | Real commands, real patterns, real gotchas | Generic advice that applies to any domain |
| Actionable examples | Commands users can copy-paste | Descriptions without examples |
| Focused scope | One topic per file | Multiple unrelated topics mixed in one file |

---

## Section 4: Spec Compliance

See `references/spec-compliance.md` for full field specifications.

| Check | Pass | Fail |
|-------|------|------|
| SKILL.md at root of skill folder | `skills/<name>/SKILL.md` | Nested SKILL.md |
| No duplicate `name` in repo | `name` is unique across all skills/ | Same name as another skill |
| No YAML front matter mid-document | Only one `---` block at top | Second `---` block in body |

---

## Section 5: Injection Safety

### 5.1 YAML Block Injection

Check for `---` separators in the body (not the frontmatter):

```bash
grep -n "^---" skills/<name>/SKILL.md | tail -n +2
# Should return nothing (only the closing --- of frontmatter is allowed)
```

### 5.2 Instruction Override Patterns

Search for patterns that could hijack AI behavior:

```bash
grep -ri "ignore previous\|disregard.*instructions\|override.*system\|you are now\|act as if" skills/<name>/
# Should return nothing
```

### 5.3 Token Injection Markers

```bash
grep -ri "\[INST\]\|\[SYS\]\|<s>\|<<SYS>>" skills/<name>/
# Should return nothing
```

### 5.4 Personal Information

```bash
grep -ri "api_key\|access_token\|secret_key\|password\|/Users/\|/home/[a-z]" skills/<name>/
# Should return nothing (no credentials, no absolute personal paths)
```

---

## Section 6: Stub Detection

### 6.1 Placeholder Text

```bash
grep -ri "TODO\|FIXME\|coming soon\|placeholder\|stub\|not yet implemented\|to be added" skills/<name>/
# Should return nothing
```

### 6.2 Empty Data Patterns

Check for empty structures that flow to rendering:

```bash
grep -n '=\[\]\|={}\|= null\|= ""\|= ""' skills/<name>/**/*.sh
# Review any matches — are they intentional defaults or stubs?
```

---

## Scoring Summary

| Section | Weight | Gates Approval |
|---------|--------|---------------|
| 1. Frontmatter | Required | Yes — missing fields block |
| 2. SKILL.md body | Required | Yes — over 500 lines blocks |
| 3. references/ structure | If referenced | Yes — broken references block |
| 4. Spec compliance | Required | Yes — spec violations block |
| 5. Injection safety | Required | Yes — any injection FAIL blocks |
| 6. Stub detection | Required | Yes — any stub FAIL blocks |

A skill with all PASS ratings and no WARN items can be merged immediately.
A skill with WARN items requires a reviewer comment explaining the exception.
A skill with any FAIL must be revised before merge.
