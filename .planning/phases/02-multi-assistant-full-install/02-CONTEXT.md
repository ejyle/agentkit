# Phase 2: Multi-Assistant & Full Install - Context

**Gathered:** 2026-06-08
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 2 delivers working adapters for all 5 target coding assistants (Copilot, Codex, Gemini, OpenCode, Pi) plus the uvx (Python) and Docker MCP install methods — enabling `agentkit install <name> --target <any>` to work end-to-end for any supported assistant and any install method.

**Requirements in scope:** AST-02, AST-03, AST-04, AST-05, AST-06, MCP-02 (via uvx), MCP-04

**⚠ SCOPE CHANGE from ROADMAP.md:** REG-03 (mcpmarket.com) and REG-04 (custom registry add) are REMOVED from v1 scope entirely. The curated-only model (D-01/D-02 from Phase 1) is confirmed and extended — the official agentkit-registry is the only source, forever in v1. ROADMAP.md success criteria 3 and 4 are invalidated and should NOT be planned against.

</domain>

<decisions>
## Implementation Decisions

### Registry Model (CONFIRMED)
- **D-01:** Curated-only model is permanent for v1: one official agentkit-registry, no user-added custom sources, no mcpmarket.com integration.
- **D-02:** REG-03 (mcpmarket.com) and REG-04 (`agentkit registry add`) removed from v1 scope entirely. Users wanting to add packages submit a feature request; agentkit team performs quality checks before listing.
- **D-03:** Phase 2 has no new registry work. The registry infrastructure from Phase 1 (REG-05, REG-06) is sufficient.

### Pi Adapter (AST-06)
- **D-04:** Best-effort partial adapter: researcher discovers what pi.dev's CLI/MCP/skill system actually supports, then implement those operations. Do NOT assume capabilities.
- **D-05:** Implement BOTH WriteMCPConfig AND WriteSkill if pi.dev has those mechanisms. Don't implement one without the other.
- **D-06:** Unsupported operations return `ErrNotSupported` with a clear error message (e.g., `"pi adapter: WriteSkill not supported — pi.dev has no CLI skill directory"`). Never silently no-op.
- **D-07:** Runtime config path detection using the same pattern as ClaudeCodeAdapter: check file existence order at runtime, fall back to primary path on first write.

### Docker Installer (MCP-04)
- **D-08:** Config entry format: `command="docker"`, `args=["run", "-i", "--rm", "image:tag", ...manifest_extra_args]`. The manifest can specify additional Docker args (volumes, env vars) that are appended to the args array.
- **D-09:** Docker install step: run `docker pull <image>` at install time (not lazy). Fail loudly if image pull fails.
- **D-10:** `DockerInstaller.IsAvailable()` checks for `docker` on PATH. If missing: error `"docker not found on PATH"` with install URL and exit code 1. Same pattern as `ErrNodeNotFound` for npx.

### Python/uvx Installer (MCP-02, replaces pip)
- **D-11:** Use `uvx` instead of `pip` for Python-based MCP server installs. uvx provides isolated environments without virtualenv management.
- **D-12:** Config entry default: `command="uvx"`, `args=["mcp-package", ...manifest_args]`. Researcher must verify the exact invocation pattern used by real Python MCP servers before finalizing.
- **D-13:** `UvxInstaller.IsAvailable()` checks for `uvx` on PATH. If missing: `ErrUvxNotFound` with install URL.

### Gemini Adapter (AST-04)
- **D-14:** Researcher must verify: (1) exact directory where Gemini CLI loads skill files, (2) whether it expects SKILL.md or GEMINI.md as the entrypoint filename. Do NOT assume ~/.gemini/skills/<name> — verify before coding.
- **D-15:** Researcher must verify: exact `settings.json` format for Gemini MCP entries, and whether a shared JSON merge base struct (reusing ClaudeCodeAdapter logic) is appropriate or if the formats diverge.

### OpenCode Adapter (AST-05)
- **D-16:** Full implementation required — not a stub. OpenCode uses `"mcp"` key (not `"mcpServers"`) and requires a `"type"` field on each entry. Researcher verifies exact schema from opencode.ai docs.
- **D-17:** Runtime path detection for `~/.config/opencode/opencode.json`. Same atomic write + post-verify pattern as ClaudeCodeAdapter.

### All New Adapters (AST-02, AST-03, AST-04, AST-05, AST-06)
- **D-18:** Every adapter must satisfy the `AssistantAdapter` interface fully. For Pi: unimplemented operations return `ErrNotSupported`. For all others: all 5 operations must be implemented (WriteMCPConfig, RemoveMCPConfig, ReadMCPConfig, WriteSkill, RemoveSkill).
- **D-19:** All adapters carry forward Phase 1 patterns: homeDir injection for testability, atomic writes via renameio, post-install verify (re-read + confirm key presence), D-07/D-08/D-09 merge behavior (foreign conflict → ErrForeignConflict, agentkit-owned → auto-overwrite).

### Claude's Discretion
- Whether to extract a shared JSON merge base struct for JSON-based adapters (Claude/Gemini/Codex) — depends on what the researcher finds about format divergence
- Exact uvx invocation args format for common Python MCP packages — researcher to verify
- Whether Copilot CLI vs VS Code Copilot requires separate sub-adapters in Phase 2 or single adapter covers both — researcher verifies

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Roadmap
- `.planning/REQUIREMENTS.md` — Full v1 requirement list; AST-02 through AST-06, MCP-02 (now uvx), MCP-04 are Phase 2 items. REG-03 and REG-04 are OUT OF SCOPE for v1.
- `.planning/ROADMAP.md` — Phase 2 scope; NOTE: success criteria 3 and 4 reference mcpmarket.com and custom registry — these are INVALIDATED by D-01/D-02. Plan only against criteria 1 and 2.
- `.planning/PROJECT.md` — Core constraints, technology stack rationale
- `.planning/phases/01-foundation/01-CONTEXT.md` — Phase 1 decisions that carry forward (D-07 through D-19 merge/conflict/uninstall patterns, architecture constraints)

### Technology Stack
- `CLAUDE.md` §MCP Config Formats — per-assistant MCP config format reference for Claude Code, Copilot, Gemini, Codex, OpenCode
- `CLAUDE.md` §Cross-Platform Path Handling — XDG conventions per platform
- `CLAUDE.md` §Technology Stack — existing dependency decisions (Cobra, Bubbletea, Lipgloss, go-retryablehttp, BurntSushi/toml)

### Existing Adapter Pattern (read before coding new adapters)
- `internal/adapter/adapter.go` — AssistantAdapter interface, ErrForeignConflict type
- `internal/adapter/claude.go` — ClaudeCodeAdapter reference implementation: homeDir injection, runtime path detection pattern, readRawConfig(), renameio writes, post-install verify
- `internal/config/paths.go` — InstalledStatePath, SkillInstallPath (add explicit cases for each new target)
- `internal/installer/installer.go` — MCPInstaller interface, NewInstaller factory (add uvx and docker cases)
- `internal/installer/npx.go` — Reference implementation for how installers check availability and handle errors

### Research Flags (from STATE.md — researcher MUST resolve these)
- Verify Copilot CLI vs VS Code Copilot adapter divergence: are they the same config or separate sub-adapters?
- Verify Codex CLI MCP config key names at latest version (key name may differ from CLAUDE.md reference)
- Verify exact pi.dev CLI MCP/skill mechanism — adapt D-04/D-05 based on findings
- Verify Gemini CLI exact skill directory path and SKILL.md vs GEMINI.md entrypoint convention
- Verify uvx invocation pattern for real Python MCP servers (command/args format)
- Verify OpenCode 'mcp' key schema and 'type' field values from opencode.ai docs

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `AssistantAdapter` interface (`internal/adapter/adapter.go`) — all new adapters implement this; no changes needed to the interface
- `ClaudeCodeAdapter` (`internal/adapter/claude.go`) — template for all new JSON-based adapters; pattern for homeDir injection, runtime path detection, atomic writes
- `MCPInstaller` interface + `NewInstaller` factory (`internal/installer/installer.go`) — add `InstallMethodUvx` and `InstallMethodDocker` to the switch
- `config.SkillInstallPath()` (`internal/config/paths.go`) — add explicit `gemini`, `copilot`, `codex`, `opencode`, `pi` cases instead of relying on the default fallback
- `ErrNodeNotFound` pattern (`internal/installer/npx.go`) — template for `ErrDockerNotFound`, `ErrUvxNotFound`

### Established Patterns
- Atomic writes: `renameio.WriteFile()` — all adapters use this, never `os.WriteFile`
- Post-install verify: re-read config after write, confirm key presence, fail if missing
- homeDir injection: `homeDir string` field + `home() (string, error)` method — enables unit tests without touching real `~`
- Per-assistant installed.json: `config.InstalledStatePath(target)` — same for all new targets (no change needed)
- `map[string]interface{}` for config merge: preserve all unmanaged keys (D-09 never clobbers)

### Integration Points
- `internal/service/install.go` (`InstallService`) — currently only calls `NewInstaller(method)` and the adapter's `WriteMCPConfig`/`WriteSkill`; new installers and adapters plug in via existing interfaces
- `cmd/root.go` — `--target` flag validation must include the 5 new targets (copilot, codex, gemini, opencode, pi)
- `internal/config/paths.go` `SkillInstallPath()` — add cases for each new target assistant

</code_context>

<specifics>
## Specific Ideas

- Docker `args` array should follow: `["run", "-i", "--rm", "image:tag"]` as base, with manifest-specified extra args appended in order
- ErrNotSupported for Pi should include the adapter name and operation in the message so the user knows exactly what failed
- Copilot adapter research flag: if CLI vs VS Code diverge, implement as two sub-adapters registered as `"copilot-cli"` and `"copilot-vscode"` rather than one confused adapter

</specifics>

<deferred>
## Deferred Ideas

- **REG-03 (mcpmarket.com API registry)** — Removed from v1 entirely. May be revisited post-GA if community demand warrants it.
- **REG-04 (custom registry add)** — Removed from v1 entirely. Quality-first curation model takes precedence.
- **`agentkit registry add/remove` commands** — Not part of v1. Deferred.
- **Background config agent (per-project facts system)** — Deferred to v2 per REQUIREMENTS.md.
- **Per-project install scope** — Deferred to v3+ per REQUIREMENTS.md.

</deferred>

---

*Phase: 2-Multi-Assistant & Full Install*
*Context gathered: 2026-06-08*
