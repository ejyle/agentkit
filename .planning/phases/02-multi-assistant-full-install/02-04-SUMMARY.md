---
phase: 02-multi-assistant-full-install
plan: "04"
subsystem: adapter-codex-opencode-factory
tags: [adapter, codex, opencode, toml, factory, mcp]
dependency_graph:
  requires:
    - 02-01 (jsonMCPAdapter base, ErrNotSupported, domain types)
  provides:
    - CodexAdapter — TOML-based MCP config at ~/.codex/config.toml
    - OpenCodeAdapter — JSON with "mcp" key, array command, "environment" key
    - NewAdapter factory dispatching all 7 targets
  affects:
    - internal/adapter/codex.go (new)
    - internal/adapter/opencode.go (new)
    - internal/adapter/factory.go (new)
    - internal/adapter/copilot_cli.go (scaffold stub)
    - internal/adapter/copilot_vscode.go (scaffold stub)
    - internal/adapter/gemini.go (scaffold stub)
    - internal/adapter/pi.go (scaffold stub)
tech_stack:
  added: []
  patterns:
    - BurntSushi/toml DecodeFile into map[string]interface{} preserves unknown TOML keys
    - toml.NewEncoder + bytes.Buffer for re-encode without full round-trip loss
    - OpenCode command-as-array: append([]string{cmd}, args...) — never a shell string
    - "environment" key (not "env") for OpenCode env vars
    - Scaffold stubs for wave-2 parallel plans so factory compiles now
key_files:
  created:
    - internal/adapter/codex.go
    - internal/adapter/opencode.go
    - internal/adapter/factory.go
    - internal/adapter/copilot_cli.go (scaffold)
    - internal/adapter/copilot_vscode.go (scaffold)
    - internal/adapter/gemini.go (scaffold)
    - internal/adapter/pi.go (scaffold)
  modified: []
decisions:
  - CodexAdapter does NOT embed jsonMCPAdapter — TOML format is incompatible with JSON base
  - OpenCodeAdapter does NOT embed jsonMCPAdapter — different schema would require overriding everything
  - Scaffold stubs for copilot_cli/vscode, gemini, pi allow factory.go to compile in this wave; full impls in 02-02/02-03
  - factory.go default error names all 7 valid targets for self-diagnosing CLI users
  - Env sub-table in TOML stored as map[string]interface{} to render as [mcp_servers.<name>.env] section
metrics:
  duration: "~12 minutes"
  completed: "2026-06-09"
  tasks: 2
  files_created: 7
  files_modified: 0
  tests_added: 0
---

# Phase 02 Plan 04: Codex Adapter, OpenCode Adapter, and NewAdapter Factory Summary

CodexAdapter (TOML), OpenCodeAdapter (distinct JSON schema), and NewAdapter factory wiring all 7 targets — completing AST-03 and AST-05 requirements.

## What Was Built

**CodexAdapter** (`internal/adapter/codex.go`):
- Reads/writes `~/.codex/config.toml` using `BurntSushi/toml` — the only TOML adapter in the stack
- `readRawConfig` decodes into `map[string]interface{}` preserving all non-`mcp_servers` TOML keys
- `writeRawConfig` re-encodes via `toml.NewEncoder` + `bytes.Buffer` → `renameio.WriteFile`
- Env vars written as a nested sub-table: `[mcp_servers.<name>.env]`
- `WriteSkill`/`RemoveSkill` return `ErrNotSupported` (no user-global skill directory)

**OpenCodeAdapter** (`internal/adapter/opencode.go`):
- Reads/writes `<UserConfigDir>/opencode/opencode.json`
- Top-level key is `"mcp"` (not `"mcpServers"`) — OpenCode-specific schema
- `command` field is a JSON array: `append([]string{entry.Command}, entry.Args...)` (T-02-11)
- Env key is `"environment"` (not `"env"`) (T-02-12)
- `ReadMCPConfig` splits array: `arr[0]` = Command, `arr[1:]` = Args
- `WriteSkill`/`RemoveSkill` return `ErrNotSupported`

**NewAdapter factory** (`internal/adapter/factory.go`):
- `NewAdapter(target, store)` dispatches all 7 targets: claude, copilot-cli, copilot-vscode, gemini, pi, codex, opencode
- Default case returns descriptive error listing all valid target names

**Scaffold stubs** (copilot_cli.go, copilot_vscode.go, gemini.go, pi.go):
- Minimal `AssistantAdapter` implementations that return `ErrNotSupported`
- Exist so `factory.go` compiles before plans 02-02 and 02-03 run
- Will be replaced by full implementations in those plans

## Test Results

- `go test ./internal/adapter/... -run "TestCodex"` — 8 cases PASS
- `go test ./internal/adapter/... -run "TestOpenCode"` — 8 cases PASS
- `go test ./...` — 96 tests PASS (full suite green, no regressions)

## Deviations from Plan

### Auto-added scaffold stubs (Rule 2 — missing critical functionality)

**Found during:** Task 2 (factory.go)
**Issue:** Plans 02-02 and 02-03 (copilot-cli, copilot-vscode, gemini, pi adapters) are parallel wave-2 plans that hadn't run yet. Factory.go references all 7 adapters but only claude, codex, opencode exist.
**Fix:** Created minimal scaffold stubs for copilot_cli.go, copilot_vscode.go, gemini.go, pi.go — each satisfies `AssistantAdapter` and returns `ErrNotSupported`. Plans 02-02/02-03 will replace these with full implementations.
**Files modified:** internal/adapter/copilot_cli.go, copilot_vscode.go, gemini.go, pi.go

## Known Stubs

| File | Description |
|------|-------------|
| internal/adapter/copilot_cli.go | All methods return ErrNotSupported — full impl in 02-02 |
| internal/adapter/copilot_vscode.go | All methods return ErrNotSupported — full impl in 02-02 |
| internal/adapter/gemini.go | All methods return ErrNotSupported — full impl in 02-03 |
| internal/adapter/pi.go | All methods return ErrNotSupported — full impl in 02-03 |

These stubs are intentional scaffolding. They will be overwritten when plans 02-02 and 02-03 execute.

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns beyond what the plan's threat model covers. All writes remain inside user home or config directories. No new trust boundaries introduced.

## Self-Check: PASSED

- internal/adapter/codex.go: EXISTS
- internal/adapter/opencode.go: EXISTS
- internal/adapter/factory.go: EXISTS
- Commits: codex.go, opencode.go, factory.go+stubs all committed on worktree-agent-afd5b1e104d922f60
- go test ./... exits 0 with 96 tests
