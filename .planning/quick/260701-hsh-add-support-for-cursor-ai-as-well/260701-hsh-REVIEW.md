---
phase: 260701-hsh-add-support-for-cursor-ai-as-well
reviewed: 2026-07-02T00:00:00Z
depth: quick
files_reviewed: 5
files_reviewed_list:
  - README.md
  - cmd/doctor.go
  - cmd/root.go
  - internal/adapter/cursor.go
  - internal/adapter/cursor_test.go
  - internal/adapter/factory.go
findings:
  critical: 0
  warning: 2
  info: 3
  total: 5
status: issues_found
---

# Phase 260701-hsh-add-support-for-cursor-ai-as-well: Code Review Report

**Reviewed:** 2026-07-02T00:00:00Z
**Depth:** quick
**Files Reviewed:** 6
**Status:** issues_found

## Summary

Reviewed the Cursor adapter addition (`internal/adapter/cursor.go`, its test file, `factory.go` wiring) plus the two command files touched to register the new target (`cmd/root.go`, `cmd/doctor.go`) and the README updates. `cursor.go` is structurally a faithful clone of the existing `GeminiAdapter`/`PiAdapter` pattern (embeds `jsonMCPAdapter`, no `type` field in MCP entries, standard `SKILL.md` folder layout) and is wired correctly into `factory.go` and `cmd/root.go`'s `validTargets`. Build, `go vet`, and the adapter test suite (62 tests) all pass. No security issues, hardcoded secrets, or dangerous functions were found via pattern scan.

The two real defects found are both formatting regressions introduced by this change (`gofmt` non-compliance in the two `cmd/*.go` files that were edited to add "cursor" support), which will fail CI on any repo that gates on `gofmt -l`. No critical/security findings.

## Warnings

### WR-01: cmd/root.go fails gofmt — misaligned struct field verticals

**File:** `cmd/root.go:10-16`
**Issue:** Editing this file (to append `"cursor"` handling) left the `cobra.Command` struct literal gofmt-dirty. `gofmt -l` flags this file. Field alignment for `Use`/`Short`/`Long`/`Version` no longer matches canonical gofmt output because the `Long` field's multi-line raw string desynced the aligner:
```go
var rootCmd = &cobra.Command{
	Use:     "agentkit",
	Short:   "AI agent skill and MCP server manager",
	Long:    `agentkit installs, updates, and manages AI agent skills, MCP servers, and
coding agents across all major AI coding assistants.`,
	Version: version.String(),
}
```
**Fix:** Run `gofmt -w cmd/root.go`. Canonical form removes the extra alignment spaces on `Use`, `Short`, `Long`:
```go
var rootCmd = &cobra.Command{
	Use:   "agentkit",
	Short: "AI agent skill and MCP server manager",
	Long: `agentkit installs, updates, and manages AI agent skills, MCP servers, and
coding agents across all major AI coding assistants.`,
	Version: version.String(),
}
```

### WR-02: cmd/doctor.go fails gofmt — misaligned struct literals in two functions

**File:** `cmd/doctor.go:171-175`, `cmd/doctor.go:237-243`
**Issue:** `gofmt -l` also flags this file (touched to add the Cursor entry to `checkAssistantDirs`). Two struct literals have inconsistent field-name alignment:
- Line ~171: the error-path `CheckResult{{Label: ..., Status: ..., Message: ...}}` literal has `Label`/`Status` misaligned relative to `Message`.
- Line ~237: the `dep` struct type (`binary`, `passLabel`, `failLabel`, `failMsg`, `hint`) has inconsistent column alignment.

Any CI step running `gofmt -l .` or `test -z "$(gofmt -l .)"` will fail on this PR.
**Fix:** Run `gofmt -w cmd/doctor.go` (or `go fmt ./...` for the whole repo before committing).

## Info

### IN-01: README "Homebrew (coming soon)" / "Scoop (coming soon)" sections not affected by this change but Cursor is missing from any package-manager-tap install instructions

**File:** `README.md:38-49`
**Issue:** Not a regression from this change, but worth flagging since the README was touched in this diff to add Cursor to the feature list and target list — the Homebrew/Scoop sections still say "coming soon" while Quick Start / Usage sections present fully-working commands, which is a minor inconsistency in polish but not a functional bug.
**Fix:** No action required for this PR; consider removing "(coming soon)" once Homebrew/Scoop taps ship, or note it's tracked in a follow-up issue.

### IN-02: `factory.go` default-case error message duplicates target list that must be kept in sync with `cmd/root.go`'s `validTargets` and `README.md`'s two "Supported targets" mentions

**File:** `internal/adapter/factory.go:34`, `cmd/root.go:18,44`, `README.md:6,13,85`
**Issue:** The list of supported target names ("claude, copilot-cli, copilot-vscode, gemini, pi, codex, opencode, cursor") is now duplicated across at least 4 locations (factory.go error string + doc comment, root.go `validTargets` slice + error string, README feature bullet, README "Supported targets" line). This PR correctly updated all of them for "cursor", but the duplication is a latent maintenance risk — a future added/removed target is easy to miss in one of the four spots (e.g., `validTargets` order differs slightly from `factory.go`'s switch order, which differs again from the README bullet order: "gemini, pi" vs "codex, gemini, opencode, pi"). This does not currently cause a bug, but increases the chance of a silent doc/behavior drift next time a target is added.
**Fix:** Consider defining a single source of truth (e.g., an exported `adapter.SupportedTargets []string` in `factory.go`) and have `cmd/root.go`'s `validTargets` and error message derive from it, so there's one list to update.

### IN-03: `checkAssistantDirs` in doctor.go silently ignores `os.UserHomeDir()` failure path formatting

**File:** `cmd/doctor.go:169-176`
**Issue:** Minor: the early-return `CheckResult` literal on `os.UserHomeDir()` error omits `Hint` (unlike other fail-path results in this file), and is the same literal flagged by WR-02 for gofmt. Not a functional bug — just an inconsistency with the rest of the file's fail-result conventions (most fail results set a `Hint`).
**Fix:** Optionally add a `Hint` (e.g., `"Ensure $HOME is set"`) for consistency with other checks in the file.

---

_Reviewed: 2026-07-02T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: quick_
