---
phase: 01-foundation
plan: "03"
subsystem: install-vertical-slice
tags: [installer, adapter, service, cli, tdd]
dependency_graph:
  requires: [01-01, 01-02]
  provides: [MCPInstaller, AssistantAdapter, InstallService, agentkit-install-command]
  affects: [cmd/install.go, internal/installer, internal/adapter, internal/service, internal/skill, internal/ui]
tech_stack:
  added: [charmbracelet/bubbles@v1.0.0]
  patterns: [arg-array exec, atomic-renameio-write, runtime-path-detection, post-install-verify, bubbletea-spinner]
key_files:
  created:
    - internal/installer/installer.go
    - internal/installer/npx.go
    - internal/installer/binary.go
    - internal/installer/npx_test.go
    - internal/installer/binary_test.go
    - internal/adapter/adapter.go
    - internal/adapter/claude.go
    - internal/adapter/claude_test.go
    - internal/skill/validate.go
    - internal/skill/validate_test.go
    - internal/service/install.go
    - internal/service/install_test.go
    - internal/ui/spinner.go
  modified:
    - cmd/install.go
    - go.mod
    - go.sum
decisions:
  - "InstallService uses local interface types (Resolver, AdapterWriter, Recorder, Installer) for full mock-injectability in tests"
  - "BinaryInstaller receives injected http.Client and binDir in tests to avoid real filesystem and network access"
  - "ClaudeCodeAdapter tests inject homeDir via NewClaudeCodeAdapterWithHome to isolate all reads and writes from real ~/.claude.json"
  - "SpinnerModel drives async install via goroutine + channel + tea.Send — avoids blocking bubbletea's event loop"
  - "InstallService.Install skips WriteMCPConfig for skill-type packages (skills use WriteSkill instead)"
metrics:
  duration: "~45 minutes"
  completed: "2026-06-08"
  tasks_completed: 2
  files_created: 13
  files_modified: 3
  tests_added: 25
  tests_passing: 48
---

# Phase 1 Plan 03: Install Vertical Slice Summary

**One-liner:** Walking skeleton install command with npx/binary installers, ClaudeCodeAdapter (runtime path detection + atomic merge + post-install verify + foreign conflict detection), InstallService orchestrating the 9-step flow, and bubbletea spinner.

---

## What Was Built

### Task 1: MCP Installers, ClaudeCodeAdapter, Skill Validator

**NpxInstaller** (`internal/installer/npx.go`):
- Exec arg-array form — never shell string interpolation (T-03-01 mitigated)
- Injectable `lookPathFunc` and `runFunc` for test isolation
- Returns `ErrNodeNotFound` when Node.js is not on PATH

**BinaryInstaller** (`internal/installer/binary.go`):
- Rejects non-HTTPS URLs with `ErrInsecureURL` (T-03-03 mitigated)
- SHA256 checksum verification against `spec.Args[0]` (T-03-02 mitigated)
- Atomic write via temp file + `os.Rename`; injectable client + binDir for testing

**ClaudeCodeAdapter** (`internal/adapter/claude.go`):
- Runtime path detection: `~/.claude.json` → `~/.claude/settings.json` → create `~/.claude.json` (Pattern 2, Pitfall 1)
- Non-destructive merge: reads full file as `map[string]interface{}`, modifies only `mcpServers` key (T-03-04)
- Atomic write via `renameio.WriteFile` (no `os.WriteFile` in production code)
- Foreign conflict: checks `installed.json` ownership before any overwrite; returns `ErrForeignConflict` (D-07, T-03-05)
- Auto-overwrite for agentkit-owned keys (D-08)
- Post-install verify: re-reads config after write, fails loudly if key absent (T-03-06, MCP-06)
- 14 tests, all isolated to `t.TempDir()` — no writes to real `~/.claude.json`

**ValidateSkill** (`internal/skill/validate.go`):
- SKILL.md presence check (SKL-01)
- Line count warning at >500 (non-blocking, SKL-02)
- `references/*.md` existence check from `manifest.Install.Args` (SKL-03)

### Task 2: InstallService, Spinner, agentkit install Command

**InstallService** (`internal/service/install.go`):
- 9-step orchestration: resolve → create installer → install → validate (skill) → write skill/MCP → record
- Local interface types (`Resolver`, `AdapterWriter`, `Recorder`, `Installer`) enable full mock injection
- `NewInstallServiceWithValidator` for injecting stub skill validator in tests
- ErrForeignConflict is propagated transparently to the CLI layer

**SpinnerModel** (`internal/ui/spinner.go`):
- `charmbracelet/bubbles` spinner with D-03 phases: `PhaseFetchRegistry` → `PhaseResolving` → `PhaseInstalling`
- Message types: `PhaseUpdateMsg`, `DoneMsg`, `ErrorMsg`
- `View()` returns empty string when done — caller prints success line

**cmd/install.go**:
- Wires `ConfigStore` + `RegistryManager` + `ClaudeCodeAdapter` + `InstallService`
- Runs spinner in goroutine; sends `DoneMsg`/`ErrorMsg` via `p.Send()`
- D-03 success: `✓ <name>@<version> installed → <path> (<target>)`
- D-04 error: `✗ Error: <message>\nRun: agentkit search <name>`, exits 1
- D-07 conflict: shows old vs new entry, prompts `Overwrite? [y/N]`

---

## Test Results

```
go test ./... — 48 passed in 10 packages
go vet ./...  — no issues
```

TDD gate compliance:
- Task 1: RED commit `f439bb2` (test) → GREEN commit `cc807f6` (feat)
- Task 2: RED commit `2220788` (test) → GREEN commit `a6f33db` (feat)

---

## Deviations from Plan

### Auto-added: `NewNpxInstallerWithRunner` test helper

**Found during:** Task 1 — npx test for no-shell-interpolation
**Issue:** Plan specified testing that `exec.Command` is called with arg-array form, but the test can't inspect exec.Command calls without a runner hook.
**Fix:** Added `NewNpxInstallerWithRunner(runFunc)` constructor alongside `NewNpxInstallerWithLookPath`. No behaviour change to production path.
**Files modified:** `internal/installer/npx.go`, `internal/installer/npx_test.go`

### Auto-added: `charmbracelet/bubbles` dependency (Rule 2 — missing critical functionality)

**Found during:** Task 2 — spinner implementation
**Issue:** `bubbles` package not in go.mod; required for `spinner.Model` used by `SpinnerModel`.
**Fix:** `go get github.com/charmbracelet/bubbles@latest` (v1.0.0 — stable official Charm package, same org as bubbletea/lipgloss already in go.mod).
**Files modified:** `go.mod`, `go.sum`

### Auto-fixed: InstallService skips WriteMCPConfig for skill-type packages (Rule 1 — bug fix)

**Found during:** Task 2 — implementing InstallService
**Issue:** Plan step 5-6 builds MCPServerEntry and calls WriteMCPConfig regardless of package type. Skill packages should call WriteSkill, not WriteMCPConfig (skills don't have MCP server entries).
**Fix:** Added `if pkg.Type != domain.PackageTypeSkill` guard around WriteMCPConfig call; skill path calls WriteSkill instead.
**Files modified:** `internal/service/install.go`

---

## Known Stubs

None — all paths in the install flow are fully wired. The walking skeleton is complete:
`agentkit install playwright --target claude` will resolve from registry, run npx adapter, write `~/.claude.json`, record to `installed.json`, show spinner, print success line.

---

## Threat Flags

No new threat surface introduced beyond what is in the plan's threat model. All T-03-xx mitigations are implemented and tested.

---

## Self-Check: PASSED

Files exist:
- internal/installer/installer.go: FOUND
- internal/installer/npx.go: FOUND
- internal/installer/binary.go: FOUND
- internal/adapter/adapter.go: FOUND
- internal/adapter/claude.go: FOUND
- internal/skill/validate.go: FOUND
- internal/service/install.go: FOUND
- internal/ui/spinner.go: FOUND
- cmd/install.go: FOUND (updated)

Commits exist:
- f439bb2: test(01-03) RED Task 1
- cc807f6: feat(01-03) GREEN Task 1
- 2220788: test(01-03) RED Task 2
- a6f33db: feat(01-03) GREEN Task 2
