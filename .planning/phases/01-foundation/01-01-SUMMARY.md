---
phase: 01-foundation
plan: 01
subsystem: cli-scaffold
tags: [go, cobra, domain-types, scaffold]
dependency_graph:
  requires: []
  provides:
    - go.mod (github.com/ejyle/agentkit)
    - internal/domain types (Package, Manifest, MCPServerEntry, InstalledRecord, InstalledState)
    - cmd cobra skeleton (install, list, search, uninstall, update)
  affects:
    - all subsequent plans import internal/domain
    - all subsequent plans extend cmd/
tech_stack:
  added:
    - github.com/spf13/cobra@v1.10.2
    - github.com/charmbracelet/bubbletea@v1.3.10
    - github.com/charmbracelet/lipgloss@v1.1.0
    - github.com/hashicorp/go-retryablehttp@v0.7.8
    - github.com/google/renameio/v2@v2.0.2
  patterns:
    - Cobra root + subcommand registration pattern via init()
    - Domain types in internal/domain with snake_case JSON tags per D-11
key_files:
  created:
    - go.mod
    - go.sum
    - main.go
    - cmd/root.go
    - cmd/install.go
    - cmd/list.go
    - cmd/search.go
    - cmd/uninstall.go
    - cmd/update.go
    - internal/domain/package.go
    - internal/domain/installed.go
    - internal/domain/package_test.go
    - internal/domain/installed_test.go
  modified: []
decisions:
  - "cobra v1.10.2 pinned (not latest) per CLAUDE.md recommendation"
  - "bubbletea v1.3.10 (v1.x stable, not v2 RC) per SKELETON.md"
  - "D-11 InstalledRecord uses snake_case JSON: install_path, source_url, installed_at"
  - "PersistentPreRunE validates --target against allowlist; returns formatted error"
  - "InstallMethod and PackageType are string types with const block for extensibility"
metrics:
  duration: "~15 minutes"
  completed_date: "2026-06-08"
  tasks_completed: 2
  tasks_total: 2
  files_created: 13
  files_modified: 0
requirements_satisfied:
  - CLI-01
  - CLI-02
---

# Phase 01 Plan 01: Go Module Scaffold and Domain Types Summary

**One-liner:** Go module with cobra CLI skeleton (5 stub subcommands, --target flag validation) and D-11-compliant domain types (Package, Manifest, InstalledRecord, InstalledState) with JSON roundtrip tests.

---

## What Was Built

A compilable Go project from greenfield providing:

1. **go.mod** — `module github.com/ejyle/agentkit`, Go 1.21, all 5 approved dependencies vendored in go.sum.

2. **cmd/root.go** — Cobra root command `agentkit` with `--target` persistent flag (default `"claude"`), validated against the allowlist `[claude, copilot, codex, gemini, opencode]` in `PersistentPreRunE`.

3. **Stub subcommands** — `install`, `list`, `search`, `uninstall`, `update` — all registered to rootCmd with meaningful `Short` descriptions; each `RunE` returns `nil` (implementation in later plans).

4. **internal/domain/package.go** — `InstallMethod` (npx/binary/custom), `PackageType` (mcp/skill/agent), `InstallSpec`, `MCPServerEntry`, `Package`, `Manifest` — all exported with json tags.

5. **internal/domain/installed.go** — `InstalledRecord` (D-11 schema: `install_path`, `source_url`, `installed_at` all snake_case), `InstalledState` (Packages map, nil-safe zero-value).

6. **Test files** — 8 tests covering JSON roundtrip, D-11 field-name enforcement, zero-value nil-safety, and constant values. All pass.

---

## Task Completion

| Task | Status | Commit |
|------|--------|--------|
| Task 1: Verify package legitimacy | Approved by user (pre-execution checkpoint) | — |
| Task 2: Scaffold Go module and domain types | Complete | 4335d31 |

---

## Verification Results

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `./agentkit --help` contains install/list/search/uninstall/update | PASS |
| `./agentkit install --help` contains `--target` | PASS |
| `go test ./internal/domain/...` | PASS (8/8) |
| `go vet ./...` | PASS |
| `head -1 go.mod` = `module github.com/ejyle/agentkit` | PASS |
| `installed.go` json tag = `install_path` | PASS |
| `installed.go` json tag = `source_url` | PASS |
| `installed.go` json tag = `installed_at` | PASS |
| `grep os.Getenv("HOME")` = empty | PASS |

---

## Deviations from Plan

None — plan executed exactly as written.

---

## Known Stubs

All 5 subcommands (`install`, `list`, `search`, `uninstall`, `update`) have `RunE` returning `nil`. This is intentional — implementations are planned in 01-02 through 01-05. The stubs do not block the plan's goal (compilable scaffold with domain types).

---

## Threat Flags

None. No new network endpoints or auth paths introduced in this plan. Package supply-chain risk was mitigated by the Task 1 human-verify checkpoint.

---

## Self-Check: PASSED

Files verified:
- go.mod EXISTS
- main.go EXISTS
- cmd/root.go EXISTS
- internal/domain/package.go EXISTS
- internal/domain/installed.go EXISTS
- commit 4335d31 EXISTS in git log
