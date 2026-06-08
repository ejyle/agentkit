# Walking Skeleton: agentkit

_Phase 1 — Foundation | Generated: 2026-06-08_

---

## What Is This?

The Walking Skeleton is the thinnest possible end-to-end slice that proves the architecture
compiles, the data flows work, and the critical correctness properties (atomic writes, runtime
path detection, post-install verify) are locked in before any feature expansion.

For agentkit, the skeleton is:

```
agentkit install playwright --target claude
```

Resolves from the official agentkit-registry → runs npx adapter → writes a real MCP entry to
`~/.claude.json` → records install to `~/.config/agentkit/claude/installed.json` → prints
success line. End-to-end. No stubs in the write path.

---

## Architectural Decisions

### Framework

| Concern | Decision | Rationale |
|---------|----------|-----------|
| CLI routing | `spf13/cobra` v1.10.2 | De facto standard; used by Kubernetes/Docker/GitHub CLI; 44K stars |
| Spinner + interactive prompts | `charmbracelet/bubbletea` v1.3.x | Only library that handles both spinner phases and D-07 interactive conflict prompt |
| Terminal styling | `charmbracelet/lipgloss` v1.x | Same Charm ecosystem; consistent table/color API |
| HTTP with retry | `hashicorp/go-retryablehttp` v0.7.8 | Auto-retry, Retry-After, CVE-2024-6104 fixed at this version |
| Atomic file write | `google/renameio/v2` | `os.Rename` is non-atomic on Windows; renameio is correct cross-platform |
| JSON config R/W | stdlib `encoding/json` | No external dep needed; all config files are JSON |
| Path handling | stdlib `os`, `path/filepath` | `os.UserHomeDir()`, `os.UserConfigDir()`, `os.UserCacheDir()` cover all platforms |

**Rejected:** `spf13/viper` (10+ transitive deps, overkill), `urfave/cli` (positional-arg style),
`adrg/xdg` (stdlib sufficient), `briandowns/spinner` (cannot do interactive prompts).

### Module

```
module github.com/ejyle/agentkit
go 1.21
```

Single binary, no runtime dependency. Cross-compiles to linux/amd64, linux/arm64,
darwin/amd64, darwin/arm64, windows/amd64.

### Directory Layout

```
agentkit/
├── cmd/
│   ├── root.go          # cobra root command; --target persistent flag (default: "claude")
│   ├── install.go       # agentkit install <name>
│   ├── list.go          # agentkit list [--target <assistant>]
│   ├── search.go        # agentkit search <query>
│   ├── uninstall.go     # agentkit uninstall <name>
│   └── update.go        # agentkit update [name]
├── internal/
│   ├── domain/
│   │   ├── package.go   # Package, Manifest, MCPServerEntry types
│   │   └── installed.go # InstalledRecord, InstalledState (D-11 schema)
│   ├── config/
│   │   ├── store.go     # ConfigStore: CRUD for installed.json, atomic writes
│   │   └── paths.go     # all path resolution via os.UserHomeDir/UserConfigDir/UserCacheDir
│   ├── registry/
│   │   ├── registry.go  # Registry interface + RegistryManager
│   │   ├── github.go    # GitHubManifestRegistry with ETag cache
│   │   └── cache.go     # CachedManifest struct + disk cache read/write
│   ├── adapter/
│   │   ├── adapter.go   # AssistantAdapter interface
│   │   └── claude.go    # ClaudeCodeAdapter: runtime path detection, skill install, MCP merge
│   ├── installer/
│   │   ├── installer.go # MCPInstaller interface
│   │   ├── npx.go       # NpxInstaller: exec "npx -y <pkg>"
│   │   └── binary.go    # BinaryInstaller: download → ~/.config/agentkit/bin/ → chmod +x
│   ├── service/
│   │   ├── install.go   # InstallService: resolve → install → write → record
│   │   └── search.go    # SearchService: fan-out, rank, deduplicate
│   ├── skill/
│   │   └── validate.go  # ValidateSkill(): SKILL.md check, line count, references/
│   └── ui/
│       ├── spinner.go   # bubbletea spinner model (D-03 phases)
│       └── table.go     # lipgloss table for list/search output (D-05, D-06)
├── main.go
└── go.mod
```

### Data Layer

**Registry manifest** (`registry.json` in the agentkit-registry GitHub repo):

```json
{
  "packages": [
    {
      "name": "playwright",
      "version": "1.2.0",
      "description": "Browser automation and E2E testing MCP server",
      "type": "mcp",
      "source": "github.com/microsoft/playwright-mcp",
      "install": {
        "method": "npx",
        "package": "@playwright/mcp"
      },
      "mcp_entry": {
        "command": "npx",
        "args": ["-y", "@playwright/mcp"]
      },
      "sha256": "abc123..."
    }
  ]
}
```

**Installed state** (`~/.config/agentkit/<target>/installed.json`):

```json
{
  "packages": {
    "playwright": {
      "name": "playwright",
      "version": "1.2.0",
      "type": "mcp",
      "install_path": "mcpServers.playwright",
      "installed_at": "2026-06-08T10:00:00Z",
      "source_url": "https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json",
      "checksum": "sha256:abc123"
    }
  }
}
```

**Claude Code MCP config** (`~/.claude.json`):

```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["-y", "@playwright/mcp"]
    }
  }
}
```

### Path Resolution

All paths resolved via stdlib — never `os.Getenv("HOME")`:

| Path | Resolution |
|------|-----------|
| `~/.config/agentkit/<target>/installed.json` | `os.UserConfigDir()` + `agentkit/<target>/installed.json` |
| `~/.cache/agentkit/<registry-id>/manifest.json` | `os.UserCacheDir()` + `agentkit/<registry-id>/manifest.json` |
| `~/.claude.json` | `os.UserHomeDir()` + `.claude.json` |
| `~/.claude/skills/<name>/` | `os.UserHomeDir()` + `.claude/skills/<name>/` |
| `~/.config/agentkit/bin/` | `os.UserConfigDir()` + `agentkit/bin/` |

### Security Properties (Phase 1 baseline)

- Shell injection prevention: `exec.Command(binary, arg1, arg2)` arg-array form always; never shell string interpolation
- HTTPS-only for all registry fetches (enforced by manifest URL scheme validation)
- SHA256 checksum verification for binary downloads (MCP-03)
- Atomic writes via `google/renameio/v2` for all config file mutations
- Post-install verify: re-read `~/.claude.json` after every write; fail loudly if parse fails or key absent
- Foreign conflict detection: compare `mcpServers` key against `installed.json` ownership before any overwrite (D-07)
- go-retryablehttp v0.7.8+ pinned (CVE-2024-6104 URL log sanitization)

### Critical Pitfall: Claude Code MCP Config Path

**The path `~/.claude/settings.json` is NOT where Claude Code reads MCP servers.**

Current path mapping (confirmed via GitHub issue #4976):

| Scope | File |
|-------|------|
| User-scoped MCP (current Claude Code) | `~/.claude.json` |
| Project-scoped MCP | `./.mcp.json` |
| General settings (not MCP) | `~/.claude/settings.json` |

ClaudeCodeAdapter uses runtime detection — stat `~/.claude.json` first, then
`~/.claude/settings.json` as fallback for legacy installs, then create `~/.claude.json`
on first use. Never hardcode.

---

## Walking Skeleton Acceptance Test

When Phase 1 Plan 01-03 completes, this command MUST work:

```bash
# Build
go build -o agentkit ./...

# Install playwright targeting Claude Code
./agentkit install playwright --target claude

# Expected output:
# [spinner] Fetching registry...
# [spinner] Resolving playwright...
# [spinner] Running install adapter...
# ✓ playwright@1.2.0 installed → ~/.claude/skills/playwright/ (claude)

# Verify state was written
cat ~/.claude.json | jq '.mcpServers.playwright'
# Returns: { "command": "npx", "args": ["-y", "@playwright/mcp"] }

cat ~/.config/agentkit/claude/installed.json | jq '.packages.playwright.version'
# Returns: "1.2.0"
```

---

_This file records architectural decisions for Phase 1. Subsequent phases build on this
skeleton without renegotiating these foundations._
