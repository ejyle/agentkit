---
quick_id: 260626-jvc
date: 2026-06-26
status: Ready for planning
---

# Quick Task 260626-jvc: Create skill-finder skill — Context

**Gathered:** 2026-06-26
**Status:** Ready for planning

<domain>
## Task Boundary

Two parts:
1. **Cleanup:** Delete all skills from `./skills/` EXCEPT `skill-author`. Skills to remove: context-mode, azure, serena, playwright, gcp, cicd, github, rtk, aws, external/ (and all sub-directories).
2. **skill-finder skill:** Create a new `./skills/skill-finder/` skill (with SKILL.md + references/) that searches for popular/maintained Claude Code skills and optionally installs them to `./skills/`.

</domain>

<decisions>
## Implementation Decisions

### Cleanup scope
- Target: `./skills/` directory in this project root (NOT `~/.claude/skills/`)
- Keep: `./skills/skill-author/` only
- Delete: all other top-level dirs and `external/` subtree

### Install target for --add
- `./skills/` in this project — so agentkit can package these skills for distribution
- NOT the global `~/.claude/skills/`

### Search sources for skill-finder
- Curated registries first: agentkit manifest, mcpmarket.com/tools/skills, known GitHub repos
- Fallback web search for anything not indexed
- Also search the specific sites user listed: claude-code-skills.com, capafy.ai, agent-skills-directory
- Validate quality signals: maintenance status (last commit date), download/install count, GitHub stars, number of users

### Skill-finder behavior
- No `--add` flag → research mode: search web, present ranked list with quality signals, ask user at the end if they want to add any
- `--add` flag → auto-add mode: still researches, but adds top results to `./skills/` without asking at the end
- Skill format: each installed skill follows the SKILL.md + references/ split (same pattern as existing skills)

### Claude's Discretion
- Exact ranking algorithm for quality signals (stars + recency + downloads)
- How to handle skills with no download count (GitHub-only)
- Whether to use a `scripts/` dir inside skill-finder for any helper logic

</decisions>

<specifics>
## Specific Ideas

- User referenced: `npx skills add https://github.com/vercel-labs/skills --skill find-skills` as inspiration for the CLI pattern
- The skill should be named `skill-finder`
- References split: heavy content (source list, validation logic, example outputs) goes in `references/` — SKILL.md body stays under 500 lines
- This skill lives in `./skills/skill-finder/` and is packaged with agentkit

</specifics>

<canonical_refs>
## Canonical References

- `./skills/skill-author/SKILL.md` — existing skill structure to follow as template
- `./skills/skill-author/references/` — example of references split pattern
- Sites to search: mcpmarket.com/tools/skills, claude-code-skills.com, capafy.ai, agent-skills-directory

</canonical_refs>
