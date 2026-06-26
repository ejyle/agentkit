---
phase: 02-multi-assistant-full-install
plan: "03"
subsystem: adapter
tags: [adapter, gemini, pi, mcp, skills, tdd]
dependency_graph:
  requires:
    - 02-01 (jsonMCPAdapter base, SkillInstallPath extensions)
  provides:
    - GeminiAdapter implementing AssistantAdapter (WriteMCPConfig + WriteSkill)
    - PiAdapter implementing AssistantAdapter (WriteMCPConfig + WriteSkill)
  affects:
    - internal/adapter/ (new gemini.go, pi.go, gemini_test.go, pi_test.go)
tech_stack:
  added: []
  patterns:
    - Struct embedding of jsonMCPAdapter for shared read-merge-atomic-write logic
    - Injected homeDir for test isolation (avoids os.UserHomeDir() in tests)
    - TDD RED/GREEN cycle per-adapter
key_files:
  created:
    - internal/adapter/gemini.go
    - internal/adapter/gemini_test.go
    - internal/adapter/pi.go
    - internal/adapter/pi_test.go
decisions:
  - GeminiAdapter and PiAdapter use injected homeDir for WriteSkill to maintain test isolation
  - Pi skills resolve to ~/.agents/skills/ per D-11 confirmed by explicit negative assertion in test
  - No ErrNotSupported in either adapter; both MCP and skill operations fully implemented
metrics:
  duration: "~3 minutes"
  completed: "2026-06-09"
  tasks: 2
  files_created: 4
  files_modified: 0
  tests_added: 16
  total_tests: 96
---

# Phase 2 Plan 03: Gemini and Pi Adapter Vertical Slices Summary

GeminiAdapter and PiAdapter implemented via jsonMCPAdapter embedding, each writing plain command/args/env mcpServers entries (no "type" field) to their respective config paths, with full WriteSkill support.

## Tasks Completed

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 (RED) | Failing tests for GeminiAdapter | 723f8b4 | internal/adapter/gemini_test.go |
| 1 (GREEN) | GeminiAdapter implementation | 7b57ca6 | internal/adapter/gemini.go |
| 2 (RED) | Failing tests for PiAdapter | 7465712 | internal/adapter/pi_test.go |
| 2 (GREEN) | PiAdapter implementation | ae44e5e | internal/adapter/pi.go |

## Verification Results

- go test ./internal/adapter/... -run TestGemini: 8 tests pass
- go test ./internal/adapter/... -run TestPi: 8 tests pass
- go test ./...: 88 tests pass (up from 80)
- gemini.go: configPath returns ~/.gemini/settings.json; no "type" field in MCP entries
- pi.go: configPath returns ~/.pi/agent/mcp.json; WriteSkill writes to ~/.agents/skills/name
- Neither adapter uses ErrNotSupported; both operations fully implemented
- TestPiAdapter_WriteSkill_CreatesDirectory: negative assertion confirms file NOT at ~/.pi/skills/name

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] WriteSkill uses injected homeDir instead of config.SkillInstallPath**

- **Found during:** Task 1 GREEN phase
- **Issue:** config.SkillInstallPath calls os.UserHomeDir() internally, ignoring the injected homeDir. Tests verify files at tmpHome paths.
- **Fix:** Both adapters construct the skill path directly from a.home() plus hardcoded relative segments.
- **Files modified:** internal/adapter/gemini.go, internal/adapter/pi.go
- **Commits:** 7b57ca6, ae44e5e

## TDD Gate Compliance

- RED gate: 723f8b4 (test(02-03): add failing tests for GeminiAdapter)
- GREEN gate: 7b57ca6 (feat(02-03): implement GeminiAdapter)
- RED gate: 7465712 (test(02-03): add failing tests for PiAdapter)
- GREEN gate: ae44e5e (feat(02-03): implement PiAdapter)

## Known Stubs

None; all implementations are fully wired.

## Threat Surface Scan

No new network endpoints. Writes use renameio.WriteFile (atomic) and os.MkdirAll. T-02-07 and T-02-08 mitigated as designed.

## Self-Check: PASSED

- internal/adapter/gemini.go: FOUND
- internal/adapter/gemini_test.go: FOUND
- internal/adapter/pi.go: FOUND
- internal/adapter/pi_test.go: FOUND
- Commits 723f8b4, 7b57ca6, 7465712, ae44e5e: confirmed in git log
