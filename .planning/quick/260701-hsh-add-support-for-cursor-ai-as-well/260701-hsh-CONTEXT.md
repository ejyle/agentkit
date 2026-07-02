# Quick Task 260701-hsh: add support for cursor AI as well. - Context

**Gathered:** 2026-07-01
**Status:** Ready for planning

<domain>
## Task Boundary

Add "Cursor AI" (the Cursor editor) as a new supported assistant target in agentkit, following the existing adapter pattern established for claude, copilot-cli, copilot-vscode, codex, gemini, opencode, and pi (see `internal/adapter/factory.go`, `internal/adapter/adapter.go`, `cmd/root.go` validTargets).

</domain>

<decisions>
## Implementation Decisions

### Skill support — CORRECTED after research + user verification
- Initial research (since superseded) proposed mapping skills to Cursor Rules (`.cursor/rules/*.mdc`) because it incorrectly found no user-global rules directory.
- **Corrected finding (confirmed via https://cursor.com/docs/skills and cross-checked via web search):** Cursor has native Agent Skills support with a genuine user-global directory at `~/.cursor/skills/<name>/`, using the exact same `SKILL.md` + optional `scripts/`/`references/`/`assets/` subdirectory structure agentkit already installs for Claude Code/Gemini/pi. **No format conversion needed.**
- Decision: `WriteSkill`/`RemoveSkill` for the Cursor adapter write standard SKILL.md skill folders to `~/.cursor/skills/<name>/` — same code path/pattern as the existing gemini/claude/pi adapters, just a different base path.

### Config scope
- MCP config: user-level only, written to `~/.cursor/mcp.json`. Matches the project's existing v1 constraint (user scope, no root/admin) and how every other adapter currently works.
- Skills: user-level only, written to `~/.cursor/skills/<name>/`. No scope conflict — this resolves the earlier (incorrect) concern about Rules being project-only.

### Target naming
- Single target: `"cursor"` (covers the Cursor IDE only, via `~/.cursor/mcp.json`). Do not split into `cursor-ide`/`cursor-cli` — that split is out of scope for this task.

### Claude's Discretion
- Exact Go file/struct naming for the new adapter (follow nearest existing adapter, e.g. `internal/adapter/cursor.go`, `CursorAdapter`).
- Whether `ReadMCPConfig`/`RemoveMCPConfig` need any Cursor-specific quirks beyond the standard JSON `mcpServers` key pattern used by claude/gemini/opencode.
- README/docs wording updates.

</decisions>

<specifics>
## Specific Ideas

No specific code references beyond the existing adapter pattern in `internal/adapter/` and `cmd/root.go`'s `validTargets` list.

</specifics>

<canonical_refs>
## Canonical References

No external specs — requirements fully captured in decisions above. Research phase should verify Cursor's actual MCP config path/format and Rules file format/location against current Cursor documentation before planning.

</canonical_refs>
