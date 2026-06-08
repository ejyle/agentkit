# Pitfalls Research: agentkit

**Domain:** Go CLI for AI agent skill/MCP management
**Researched:** 2026-06-08
**Confidence:** MEDIUM-HIGH (verified across official docs, GitHub issues, and community sources)

---

## Cross-Platform Go

### Critical: HOME vs USERPROFILE Resolution
**What goes wrong:** Hardcoding `~` or `os.Getenv("HOME")` fails silently on Windows. `HOME` is not always set on Windows; `USERPROFILE` is the Windows equivalent. Code that constructs paths like `os.Getenv("HOME") + "/.claude"` returns empty string on Windows, producing paths like `"/.claude"` (root-relative), which either silently no-ops or writes to unexpected locations.

**Prevention:** Always use `os.UserHomeDir()` — it is the only cross-platform API. Never concatenate path segments with `/` or `\`; use `filepath.Join()` exclusively. Assert that `UserHomeDir()` returns a non-empty value at startup and abort with a clear error if not.

**Warning signs:** Any string containing `os.Getenv("HOME")` or a raw `/` path separator in path construction code.

**Phase:** Phase 1 (CLI Core) — establish `paths.go` abstraction before writing any install logic.

---

### Critical: Symlink Handling on Windows
**What goes wrong:** Windows requires elevated privilege or Developer Mode for `os.Symlink()`. Skills and MCP binaries that use symlinks for aliasing or shim creation will fail on standard Windows installs. `filepath.WalkDir` follows symlinks differently than on Unix, risking infinite loops or missed entries.

**Prevention:** Avoid symlinks entirely for v1. Use hard copies or wrapper scripts. When traversal is needed, use `filepath.WalkDir` with explicit symlink detection (`os.Lstat` + `os.ModeSymlink` check) to avoid following cycles.

**Phase:** Phase 1 (binary install/uninstall logic).

---

### Moderate: File Permission Bits Ignored on Windows
**What goes wrong:** `os.Chmod(path, 0755)` is a no-op on Windows. Executable scripts copied with `0755` silently lose the execute bit on Windows, then fail to run when invoked via `exec.Command`. On Linux/macOS, scripts written without `0755` are non-executable.

**Prevention:** After writing any script or binary, call `os.Chmod(path, 0755)` on Unix (guarded by `runtime.GOOS != "windows"`). On Windows, rely on file extension (`.exe`, `.bat`, `.cmd`) to signal executability. Document this contract in SKILL.md authoring guide.

**Phase:** Phase 2 (MCP binary install / skill scripts/).

---

### Moderate: Path Separator in Stored Config
**What goes wrong:** Config JSON written on Windows with `filepath.Join` produces `C:\Users\foo\.claude\skills\aws` — if that path is then read on macOS (e.g. shared dotfiles via chezmoi), or compared against a Unix-constructed path, matching fails. Skill version tracking and "is installed" checks break silently.

**Prevention:** Store all paths in config as forward-slash-normalized form using `filepath.ToSlash()` before writing, and `filepath.FromSlash()` when reading back for OS use. Never store absolute paths if avoidable; store relative paths from `~` instead.

**Phase:** Phase 1 (config schema design in `.agent-utils/config.json`).

---

## MCP Config Format Drift

### Critical: Each Assistant Uses a Different Schema and File Location
**What goes wrong:** Writing MCP config in Claude Code's format and blindly copying it to other assistants causes silent failures or assistant startup errors. Confirmed differences (HIGH confidence from official docs and GitHub issues):

| Assistant | Config Path | Top-Level Key | Format |
|-----------|------------|---------------|--------|
| Claude Code (user) | `~/.claude.json` | `mcpServers` | JSON |
| Claude Code (project) | `.mcp.json` | `mcpServers` | JSON |
| GitHub Copilot CLI | `~/.copilot/mcp-config.json` | `mcpServers` | JSON |
| VS Code Copilot | `%APPDATA%\Code\User\mcp.json` | `servers` (not `mcpServers`) | JSON |
| Gemini CLI | `~/.gemini/settings.json` | `mcpServers` | JSON (nested) |
| Codex CLI | `~/.codex/config.toml` | varies | TOML |

VS Code Copilot removed `.vscode/mcp.json` support (breaking change tracked in `github/copilot-cli#3019`). Copilot CLI and VS Code Copilot use different top-level keys despite being the same product.

**Prevention:** Model each assistant as a typed `Adapter` interface with `ConfigPath() string`, `Schema() SchemaVersion`, and `Write(servers []MCPServer) error`. Never share a raw config struct across adapters. Validate written config by re-reading and parsing.

**Warning signs:** A single `writeConfig()` function that takes a path and an untyped map.

**Phase:** Phase 2 (MCP install). Design adapters in Phase 1.

---

### Moderate: Schema Versions Change Without Warning
**What goes wrong:** Claude Code has silently migrated MCP server config between `~/.claude/settings.json` and `~/.claude.json` across versions (tracked in `anthropics/claude-code#4976`). An agentkit version that writes to the old location will appear to succeed but the MCP server will not be loaded.

**Prevention:** Read the existing config first, detect the schema version or file presence, and write to whichever file currently exists. Log the detected path at `--verbose`. Add an integration test that launches the target assistant's config loader (or parses its schema) to verify the written config is found.

**Warning signs:** Users reporting MCP servers not appearing after install despite no error.

**Phase:** Phase 2. Add a "verify install" step (`agentkit verify <name>`) in Phase 3.

---

### Moderate: Environment Variable Interpolation in MCP Config
**What goes wrong:** Some assistants expand `${ENV_VAR}` in MCP config values; others do not. Writing a config with raw API key values leaks secrets to disk. Writing `${ANTHROPIC_API_KEY}` on an assistant that does not expand variables breaks the MCP server at startup.

**Prevention:** AgentLink and similar tools confirm: never write raw secret values to config files. Use `${ENV_VAR}` syntax and document the required environment variables in SKILL.md. Test interpolation behavior for each adapter before writing secrets-bearing configs.

**Phase:** Phase 2.

---

## Registry Reliability

### Critical: Registry Down = Silent Install Failure Without Fallback
**What goes wrong:** `agentkit install aws` fetches the registry manifest from GitHub. If the registry is down, slow (>5s), or returns a 5xx, the CLI hangs or emits a generic network error with no fallback. Users on restricted corporate networks (no github.com access) are permanently blocked.

**Prevention:**
- Cache the last-successful manifest locally in `~/.agentkit/cache/<registry-id>/manifest.json` with a timestamp.
- On fetch failure, fall back to cache with a visible warning: `[warn] registry unreachable; using cached manifest from 2 days ago`.
- Set a short connect timeout (3s) and a longer read timeout (10s). Do not use Go's default (no timeout).
- Support `--offline` flag that forces cache-only resolution.

**Warning signs:** `http.Get()` or `http.DefaultClient` usage anywhere without an explicit timeout.

**Phase:** Phase 2 (registry system). Cache schema must be decided in Phase 1.

---

### Moderate: Stale Manifests After Package Deletion
**What goes wrong:** A registry manifest cached locally lists version 1.2.0 of a package. The maintainer deletes that release from GitHub. `agentkit install` resolves from the cached manifest, then 404s on the download. Error message blames network, not the stale cache.

**Prevention:** On any 404 during download, invalidate the manifest cache for that registry and re-fetch before surfacing a user-visible error. Include the manifest cache age in error messages to help users self-diagnose.

**Phase:** Phase 2.

---

### Minor: Multiple Registries With Divergent Data for Same Package Name
**What goes wrong:** mcpmarket.com and the GitHub manifest both list a package called `playwright`. They may have different versions, different install methods, or conflicting metadata. Without explicit priority, the resolution is ambiguous.

**Prevention:** Implement a deterministic priority order (e.g., explicit `--registry` flag > custom user registries > open-gsd/gsd-core > mcpmarket.com > GitHub manifest). Document this order. When a package appears in multiple registries, show a disambiguation prompt or use highest-version semantics. Never silently pick one.

**Phase:** Phase 2 (registry resolution).

---

## Skill Install Conflicts

### Moderate: Same Skill Name From Different Registries
**What goes wrong:** User runs `agentkit install aws`. Both the open-gsd registry and mcpmarket.com have an "aws" entry. The CLI picks one silently. Later `agentkit update aws` fetches from a different registry (the higher-version one), clobbering install-time metadata. The install record no longer matches the actual files.

**Prevention:** Store `source_registry` in the install record in `.agent-utils/config.json`. Pin updates to the original registry by default. For `agentkit update`, respect the stored registry; add `--registry <id>` override for intentional migration. Show registry provenance in `agentkit list`.

**Phase:** Phase 1 (install record schema), Phase 2 (enforcement).

---

### Moderate: Skill Files Collide in ~/.claude/skills/
**What goes wrong:** Two registries both install a skill that writes `~/.claude/skills/aws/SKILL.md`. Second install silently overwrites the first. No conflict detection, no backup, no user notification.

**Prevention:** Check for existing install records before writing files. If a file exists but wasn't installed by agentkit (or by a different registry entry), prompt the user. Offer `--force` to override. Never silently overwrite.

**Phase:** Phase 2.

---

## Background Agent Token Leaks

### Critical: Context Accumulation in Agentic Loops
**What goes wrong:** Background config agents that re-invoke themselves (or spawn sub-agents) accumulate the entire conversation history on every LLM call. At step N, you are paying for N copies of the prior context. A "cheap" 500-token config diff agent can balloon to 50,000+ tokens in 10 iterations. At Claude Sonnet 4.6 rates ($3/M input), a runaway 50-step loop can cost $5+ per invocation.

**Prevention:** The agentkit background agent is XML-prompted and scoped to config diff only — enforce this through the system prompt design and a hard iteration cap (max 3 tool calls). Never allow the agent to enter a retry loop. If the diff cannot be resolved in 3 steps, surface a user-visible error and exit.

**Warning signs:** Any background agent prompt that includes phrases like "if unsure, ask again" or recursive tool call patterns.

**Phase:** Phase 3 (background agent design). Establish token budget constraints in the agent spec before implementation.

---

### Moderate: Loading Full Project Context Into Background Agent
**What goes wrong:** The background agent, trying to be thorough, reads all files in the project to detect "new infrastructure facts." Reading even a modest project (50 files, 20KB each) dumps 1MB+ into the agent's context, costing $3+ in input tokens before any reasoning occurs.

**Prevention:** The agent must receive only a structured diff: "here are the new facts discovered this session, here is the current config.json." It must never traverse the filesystem independently. The calling skill must extract and summarize the delta before handing off to the agent.

**Phase:** Phase 3. The SKILL.md authoring contract must specify exactly what the skill passes to the background agent.

---

### Minor: Token Budget Check Timing
**What goes wrong:** The background agent triggers and runs to completion before checking whether token cost exceeded the threshold — then asks for user confirmation after the fact, which is useless.

**Prevention:** Estimate token cost before invocation using a simple heuristic (config diff size in bytes / 4 as proxy for tokens). If estimate exceeds threshold, prompt user first, then invoke. If below threshold, invoke inline and report cost in a one-liner.

**Phase:** Phase 3.

---

## Config Drift

### Moderate: .agent-utils/config.json Gets Out of Sync With Reality
**What goes wrong:** A project's `.agent-utils/config.json` records that the AWS region is `us-east-1`. Six months later the infrastructure moves to `eu-west-1`. Every skill that reads the config operates on stale data without warning. The skill appears to work (no errors), but takes actions against wrong resources.

**Prevention:** Store a `recorded_at` timestamp on each config entry. Skills should check entry age at load time: entries older than a configurable TTL (default: 30 days) should trigger a silent re-discovery run or at minimum a visible `[stale]` warning. Provide `agentkit config refresh` as an explicit command.

**Warning signs:** Config entries with no timestamp field in the schema.

**Phase:** Phase 1 (config JSON schema must include `recorded_at` from day one — retrofitting timestamps is painful).

---

### Minor: Config Written Mid-Session Causing Race Conditions
**What goes wrong:** Two concurrent Claude Code sessions run skills that both discover config facts and write to `.agent-utils/config.json` simultaneously. One write clobbers the other. The final config contains partial data from whichever write won the race.

**Prevention:** Use atomic file writes (write to temp file, `os.Rename` to target). On Linux/macOS, `os.Rename` is atomic. On Windows, use `github.com/google/renameio` or equivalent (standard `os.Rename` is not atomic on Windows per `golang/go#8914`). Use a file lock (`flock` on Unix) if concurrent writes are expected in the same session.

**Phase:** Phase 1 (config write utility).

---

## Go Module Management

### Moderate: Transitive Dependency Version Conflicts in CLI Binary
**What goes wrong:** agentkit depends on a HTTP client library at v2.1, and also depends on a registry SDK that internally requires the same library at v1.8. Go's MVS (Minimum Version Selection) picks v2.1, but the registry SDK was tested against v1.8. Subtle behavioral differences (redirect handling, timeout semantics) cause intermittent failures in the registry SDK path that are hard to reproduce.

**Prevention:** Keep dependencies minimal and shallow. For a CLI tool, prefer stdlib over third-party HTTP clients. For required third-party libraries, vendor them (`go mod vendor`) in the final release build. Run `go mod why <package>` regularly to audit transitive pull-ins. Flag any dependency tree deeper than 3 levels as a risk.

**Warning signs:** `go mod graph | wc -l` growing past ~50 lines in a simple CLI.

**Phase:** Phase 1 (module setup). Revisit before each phase when new dependencies are added.

---

### Minor: GOPROXY Unavailability in Air-Gapped Environments
**What goes wrong:** `proxy.golang.org` is unavailable in corporate air-gapped networks. `go install github.com/ejyle/agentkit@latest` fails. Users in these environments are blocked from installing agentkit itself.

**Prevention:** Provide pre-compiled binaries via GitHub Releases (the primary distribution path for a Go CLI). Document `GOPROXY=off` / `GOFLAGS=-mod=vendor` for source builds. Do not rely on `go install` as the primary distribution method.

**Phase:** Phase 4 (distribution/release). But pre-compiled binary release must be designed from Phase 1 (Makefile / goreleaser config).

---

## AI Assistant Config Breaking Changes

### Critical: Undocumented Config Location Changes
**What goes wrong:** Claude Code silently moved MCP server configuration from `settings.json` to `~/.claude.json` between versions (confirmed by `anthropics/claude-code#4976`). GitHub Copilot removed `.vscode/mcp.json` support without major announcement (tracked in `github/copilot-cli#3019`). agentkit's adapter writes to the old path, the install appears to succeed, but the MCP server is never loaded by the assistant.

**Prevention:**
- Each adapter must declare the config path it targets with a version comment: `// Verified against Claude Code v1.x — re-verify on major version bumps`
- Add a `--verify` post-install step that reads back the config using the assistant's own config discovery logic (or parsing the target file and checking the server appears)
- Subscribe to release notes/changelogs for each supported assistant; pin adapter versions to assistant versions if format changes are detected
- Fail loudly with actionable messages: `Wrote config to ~/.claude.json but assistant version X.Y uses a different path. Run: agentkit repair <target>`

**Warning signs:** User reports "MCP server installed but not showing up in assistant" — this is almost always a path mismatch.

**Phase:** Phase 2. A `verify` subcommand should be scoped into Phase 3.

---

### Moderate: Copilot CLI vs VS Code Copilot Config Divergence
**What goes wrong:** Copilot CLI uses `mcpServers` as the top-level key; VS Code Copilot uses `servers`. A tool that writes Copilot CLI format breaks VS Code Copilot integration and vice versa. This is an active known divergence the Copilot team has not yet resolved (see `community#187954`).

**Prevention:** Treat Copilot CLI and VS Code Copilot as two separate adapters. When `--target copilot` is specified, prompt: "Install for Copilot CLI, VS Code Copilot, or both?" and write both config files with the correct schema for each.

**Phase:** Phase 2.

---

### Moderate: Codex Uses TOML, All Others Use JSON
**What goes wrong:** A shared "write MCP config" utility that produces JSON cannot write Codex config (`~/.codex/config.toml`). Developers assume all assistants use JSON and ship a Codex adapter that silently produces an unreadable config.

**Prevention:** The adapter interface must not assume output format. Provide a TOML marshaller (e.g. `github.com/BurntSushi/toml`) for the Codex adapter. Test each adapter against a real (or mocked) assistant config loader.

**Phase:** Phase 2.

---

## Permission / Multi-User Issues

### Moderate: Writing to ~/.claude/ Without Ownership Check
**What goes wrong:** On a shared Linux machine, user A runs `sudo agentkit install` (to get a binary into `/usr/local/bin`). The sudo context sets `$HOME` to root's home (`/root`). The skill installs into `/root/.claude/` instead of `/home/usera/.claude/`. User A cannot see the installed skill in their non-root session.

**Prevention:** Never run agentkit as root for user-scope installs. Detect `os.Geteuid() == 0` and abort with: `Do not run agentkit with sudo for user-scope installs. If you need a system-wide binary, install agentkit itself with sudo, then run skill installs as your normal user.`

Separately, for binary installs (MCP servers), install to `~/.local/bin` (user-owned, no sudo needed) and ensure this is on the user's PATH. Detect if `~/.local/bin` is not in PATH and warn.

**Warning signs:** Any `sudo agentkit` invocation in documentation or install scripts.

**Phase:** Phase 1 (binary install path design), Phase 2 (sudo detection).

---

### Minor: ~/.claude/ Permissions After Manual edits
**What goes wrong:** If a user manually creates `~/.claude/skills/` with `sudo mkdir`, the directory is root-owned. Subsequent `agentkit install` (run as normal user) fails with `permission denied` writing skill files. The error message shows the path but not the ownership cause.

**Prevention:** Before writing to any target directory, check `os.Stat()` and compare ownership against `os.Getuid()`. If directory exists but is not owned by the current user, emit: `Directory ~/.claude/skills/ exists but is owned by root. Fix with: sudo chown -R $USER ~/.claude/`

**Phase:** Phase 2 (pre-install checks).

---

## Phase Mapping Summary

| Phase | Key Pitfalls to Address |
|-------|------------------------|
| Phase 1 (CLI Core & Config Schema) | `os.UserHomeDir()` abstraction, `filepath.Join` throughout, config schema with `recorded_at` and `source_registry`, atomic file writes, module setup |
| Phase 2 (Registry & MCP Install) | Multi-adapter MCP config writers, registry timeout+cache, name conflict resolution, sudo detection, directory ownership checks |
| Phase 3 (Background Agent) | Token budget cap (max 3 tool calls), structured delta input only, pre-invocation cost estimate |
| Phase 4 (Distribution) | Pre-compiled binary releases, GOPROXY docs, PATH detection for ~/.local/bin |
| Ongoing | Subscribe to assistant changelogs; re-verify adapter config paths on each Claude Code / Copilot major version |
