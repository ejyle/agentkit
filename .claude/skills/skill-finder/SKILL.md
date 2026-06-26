---
name: skill-finder
description: >
  Use when you want to discover skills and MCP servers to add to the agentkit registry.
  Generates properly formatted registry entries (domain.Package JSON) ready to paste into
  testdata/registry.json or the live agentkit-registry. Knows install methods, mcp_entry
  format, multi_skill flag, and path verification rules.
license: Apache-2.0
---

## When to Use

Activate this skill when working inside the agentkit project and:

- The user wants to find a skill or MCP server to add to the agentkit registry
- The user wants to know how to add something to `testdata/registry.json`
- The user needs to figure out the correct install method, path, or mcp_entry for a package
- The user invokes /skill-finder with or without the --add flag

> **This is the agentkit project skill-finder.** Output is registry entries for the
> agentkit distribution registry, not direct installs to Claude Code. To install skills
> directly to `~/.claude/skills/`, use the global skill-finder instead.

## Behavior

### Research Mode (no --add flag)

**Step 1: Search for skills and MCP servers.**
See `references/registry-sources.md` for the full list and priority order. Search the same
registries as the global skill-finder, but also check official MCP sources:
- `awslabs/mcp` for AWS official MCP servers
- `googleapis/gcloud-mcp` and `google/mcp` for Google Cloud official MCP servers
- `microsoft/mcp` and `microsoft/azure-skills` for Azure official MCP servers and skills

**Step 2: Deduplicate and rank.**
Apply the quality scoring formula from `references/quality-signals.md`. Include both skills
(type: `skill`) and MCP servers (type: `mcp`) in results.

**Step 3: Present a ranked table.**
Display columns: Rank, Name, Type (skill/mcp), Source (✅ Official / community), Stars,
Last Commit, Quality Score, Install Method.
Prefix quality score with HIGH (>=70), MED (40-69), or LOW (<40).

**Step 4: Prompt the user.**
Ask: "Would you like to generate registry entries for any of these? (enter numbers, e.g. '1 3 5', or 'none')"

**Step 5: Generate registry entries on confirmation.**
For each selection, produce a complete `domain.Package` JSON entry following
`references/registry-format.md`. Verify the repo structure before setting `path`.
Output entries ready to paste into `testdata/registry.json`.

### Auto-Add Mode (--add flag)

**Step 1: Same discovery flow as research mode steps 1-3.**

**Step 2: Generate entries for top 5 results.**
Select the top 5 by quality score, verify repo structure for each, and produce complete
`domain.Package` JSON entries.

**Step 3: Append to testdata/registry.json.**
Insert the generated entries into the `packages` array in `testdata/registry.json`.
Check for duplicate `name` fields before inserting — skip if already present.

**Step 4: Print a summary.**
List each result as "Added: NAME (type: TYPE, method: METHOD)" or "Skipped: NAME — REASON".

## Quick Reference

| Command | Behavior |
|---------|----------|
| `/skill-finder` | Research mode: ranked table of skills/MCPs + prompt to generate registry entries |
| `/skill-finder --add` | Auto-add mode: generate + append top 5 entries to testdata/registry.json |

**Slopsquatting warning:** Before generating a registry entry for any skill/MCP discovered
via web search, verify the GitHub repo URL resolves to the claimed organization. See
`references/registry-format.md` for the full verification checklist.

## Reference Files

| Task | Reference file |
|------|---------------|
| Full registry list with URLs, scrape approach, and priority order | `references/registry-sources.md` |
| Quality ranking rubric — formula, weights, factor definitions, edge cases | `references/quality-signals.md` |
| domain.Package JSON schema, install methods, mcp_entry format, path rules | `references/registry-format.md` |
