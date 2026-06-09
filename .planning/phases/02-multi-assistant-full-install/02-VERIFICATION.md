---
phase: 02-multi-assistant-full-install
verified: 2026-06-09T06:30:00Z
status: passed
score: 2/2
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 1/2
  gaps_closed:
    - "cmd/install.go now calls adapter.NewAdapter(target, store) — all 7 adapters dispatched correctly from the CLI entry point"
  gaps_remaining: []
  regressions: []
---

# Phase 2: Multi-Assistant & Full Install — Verification Report

**Phase Goal:** Users can install skills and MCP servers targeting any of the 5 supported coding assistants using any of the 4 install methods.
**Verified:** 2026-06-09T06:30:00Z
**Status:** passed
**Re-verification:** Yes — after gap closure (commit: fix(02): wire NewAdapter factory in cmd/install.go)

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `agentkit install <name> --target <copilot-cli/copilot-vscode/codex/gemini/opencode/pi>` writes MCP config to the correct assistant path (runtime-detected, not hardcoded) | VERIFIED | `cmd/install.go:44` now calls `ad, err := adapter.NewAdapter(target, store)`. All 7 targets dispatched through factory.go. 126 tests pass. |
| 2 | `agentkit install <name>` using a uvx-based or Docker-based MCP server completes without error and produces a valid config entry | VERIFIED | `installer.NewInstaller` switch handles `InstallMethodUvx` and `InstallMethodDocker`. `UvxInstaller` and `DockerInstaller` fully implemented with injected-runner pattern; 22 installer tests pass. |

**Score:** 2/2 truths verified

---

## Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/domain/package.go` | InstallMethodUvx and InstallMethodDocker constants | VERIFIED | Lines 15-18: both constants defined |
| `internal/installer/uvx.go` | UvxInstaller + ErrUvxNotFound | VERIFIED | Full implementation with lookPathFunc/runFunc injection; 5 test cases pass |
| `internal/installer/docker.go` | DockerInstaller + ErrDockerNotFound | VERIFIED | Full implementation; "docker pull" eager pattern; 5 test cases pass |
| `internal/adapter/jsonbase.go` | jsonMCPAdapter base struct for mcpServers adapters | VERIFIED | WriteMCPConfig, RemoveMCPConfig, ReadMCPConfig methods fully implemented |
| `internal/config/paths.go` | SkillInstallPath with gemini, pi, and ErrNotSupported cases | VERIFIED | `case "pi"` returns `~/.agents/skills/<name>` |
| `internal/adapter/copilot_cli.go` | CopilotCLIAdapter with type:local + tools:["*"] | VERIFIED | mcpKey="mcpServers", extraFields injects type/tools; 7 test cases pass |
| `internal/adapter/copilot_vscode.go` | CopilotVSCodeAdapter with servers key + edition detection | VERIFIED | mcpKey="servers"; vsCodeEditions detected; 7 test cases pass |
| `internal/adapter/gemini.go` | GeminiAdapter with full WriteSkill | VERIFIED | configPath=~/.gemini/settings.json; WriteSkill to ~/.gemini/skills/<name>/; 8 test cases pass |
| `internal/adapter/pi.go` | PiAdapter with full WriteSkill | VERIFIED | configPath=~/.pi/agent/mcp.json; WriteSkill resolves to ~/.agents/skills/<name>/; 8 test cases pass |
| `internal/adapter/codex.go` | CodexAdapter via TOML | VERIFIED | Uses BurntSushi/toml; DecodeFile preserves non-mcp_servers keys; 8 test cases pass |
| `internal/adapter/opencode.go` | OpenCodeAdapter with mcp key + array command | VERIFIED | Uses "mcp" key (not "mcpServers"); command is JSON array; 8 test cases pass |
| `internal/adapter/factory.go` | NewAdapter factory for all 7 targets | VERIFIED | All 7 targets: claude, copilot-cli, copilot-vscode, gemini, pi, codex, opencode — factory.go lines 17-30 |
| `cmd/root.go` | Expanded --target validation (7 targets) | VERIFIED | validTargets = ["claude","copilot-cli","copilot-vscode","codex","gemini","opencode","pi"] |
| `cmd/install.go` | Calls NewAdapter(target, store) factory | VERIFIED | Line 44: `ad, err := adapter.NewAdapter(target, store)` — gap closed |

---

## Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `installer.NewInstaller()` | `UvxInstaller` | `case domain.InstallMethodUvx` | WIRED | installer.go line 44 |
| `installer.NewInstaller()` | `DockerInstaller` | `case domain.InstallMethodDocker` | WIRED | installer.go line 46 |
| `cmd/install.go --target` | `NewAdapter(target, store)` | cmd/install.go:44 | WIRED | Gap closed — `ad, err := adapter.NewAdapter(target, store)` |
| `NewAdapter("claude")` | `ClaudeCodeAdapter` | factory.go case "claude" | WIRED | factory.go line 18 |
| `NewAdapter("copilot-cli")` | `CopilotCLIAdapter` | factory.go case "copilot-cli" | WIRED | factory.go line 20 |
| `NewAdapter("copilot-vscode")` | `CopilotVSCodeAdapter` | factory.go case "copilot-vscode" | WIRED | factory.go line 22 |
| `NewAdapter("codex")` | `CodexAdapter` | factory.go case "codex" | WIRED | factory.go line 28 |
| `NewAdapter("gemini")` | `GeminiAdapter` | factory.go case "gemini" | WIRED | factory.go line 24 |
| `NewAdapter("opencode")` | `OpenCodeAdapter` | factory.go case "opencode" | WIRED | factory.go line 30 |
| `NewAdapter("pi")` | `PiAdapter` | factory.go case "pi" | WIRED | factory.go line 26 |

---

## Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| All 126 tests pass | `go test ./...` | 126 passed, 0 failed | PASS |
| `cmd/install.go` calls `NewAdapter(target, store)` | Read cmd/install.go:44 | `ad, err := adapter.NewAdapter(target, store)` | PASS |
| Factory dispatches all 7 targets | Read internal/adapter/factory.go | cases: claude, copilot-cli, copilot-vscode, gemini, pi, codex, opencode | PASS |

---

## Requirements Coverage

| Requirement | Phase Plan | Description | Status | Evidence |
|-------------|-----------|-------------|--------|----------|
| AST-02 | 02-02 | GitHub Copilot CLI adapter | SATISFIED | CopilotCLIAdapter and CopilotVSCodeAdapter implemented, tested, and wired through factory from CLI. |
| AST-03 | 02-04 | OpenAI Codex adapter (TOML) | SATISFIED | CodexAdapter with TOML; 8 tests pass; factory wired from CLI. |
| AST-04 | 02-03 | Gemini CLI adapter | SATISFIED | GeminiAdapter with WriteSkill; 8 tests pass; factory wired from CLI. |
| AST-05 | 02-04 | OpenCode adapter | SATISFIED | OpenCodeAdapter with mcp key + array command; 8 tests pass; factory wired from CLI. |
| AST-06 | 02-03 | Pi adapter | SATISFIED | PiAdapter with ~/.pi/agent/mcp.json; 8 tests pass; factory wired from CLI. |
| MCP-02 | 02-01 | pip/uvx install adapter | SATISFIED | UvxInstaller fully implemented; factory routes InstallMethodUvx; 5 tests pass. |
| MCP-04 | 02-01 | Docker adapter | SATISFIED | DockerInstaller eager-pulls; factory routes InstallMethodDocker; 5 tests pass. |
| REG-03 | — | mcpmarket.com registry | DESCOPED | ROADMAP.md note: "removed from v1 scope per D-01/D-02". Not a Phase 2 requirement. |
| REG-04 | — | Custom registry add | DESCOPED | ROADMAP.md note: "removed from v1 scope per D-01/D-02". Not a Phase 2 requirement. |

---

## Anti-Patterns Found

None. The previously identified blocker (hardcoded `NewClaudeCodeAdapter` in `cmd/install.go:44`) is resolved.

---

## Human Verification Required

None.

---

## Gaps Summary

No gaps. All Phase 2 deliverables are verified end-to-end:

- 7 adapters implemented, tested, and dispatched through the factory from `cmd/install.go`
- 4 install methods (npm/npx, pip/uvx, Docker, direct binary) routed through `installer.NewInstaller`
- 126 tests pass across all packages
- CLI `--target` flag correctly dispatches to the right adapter at runtime

---

_Verified: 2026-06-09T06:30:00Z_
_Verifier: Claude (gsd-verifier)_
