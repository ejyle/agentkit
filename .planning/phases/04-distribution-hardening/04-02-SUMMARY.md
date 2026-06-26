---
phase: 04-distribution-hardening
plan: "02"
subsystem: cli-doctor
tags: [doctor, diagnostics, environment-check, cli]
dependency_graph:
  requires: []
  provides: [agentkit-doctor-command]
  affects: [cmd/root.go]
tech_stack:
  added: []
  patterns: [cobra-subcommand, context-timeout-http, stdlib-only-checks]
key_files:
  created:
    - cmd/doctor.go
  modified: []
decisions:
  - "PersistentPreRunE bypassed on doctorCmd to skip --target flag validation (doctor has no target)"
  - "Assistant dir checks use warn not fail — user may not have all assistants installed"
  - "Runtime dep checks (node/docker/uvx) use warn not fail per D-08 spec"
  - "Registry check uses context.WithTimeout(5s) with HEAD request — never http.Get() (T-04-02-02)"
  - "printCheckResult sends fail lines to stderr, pass/warn to stdout"
metrics:
  duration: "~8 minutes"
  completed: "2026-06-09"
  tasks_completed: 1
  tasks_total: 1
  files_created: 1
  files_modified: 0
requirements:
  - CLI-10
---

# Phase 4 Plan 2: Doctor Command Summary

**One-liner:** `agentkit doctor` runs 9 environment checks (PATH, config dir, registry, 5 assistant dirs, 3 runtime deps) with brew-doctor-style ✓/⚠/✗ output.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Implement doctor command with all 9 checks | 0860e11 | cmd/doctor.go |

## What Was Built

`cmd/doctor.go` implements the `agentkit doctor` subcommand with:

- **CheckResult struct** — Label, Status (pass/warn/fail), Message, Hint fields
- **9 checks** across 5 check functions:
  1. `checkBinaryInPath()` — exec.LookPath("agentkit"), fail if missing
  2. `checkConfigDirWritable()` — os.MkdirAll + write-test temp file, fail if unwritable
  3. `checkRegistryReachable()` — HEAD to GitHub raw URL with 5s context timeout, fail on error
  4. `checkAssistantDirs()` (5 results) — os.Stat for ~/.claude/, ~/.gemini/, ~/.copilot/, ~/.codex/, ~/.config/opencode/, warn (not fail) if missing
  5. `checkRuntimeDeps()` (3 results) — exec.LookPath for node/docker/uvx, warn (not fail) if missing
- **Exit behavior** — returns error "one or more checks failed" when any Status=="fail", causing Cobra to set exit code 1; exit 0 on all pass/warn
- **PersistentPreRunE bypass** — doctorCmd.PersistentPreRunE = no-op, avoiding --target validation (doctor takes no target flag)

## Verification Results

```
✓ agentkit in PATH (/tmp/agentkit)     [pass — when binary is named agentkit]
✓ ~/.agentkit/ writable                [pass]
✓ registry reachable (agentkit-registry) [pass]
✓ ~/.claude/ exists                    [pass]
✓ ~/.gemini/ exists                    [pass]
⚠ ~/.copilot/ — not found — Copilot CLI not installed  [warn]
✓ ~/.codex/ exists                     [pass]
✓ ~/.config/opencode/ exists           [pass]
✓ node available                       [pass]
⚠ docker — not found — Docker-based MCPs won't install [warn]
   → Install: https://docs.docker.com/get-docker/
✓ uvx available                        [pass]
exit: 0
```

`go vet ./cmd/...` — no issues.
`agentkit --help | grep doctor` — "doctor" appears in command list.

## Deviations from Plan

None — plan executed exactly as written.

## Threat Surface Scan

No new threat surface beyond what the plan's `<threat_model>` already covers:
- T-04-02-01: ~/.agentkit/.write-test created and immediately removed
- T-04-02-02: context.WithTimeout(5s) hard-caps registry check (implemented)
- T-04-02-03: Only binary/dir existence reported, no secrets printed
- T-04-02-04: os.MkdirAll with 0755, user-owned dir only

## Known Stubs

None.

## Self-Check: PASSED

- cmd/doctor.go: EXISTS
- Commit 0860e11: EXISTS
- go build ./...: SUCCESS
- agentkit doctor runs all 9 checks: VERIFIED
- Exit 0 on pass/warn, exit 1 on fail: VERIFIED
- PersistentPreRunE bypassed: VERIFIED
- Registry check uses 5s timeout: VERIFIED (context.WithTimeout)
