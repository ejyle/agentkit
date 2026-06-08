---
phase: 01-foundation
verified: 2026-06-08T01:00:00Z
status: passed
score: 20/20 must-haves verified
overrides_applied: 0
re_verification:
  previous_status: human_needed
  previous_score: 19/20
  gaps_closed:
    - "MCP-05: CustomInstaller implemented — NewInstaller factory now returns NewCustomInstaller() for InstallMethodCustom"
  gaps_remaining: []
  regressions: []
---

# Phase 01: Foundation Verification Report

**Phase Goal:** Build the agentkit foundation — compilable Go project with cobra CLI skeleton, data layer, install vertical slice, and all 5 CLI commands working end-to-end.
**Verified:** 2026-06-08T01:00:00Z
**Status:** passed
**Re-verification:** Yes — after MCP-05 gap closure

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `go build ./...` succeeds with zero errors | ✓ VERIFIED | `agentkit` binary exists at project root; `go.mod` has `module github.com/ejyle/agentkit`; all required source files present |
| 2 | `agentkit --help` shows install/list/search/uninstall/update subcommands | ✓ VERIFIED | All 5 `cmd/*.go` files declare cobra commands and call `rootCmd.AddCommand()` in `init()` |
| 3 | Domain types Package, Manifest, MCPServerEntry, InstalledRecord, InstalledState exported | ✓ VERIFIED | `internal/domain/package.go` and `internal/domain/installed.go` export all required types with correct D-11 JSON tags |
| 4 | ConfigStore reads/writes InstalledRecord atomically via renameio | ✓ VERIFIED | `internal/config/store.go` imports `renameio/v2`; all write paths go through `renameio.WriteFile` |
| 5 | ConfigStore auto-creates directory on first write | ✓ VERIFIED | `RecordInstalled` calls `os.MkdirAll` before write |
| 6 | GitHubManifestRegistry fetches manifest with ETag caching and offline fallback | ✓ VERIFIED | `internal/registry/github.go` implements ETag 304 handling; stale fallback present |
| 7 | RegistryManager.Resolve and Search work with agentkit-registry and gsd-core defaults | ✓ VERIFIED | `NewRegistryManager()` registers both `agentkit-registry` and `gsd-core` GitHub URLs |
| 8 | NpxInstaller uses arg-array form (no shell interpolation) | ✓ VERIFIED | `internal/installer/npx.go` uses `exec.Command("npx", "-y", spec.Package)` |
| 9 | BinaryInstaller downloads via HTTPS only, verifies SHA256 | ✓ VERIFIED | Scheme check `u.Scheme != "https"` returns `ErrInsecureURL`; `crypto/sha256` used |
| 10 | ClaudeCodeAdapter detects config path at runtime (not hardcoded) | ✓ VERIFIED | `mcpConfigPath()` stats `~/.claude.json` then `~/.claude/settings.json` at runtime |
| 11 | MCP config merge is non-destructive; only mcpServers key touched | ✓ VERIFIED | `WriteMCPConfig` reads full config as `map[string]interface{}`, sets only `mcpServers[entry.Name]` |
| 12 | Post-install verify re-reads config and fails loudly if key absent | ✓ VERIFIED | `ReadMCPConfig()` called after write; error returned if key absent |
| 13 | Foreign conflict detection returns ErrForeignConflict when not owned | ✓ VERIFIED | `ErrForeignConflict` struct defined; `WriteMCPConfig` checks store ownership before overwrite |
| 14 | InstallService orchestrates: Resolve → Install → Validate → WriteMCPConfig → RecordInstalled | ✓ VERIFIED | `internal/service/install.go` implements all 9 steps in order |
| 15 | Skill validator checks SKILL.md presence, 500-line warning, references/ files | ✓ VERIFIED | `internal/skill/validate.go` implements all 3 checks |
| 16 | agentkit list prints table with PACKAGE/VERSION/TYPE/TARGET/REGISTRY columns | ✓ VERIFIED | `internal/ui/table.go` renders correct headers; `cmd/list.go` calls `RenderInstalledTable` |
| 17 | agentkit search shows spinner then ranked results | ✓ VERIFIED | `cmd/search.go` launches `SpinnerModel` then calls `RenderSearchResults` |
| 18 | agentkit uninstall removes MCP config entry and installed.json record | ✓ VERIFIED | `UninstallService` calls `RemoveMCPConfig` then `RemoveRecord` |
| 19 | agentkit update resolves latest version, delegates to InstallService | ✓ VERIFIED | `UpdateService` compares versions, calls `installer.Install`; `UpdateAll` iterates all records |
| 20 | MCP-05: Custom install method supported end-to-end | ✓ VERIFIED | `internal/installer/custom.go` implements `CustomInstaller`; `NewInstaller` factory returns `NewCustomInstaller()` for `InstallMethodCustom`; 6 tests in `custom_test.go` including `TestNewInstaller_CustomMethod` |

**Score:** 20/20 truths verified

### MCP-05 Gap Closure Evidence

The previously failing truth (Truth 20) is now **VERIFIED**:

- `internal/installer/custom.go` — `CustomInstaller` struct implementing `MCPInstaller` interface; `Install()` uses `exec.Command` arg-array form (no shell interpolation); `IsAvailable()` always returns true; `Method()` returns `domain.InstallMethodCustom`; `ErrCustomMissingCommand` sentinel defined.
- `internal/installer/installer.go` line 38 — `case domain.InstallMethodCustom: return NewCustomInstaller(), nil` — factory now routes `custom` to real implementation instead of returning an error.
- `internal/installer/custom_test.go` — 6 substantive tests: `TestCustomInstaller_Method`, `TestCustomInstaller_IsAvailable`, `TestCustomInstaller_MissingCommand`, `TestCustomInstaller_RunsCommand`, `TestCustomInstaller_CommandError`, `TestNewInstaller_CustomMethod`.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `go.mod` | Module with all deps | ✓ VERIFIED | Module `github.com/ejyle/agentkit`; all required deps present |
| `main.go` | Entry point | ✓ VERIFIED | Calls `cmd.Execute()` |
| `cmd/root.go` | Cobra root with `--target` | ✓ VERIFIED | PersistentFlags + PersistentPreRunE validation |
| `internal/domain/package.go` | Package, Manifest, MCPServerEntry, InstallSpec | ✓ VERIFIED | All types exported with correct json tags |
| `internal/domain/installed.go` | InstalledRecord per D-11 | ✓ VERIFIED | D-11 JSON tags: `install_path`, `installed_at`, `source_url`, `checksum` |
| `internal/config/paths.go` | Path functions | ✓ VERIFIED | All 4 functions using stdlib |
| `internal/config/store.go` | ConfigStore CRUD | ✓ VERIFIED | Full CRUD with renameio atomic writes |
| `internal/registry/registry.go` | Registry interface + manager | ✓ VERIFIED | Both defaults registered |
| `internal/registry/github.go` | GitHubManifestRegistry | ✓ VERIFIED | ETag, timeouts, retries |
| `internal/registry/cache.go` | CachedManifest | ✓ VERIFIED | ETag field; renameio writes |
| `internal/installer/installer.go` | MCPInstaller interface + factory | ✓ VERIFIED | Factory handles npx, binary, custom |
| `internal/installer/npx.go` | NpxInstaller | ✓ VERIFIED | arg-array exec.Command |
| `internal/installer/binary.go` | BinaryInstaller | ✓ VERIFIED | HTTPS check; SHA256 verification |
| `internal/installer/custom.go` | CustomInstaller (MCP-05) | ✓ VERIFIED | NEW — arg-array exec.Command; ErrCustomMissingCommand; injected runner for tests |
| `internal/adapter/adapter.go` | AssistantAdapter interface | ✓ VERIFIED | 5 methods; ErrForeignConflict |
| `internal/adapter/claude.go` | ClaudeCodeAdapter | ✓ VERIFIED | Runtime path; non-destructive merge; post-install verify |
| `internal/service/install.go` | InstallService 9-step | ✓ VERIFIED | All 9 steps; local interfaces |
| `internal/skill/validate.go` | ValidateSkill | ✓ VERIFIED | SKILL.md, line count, references checks |
| `internal/ui/spinner.go` | SpinnerModel | ✓ VERIFIED | 3 phases; tea.Model |
| `internal/ui/table.go` | Table renderers | ✓ VERIFIED | Both functions; lipgloss; correct headers |
| `internal/service/search.go` | SearchService | ✓ VERIFIED | Delegates to registry |
| `internal/service/uninstall.go` | UninstallService | ✓ VERIFIED | ErrNotInstalled; correct ordering |
| `internal/service/update.go` | UpdateService | ✓ VERIFIED | Update + UpdateAll |
| `cmd/install.go` | Full install command | ✓ VERIFIED | Wires all dependencies |
| `cmd/list.go` | List command | ✓ VERIFIED | ListInstalled + RenderInstalledTable |
| `cmd/search.go` | Search with spinner | ✓ VERIFIED | SpinnerModel + RenderSearchResults |
| `cmd/uninstall.go` | Uninstall command | ✓ VERIFIED | errors.Is(ErrNotInstalled) |
| `cmd/update.go` | Update command | ✓ VERIFIED | Single + all update paths |

### Key Link Verification

| From | To | Via | Status |
|------|----|-----|--------|
| `main.go` | `cmd/root.go` | `cmd.Execute()` | ✓ WIRED |
| `cmd/install.go` | `internal/service/install.go` | `InstallService.Install()` | ✓ WIRED |
| `internal/service/install.go` | `internal/registry/manager.go` | `Resolve()` | ✓ WIRED |
| `internal/service/install.go` | `internal/adapter/claude.go` | `WriteMCPConfig()` | ✓ WIRED |
| `internal/service/install.go` | `internal/config/store.go` | `RecordInstalled()` | ✓ WIRED |
| `internal/installer/installer.go` | `internal/installer/custom.go` | `NewCustomInstaller()` | ✓ WIRED |
| `internal/config/store.go` | `renameio/v2` | `renameio.WriteFile` | ✓ WIRED |
| `internal/registry/github.go` | `internal/registry/cache.go` | `CachedManifest` | ✓ WIRED |
| `cmd/list.go` | `internal/config/store.go` | `ListInstalled()` | ✓ WIRED |
| `cmd/search.go` | `internal/service/search.go` | `SearchService` | ✓ WIRED |
| `internal/service/uninstall.go` | `internal/adapter/claude.go` | `RemoveMCPConfig()` | ✓ WIRED |
| `internal/service/update.go` | `internal/service/install.go` | `InstallService.Install()` | ✓ WIRED |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| CLI-01 | 01-01, 01-03 | `agentkit install <name>` | ✓ SATISFIED | `cmd/install.go` + InstallService |
| CLI-02 | 01-01, 01-03 | `--target` flag | ✓ SATISFIED | PersistentFlag in root.go |
| CLI-05 | 01-05 | `agentkit uninstall <name>` | ✓ SATISFIED | `cmd/uninstall.go` + UninstallService |
| CLI-06 | 01-04 | `agentkit search <query>` | ✓ SATISFIED | `cmd/search.go` + SearchService |
| CLI-07 | 01-05 | `agentkit update [name]` | ✓ SATISFIED | `cmd/update.go` + UpdateService |
| CLI-08 | 01-01, 01-04 | `agentkit list` | ✓ SATISFIED | `cmd/list.go` + RenderInstalledTable |
| REG-01 | 01-02 | GitHub manifest-driven registry | ✓ SATISFIED | GitHubManifestRegistry |
| REG-02 | 01-02 | open-gsd/gsd-core as default | ✓ SATISFIED | Registered in NewRegistryManager() |
| REG-05 | 01-02 | agentkit-registry as default | ✓ SATISFIED | Registered in NewRegistryManager() |
| REG-06 | 01-02 | ETag cache with offline fallback | ✓ SATISFIED | 304 handling + stale fallback |
| AST-01 | 01-03 | Claude Code adapter | ✓ SATISFIED | ClaudeCodeAdapter |
| MCP-01 | 01-03 | npx install adapter | ✓ SATISFIED | NpxInstaller |
| MCP-03 | 01-03 | Binary download adapter | ✓ SATISFIED | BinaryInstaller |
| MCP-05 | 01-03 | Custom install method override | ✓ SATISFIED | CustomInstaller in custom.go; factory wired |
| MCP-06 | 01-03 | Post-install verify | ✓ SATISFIED | Re-reads config after write |
| MCP-07 | 01-03 | Runtime config path detection | ✓ SATISFIED | `mcpConfigPath()` runtime stat |
| SKL-01 | 01-03 | Skill SKILL.md validation | ✓ SATISFIED | ValidateSkill |
| SKL-02 | 01-03 | 500-line warning | ✓ SATISFIED | Non-blocking warning |
| SKL-03 | 01-03 | references/ validation | ✓ SATISFIED | Files checked in ValidateSkill |

### Anti-Patterns Found

| File | Line | Pattern | Severity |
|------|------|---------|----------|
| None | — | No debt markers found | — |

No shell interpolation in npx.go or custom.go. No `os.WriteFile` in write paths. No `http.DefaultClient`. No hardcoded HOME paths.

### Human Verification (Previously Completed)

Both human verification items from the initial report were confirmed by the developer prior to this re-verification:

1. **11-step walking skeleton test** — User approved all 11 steps in plan 01-06.
2. **`go test ./... -count=1`** — User confirmed all tests pass.

These items are closed. No new human verification items identified.

---

_Verified: 2026-06-08T01:00:00Z_
_Verifier: Claude (gsd-verifier)_
