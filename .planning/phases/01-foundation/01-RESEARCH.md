# Phase 1: Foundation - Research

**Researched:** 2026-06-08
**Domain:** Go CLI — greenfield project scaffold, registry client, Claude Code MCP adapter, install/list/search/uninstall commands
**Confidence:** HIGH (stack verified via official docs and GitHub releases; Claude Code MCP path verified via GitHub issues and multiple independent sources)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** agentkit is a **curated registry**, not a general package manager. ONE official curated agentkit-registry. Curation and benchmarking is the core product value.
- **D-02:** v1 installs only from the official curated registry. REG-01 narrowed: only the single official agentkit-registry repo.
- **D-03:** `agentkit install`: bubbletea spinner for each phase (fetching registry → resolving package → running install adapter), then clean single-line success: `✓ playwright@1.2.0 installed → ~/.claude/skills/playwright/ (claude)`
- **D-04:** Error output: single clear error line + context + suggested next command. Exit code 1. No stack traces. Example: `✗ Error: playwright not found in agentkit-registry` / `Run: agentkit search playwright`
- **D-05:** `agentkit list`: table format — `PACKAGE`, `VERSION`, `TYPE`, `TARGET`, `REGISTRY` columns. `go list -m all` style.
- **D-06:** `agentkit search <query>`: spinner while fetching registry, then deterministic ranked result list (name, type, source label, one-line description).
- **D-07:** Foreign MCP conflict: stop and prompt — show old vs new entry, ask `Overwrite? [y/N]`.
- **D-08:** Upgrade (agentkit-owned): auto-overwrite with one-line notice `⚠ playwright: upgrading 0.9 → 1.2.0`.
- **D-09:** Uninstall merge: read `settings.json` → delete only agentkit's key → atomic write. Never clobber user-written config.
- **D-10:** Per-assistant state files: `~/.config/agentkit/claude/installed.json`, one file per target assistant.
- **D-11:** Full install entry schema (name, version, type, install_path, installed_at, source_url, checksum).
- **D-12:** Auto-create `~/.config/agentkit/claude/` and `installed.json` on first install — no `agentkit init` required.
- Architecture constraints from STATE.md: `os.UserHomeDir()` everywhere, `filepath.Join()` throughout, atomic file writes via temp+rename, runtime path detection, post-install verify, 3s connect / 10s read timeouts, ETag-based manifest cache, `--offline` fallback.

### Claude's Discretion

- Ranking algorithm for `agentkit search` results (exact match, then fuzzy, then description keywords)
- Exact bubbletea component structure for the spinner (style, frame rate, color)
- JSON schema versioning strategy for `installed.json` (whether to include a `schema_version` field)

### Deferred Ideas (OUT OF SCOPE)

- CLI-09 (`agentkit registry add/remove`) — removed from v1
- REG-03 (mcpmarket.com API registry) — removed from v1
- REG-04 (custom registry sources) — removed from v1
- `--yes` / `--non-interactive` flag — noted, not required for Phase 1
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CLI-01 | Install skill/agent/MCP by name | Install flow: registry resolve → MCP installer → adapter write → state record |
| CLI-02 | Specify target assistant with `--target` | AssistantAdapter interface; ClaudeCode adapter is Phase 1's only full impl |
| CLI-05 | Uninstall cleanly, no leftover artifacts | Uninstall: read installed.json → remove skill dir or MCP entry → atomic write → delete record |
| CLI-06 | Search all registries with ranked results | SearchService: fan-out (single registry in Phase 1), rank by exact/fuzzy/description, return with source label |
| CLI-07 | Update one or all installed packages | UpdateService: compare installed version vs manifest, run install adapter for newer version |
| CLI-08 | List installed packages (version, source, target) | Read all `~/.config/agentkit/<target>/installed.json` files, format as table |
| REG-01 | GitHub manifest-driven registry (narrowed: single curated registry) | GitHubManifestRegistry with ETag caching; fetches `registry.json` from agentkit-registry repo |
| REG-02 | open-gsd/gsd-core as built-in default source | Treated as a GitHubManifestRegistry with known URL; registered at startup |
| REG-05 | agentkit-registry as default curated registry | Separate GitHub repo with `registry.json`; needs to exist before Phase 1 tests pass |
| REG-06 | Registry manifests cached locally with ETag validation | ETag stored alongside manifest in `~/.cache/agentkit/<registry-id>/`; stale cache with warning on network fail |
| AST-01 | Claude Code adapter fully implemented | Skills → `~/.claude/skills/`, MCP config → `~/.claude.json` (user-scope) or `.mcp.json` (project-scope); runtime path detection |
| MCP-01 | npx install adapter | `npx -y <package>` via `exec.Command`; prereq check: `node --version` |
| MCP-03 | Binary download adapter | Download from manifest URL → `~/.config/agentkit/bin/` → chmod +x |
| MCP-05 | Provider manifest can override install with custom steps | `install.override` field in manifest; CustomInstaller wraps as shell script |
| MCP-06 | Post-install verify: re-read written MCP config | Each adapter's WriteMCPConfig calls ReadMCPConfig after write; fail loudly on parse error |
| MCP-07 | Runtime config path detection; merge-write, atomic | Read → merge → write to temp → rename; never hardcode paths |
| SKL-01 | Validate agentskills.io spec: SKILL.md + references/ + scripts/ | ValidateSkill() function: check required files, parse YAML frontmatter |
| SKL-02 | SKILL.md line count warning if >500 lines | Count lines on install; warning emitted, install proceeds (non-blocking) |
| SKL-03 | Multi-domain skills organized under references/ | Enforced by registry manifest schema `skill.references[]` field |
</phase_requirements>

---

## Summary

Phase 1 is a greenfield Go project that delivers the walking skeleton: `agentkit install <name> --target claude` end-to-end. This means Go module scaffold, Cobra CLI wiring, a single registry client (agentkit-registry on GitHub), a Claude Code adapter, npx and binary install adapters, an `installed.json` state tracker, and all five CLI commands (install, list, search, uninstall, update).

The most consequential discovery in this research is a **correction to the existing STACK.md and ARCHITECTURE.md**: Claude Code's MCP server configuration does NOT live in `~/.claude/settings.json`. As of current Claude Code versions, user-scoped MCP servers are configured in `~/.claude.json` (in the home directory, not inside `.claude/`), and project-scoped servers use `.mcp.json` in the project root. The `~/.claude/settings.json` file is for general Claude Code settings but is NOT read for MCP server definitions. This is confirmed by GitHub issue #4976, multiple 2025/2026 blog posts, and the Claude Code MCP documentation. The ClaudeCode adapter must detect which file is present at runtime and write to the correct location.

Walking skeleton scope: the thinnest vertical slice that proves the architecture is `agentkit install playwright --target claude` working end-to-end — registry fetch, npx adapter, `~/.claude.json` write, `installed.json` write, spinner output, success line. Every other command (list, search, uninstall, update) reuses the same domain types and interfaces.

**Primary recommendation:** Build strictly bottom-up — domain types → config store → registry client → Claude Code adapter → install service → CLI commands. Do not start CLI commands before interfaces are stable. The registry manifest schema and `installed.json` schema must be finalized in Wave 1 because every other piece depends on them.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| CLI routing and flag parsing | CLI Layer (Cobra) | — | Cobra owns all command dispatch; service layer has no Cobra knowledge |
| Registry resolution and caching | Registry Layer | Config Store (cache path) | Registry manager owns manifest fetch/cache; config store provides path |
| MCP server config write | Assistant Adapter (Claude Code) | Config Store (state record) | Adapter owns per-assistant format; config store tracks what was installed |
| MCP server install (npx/binary) | MCP Installer | Assistant Adapter (post-install) | Installer runs the install command; adapter writes the config entry after |
| State tracking (installed.json) | Config Store | — | Config store is the single source of truth for installed packages |
| Output formatting (spinner, table) | CLI Layer | Bubbletea/Lipgloss | Presentation is CLI concern only; service layer returns structured data |
| Skill structure validation | Skill Validator | Registry Layer | Validator runs at install time; registry provides the manifest |

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `spf13/cobra` | v1.10.2 | CLI command routing and flag parsing | De facto Go CLI standard; 44K stars; used by Kubernetes, Docker, GitHub CLI [VERIFIED: github.com/spf13/cobra] |
| `charmbracelet/bubbletea` | v1.3.x | Spinner and interactive confirm prompts | Best-in-class Go TUI; pairs cleanly with Cobra; v1.3.0 added refined interrupt handling [VERIFIED: github.com/charmbracelet/bubbletea] |
| `charmbracelet/lipgloss` | v1.x | Terminal styling (table formatting, colored output) | Same Charm ecosystem as Bubbletea; consistent API [ASSUMED — version not individually verified beyond v1.x] |
| `hashicorp/go-retryablehttp` | v0.7.8 | Resilient HTTP for registry manifest fetches | Auto-retry, exponential backoff, Retry-After header; v0.7.7 fixed CVE-2024-6104 (log URL sanitization) [VERIFIED: github.com/hashicorp/go-retryablehttp] |
| stdlib `encoding/json` | — | All JSON config read/write | No external dependency; sufficient for all config formats |
| stdlib `os`, `path/filepath` | — | Cross-platform path handling | `os.UserHomeDir()`, `os.UserConfigDir()`, `os.UserCacheDir()` cover all platforms |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `google/renameio` | v2.x | Atomic file write on Windows | Every config file write; v2 respects umask; v1 is available but v2 preferred [VERIFIED: github.com/google/renameio] |
| `BurntSushi/toml` | v1.x | Read/write Codex config.toml | Phase 2 only — not needed in Phase 1 (Claude Code only); include in go.mod for Phase 2 readiness [ASSUMED — not independently re-verified but widely cited] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `spf13/cobra` | `urfave/cli` | urfave uses positional-arg style; worse for subcommand-heavy tools |
| `hashicorp/go-retryablehttp` | `net/http` directly | No auto-retry; must implement backoff manually; no Retry-After handling |
| `google/renameio` | `os.Rename` directly | `os.Rename` is not atomic on Windows (golang/go#8914); renameio provides correct cross-platform behavior |
| `charmbracelet/bubbletea` | `briandowns/spinner` | bubbletea provides interactive prompts too (needed for D-07 foreign conflict); spinner-only library cannot do both |

**Installation:**
```bash
go mod init github.com/ejyle/agentkit
go get github.com/spf13/cobra@v1.10.2
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/hashicorp/go-retryablehttp@v0.7.8
go get github.com/google/renameio/v2@latest
```

---

## Package Legitimacy Audit

slopcheck was not available at research time (install failed). All packages below are tagged `[ASSUMED]` per graceful degradation protocol. The planner must gate each `go get` behind a `checkpoint:human-verify` task confirming package identity.

| Package | Registry | Age | Downloads/Use | Source Repo | slopcheck | Disposition |
|---------|----------|-----|---------------|-------------|-----------|-------------|
| `github.com/spf13/cobra` | pkg.go.dev | 10+ yrs | Used by Kubernetes, Docker, GitHub CLI | github.com/spf13/cobra | [ASSUMED] | Approved — extremely well-known |
| `github.com/charmbracelet/bubbletea` | pkg.go.dev | 4+ yrs | 30K+ GitHub stars | github.com/charmbracelet/bubbletea | [ASSUMED] | Approved — well-known Charm ecosystem |
| `github.com/charmbracelet/lipgloss` | pkg.go.dev | 4+ yrs | Same Charm ecosystem | github.com/charmbracelet/lipgloss | [ASSUMED] | Approved — same org as bubbletea |
| `github.com/hashicorp/go-retryablehttp` | pkg.go.dev | 8+ yrs | HashiCorp production | github.com/hashicorp/go-retryablehttp | [ASSUMED] | Approved — HashiCorp production library |
| `github.com/google/renameio` | pkg.go.dev | 5+ yrs | Google-maintained | github.com/google/renameio | [ASSUMED] | Approved — Google-maintained |
| `github.com/BurntSushi/toml` | pkg.go.dev | 10+ yrs | Widely used Go TOML | github.com/BurntSushi/toml | [ASSUMED] | Approved — long-standing, stable |

*All packages above are tagged `[ASSUMED]` due to slopcheck unavailability. Planner must add a `checkpoint:human-verify` task before the `go get` wave.*

**Packages removed due to [SLOP] verdict:** none
**Packages flagged [SUS]:** none identified by manual review

---

## Architecture Patterns

### System Architecture Diagram

```
agentkit CLI (Cobra)
  │
  ├── install <name> --target claude
  │     │
  │     ▼
  │   InstallService
  │     ├── RegistryManager.Resolve("name")
  │     │     └── GitHubManifestRegistry
  │     │           ├── fetch registry.json (go-retryablehttp, ETag cache)
  │     │           └── return Package{manifest}
  │     ├── MCPInstaller.Install(manifest.install)
  │     │     ├── NpxInstaller: exec "npx -y <pkg>"
  │     │     └── BinaryInstaller: download → ~/.config/agentkit/bin/ → chmod
  │     ├── ClaudeCodeAdapter.WriteMCPConfig(entry)
  │     │     ├── detect path: ~/.claude.json OR .mcp.json
  │     │     ├── read existing → merge → write to temp → os.Rename
  │     │     └── ReadMCPConfig() verify (MCP-06)
  │     └── ConfigStore.RecordInstalled(pkg)
  │           └── ~/.config/agentkit/claude/installed.json (atomic write)
  │
  ├── list [--target claude]
  │     └── ConfigStore.ListInstalled() → table output (lipgloss)
  │
  ├── search <query>
  │     └── RegistryManager.Search(query) → ranked results
  │
  ├── uninstall <name>
  │     ├── ClaudeCodeAdapter.RemoveMCPConfig(name) (merge-write)
  │     ├── remove ~/.claude/skills/<name>/ if skill type
  │     └── ConfigStore.RemoveRecord(name)
  │
  └── update [name]
        ├── RegistryManager.Resolve (get latest version)
        ├── compare with installed.json
        └── if newer → InstallService.Install (D-08 auto-overwrite)
```

### Recommended Project Structure
```
agentkit/
├── cmd/
│   ├── root.go          # cobra root command, --target flag
│   ├── install.go       # agentkit install
│   ├── list.go          # agentkit list
│   ├── search.go        # agentkit search
│   ├── uninstall.go     # agentkit uninstall
│   └── update.go        # agentkit update
├── internal/
│   ├── domain/
│   │   ├── package.go   # Package, Manifest, MCPServerEntry types
│   │   └── installed.go # InstalledRecord type (D-11 schema)
│   ├── config/
│   │   ├── store.go     # ConfigStore: read/write installed.json + atomic writes
│   │   └── paths.go     # all path resolution via os.UserHomeDir/UserConfigDir
│   ├── registry/
│   │   ├── registry.go  # Registry interface
│   │   ├── manager.go   # RegistryManager: ordered resolution, search fan-out
│   │   ├── github.go    # GitHubManifestRegistry with ETag cache
│   │   └── local.go     # LocalRegistry for testing
│   ├── adapter/
│   │   ├── adapter.go   # AssistantAdapter interface
│   │   └── claude.go    # ClaudeCodeAdapter: ~/.claude.json + ~/.claude/skills/
│   ├── installer/
│   │   ├── installer.go # MCPInstaller interface
│   │   ├── npx.go       # NpxInstaller
│   │   └── binary.go    # BinaryInstaller
│   ├── service/
│   │   ├── install.go   # InstallService: orchestrate resolve→install→write→record
│   │   └── search.go    # SearchService: fan-out, rank, deduplicate
│   ├── skill/
│   │   └── validate.go  # ValidateSkill(): frontmatter parse, line count, paths
│   └── ui/
│       ├── spinner.go   # bubbletea spinner model
│       └── table.go     # lipgloss table rendering for list/search output
├── main.go
└── go.mod
```

### Pattern 1: Walking Skeleton — Minimal End-to-End Slice

**What:** Build just enough of each layer to get `agentkit install playwright --target claude` passing the success criteria test before expanding any layer.
**When to use:** First task of Wave 1 (Skeleton). Proves the architecture compiles and the data flow works before investing in full implementations.
**Example:**
```go
// Source: standard Go CLI walking skeleton pattern [ASSUMED]
// Minimal install command that proves end-to-end connectivity:
// 1. Cobra command wired with --target flag
// 2. LocalRegistry stub returns a hardcoded playwright package
// 3. NpxInstaller executes (or is stubbed)
// 4. ClaudeCodeAdapter writes a real ~/.claude.json entry
// 5. ConfigStore writes a real installed.json entry
// Swap LocalRegistry for GitHubManifestRegistry in next wave
func runInstall(cmd *cobra.Command, args []string) error {
    name := args[0]
    target, _ := cmd.Flags().GetString("target")
    // ... delegate to InstallService
}
```

### Pattern 2: Runtime MCP Config Path Detection

**What:** Detect whether `~/.claude.json` or `.mcp.json` exists; write to the correct location. Never hardcode the path.
**When to use:** Every write in ClaudeCodeAdapter.
**Example:**
```go
// Source: verified from Claude Code issue #4976 and current docs [CITED: github.com/anthropics/claude-code/issues/4976]
func (a *ClaudeCodeAdapter) mcpConfigPath() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    // User-scoped: ~/.claude.json (current Claude Code)
    userPath := filepath.Join(home, ".claude.json")
    if _, err := os.Stat(userPath); err == nil {
        return userPath, nil
    }
    // Fallback: ~/.claude/settings.json (older versions — still check)
    legacyPath := filepath.Join(home, ".claude", "settings.json")
    if _, err := os.Stat(legacyPath); err == nil {
        return legacyPath, nil
    }
    // Create user-scoped path on first use
    return userPath, nil
}
```

### Pattern 3: Atomic Config File Write

**What:** Read existing config → modify in memory → write to temp file → os.Rename (atomic on Unix/Linux; renameio on Windows).
**When to use:** Every file write that touches a config file users also edit (MCP config, installed.json).
**Example:**
```go
// Source: google/renameio v2 [CITED: github.com/google/renameio]
import "github.com/google/renameio/v2"

func writeJSON(path string, data interface{}) error {
    b, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }
    return renameio.WriteFile(path, b, 0644)
}
```

### Pattern 4: Bubbletea Spinner + Cobra Handoff

**What:** Cobra command handler launches a Bubbletea program for the spinner phases, gets the result back, then prints final output.
**When to use:** `install`, `update`, `search` — any command with a network call.
**Example:**
```go
// Source: charmbracelet/bubbletea docs [CITED: github.com/charmbracelet/bubbletea]
type spinnerModel struct {
    spinner  spinner.Model
    phase    string
    result   *domain.Package
    err      error
    done     bool
}

func (m spinnerModel) Init() tea.Cmd {
    return tea.Batch(m.spinner.Tick, resolvePackageCmd(m.name))
}
```

### Pattern 5: Registry ETag Cache

**What:** Store `ETag` response header alongside cached manifest; send `If-None-Match` on next fetch; use cache on 304.
**When to use:** Every registry manifest fetch.
**Example:**
```go
// Source: HTTP spec + go-retryablehttp [ASSUMED pattern]
type cachedManifest struct {
    ETag     string    `json:"etag"`
    FetchedAt time.Time `json:"fetched_at"`
    Manifest  Manifest  `json:"manifest"`
}
// On fetch failure, load cache with visible warning:
// [warn] registry unreachable; using cached manifest from 2h ago
```

### Anti-Patterns to Avoid

- **Hardcoded MCP config path:** `~/.claude/settings.json` is NOT where Claude Code reads MCP servers. Use runtime detection. This is the #1 pitfall in this phase.
- **`os.Getenv("HOME")` for path construction:** Returns empty string on Windows. Use `os.UserHomeDir()` exclusively.
- **`filepath.Join` with "/" literal:** Paths constructed with "/" separators fail on Windows. Use `filepath.Join` throughout.
- **Direct file.Write without rename:** Non-atomic; corrupts config if process is killed mid-write.
- **`http.DefaultClient` without timeout:** Corporate firewalls and slow networks hang indefinitely. Set 3s connect / 10s read.
- **Starting CLI commands before interfaces are stable:** Cobra commands are thin wrappers — if the service/adapter interfaces change, command wiring must change too. Finalize domain types and interfaces first.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| CLI command routing | Custom arg parser | `spf13/cobra` | Subcommand routing, shell completion, help generation — thousands of edge cases |
| Spinner / interactive prompts | Raw ANSI escape codes | `charmbracelet/bubbletea` | Terminal state management, resize, interrupt handling — non-trivial |
| HTTP retry with backoff | Custom retry loop | `hashicorp/go-retryablehttp` | Retry-After header, exponential jitter, request body rewind — subtle correctness |
| Atomic file write | `ioutil.WriteFile` | `google/renameio/v2` | Not atomic on Windows; partial write on crash corrupts config |
| Table formatting | Manual padding with fmt | `charmbracelet/lipgloss` | Column alignment, color, terminal width detection |

**Key insight:** In a CLI tool, output formatting and file safety are where hand-rolled solutions silently fail users. Use libraries for both.

---

## Common Pitfalls

### Pitfall 1: Wrong Claude Code MCP Config Path

**What goes wrong:** Writing MCP config to `~/.claude/settings.json` (the old documented location) causes the MCP server to never load. Claude Code currently reads user-scoped MCP servers from `~/.claude.json` (in home directory root, not inside `.claude/`). The install appears to succeed, but the MCP server is invisible.
**Why it happens:** The existing STACK.md and ARCHITECTURE.md in this project document the old path. Claude Code changed this without a major announcement (confirmed by issue #4976). The CLI docs may still show the old path.
**How to avoid:** Use runtime path detection as shown in Pattern 2. Check which file exists; create `~/.claude.json` on first use.
**Warning signs:** "Installed but not showing up in Claude Code" — almost always a path mismatch.

**Corrected mapping:**

| Claude Code Scope | File | Key |
|-------------------|------|-----|
| User-scoped MCP (current) | `~/.claude.json` | `mcpServers` |
| Project-scoped MCP | `./.mcp.json` | `mcpServers` |
| General settings (not MCP) | `~/.claude/settings.json` | various |
| Skill files | `~/.claude/skills/<name>/` | — |

### Pitfall 2: Non-Atomic Config Write Corrupts User Config

**What goes wrong:** Process killed mid-write leaves a 0-byte or partial `~/.claude.json`. Claude Code fails to start. User loses all MCP config.
**Why it happens:** `os.WriteFile` is not atomic.
**How to avoid:** Write to temp file then `os.Rename` (Unix) or `renameio.WriteFile` (cross-platform).
**Warning signs:** Any call to `os.WriteFile` or `ioutil.WriteFile` on a config file.

### Pitfall 3: Foreign MCP Conflict Not Detected

**What goes wrong:** User has manually configured a `playwright` entry in `~/.claude.json`. agentkit's install silently overwrites it with a different config. User's custom configuration is lost.
**Why it happens:** Read-before-write skipped or `source_url` ownership check skipped.
**How to avoid:** Implement D-07 — compare key in existing config against `installed.json` ownership record. If key exists but not in `installed.json`, prompt with old vs new entry.
**Warning signs:** Install path that doesn't read existing config before writing.

### Pitfall 4: Registry Fetch Hangs on Corporate Network

**What goes wrong:** `http.DefaultClient` has no timeout. On a network that blocks GitHub raw content, `agentkit install` hangs forever.
**Why it happens:** Go's default HTTP client has no timeout configured.
**How to avoid:** Use `go-retryablehttp` with explicit 3s connect / 10s read timeouts. Implement ETag cache so offline mode works.
**Warning signs:** Any `http.Get()` call without a custom client with timeout.

### Pitfall 5: Skill Directory Not Created Before Writing

**What goes wrong:** `os.WriteFile("~/.claude/skills/playwright/SKILL.md")` fails because the parent directory doesn't exist.
**Why it happens:** `os.WriteFile` doesn't create intermediate directories.
**How to avoid:** Call `os.MkdirAll(dir, 0755)` before writing any file in a skill directory.
**Warning signs:** Any WriteFile call without a preceding MkdirAll.

### Pitfall 6: go.mod Module Name

**What goes wrong:** Module named `agentkit` (no org path) causes `go install` issues and makes import paths ambiguous.
**Why it happens:** `go mod init agentkit` without full path.
**How to avoid:** Use `go mod init github.com/ejyle/agentkit` (or the actual org/repo path).
**Warning signs:** `go.mod` with a single-word module name.

---

## Code Examples

### Cobra Root Command with --target Flag
```go
// Source: spf13/cobra v1.10.x [CITED: github.com/spf13/cobra]
var rootCmd = &cobra.Command{
    Use:   "agentkit",
    Short: "AI agent skill and MCP server manager",
}

func init() {
    rootCmd.PersistentFlags().StringP("target", "t", "claude", 
        "Target coding assistant (claude|copilot|codex|gemini|opencode)")
}
```

### installed.json Entry Schema (D-11)
```go
// Source: CONTEXT.md D-11 [CITED: .planning/phases/01-foundation/01-CONTEXT.md]
type InstalledRecord struct {
    Name        string    `json:"name"`
    Version     string    `json:"version"`
    Type        string    `json:"type"` // "mcp" | "skill" | "agent"
    InstallPath string    `json:"install_path"` // e.g. "mcpServers.playwright"
    InstalledAt time.Time `json:"installed_at"`
    SourceURL   string    `json:"source_url"`
    Checksum    string    `json:"checksum"` // "sha256:<hex>"
}

type InstalledState struct {
    Packages map[string]InstalledRecord `json:"packages"`
}
```

### Post-Install Verify (MCP-06)
```go
// Source: CONTEXT.md D-11; MCP-06 requirement [ASSUMED pattern]
func (a *ClaudeCodeAdapter) WriteMCPConfig(entry MCPServerEntry) error {
    if err := a.mergeAndWrite(entry); err != nil {
        return fmt.Errorf("write MCP config: %w", err)
    }
    // Post-install verify: re-read and confirm entry present
    config, err := a.ReadMCPConfig()
    if err != nil {
        return fmt.Errorf("post-install verify failed (config unreadable): %w", err)
    }
    if _, ok := config.MCPServers[entry.Name]; !ok {
        return fmt.Errorf("post-install verify failed: %q not found after write", entry.Name)
    }
    return nil
}
```

### Registry Manifest Fetch with ETag
```go
// Source: HTTP spec + go-retryablehttp [ASSUMED pattern]
func (r *GitHubManifestRegistry) fetch() (*Manifest, error) {
    req, _ := http.NewRequest("GET", r.manifestURL, nil)
    if r.cache.ETag != "" {
        req.Header.Set("If-None-Match", r.cache.ETag)
    }
    resp, err := r.client.Do(req)
    if err != nil {
        if r.cache.Manifest != nil {
            log.Warn("registry unreachable; using cached manifest from", r.cache.FetchedAt)
            return r.cache.Manifest, nil
        }
        return nil, fmt.Errorf("registry unreachable and no cache: %w", err)
    }
    if resp.StatusCode == 304 {
        return r.cache.Manifest, nil
    }
    // parse and cache with new ETag
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| MCP config at `~/.claude/settings.json` | MCP config at `~/.claude.json` (user-scope) | Approx mid-2025 | Existing research in STACK.md/ARCHITECTURE.md is wrong; ClaudeCode adapter must detect at runtime |
| `bubbletea` v0.x | `bubbletea` v1.3.x (stable API) | v1.0.0 released 2024 | v1 API is stable; v2 is in pre-release (RC) with breaking changes — use v1 |
| `go-retryablehttp` v0.7.4 | v0.7.8 (CVE-2024-6104 fix) | Jun 2025 | Pin to v0.7.8+ to avoid log URL credential leak vulnerability |

**Deprecated/outdated in existing research:**
- `~/.claude/settings.json` as MCP config path: confirmed incorrect per GitHub issue #4976 and 2025-2026 blog sources. The existing STACK.md and ARCHITECTURE.md must be treated as wrong for this specific detail.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `charmbracelet/lipgloss` v1.x is the current stable API | Standard Stack | Minor — if API changed, table rendering code needs updating |
| A2 | `BurntSushi/toml` v1.x is current and stable | Standard Stack | Minor — Phase 1 doesn't use it; risk is Phase 2 |
| A3 | `go-retryablehttp` v0.7.8 is the latest non-breaking release | Standard Stack | Low — if newer patch exists, update pin |
| A4 | `google/renameio/v2` is the recommended version for new projects | Standard Stack | Low — v1 also works on Unix/macOS; v2 only matters for umask behavior |
| A5 | Ranking algorithm for search: exact name match → fuzzy → description keyword | Claude's Discretion | Low — planner/implementer can choose; correct by convention |
| A6 | `~/.claude.json` is the CURRENT user-scoped MCP path for Claude Code | **Critical pitfall** | High — if this changes again, adapter breaks silently. Always detect at runtime, never hardcode. |

---

## Open Questions

1. **Does the agentkit-registry GitHub repo exist yet?**
   - What we know: ROADMAP.md TODO: "Create agentkit-registry GitHub repo with initial registry.json listing 9 bundled skills"
   - What's unclear: Repo not yet created; Phase 1 tests require it to pass REG-05
   - Recommendation: Wave 0 task — create the repo with a minimal registry.json containing at least playwright so the walking skeleton install test can run.

2. **`~/.claude.json` format: does it contain ONLY mcpServers or other keys?**
   - What we know: Multiple sources confirm `mcpServers` key is present; the file may contain other Claude Code state
   - What's unclear: If the file contains non-MCP keys, the merge must be truly non-destructive (read ALL keys, modify only mcpServers)
   - Recommendation: Implement merge as generic JSON map merge — read as `map[string]interface{}`, modify `mcpServers` key only, write back.

3. **Does `bubbletea` v2 have any features needed for Phase 1?**
   - What we know: v2 is in RC; v1 is stable. v2 has breaking API changes.
   - What's unclear: Whether v2 offers meaningful improvements for the spinner + table use cases
   - Recommendation: Use v1.3.x for Phase 1. Evaluate v2 after GA.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Building the binary | Yes | go1.26.3 | — |
| `node` / `npx` | NpxInstaller.IsAvailable() check | Not checked | — | NpxInstaller.IsAvailable() returns false; install proceeds with warning |
| GitHub network access | Registry manifest fetch | Likely (dev machine) | — | ETag cache + --offline flag |
| `~/.claude/` directory | ClaudeCodeAdapter target | Not checked | — | Created on first install |

**Missing dependencies with no fallback:** None for building. Node.js is a runtime prereq for npx installs but agentkit itself does not require it.

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | No auth in Phase 1 (public registry, public GitHub) |
| V3 Session Management | No | CLI tool, no sessions |
| V4 Access Control | No | User-scope only, single user |
| V5 Input Validation | Yes | Package name input validated before use in exec.Command (shell injection risk) |
| V6 Cryptography | Yes | SHA256 checksum verification of downloaded binaries (MCP-03 BinaryInstaller) |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Shell injection via package name | Tampering | Pass package name as argument array to `exec.Command`, never interpolate into shell string |
| Registry manifest MITM | Spoofing | HTTPS only for manifest fetch; SHA256 checksum verification for binary downloads |
| Partial write corrupting user config | Tampering | Atomic write via renameio; post-install verify re-read |
| Log URL credential leak | Information Disclosure | Use go-retryablehttp v0.7.8+ (CVE-2024-6104 fix) |

---

## Sources

### Primary (HIGH confidence)
- [github.com/spf13/cobra](https://github.com/spf13/cobra) — v1.10.2 verified, Dec 2025 release
- [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — v1.3.x stable, v2 in RC
- [github.com/hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp) — v0.7.8, Jun 2025 (CVE fix)
- [github.com/google/renameio](https://github.com/google/renameio) — v2 current
- [anthropics/claude-code#4976](https://github.com/anthropics/claude-code/issues/4976) — MCP config path instability confirmed
- Claude Code official MCP docs — [code.claude.com/docs/en/mcp](https://code.claude.com/docs/en/mcp) — `~/.claude.json` for user-scoped MCP

### Secondary (MEDIUM confidence)
- [inventivehq.com Claude config guide](https://inventivehq.com/knowledge-base/claude/where-configuration-files-are-stored) — confirmed ~/.claude.json vs settings.json distinction
- [nimbalyst.com Claude Code MCP 2026 guide](https://nimbalyst.com/blog/claude-code-mcp-setup/) — confirms current path

### Tertiary (LOW confidence)
- BurntSushi/toml v1.x — version not independently re-verified beyond training knowledge
- lipgloss v1.x — version not independently re-verified

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — verified via GitHub releases and pkg.go.dev; CVE fix version confirmed
- Architecture patterns: HIGH — standard Go CLI patterns; walking skeleton approach is standard for greenfield
- Claude Code MCP path: HIGH (corrected from existing research) — confirmed via GitHub issue #4976, official docs, multiple 2025-2026 sources
- Pitfalls: HIGH — based on verified GitHub issues and documented CVEs

**Research date:** 2026-06-08
**Valid until:** 2026-08-08 (60 days — Go libraries stable; Claude Code config paths should be re-verified before any adapter code is written)

**Critical correction from existing project research:** `.planning/research/STACK.md` and `.planning/research/ARCHITECTURE.md` both document `~/.claude/settings.json` as the Claude Code MCP config path. **This is incorrect as of current Claude Code versions.** The correct path is `~/.claude.json`. The planner must ensure the ClaudeCodeAdapter uses runtime detection (Pattern 2 above), not the path from the existing research files.
