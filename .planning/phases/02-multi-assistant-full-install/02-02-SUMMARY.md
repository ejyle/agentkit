---
phase: 02-multi-assistant-full-install
plan: "02"
subsystem: copilot-adapters
tags: [adapter, copilot, copilot-cli, copilot-vscode, mcp, tdd]
dependency_graph:
  requires:
    - 02-01 (jsonMCPAdapter base, ErrNotSupported, domain types)
  provides:
    - CopilotCLIAdapter implementing AssistantAdapter (mcpServers key, type:local, tools:[*])
    - CopilotVSCodeAdapter implementing AssistantAdapter (servers key, edition detection)
  affects:
    - internal/adapter/ (new copilot_cli.go, copilot_vscode.go)
tech_stack:
  added: []
  patterns:
    - jsonMCPAdapter struct embedding for Copilot adapters
    - COPILOT_HOME env var path override (T-02-04)
    - VS Code edition detection via runtime os.Stat (Code → Code-Insiders → code-server)
    - NewXxxWithConfigDir constructor for hermetic test injection
key_files:
  created:
    - internal/adapter/copilot_cli.go
    - internal/adapter/copilot_cli_test.go
    - internal/adapter/copilot_vscode.go
    - internal/adapter/copilot_vscode_test.go
decisions:
  - CopilotCLIAdapter uses extraFields to inject type:local and tools:[*] per Copilot CLI MCP format
  - CopilotVSCodeAdapter uses configDir injection (not homeDir) because VS Code config lives in os.UserConfigDir(), not home
  - vsCodeEditions slice ordered Code → Code-Insiders → code-server; first existing User/ wins
  - WriteSkill/RemoveSkill return wrapped ErrNotSupported for both adapters (no user-global skill dir)
metrics:
  duration: "~6 minutes"
  completed: "2026-06-09"
  tasks: 2
  files_created: 4
  files_modified: 0
  tests_added: 14
  total_tests: 94
---

# Phase 2 Plan 02: Copilot Adapters (CLI + VS Code) Summary

CopilotCLIAdapter writes mcpServers entries with type:local and tools:[*] to ~/.copilot/mcp-config.json; CopilotVSCodeAdapter writes under the servers key to VS Code's mcp.json with runtime edition detection.

## Tasks Completed

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 (RED) | Failing tests for CopilotCLIAdapter (7 cases) | 63a2cbd | copilot_cli_test.go |
| 1 (GREEN) | CopilotCLIAdapter implementation | 9bd9807 | copilot_cli.go |
| 2 (RED) | Failing tests for CopilotVSCodeAdapter (7 cases) | 5ddcf35 | copilot_vscode_test.go |
| 2 (GREEN) | CopilotVSCodeAdapter implementation | 4ae06ce | copilot_vscode.go |

## Verification Results

- `go test ./internal/adapter/... -run TestCopilotCLI` — 7 tests PASS
- `go test ./internal/adapter/... -run TestCopilotVSCode` — 7 tests PASS
- `go test ./...` — 94 tests PASS (no regressions from 80 pre-plan baseline)
- CopilotCLIAdapter.WriteMCPConfig injects "type":"local" and "tools":["*"] — CONFIRMED
- CopilotVSCodeAdapter uses mcpKey="servers" (not "mcpServers") — CONFIRMED
- CopilotVSCodeAdapter edition detection: Code → Code-Insiders → code-server via os.Stat — CONFIRMED
- COPILOT_HOME env var path override (T-02-04 mitigated) — CONFIRMED
- WriteSkill returns errors.Is(err, ErrNotSupported)==true for both adapters — CONFIRMED

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — all implementations are fully wired.

## Threat Surface Scan

No new network endpoints, auth paths, or trust boundary changes introduced.

- T-02-04 (COPILOT_HOME tampering): mitigated — filepath.Join treats env value as a directory path only; no shell execution.
- T-02-05 (foreign key clobber): mitigated — inherited from jsonMCPAdapter conflict check; ErrForeignConflict returned for non-owned keys.
- T-02-06 (edition detection information disclosure): accepted — read-only stat(); defaults to Code path safely.

## Self-Check: PASSED

- internal/adapter/copilot_cli.go: FOUND
- internal/adapter/copilot_cli_test.go: FOUND
- internal/adapter/copilot_vscode.go: FOUND
- internal/adapter/copilot_vscode_test.go: FOUND
- Commits 63a2cbd, 9bd9807, 5ddcf35, 4ae06ce: FOUND in git log
