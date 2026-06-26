---
phase: 02-multi-assistant-full-install
plan: "01"
subsystem: installer-foundation
tags: [installer, adapter, domain, uvx, docker, paths, targets]
dependency_graph:
  requires: []
  provides:
    - InstallMethodUvx and InstallMethodDocker domain constants
    - UvxInstaller and DockerInstaller with injected-runner pattern
    - jsonMCPAdapter shared base for all mcpServers adapters
    - SkillInstallPath support for gemini and pi targets
    - validTargets expansion (7 targets including copilot-cli, copilot-vscode, pi)
  affects:
    - internal/installer/ (new uvx.go, docker.go)
    - internal/adapter/ (new jsonbase.go, adapter.go ErrNotSupported)
    - internal/config/paths.go (path routing)
    - cmd/root.go (CLI flag validation)
tech_stack:
  added: []
  patterns:
    - Injected lookPathFunc/runFunc for testable subprocess invocations
    - exec.Command arg-array form only (no shell string interpolation)
    - Shared jsonMCPAdapter base via struct embedding
    - Atomic writes via renameio + post-install verify
key_files:
  created:
    - internal/installer/uvx.go
    - internal/installer/docker.go
    - internal/installer/uvx_test.go
    - internal/installer/docker_test.go
    - internal/adapter/jsonbase.go
  modified:
    - internal/domain/package.go
    - internal/installer/installer.go
    - internal/adapter/adapter.go
    - internal/config/paths.go
    - cmd/root.go
decisions:
  - NewUvxInstallerWithRunner and NewDockerInstallerWithRunner use stub lookPath (always succeeds) so tests do not require uvx/docker installed on CI
  - extractEntryFromRaw handles both string command and []interface{} array command (OpenCode compatibility)
  - pi target resolves to ~/.agents/skills/<name>, not ~/.pi/skills/<name>
  - copilot-cli, copilot-vscode, codex, opencode return ErrFormat from SkillInstallPath (no user-global skill dir)
metrics:
  duration: "~4 minutes"
  completed: "2026-06-09"
  tasks: 2
  files_created: 5
  files_modified: 5
  tests_added: 9
  total_tests: 80
---

# Phase 2 Plan 01: Multi-Assistant Foundation Infrastructure Summary

Wave 1 foundation infrastructure: UvxInstaller + DockerInstaller with injected runner pattern, jsonMCPAdapter shared base extracting claude.go's read-merge-atomic-write loop, SkillInstallPath extensions for gemini/pi, and --target flag expansion to 7 assistants.

## Tasks Completed

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 (RED) | Failing tests + domain constants | ec7d4a7 | domain/package.go, uvx_test.go, docker_test.go |
| 1 (GREEN) | UvxInstaller + DockerInstaller implementation | da16233 | uvx.go, docker.go, installer.go |
| 2 | jsonMCPAdapter base + paths + targets | 84dac08 | jsonbase.go, adapter.go, paths.go, root.go |

## Verification Results

- `go build ./...` — exits 0, zero errors
- `go test ./internal/installer/... -run "TestUvx|TestDocker"` — 9 tests pass
- `go test ./...` — 80 tests pass (no Phase 1 regressions)
- UvxInstaller.Install uses exec.Command arg-array form (T-02-01 mitigated)
- DockerInstaller.Install runs "docker pull" eagerly (D-09)
- SkillInstallPath("pi", "x") returns `~/.agents/skills/x` (not `~/.pi/skills/x`)
- cmd/root.go accepts copilot-cli, copilot-vscode, codex, gemini, opencode, pi

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] NewXxxWithRunner constructors use stub lookPath**

- **Found during:** Task 1 GREEN phase
- **Issue:** `NewDockerInstallerWithRunner` used real `exec.LookPath` so `TestDockerInstaller_Install_PullsEagerly` failed on machines without Docker installed
- **Fix:** Both `NewUvxInstallerWithRunner` and `NewDockerInstallerWithRunner` inject a stub lookPath that returns success, making tests hermetic
- **Files modified:** internal/installer/uvx.go, internal/installer/docker.go
- **Commit:** da16233

## Known Stubs

None — all implementations are fully wired.

## Threat Surface Scan

No new network endpoints or trust boundary changes introduced. All subprocess invocations use exec.Command arg-array form per T-02-01.

## Self-Check: PASSED

- internal/installer/uvx.go: FOUND
- internal/installer/docker.go: FOUND
- internal/installer/uvx_test.go: FOUND
- internal/installer/docker_test.go: FOUND
- internal/adapter/jsonbase.go: FOUND
- internal/domain/package.go contains InstallMethodUvx: CONFIRMED
- internal/config/paths.go contains case "pi": CONFIRMED
- cmd/root.go validTargets contains "copilot-cli": CONFIRMED
- Commits ec7d4a7, da16233, 84dac08: FOUND in git log
