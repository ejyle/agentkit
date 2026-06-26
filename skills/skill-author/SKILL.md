---
name: skill-author
description: >
  Use when writing a new agentkit skill from scratch, reviewing a skill pull request,
  improving an existing skill for quality and spec compliance, or validating that a skill
  meets the agentskills.io standard before submission.
license: Apache-2.0
---

## When to Use

Activate this skill for:

- **Authoring a new skill** — from domain research through SKILL.md + references/ creation
- **Reviewing a skill PR** — evaluating frontmatter, line count, injection safety, spec compliance
- **Improving an existing skill** — restructuring, adding references/, reducing line count
- **Validating before submission** — running the validation script and interpreting results

## Authoring Workflow

Follow these steps in order:

1. **Identify the domain and activation triggers** — what task does an AI agent do to need this skill?
2. **Research domain content** — use the `auto-researcher` agent for unfamiliar domains (see below)
3. **Write `SKILL.md` entrypoint** — frontmatter (name, description) + When to Use + Quick Reference + Reference Files table (under 500 lines total)
4. **Identify sub-domains** — any major topic area that needs more than 50 lines of detail goes in `references/`
5. **Create reference files** — one file per sub-domain, 200–400 lines each, linked from SKILL.md Reference Files table
6. **Add scripts if needed** — `scripts/` for setup automation, detection scripts, or validation helpers
7. **Run validation** — `bash skills/<name>/scripts/validate-skill.sh skills/<name>/`
8. **Stub check** — search for TODO, placeholder, stub, "coming soon" in all authored files
9. **Submit PR** — include skill folder, references/, and scripts/

## Evaluation Checklist (Quick Summary)

Before approving or submitting a skill, verify:

- [ ] `SKILL.md` has valid frontmatter (`name`, `description` required)
- [ ] `name` field matches folder name — lowercase, hyphen-separated, max 64 chars
- [ ] `description` starts with "Use when" — tells agent WHEN to activate, not what it is
- [ ] SKILL.md body is under 500 lines
- [ ] Heavy content (>50 lines per sub-domain) is in `references/` not inline
- [ ] `references/` files are 200–400 lines each (not stubs, not bloated)
- [ ] No `TODO`, `stub`, `placeholder`, or `coming soon` text in any file
- [ ] No prompt injection patterns: no `---` YAML blocks mid-file, no `[INST]` tokens, no "ignore previous instructions"
- [ ] No personal credentials, API keys, or absolute home paths

## Quick Reference

### Starting a New Skill

```bash
# Create directory structure
mkdir -p skills/<name>/references skills/<name>/scripts

# Write entrypoint
# skills/<name>/SKILL.md (see spec-compliance.md for required frontmatter)

# Validate
bash skills/<name>/scripts/validate-skill.sh skills/<name>/
# Or use the shared script
bash skills/skill-author/scripts/validate-skill.sh skills/<name>/
```

### Using the Auto-Researcher Agent

For domains where you need research before writing:

```
See: agents/auto-researcher/AGENT.md

Invocation: "Research the [domain] domain for a new skill. 
Output a draft SKILL.md and references/ skeleton to skills/<name>/"

Budget: max 5 web searches, output to files only (not inline)
```

The auto-researcher produces a draft with correct structure. Review and refine before
treating the draft as final.

## Reference Files

| Task | Reference file |
|------|---------------|
| Full scoring rubric — frontmatter, structure, injection, line count | `references/evaluation-rubric.md` |
| agentskills.io spec — required vs. optional fields, body conventions | `references/spec-compliance.md` |
| Step-by-step authoring guide with examples | `references/authoring-guide.md` |

## Validation Script

```bash
bash skills/skill-author/scripts/validate-skill.sh <skill-dir>

# Example:
bash skills/skill-author/scripts/validate-skill.sh skills/aws/
```

Output: PASS / WARN / FAIL per check. Exit code 1 if any FAIL.
