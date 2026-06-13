---
phase: 01-foundation
plan: 05
subsystem: service/cmd
tags: [uninstall, update, service, cobra, tdd]
dependency_graph:
  requires: [01-02, 01-03]
  provides: [UninstallService, UpdateService, agentkit-uninstall, agentkit-update]
  affects: [internal/service, cmd]
tech_stack:
  added: []
  patterns: [local-interface-injection, tdd-red-green, installservice-delegate]
key_files:
  created:
    - internal/service/uninstall.go
    - internal/service/uninstall_test.go
    - internal/service/update.go
    - internal/service/update_test.go
  modified:
    - cmd/uninstall.go
    - cmd/update.go
decisions:
  - "ErrNotInstalled shared sentinel: defined once in uninstall.go, reused by UpdateService"
  - "UpdateService uses local updateInstaller interface so InstallService is injected without package coupling"
  - "installServiceAdapter in cmd/update.go bridges InstallService.Install to updateInstaller interface"
  - "Duplicate --target flag removed from cmd init; uses root PersistentFlag"
metrics:
  duration: "~15 minutes"
  completed: "2026-06-08T15:20:00Z"
  tasks: 2
  files_created: 4
  files_modified: 2
---

# Phase 01 Plan 05: Uninstall and Update Commands Summary

**One-liner:** D-09 non-destructive uninstall and D-08 InstallService-delegating update — completes Phase 1 CLI surface.

---

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | UninstallService and UpdateService (TDD) | RED: test(01-05), GREEN: feat(01-05) | uninstall.go, uninstall_test.go, update.go, update_test.go |
| 2 | Wire cmd/uninstall.go and cmd/update.go | feat(01-05) | cmd/uninstall.go, cmd/update.go |

---

## What Was Built

### UninstallService (`internal/service/uninstall.go`)

Implements D-09 non-destructive uninstall:
1. GetRecord from store — return `ErrNotInstalled` if absent
2. RemoveMCPConfig via adapter — halt on error (RemoveRecord is NOT called if this fails)
3. RemoveSkill via adapter (skill-type only) — halt on error
4. RemoveRecord to clean up installed.json

Sentinel `ErrNotInstalled` defined here; shared with UpdateService.

### UpdateService (`internal/service/update.go`)

Implements D-08 auto-overwrite update:
- `Update(name, target)`: compares installed vs registry version; returns "already up to date" on match; calls `installer.Install()` for upgrade
- `UpdateAll(target)`: iterates all installed records; continues on single failure; returns first error + all successful messages
- Delegates actual install to `InstallService` which owns D-08 ownership check

### Commands

- `cmd/uninstall.go`: `ExactArgs(1)`, uses `errors.Is(err, service.ErrNotInstalled)` for D-04 error with `agentkit list` suggestion. Success: `✓ <name> uninstalled (<target>)`
- `cmd/update.go`: `MaximumNArgs(1)`, optional name (UpdateAll when absent). D-08 format: `⚠ updated <name>: <old> → <new>`

---

## Test Results

- 10 new tests in `internal/service/uninstall_test.go` + `update_test.go` — all pass (GREEN)
- Full suite: **65 tests pass**, `go vet ./...` clean

---

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed duplicate --target flag definition**
- **Found during:** Task 2
- **Issue:** Init functions in cmd/uninstall.go and cmd/update.go defined a local `--target` flag that conflicted with the persistent `--target` flag already defined on rootCmd.
- **Fix:** Removed duplicate `Flags().String("target", ...)` calls; both commands inherit the persistent global flag.
- **Files modified:** cmd/uninstall.go, cmd/update.go

---

## Known Stubs

None — all code paths are fully implemented.

---

## Threat Flags

None — no new network endpoints, auth paths, file access patterns, or schema changes beyond the threat model (T-05-01, T-05-02, T-05-03 all addressed by existing adapter and service layer).

---

## Self-Check: PASSED

- internal/service/uninstall.go: FOUND
- internal/service/uninstall_test.go: FOUND
- internal/service/update.go: FOUND
- internal/service/update_test.go: FOUND
- cmd/uninstall.go: FOUND (modified)
- cmd/update.go: FOUND (modified)
- All commits present: `git log --oneline -6` shows test(01-05), feat(01-05) x2
