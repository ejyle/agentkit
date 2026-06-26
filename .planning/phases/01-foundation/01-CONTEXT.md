# Phase 1: Foundation - Context

**Gathered:** 2026-06-08
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 1 delivers a working CLI binary where a user can run `agentkit install <name> --target claude` and get a skill or MCP server installed into Claude Code's config — resolving from the official curated agentkit-registry, tracking the install in per-assistant state files, and producing clean bubbletea-driven output. Also includes: `agentkit list`, `agentkit search`, `agentkit uninstall`, and `agentkit update`.

**Requirements in scope:** CLI-01, CLI-02, CLI-05, CLI-06, CLI-07, CLI-08, REG-01 (narrowed), REG-02, REG-05, REG-06, AST-01, MCP-01, MCP-03, MCP-05, MCP-06, MCP-07, SKL-01, SKL-02, SKL-03

**Requirements removed from scope (direction change):** CLI-09, REG-03, REG-04 — see Deferred section.

</domain>

<decisions>
## Implementation Decisions

### Product Philosophy (direction decision — shapes everything)
- **D-01:** agentkit is a **curated registry**, not a general package manager. There is ONE official curated agentkit-registry. Curation and benchmarking is the core product value. Skills/MCPs/agents are tested before being listed.
- **D-02:** v1 installs only from the official curated registry. No user-added custom sources, no mcpmarket.com. REG-01 narrowed: only the single official agentkit-registry repo, not "any GitHub repo with a manifest."

### Install Output Style
- **D-03:** `agentkit install`: bubbletea spinner for each phase (fetching registry → resolving package → running install adapter), then a clean single-line success: `✓ playwright@1.2.0 installed → ~/.claude/skills/playwright/ (claude)`
- **D-04:** Error output: single clear error line stating what failed and why, followed by context (registries checked, or what was missing). A suggested next command on the last line. Exit code 1. No stack traces.
  - Example: `✗ Error: playwright not found in agentkit-registry` / `Run: agentkit search playwright`
- **D-05:** `agentkit list`: table format with aligned columns — `PACKAGE`, `VERSION`, `TYPE`, `TARGET`, `REGISTRY`. Matches `go list -m all` style.
- **D-06:** `agentkit search <query>`: spinner while fetching registry (parallel across any future sources), then a deterministic ranked result list. Each entry: name, type, source label, one-line description.

### MCP Config Merge Behavior
- **D-07:** **Foreign conflict** (same `mcpServers` key exists but agentkit didn't install it): stop and prompt — show old vs new entry, ask `Overwrite? [y/N]`.
- **D-08:** **Upgrade** (agentkit installed this package, ownership confirmed via `installed.json`): auto-overwrite with a one-line notice: `⚠ playwright: upgrading 0.9 → 1.2.0`.
- **D-09:** **Uninstall merge**: read `settings.json` → delete only agentkit's `mcpServers` key → atomic write. All other keys and entries untouched. Never clobber user-written config.

### Installed State Tracking
- **D-10:** **Per-assistant files**: `~/.config/agentkit/claude/installed.json`, one file per target assistant. Cleaner separation; supports `agentkit list --target claude` queries without filtering a global file.
- **D-11:** **Full entry schema** — each installed package record includes:
  ```json
  {
    "name": "playwright",
    "version": "1.2.0",
    "type": "mcp",
    "install_path": "mcpServers.playwright",
    "installed_at": "2026-06-08T10:00:00Z",
    "source_url": "https://raw.githubusercontent.com/.../registry.json",
    "checksum": "sha256:abc123"
  }
  ```
- **D-12:** **Auto-create on first install**: agentkit creates `~/.config/agentkit/claude/` and `installed.json` transparently on first use. No `agentkit init` required. Matches `brew`/`cargo` behavior.

### Architecture Constraints (from STATE.md — already decided, not re-discussed)
- Use `os.UserHomeDir()` for all path resolution — never `os.Getenv("HOME")`
- Use `filepath.Join()` for all path construction
- Atomic file writes via temp + rename (`google/renameio` on Windows)
- Each assistant adapter detects config path at runtime — never hardcoded
- Post-install verify: re-read written MCP config after every write; fail loudly on invalid config
- Registry HTTP: explicit 3s connect / 10s read timeouts via `go-retryablehttp`
- ETag-based manifest cache; `--offline` fallback with visible warning

### Claude's Discretion
- Ranking algorithm for `agentkit search` results (exact match, then fuzzy, then description keywords — researcher/planner to decide)
- Exact bubbletea component structure for the spinner (style, frame rate, color)
- JSON schema versioning strategy for `installed.json` (whether to include a `schema_version` field)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Roadmap
- `.planning/REQUIREMENTS.md` — Full v1 requirement list with IDs (CLI-xx, REG-xx, AST-xx, MCP-xx, SKL-xx, BND-xx) and phase traceability
- `.planning/ROADMAP.md` — Phase 1 success criteria (5 acceptance tests that define "done")
- `.planning/PROJECT.md` — Core value, constraints, key decisions table, and technology stack rationale

### Technology Stack (from CLAUDE.md)
- `CLAUDE.md` §Technology Stack — Cobra v1.10.x, Bubbletea v1.x, Lipgloss v1.x, go-retryablehttp v0.7.x, BurntSushi/toml v1.x, and what was rejected and why
- `CLAUDE.md` §MCP Config Formats — per-assistant MCP config format reference for Claude Code, Copilot, Gemini, Codex, OpenCode
- `CLAUDE.md` §Cross-Platform Path Handling — XDG conventions per platform

### Registry Design
- `CLAUDE.md` §Registry Client — manifest schema (`registry.json` flat JSON index with `packages[]`), fetch URL pattern, ETag caching behavior
- `CLAUDE.md` §Version Management — `installed.json` lock file design, how `agentkit update` uses it

No external ADRs or specs yet — this is a greenfield project. Requirements fully captured above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- None — this is a greenfield project. No existing code.

### Established Patterns
- None yet. Phase 1 establishes the patterns all subsequent phases will follow.

### Integration Points
- `~/.claude/settings.json` — agentkit writes `mcpServers` entries and reads to detect foreign conflicts
- `~/.claude/skills/` — agentkit writes skill directories here for Claude Code
- `~/.config/agentkit/claude/installed.json` — agentkit's own state (created on first install)

</code_context>

<specifics>
## Specific Ideas

- The success line format for install: `✓ playwright@1.2.0 installed → ~/.claude/skills/playwright/ (claude)` — shows version, path, and target in one line
- Error output should always end with a suggested command the user can run next (not just a description of the problem)
- `agentkit list` table should print nothing (and a helpful message) when no packages are installed for the target, rather than an empty table

</specifics>

<deferred>
## Deferred Ideas

- **CLI-09 (`agentkit registry add/remove`)** — Removed from v1. agentkit uses a single curated registry; custom sources are not part of the v1 model.
- **REG-03 (mcpmarket.com API registry)** — Removed from v1. Curated-only model.
- **REG-04 (custom registry sources)** — Removed from v1. Curated-only model.
- **`--yes` / `--non-interactive` flag** — Could be added to skip MCP conflict prompts for CI use. Noted but not required for Phase 1.

</deferred>

---

*Phase: 1-Foundation*
*Context gathered: 2026-06-08*
