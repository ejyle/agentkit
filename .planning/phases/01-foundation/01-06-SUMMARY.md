---
plan: 01-06
phase: 01-foundation
status: complete
completed_at: 2026-06-08
---

## Summary

Human verification checkpoint for Phase 1 walking skeleton — confirmed all 5 CLI commands
work end-to-end against real infrastructure.

## What Was Verified

All 11 acceptance criteria steps passed:

1. `go build -o agentkit .` — exits 0
2. `agentkit install playwright --target claude` — `✓ playwright@1.40.0 installed → ~/.claude.json#mcpServers.playwright (claude)`
3. `~/.claude.json` valid JSON, `mcpServers.playwright` present with `command: npx`
4. `~/Library/Application Support/agentkit/claude/installed.json` has D-11 schema fields
5. `agentkit list` — table with playwright row (PACKAGE VERSION TYPE TARGET REGISTRY)
6. `agentkit search playwright` — ranked result from local registry
7. `agentkit update playwright` — `✓ playwright: already up to date`
8. `agentkit uninstall playwright --target claude` — `✓ playwright uninstalled (claude)`
9. `~/.claude.json` after uninstall — playwright key absent, other keys unchanged
10. `agentkit list` — `No packages installed for target: claude`
11. `go test ./... -count=1` — all 8 packages pass

## Bugs Found and Fixed

Three bugs discovered during verification:

1. **TTY deadlock** (`cmd/search.go`, `cmd/install.go`) — bubbletea crashed with
   "could not open a new TTY" in non-TTY environments, causing goroutine deadlock.
   Fixed: added `ui.IsTerminal()` check; non-TTY paths run synchronously.

2. **Nil NpxInstaller** (`internal/installer/installer.go`) — `NewInstaller()` returned
   `&NpxInstaller{}` (zero struct) instead of `NewNpxInstaller()`, leaving `lookPath`
   field nil and causing a panic on first install.

3. **Missing ejyle/agentkit-registry** — the primary registry repo does not yet exist.
   Fixed: added `AGENTKIT_REGISTRY_FILE` env override and `LocalFileRegistry`; bundled
   `testdata/registry.json` with playwright and context7 entries for development.

## Deviations

- macOS installed.json path is `~/Library/Application Support/agentkit/claude/installed.json`
  (not `~/.config/agentkit/...` as stated in the plan). This is correct — `os.UserConfigDir()`
  returns `~/Library/Application Support` on macOS per XDG-equivalent conventions.
- Verification used `AGENTKIT_REGISTRY_FILE=testdata/registry.json` because
  `ejyle/agentkit-registry` does not exist yet. Creating this repo is a follow-up task.

## Self-Check: PASSED

All must_haves confirmed:
- [x] agentkit install playwright resolves, installs, writes ~/.claude.json, records installed.json
- [x] agentkit list shows playwright
- [x] agentkit search playwright returns results
- [x] agentkit uninstall removes entry, no leftover artifacts
- [x] agentkit update shows up-to-date status
- [x] ~/.claude.json is valid JSON after every operation
- [x] No other keys in ~/.claude.json were modified

## Key Files

- `internal/ui/tty.go` — new: IsTerminal() helper
- `internal/registry/local.go` — new: LocalFileRegistry for dev/testing
- `testdata/registry.json` — new: bootstrap registry with playwright + context7
- `internal/installer/installer.go` — fixed: NewInstaller uses NewNpxInstaller()
- `cmd/search.go`, `cmd/install.go` — fixed: non-TTY synchronous path
