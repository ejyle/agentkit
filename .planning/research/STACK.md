# Stack Research: agentkit

**Researched:** 2026-06-08
**Overall confidence:** HIGH (verified via official docs and active GitHub repos)

---

## CLI Framework

**Recommendation: Cobra v1.10.x + Bubbletea (Charm) for interactive prompts**

Cobra is the de facto standard for Go CLIs in 2025-2026. 44K GitHub stars, used by Kubernetes, Docker, GitHub CLI, Hugo. Latest stable: v1.10.2 (Dec 2025). The API is stable and battle-tested for subcommand-heavy tools exactly like agentkit (`install`, `search`, `update`, `list`).

**Why Cobra over alternatives:**

| Framework | Stars | Verdict |
|-----------|-------|---------|
| spf13/cobra | 44K | **Use this.** Proven, composable, widely understood |
| urfave/cli | 22K | Older API; positional-arg style doesn't fit subcommand-heavy tools |
| alecthomas/kong | 7K | Struct-tag based, elegant for smaller tools, but less ecosystem support for shell completion and plugin extension |
| charmbracelet/bubbletea | 30K | TUI framework, not a CLI router. Use alongside Cobra for interactive confirm prompts |

**Pattern:** Cobra handles all command routing and flag parsing. Bubbletea handles interactive UX moments — e.g., `agentkit search` displaying an interactive results list, or confirming overwrite of an existing config. They compose naturally: a Cobra command handler launches a Bubbletea program, returns to normal output on completion.

**Do NOT use:** `kingpin` (deprecated in favor of kong), `flag` stdlib (no subcommands).

**Packages:**
```
github.com/spf13/cobra v1.10.x
github.com/charmbracelet/bubbletea v1.x
github.com/charmbracelet/lipgloss v1.x  (styling for TUI output)
```

Confidence: HIGH — verified via cobra.dev, GitHub stars, and active 2025-2026 releases.

---

## Cross-Platform Path Handling

**Recommendation: stdlib `os.UserConfigDir` + `os.UserHomeDir` + hardcoded assistant path map**

Go's `os.UserConfigDir()` returns:
- Linux: `$XDG_CONFIG_HOME` or `~/.config`
- macOS: `~/Library/Application Support`
- Windows: `%APPDATA%`

For agentkit's own config (`~/.config/agentkit/` on Linux/macOS), use `os.UserConfigDir()`.

For target-assistant paths, **hardcode the per-assistant paths as a map** — they are fixed by the vendor:

| Assistant | Skills Path | MCP Config Path |
|-----------|-------------|-----------------|
| Claude Code | `~/.claude/skills/` | `~/.claude/settings.json` (mcpServers key) |
| GitHub Copilot | `~/.copilot/skills/` | `~/.copilot/mcp-config.json` |
| OpenAI Codex | `~/.codex/skills/` | `~/.codex/config.toml` (mcp section) |
| Gemini CLI | `~/.gemini/skills/` | `~/.gemini/settings.json` (mcpServers key) |
| OpenCode | `~/.config/opencode/skills/` | `~/.config/opencode/opencode.json` (mcp key) |

Use `os.UserHomeDir()` as the root for all `~/` paths — never hardcode `/home/user` or `/Users/user`.

**Do NOT use:** `adrg/xdg` third-party library — stdlib is sufficient and adds no dependency.

**Windows note:** All `~` paths resolve correctly via `os.UserHomeDir()`. The Codex CLI uses `%USERPROFILE%\.codex\` which `os.UserHomeDir()` returns. Test with `filepath.Join(os.UserHomeDir(), ".codex")`.

Confidence: HIGH for Claude/Codex/Gemini/OpenCode (official docs verified). MEDIUM for Copilot skills path (`~/.copilot/skills/` — confirmed by GitHub docs but path may vary by VS Code extension vs CLI context).

---

## Registry Client

**Recommendation: `hashicorp/go-retryablehttp` for registry HTTP calls + stdlib `net/http` for simple fetches**

```
github.com/hashicorp/go-retryablehttp v0.7.x
```

Provides:
- Automatic retries with exponential backoff
- HTTP 429 / 503 handling with `Retry-After` header support (rate limiting)
- Wraps `net/http` — same interface, zero learning curve

For caching manifest files locally: use stdlib `os.MkdirTemp` pattern with a disk cache in `~/.cache/agentkit/` (or `os.UserCacheDir()`). Cache manifests for TTL (e.g., 1 hour), invalidate on `update`.

**Registry protocol design** — follow the mise pattern, not npm:
- Each registry is a GitHub repo with a `registry.json` (or `manifest.json`) at the root
- The manifest is a flat JSON index: `{ "packages": [ { "name": "gsd", "version": "1.2.0", "description": "...", "type": "skill|mcp|agent", "source": "github.com/open-gsd/gsd-core", "path": "skills/gsd/" } ] }`
- agentkit fetches `https://raw.githubusercontent.com/{owner}/{repo}/{ref}/registry.json`
- No registry server needed — raw GitHub serves as the CDN

Do NOT use a live HTTP API as the primary registry protocol — raw GitHub manifests are zero-infrastructure and work behind corporate firewalls. mcpmarket.com API is a secondary source, fetched only when explicitly configured.

Confidence: HIGH for go-retryablehttp (HashiCorp production-grade). MEDIUM for manifest protocol design (pattern derived from mise/aqua observation; no single authoritative spec).

---

## MCP Config Formats (per assistant)

All assistants use JSON-based MCP config with a `mcpServers` key (same schema), except Codex which uses TOML. agentkit must write to all formats.

### Claude Code — `~/.claude/settings.json`
```json
{
  "mcpServers": {
    "server-name": {
      "command": "npx",
      "args": ["-y", "@package/server"]
    }
  }
}
```

### GitHub Copilot CLI — `~/.copilot/mcp-config.json`
```json
{
  "mcpServers": {
    "server-name": {
      "command": "npx",
      "args": ["-y", "@package/server"]
    }
  }
}
```
Location can be overridden with `COPILOT_HOME` env var.

### Gemini CLI — `~/.gemini/settings.json`
```json
{
  "mcpServers": {
    "server-name": {
      "command": "npx",
      "args": ["-y", "@package/server"],
      "env": { "KEY": "value" }
    }
  }
}
```

### OpenAI Codex — `~/.codex/config.toml`
```toml
[mcp.servers.server-name]
command = "npx"
args = ["-y", "@package/server"]
```
Uses TOML. Use `github.com/BurntSushi/toml` for reading/writing.

### OpenCode — `~/.config/opencode/opencode.json`
```json
{
  "mcp": {
    "server-name": {
      "type": "local",
      "command": ["npx", "-y", "@package/server"],
      "env": { "KEY": "value" }
    }
  }
}
```
OpenCode uses a different key (`mcp` not `mcpServers`) and `"type": "local"|"remote"` field.

**Implementation note:** Model each config format as a Go struct + a `Read(path) → Config` / `Write(path, Config)` pair. Converge on a shared `MCPServerEntry` internal type and transform to/from each format.

Confidence: HIGH for Claude/Gemini/Codex (official docs verified). MEDIUM for Copilot and OpenCode (community docs + GitHub issues verified; may shift with product updates).

---

## Distribution / Release Tooling

**Recommendation: GoReleaser v2 + GitHub Actions**

GoReleaser handles:
- Cross-compile for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- GitHub Releases with checksums and SBOMs
- Homebrew tap auto-update (publish `agentkit` to a `homebrew-tap` repo)
- Scoop manifest for Windows (generated automatically)
- `.goreleaser.yaml` in repo root

No CGO in agentkit (pure Go + JSON/TOML parsing) means cross-compilation is trivial — no Docker required.

**Install paths for users:**
- macOS/Linux: `brew install yourorg/tap/agentkit` or `curl | sh` script
- Windows: `scoop install agentkit` or direct binary download
- Any: `go install github.com/yourorg/agentkit@latest`

```
github.com/goreleaser/goreleaser  (dev tool, not a Go package dep)
```

GoReleaser v2 is current as of 2025. The free tier covers all needs for an open-source CLI.

Confidence: HIGH — GoReleaser is the unambiguous standard for Go CLI distribution.

---

## Config File Format

**Recommendation: JSON for all agentkit-managed files**

Rationale:
- `.agent-utils/config.json` is already specified in PROJECT.md — match the ecosystem
- Claude Code (`settings.json`), Gemini CLI (`settings.json`), Copilot (`mcp-config.json`), OpenCode (`opencode.json`) all use JSON
- Go's `encoding/json` is stdlib — zero dependencies
- Grep/jq-friendly for power users
- Human-readable enough for the developer audience

**For agentkit's own user config** (`~/.config/agentkit/config.json`):
- JSON — consistent with the rest of the ecosystem
- Schema: `{ "registries": [...], "defaults": { "target": "claude" }, "auth": { ... } }`

**Do NOT use TOML** for agentkit's own config — Codex is the only assistant using TOML, and mixing formats confuses users. Keep JSON as the single format across all agentkit-owned files.

**Do NOT use YAML** — indentation-sensitive, difficult to generate programmatically, adds `gopkg.in/yaml.v3` dependency for marginal DX gain.

Confidence: HIGH — JSON is the dominant format in the target ecosystem.

---

## Version Management

**Recommendation: SemVer pinning in registry manifest + SHA256 integrity check + local lock file**

agentkit is both a Go module (its own `go.mod`/`go.sum` for deps) and a package manager (manages skills/MCP). Two separate concerns:

**For agentkit's own Go dependencies:** Use `go.mod` + `go.sum` (stdlib mechanism). Commit both files. Use `go mod tidy` in CI.

**For packages agentkit installs (skills/MCP):**
- Each registry manifest entry includes `version` (SemVer) and `sha256` of the installable artifact
- On install, agentkit writes an entry to `~/.config/agentkit/installed.json`:
  ```json
  {
    "packages": {
      "gsd": {
        "version": "1.2.0",
        "sha256": "abc123...",
        "source": "github.com/open-gsd/gsd-core",
        "installed_at": "2026-06-08T12:00:00Z",
        "targets": ["claude", "copilot"]
      }
    }
  }
  ```
- This is the "lock file" for user installs — analogous to `package-lock.json`
- `agentkit update` checks current version against registry and upgrades if newer
- No project-scope lock file in v1 (user scope only per PROJECT.md)

**Do NOT implement** a full dependency resolution graph in v1 — skills have no interdependencies. Pin by exact version, no ranges.

Confidence: MEDIUM — the exact lock file schema is original design, but the pattern is validated by mise/npm/brew precedents.

---

## Supporting Libraries

| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| `spf13/cobra` | v1.10.x | CLI command routing | De facto standard, 44K stars |
| `charmbracelet/bubbletea` | v1.x | Interactive TUI prompts | Best-in-class Go TUI, pairs with Cobra |
| `charmbracelet/lipgloss` | v1.x | Terminal styling | Same ecosystem as Bubbletea |
| `hashicorp/go-retryablehttp` | v0.7.x | Resilient HTTP for registry | Auto-retry, rate-limit handling, HashiCorp battle-tested |
| `BurntSushi/toml` | v1.x | Read/write Codex config.toml | Only TOML dep; high quality, stable API |
| stdlib `encoding/json` | — | All JSON config R/W | No dep needed |
| stdlib `os`, `path/filepath` | — | Cross-platform path handling | Sufficient, avoids adrg/xdg overhead |

**Deliberately excluded:**
- `spf13/viper` — overkill for a tool with one config file; adds 10+ transitive deps
- `adrg/xdg` — stdlib is sufficient for the paths needed
- `gopkg.in/yaml.v3` — no YAML needed in this stack
- `urfave/cli` — inferior subcommand ergonomics vs Cobra for this use case

---

## Recommended Stack Summary

```
Language:     Go 1.22+
CLI Router:   github.com/spf13/cobra v1.10.x
TUI:          github.com/charmbracelet/bubbletea v1.x
HTTP:         github.com/hashicorp/go-retryablehttp v0.7.x
TOML:         github.com/BurntSushi/toml v1.x
JSON:         stdlib encoding/json
Paths:        stdlib os.UserHomeDir / os.UserConfigDir / os.UserCacheDir
Release:      GoReleaser v2 + GitHub Actions
Config fmt:   JSON (.json) for all agentkit-owned files
Lock format:  ~/.config/agentkit/installed.json (SemVer + SHA256)
```

Single binary, no CGO, no runtime deps. Cross-compiles cleanly to all 5 target platforms.

---

## Sources

- [Cobra GitHub](https://github.com/spf13/cobra) — v1.10.2, Dec 2025
- [Bubbletea GitHub](https://github.com/charmbracelet/bubbletea)
- [hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp)
- [GoReleaser](https://goreleaser.com/)
- [adrg/xdg Go Packages](https://pkg.go.dev/github.com/adrg/xdg) — evaluated and rejected
- [Codex CLI MCP config](https://developers.openai.com/codex/mcp)
- [Gemini CLI MCP setup](https://geminicli.com/docs/tools/mcp-server/)
- [OpenCode MCP servers](https://opencode.ai/docs/mcp-servers/)
- [GitHub Copilot skills](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills)
- [Mise registry system](https://deepwiki.com/jdx/mise/6.1-registry-system)
