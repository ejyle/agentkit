# Research Summary: agentkit

_Synthesized: 2026-06-08_

---

## Executive Summary

agentkit is a Go CLI package manager for AI agent skills, MCP servers, and coding agents — analogous to Homebrew or mise but targeting AI assistant ecosystems (Claude Code, GitHub Copilot CLI, Codex, Gemini CLI, OpenCode) instead of system tools. The domain is immature and fragmented: each assistant uses a different config file path and format, the open agentskills.io spec is recent, and MCP config locations have already broken silently between assistant versions. The recommended architecture mirrors established package managers (ordered registry sources, typed adapters per target, local install manifest) but must treat every config path as runtime-discovered, not hardcoded.

The core technical challenge is not installation logic — that is straightforward Go — but defensive multi-adapter config writing. MCP config paths have moved without notice (Claude Code issue #4976, Copilot breaking change in #3019), and each assistant uses a different schema. Every adapter must read before writing (merge, not overwrite), validate after writing, and fail loudly when the written config cannot be re-read. The second major risk is the background config agent: without strict token budget enforcement (max 3 tool calls, structured delta input only, pre-invocation cost estimate), a per-skill config writer can balloon to $5+ per invocation.

Build order strictly follows dependency layers: domain types and config I/O first, then registry and adapter interfaces, then service orchestration, then CLI commands, then the background agent, then bundled skills. No layer should be started before its dependencies have clean interfaces — especially the registry abstraction, which must be stable before any CLI commands are wired.

---

## Recommended Stack (prescriptive)

| Library | Version | Why Chosen | Do NOT Use Instead |
|---------|---------|------------|-------------------|
| `spf13/cobra` | v1.10.x | De facto Go CLI standard; 44K stars; used by Kubernetes, Docker, GitHub CLI; best subcommand + shell completion ergonomics | `urfave/cli` (positional-arg style), `flag` stdlib (no subcommands) |
| `charmbracelet/bubbletea` | v1.x | Best-in-class Go TUI; pairs with Cobra for interactive confirm prompts only | Full TUI in non-interactive paths |
| `charmbracelet/lipgloss` | v1.x | Terminal styling from same Charm ecosystem | Raw ANSI escape codes |
| `hashicorp/go-retryablehttp` | v0.7.x | Auto-retry + exponential backoff + Retry-After support; HashiCorp production-grade | `http.DefaultClient` (no timeout), custom retry loops |
| `BurntSushi/toml` | v1.x | Only TOML dep; Codex is the only assistant using TOML config | Manual TOML string building |
| `kaptinlin/jsonschema` | latest | Draft 2020-12, struct validation, default-aware unmarshaling | `gojsonschema` (older, less active) |
| stdlib `encoding/json` | — | Sufficient for all JSON config R/W; zero deps | `spf13/viper` (10+ transitive deps, overkill) |
| stdlib `os`, `path/filepath` | — | `os.UserHomeDir()` / `os.UserConfigDir()` / `os.UserCacheDir()` cover all platforms | `adrg/xdg` (unnecessary dep), `os.Getenv("HOME")` (broken on Windows) |
| `github.com/google/renameio` | latest | Atomic file write on Windows (stdlib `os.Rename` is not atomic on Windows per golang/go#8914) | Direct file.Write without rename pattern |
| GoReleaser v2 | v2.x | Unambiguous standard for Go CLI cross-compilation and Homebrew tap + Scoop auto-update | Custom Makefile release scripts |
| Go | 1.22+ | No CGO needed; clean cross-compile to 5 platforms | Any CGO dependency |

**Config format:** JSON for all agentkit-owned files; TOML only for writing Codex config.

---

## Architecture in One Page

```
+-------------------------------------------------------------+
|                     CLI Layer (Cobra)                       |
|   install / search / update / list / registry / bundle      |
+------------------------+------------------------------------+
                         | calls
+------------------------v------------------------------------+
|                    Service Layer                            |
|   InstallService  SearchService  UpdateService             |
|   (orchestrates; no direct I/O knowledge)                  |
+----+------------------+-------------------+----------------+
     |                  |                   |
+----v----+   +---------v------+   +--------v-------------+
|Registry |   |  Assistant     |   |   Config Store        |
|Manager  |   |  Adapter       |   |  ~/.agentkit/         |
|(ordered |   |  Registry      |   |  state.json           |
| sources)|   |  (per-target)  |   |  .agent-utils/        |
+----+----+   +---------+------+   |  config.json          |
     |                  |          +------------------------+
+----+---------+  +-----v-----------------+
| Registry     |  | MCP Install Adapters  |
| Sources      |  | npx / pip /           |
| GitHub/API/  |  | binary / docker       |
| Local        |  +-------+---------------+
+--------------+          | post-install
               +----------v-----------+
               | Background Config    |
               | Agent (XML-scoped,   |
               | 3-call max, delta    |
               | input only)          |
               +----------------------+
```

**Data flow (install):**

```
agentkit install playwright --target claude
  -> RegistryManager.Resolve("playwright")   [GitHub > GSD > mcpmarket, first-match wins]
  -> MCPInstaller(npx).Install()             [exec: npx -y @playwright/mcp]
  -> ClaudeCodeAdapter.WriteMCPConfig()      [read existing -> merge -> atomic write -> verify]
  -> ConfigStore.RecordInstalled()           [~/.agentkit/state.json: version + sha256 + source_registry]
  -> print "Installed playwright@1.2.0 -> Claude Code"
```

**Build order (strict dependency sequence):**

1. Domain types: `Package`, `Manifest`, `Fact`, `MCPServerEntry`; path utilities wrapping `os.UserHomeDir`
2. Config store: read/write `state.json` and `.agent-utils/config.json` with atomic writes, `recorded_at` timestamps, `source_registry` field
3. Registry interface + `GitHubManifestRegistry` with ETag cache + `LocalRegistry`
4. Assistant adapter interface + `ClaudeCodeAdapter` (full); stubs for others
5. MCP installer interface + `NpxInstaller` + `BinaryInstaller`
6. `RegistryManager` (ordered resolution, search fan-out, dedup) + `InstallService`
7. CLI commands: `install`, `search`, `list`, `update`, `registry add/remove`
8. Remaining adapters (Copilot CLI, Gemini, Codex TOML, OpenCode) + `PipInstaller`, `DockerInstaller`
9. Background config-writer agent prompt + skill-end hook with pre-cost estimate gate
10. Bundled skills: AWS, GCP, Azure, Playwright, context-mode, RTK, Serena, GitHub, CI/CD

---

## Table Stakes Features

Users leave or never adopt if these are missing:

- `agentkit install <name>` — single command from any registered registry
- `agentkit list` — installed packages with version, source, target assistant
- `agentkit update [name]` — update one or all; zero-friction currency
- `agentkit search <query>` — cross-registry ranked discovery before committing to install
- `agentkit uninstall <name>` — clean removal; partial artifacts are a known pain in manual installs
- MCP config written correctly per target assistant — wrong format means skill silently never loads
- User-scope default, no root/sudo required — developers refuse tools needing admin
- Single binary, no runtime dependency — Python CLIs with venv friction is well-documented
- Progress output with clear success/failure — silent tools feel broken
- SKILL.md + `references/` + `scripts/` structure honoured on install — deviating from the open standard breaks cross-assistant portability

---

## Key Differentiators

What agentkit does that nothing else does:

- **Cross-assistant atomic install with `--target`** — no other tool normalises Claude/Copilot/Codex/Gemini/OpenCode in one scriptable CLI; GSD's installer is interactive-only
- **`--bundle cloud/ci/gsd/context`** — curated preset collections installed atomically; new machine to fully instrumented in one command
- **Per-project `.agent-utils/config.json` with drift detection** — discovered infra facts persist; skills skip re-discovery and run cheaper every subsequent session
- **Background config agent (token-adaptive)** — auto-writes cheap discoveries inline, prompts only when expensive; no existing skill manager has this UX
- **Progressive disclosure enforced at install time** — SKILL.md line count validated (<500 lines), references/ loaded on demand; 40-60% cold-start token reduction
- **Manifest-driven extensible registries** — providers publish their own `registry.json`; no CLI PR needed to add packages (Homebrew tap model)
- **`agentkit install gsd`** — atomic install of the full GSD workflow suite, replacing the current interactive multi-step setup

---

## Top 5 Pitfalls to Avoid (ranked by severity)

1. **MCP config path instability (Critical)** — Claude Code silently moved from `settings.json` to `~/.claude.json`; Copilot broke `.vscode/mcp.json` without announcement. Prevention: each adapter detects existing file locations at runtime, never compiles paths as constants, and runs a post-install verify step that re-reads the written config.

2. **Background agent token runaway (Critical)** — agentic loops accumulate full conversation history; 10 iterations of a 500-token agent can cost $5+. Prevention: hard cap of 3 tool calls max, structured delta input only (never filesystem traversal), pre-invocation cost estimate gates auto-write vs user prompt.

3. **Windows path construction failures (Critical)** — `os.Getenv("HOME")` is unset on Windows; raw `/` separators produce root-relative paths. Prevention: `os.UserHomeDir()` everywhere, `filepath.Join()` for all path construction, assert non-empty at startup.

4. **Registry fetch with no timeout and no cache (Critical)** — `http.DefaultClient` has no timeout; corporate firewalls block GitHub. Prevention: explicit 3s connect / 10s read timeouts via `go-retryablehttp`, ETag-based manifest cache with visible warning on fallback, `--offline` flag.

5. **Concurrent config file writes causing data loss (Moderate)** — two concurrent Claude Code sessions writing `.agent-utils/config.json` simultaneously clobber each other. Prevention: atomic write via temp file + `os.Rename` (Unix) or `github.com/google/renameio` (Windows), file lock for concurrent process safety.

---

## Phase Recommendations

### Phase 1 — Foundation (CLI core + config + registry backbone)
**Delivers:** Working `agentkit install <name> --target claude` end-to-end for a single assistant with full registry resolution.
**Covers:** Domain types, path utilities, config store with atomic writes and timestamps, `GitHubManifestRegistry` with ETag cache, `ClaudeCodeAdapter` (full implementation), `NpxInstaller` + `BinaryInstaller`, `RegistryManager`, `InstallService`, CLI commands: `install`, `list`, `registry add/remove`.
**Must avoid:** Use `os.UserHomeDir()` before any other path code; establish `recorded_at` and `source_registry` in config schema from day one — retrofitting timestamps is painful.
**Research flag:** Standard patterns; no additional research phase needed.

### Phase 2 — Multi-assistant + full registry
**Delivers:** All 5 target assistants supported; all 4 install methods; mcpmarket.com and GSD registries; `search`, `update`, `uninstall` commands.
**Covers:** Copilot CLI adapter, Gemini adapter, Codex adapter (TOML), OpenCode adapter; `PipInstaller`, `DockerInstaller`; `MCPMarketRegistry`, `GSDCoreRegistry`; sudo detection; directory ownership pre-check; post-install verify step; name conflict resolution with deterministic priority.
**Must avoid:** VS Code Copilot and Copilot CLI are separate adapters (different top-level keys); schema version detection must be runtime, not compile-time.
**Research flag:** Verify Codex MCP config key names and Copilot divergence state against latest versions before coding these adapters.

### Phase 3 — Background agent + project config
**Delivers:** Per-project `.agent-utils/config.json` auto-populated; token-adaptive background config-writer agent; drift detection.
**Covers:** XML-prompted config-writer agent prompt, skill-end hook with pre-cost estimate, 3-call hard cap, structured delta input protocol (no filesystem traversal by agent), `agentkit config refresh` command.
**Must avoid:** Agent must receive structured diff only — calling skill extracts and summarizes delta before handoff; pre-invocation cost gate must fire before invoking, not after.
**Research flag:** Needs research phase — per-assistant CLI flags for subagent invocation (only `claude --agent` confirmed; Copilot/Gemini/Codex equivalents unknown).

### Phase 4 — Bundled skills + bundle command
**Delivers:** `agentkit install --bundle cloud/ci/gsd/context`; all 9 initial bundled skills authored and validated; `bundle dump` snapshot command.
**Covers:** Bundle manifest format in `registry.json`, `bundle install/dump/check` commands, AWS/GCP/Azure/Playwright/context-mode/RTK/Serena/GitHub/CI-CD skills authored to SKILL.md spec with line count validation.
**Must avoid:** GSD slash-command rewrite (hyphen for Claude, colon for Gemini) must be handled in the adapter layer, not in the skill files themselves.
**Research flag:** Standard authoring patterns; no additional research phase needed.

### Phase 5 — Distribution + hardening
**Delivers:** GoReleaser v2 pipeline; Homebrew tap; `agentkit doctor`; Windows-safe binary distribution.
**Covers:** `.goreleaser.yaml`, GitHub Actions release workflow, Homebrew tap repo, Scoop manifest, `doctor` diagnostic checks, `~/.local/bin` PATH detection and warning, `agentkit verify <name>` post-install validator.
**Must avoid:** Pre-compiled binary must be primary distribution path (not `go install`) for corporate/air-gapped environments.
**Research flag:** Standard GoReleaser patterns; no additional research phase needed.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All libraries verified via official docs and active 2025-2026 releases |
| Features | HIGH | Grounded in agentskills.io spec, GSD source, Microsoft Agent Skills, brew/npm/mise precedents |
| Architecture | HIGH (patterns) / MEDIUM (config paths) | Registry and adapter patterns are solid; Codex and OpenCode config paths need runtime verification |
| Pitfalls | MEDIUM-HIGH | Windows paths and token runaway well-documented; config path instability confirmed via GitHub issues |

**Gaps to address during planning:**

- Codex CLI MCP config key structure — LOW confidence; verify at Phase 2 start before coding adapter
- Copilot CLI vs VS Code Copilot adapter divergence — active known issue with no published resolution; design for two separate adapters now
- Per-assistant subagent invocation flags (Phase 3) — `claude --agent` confirmed; Copilot/Gemini/Codex/OpenCode equivalents unknown; needs research phase before Phase 3
- mcpmarket.com API pagination and rate limits — no official documentation found; design for graceful degradation and treat as optional registry source

---

## Sources (aggregated)

- Cobra v1.10.x: https://github.com/spf13/cobra
- Bubbletea: https://github.com/charmbracelet/bubbletea
- hashicorp/go-retryablehttp: https://github.com/hashicorp/go-retryablehttp
- GoReleaser: https://goreleaser.com/
- agentskills.io Specification: https://agentskills.io/specification
- Microsoft Agent Skills: https://github.com/MicrosoftDocs/Agent-Skills
- open-gsd/gsd-core: https://github.com/open-gsd/gsd-core
- Claude Code MCP config instability: anthropics/claude-code#4976, #32398
- Copilot MCP breaking change: github/copilot-cli#3019
- kaptinlin/jsonschema: https://github.com/kaptinlin/jsonschema
- Codex CLI MCP config: https://developers.openai.com/codex/mcp
- Gemini CLI MCP setup: https://geminicli.com/docs/tools/mcp-server/
- OpenCode MCP servers: https://opencode.ai/docs/mcp-servers/
- mcpmarket.com: https://mcpmarket.com/
- Homebrew Bundle: https://docs.brew.sh/Brew-Bundle-and-Brewfile
- mise configuration: https://mise.jdx.dev/configuration.html
- Progressive disclosure token savings: https://ardalis.com/optimizing-ai-agents-with-progressive-disclosure/
- golang/go#8914 (os.Rename not atomic on Windows)
- google/renameio: https://github.com/google/renameio
