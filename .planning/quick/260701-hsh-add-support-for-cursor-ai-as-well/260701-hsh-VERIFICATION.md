---
phase: 260701-hsh-add-support-for-cursor-ai-as-well
verified: 2026-07-02T00:00:00Z
status: passed
score: 4/4 must-haves verified
behavior_unverified: 0
overrides_applied: 0
---

# Quick Task 260701-hsh: Add Cursor AI Support Verification Report

**Task Goal:** Add support for Cursor AI as well — extend agentkit's multi-assistant support so `agentkit install <pkg> --target cursor` writes both MCP server config (`~/.cursor/mcp.json`) and skill folders (`~/.cursor/skills/<name>/`), identical in shape to Claude Code/Gemini/pi.

**Verified:** 2026-07-02
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Running `agentkit install <pkg> --target cursor` writes an MCP server entry to `~/.cursor/mcp.json` under the `mcpServers` key with no extra `type` field | VERIFIED | `internal/adapter/cursor.go:29-41` sets `configPath` to `filepath.Join(home, ".cursor", "mcp.json")`, `mcpKey: "mcpServers"`, `extraFields: nil`. `TestCursorAdapter_WriteMCPConfig_CreatesFile` explicitly asserts no `type` field is present and passes (`go test ./internal/adapter/... -run TestCursorAdapter -v` → 8/8 pass). |
| 2 | Running `agentkit install <skill> --target cursor` writes a standard SKILL.md-based skill folder to `~/.cursor/skills/<name>/`, identical in shape to Claude/Gemini/pi skill folders | VERIFIED | `internal/adapter/cursor.go:46-63` `WriteSkill` is byte-for-byte structurally identical to `internal/adapter/gemini.go:44-61` (only base path differs: `.cursor/skills` vs `.gemini/skills`). `TestCursorAdapter_WriteSkill_CreatesDirectory` and `TestCursorAdapter_RemoveSkill` pass. |
| 3 | Passing `--target cursor` no longer fails root command validation | VERIFIED | `cmd/root.go:18` — `validTargets` includes `"cursor"`; help string (line 30) and error message (line 44) both updated. Live binary check: `agentkit --target cursor --help` exits 0 with no "invalid target" error. |
| 4 | `agentkit doctor` reports whether `~/.cursor/` exists, matching the pattern used for other assistants | VERIFIED | `cmd/doctor.go:205-209` adds `{dirLabel: "~/.cursor/", path: filepath.Join(home, ".cursor"), descName: "Cursor"}` to the `entries` slice used identically for Claude/Gemini/Copilot/Codex/OpenCode, placed after Codex and before OpenCode as specified in the plan. |

**Score:** 4/4 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/adapter/cursor.go` | New `CursorAdapter` embedding `jsonMCPAdapter`, with `WriteSkill`/`RemoveSkill` | VERIFIED | Exists, 74 lines, structurally mirrors `gemini.go` exactly with Cursor-specific paths. Doc comment present as specified in plan. |
| `internal/adapter/cursor_test.go` | 8 tests mirroring `gemini_test.go` | VERIFIED | Exists, 270 lines, all 8 named tests present exactly as specified in PLAN.md (`TestCursorAdapter_WriteMCPConfig_CreatesFile`, `_PreservesExistingKeys`, `_ErrForeignConflict`, `_AutoOverwrite`, `TestCursorAdapter_ReadMCPConfig_AfterWrite`, `TestCursorAdapter_RemoveMCPConfig`, `TestCursorAdapter_WriteSkill_CreatesDirectory`, `TestCursorAdapter_RemoveSkill`). All 8 pass. |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `cmd/root.go` validTargets | `internal/adapter/factory.go` NewAdapter | `"cursor"` string routed through switch | VERIFIED | `factory.go:31-32` — `case "cursor": return NewCursorAdapter(store), nil`. Doc comment (line 11) and default error message (line 34) both list `cursor`. |
| `internal/adapter/factory.go` NewCursorAdapter | `internal/adapter/cursor.go` CursorAdapter | Constructor delegation | VERIFIED | `NewCursorAdapter` delegates to `NewCursorAdapterWithHome(store, "")`, which builds the `jsonMCPAdapter` with `configPath` → `~/.cursor/mcp.json` and defines `WriteSkill`/`RemoveSkill` targeting `~/.cursor/skills/<name>/`. |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| CursorAdapter unit tests | `go test ./internal/adapter/... -run TestCursorAdapter -v` | 8 passed in 1 package | PASS |
| No regressions in cmd/adapter packages | `go test ./cmd/... ./internal/adapter/...` | 62 passed in 2 packages | PASS |
| Full project build | `go build ./...` | Success | PASS |
| Full project vet | `go vet ./...` | No issues found | PASS |
| Full project test suite (run once) | `go test ./...` | 149 passed in 13 packages | PASS |
| gofmt clean on modified files | `gofmt -l cmd/root.go cmd/doctor.go internal/adapter/cursor.go internal/adapter/cursor_test.go internal/adapter/factory.go` | No output (clean) | PASS — confirms follow-up commit 479b8dd's gofmt fix landed correctly |
| CLI validation accepts cursor target | `agentkit --target cursor --help` | Exit 0, no invalid-target error, help text lists `cursor` | PASS |
| README documents cursor | `grep -in cursor README.md` | 3 matches (lines 6, 13, 85) — intro paragraph, features bullet, Supported targets line | PASS |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | None found | — | Scanned `cursor.go` and `cursor_test.go` for TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER/stub markers — none present. No empty implementations, no hardcoded stubs. |

### Requirements Coverage

No formal `requirements` IDs declared in PLAN.md frontmatter (`requirements: []`) — this is a quick task without REQUIREMENTS.md linkage. Not applicable.

### Human Verification Required

None. The plan's `checkpoint:human-verify` task was substituted by the executor running the same verification commands directly (documented in SUMMARY.md's "Issues Encountered" section) since no human was present during the dispatch. This verifier independently re-ran all of those commands (build, targeted tests, full package tests, CLI `--help` invocation, README grep) from a fresh process and confirmed identical passing results — the substitution did not leave any gap requiring human judgment for this quick task.

### Gaps Summary

No gaps found. All 4 must-have truths are verified against live code and passing tests, not merely SUMMARY.md narrative:

- `cursor.go` is a faithful structural mirror of the proven `gemini.go` pattern (embedded `jsonMCPAdapter`, no extra `type` field, standard SKILL.md folder writes).
- All 8 planned unit tests exist verbatim and pass.
- Wiring is confirmed at every hop: `cmd/root.go` validTargets → `factory.go` switch → `cursor.go` constructors — verified by reading each file's actual content, not just grep counts.
- `cmd/doctor.go` has the Cursor entry in the correct position with the same struct shape as every other assistant.
- README has the 3 claimed cursor references at the claimed locations.
- The orchestrator's follow-up gofmt-only commit (479b8dd) is confirmed to have introduced no logic changes (`git show --stat` shows only whitespace/alignment diffs in `cmd/doctor.go` and `cmd/root.go`), and `gofmt -l` now reports clean on all modified files.
- Full test suite (149 tests, 13 packages) passes with zero failures — no regressions to any existing adapter (claude/gemini/copilot/codex/opencode/pi).

---

_Verified: 2026-07-02_
_Verifier: Claude (gsd-verifier)_
