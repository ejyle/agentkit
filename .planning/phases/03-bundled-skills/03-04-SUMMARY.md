---
phase: 03-bundled-skills
plan: "04"
subsystem: skills
tags: [skills, context-mode, rtk, serena, skill-author, auto-researcher]
dependency_graph:
  requires: [03-01]
  provides: [skills/context-mode, skills/rtk, skills/serena, skills/skill-author, agents/auto-researcher]
  affects: [bundle/bundles.json context bundle, registry entries for context bundle]
tech_stack:
  added: []
  patterns: [agentskills.io progressive disclosure, SKILL.md + references/ structure]
key_files:
  created:
    - skills/context-mode/SKILL.md
    - skills/context-mode/references/routing-rules.md
    - skills/rtk/SKILL.md
    - skills/rtk/references/commands.md
    - skills/serena/SKILL.md
    - skills/serena/references/lsp-usage.md
    - skills/skill-author/SKILL.md
    - skills/skill-author/references/evaluation-rubric.md
    - skills/skill-author/references/spec-compliance.md
    - skills/skill-author/references/authoring-guide.md
    - skills/skill-author/scripts/validate-skill.sh
    - agents/auto-researcher/AGENT.md
  modified: []
decisions:
  - "context-mode skill adapted from CLAUDE.md routing rules in this repo (confirmed source per RESEARCH.md BND-07)"
  - "RTK skill adapted from ~/.dotfiles/.claude/RTK.md (meta commands, hook-based usage, name collision warning)"
  - "Serena skill built from Serena MCP plugin tool list (initial_instructions source; ~/.claude/skills/serena/ not present)"
  - "skill-author includes validate-skill.sh with 11 checks covering frontmatter, line count, references, injection, stubs, credentials"
  - "auto-researcher agent uses XML system prompt with 6-step workflow and 5-fetch budget"
metrics:
  duration: "~25 minutes"
  completed: "2026-06-09"
  tasks_completed: 1
  files_created: 12
---

# Phase 3 Plan 4: Context Bundle Skills and Skill-Author Meta-Skill Summary

**One-liner:** Four production-quality skills authored from real personal installs — context-mode (routing rules), RTK (token-optimized CLI proxy), Serena (LSP code intelligence), and skill-author meta-skill with validation script and bundled auto-researcher agent.

---

## What Was Built

### context-mode skill (`skills/context-mode/`)

Adapted from the CLAUDE.md routing rules in this repo (confirmed source per RESEARCH.md BND-07).
- `SKILL.md`: tool selection hierarchy table, blocked commands table, Quick Reference with call patterns, ctx management commands
- `references/routing-rules.md`: full routing rules — blocked patterns with alternatives, Bash allowed vs. redirected, ctx_batch_execute template, ctx_execute examples (shell/JS/Python), ctx_fetch_and_index pipeline, output constraints

### RTK skill (`skills/rtk/`)

Adapted from `~/.dotfiles/.claude/RTK.md` (read directly — file confirmed present).
- `SKILL.md`: meta commands (rtk gain, gain --history, discover, proxy), installation verification, hook-based usage explanation, name collision warning
- `references/commands.md`: full command reference with token savings percentages per command type (git, npm, build, test), hook configuration notes, filtering behavior description

### Serena skill (`skills/serena/`)

Built from the Serena MCP plugin tool list (the `~/.claude/skills/serena/` directory was not present; used the confirmed tool list from the objective and system prompt).
- `SKILL.md`: session setup, navigation tools, file/pattern search, structural edits, diagnostics, memory — all grouped into tool-category tables with When to use each
- `references/lsp-usage.md`: detailed tool parameters, full parameter signatures, 4 common workflows (refactoring, safe interface evolution, cross-package symbol search, rename workflow)

### skill-author meta-skill (`skills/skill-author/`)

Complete meta-skill for evaluating and authoring agentkit skills.
- `SKILL.md`: 8-step authoring workflow, evaluation checklist, Quick Reference with validation script usage, reference to auto-researcher agent
- `references/evaluation-rubric.md`: 6-section rubric (frontmatter quality, SKILL.md structure, references/ structure, spec compliance, injection safety, stub detection) with PASS/WARN/FAIL criteria and bash verification commands
- `references/spec-compliance.md`: full agentskills.io spec — required/optional fields with examples, body structure conventions, directory structure (minimal/standard/full), naming conventions, differences from anthropics/skills spec
- `references/authoring-guide.md`: 8-step guide from domain identification through PR submission; auto-researcher invocation; research source priority; PR description template; examples table pointing to existing skills
- `scripts/validate-skill.sh`: 11 checks — SKILL.md existence, line count, name field, description activation signal, name format, references/ presence, reference file line counts, stub detection, YAML separator injection, instruction-override patterns, credentials/personal paths

### auto-researcher agent (`agents/auto-researcher/`)

- `AGENT.md`: agent definition with XML system prompt, 6-step workflow (understand-domain → research-primary → research-sub-domains → write-SKILL.md → write-reference-files → report), output format spec, 5-fetch budget constraint, post-agent review checklist, when-NOT-to-use guidance

---

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing] Serena skill source: ~/.claude/skills/serena/ not present**
- **Found during:** Task 2 (fallback check as required)
- **Issue:** The plan specified `~/.claude/skills/serena/SKILL.md` as the source. This path does not exist. The objective confirmed this case: "Serena source: Serena is an MCP plugin (not a skill file). Call mcp__plugin_serena_serena__initial_instructions."
- **Fix:** Built the skill from the confirmed Serena tool list provided in the objective (19 tools listed). The skill accurately reflects the real Serena MCP plugin capabilities.
- **Files modified:** skills/serena/SKILL.md, skills/serena/references/lsp-usage.md
- **Commit:** 08d79b6

### T-03-11 / T-03-13 Threat Mitigations Applied

- context-mode: No API keys, no personal paths, no user-specific config values — extracted only routing rules and tool hierarchy
- RTK: Stripped the `@RTK.md` include reference from CLAUDE.md; included only meta commands, hook behavior, installation verification, and name collision warning — no personal tokens or absolute paths
- Serena: Built from tool list, not from personal ~/.claude/ data — no personal information possible

---

## Known Stubs

None — all skill files contain real, actionable content. The word "stub" appears in `evaluation-rubric.md` and `authoring-guide.md` only as documentation about how to detect stubs in skills being reviewed, which is the intended content for a skill-author meta-skill.

---

## Threat Flags

None — no new network endpoints, auth paths, or schema changes introduced. All files are documentation/script only.

---

## Self-Check

Files exist:
- skills/context-mode/SKILL.md: FOUND
- skills/context-mode/references/routing-rules.md: FOUND
- skills/rtk/SKILL.md: FOUND
- skills/rtk/references/commands.md: FOUND
- skills/serena/SKILL.md: FOUND
- skills/serena/references/lsp-usage.md: FOUND
- skills/skill-author/SKILL.md: FOUND
- skills/skill-author/references/evaluation-rubric.md: FOUND
- skills/skill-author/references/spec-compliance.md: FOUND
- skills/skill-author/references/authoring-guide.md: FOUND
- skills/skill-author/scripts/validate-skill.sh: FOUND (executable)
- agents/auto-researcher/AGENT.md: FOUND

Line counts (all under 500): context-mode 99, rtk 77, serena 103, skill-author 97

Commit 08d79b6: FOUND

## Self-Check: PASSED
