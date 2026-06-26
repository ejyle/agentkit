---
phase: 04-distribution-hardening
plan: "01"
subsystem: version
tags: [version, ldflags, goreleaser, cobra]
dependency_graph:
  requires: []
  provides: [internal/version package, agentkit --version command]
  affects: [cmd/root.go, .goreleaser.yaml (future)]
tech_stack:
  added: []
  patterns: [ldflags injection, cobra Version field, SetVersionTemplate]
key_files:
  created:
    - internal/version/version.go
  modified:
    - cmd/root.go
decisions:
  - "No runtime.GOOS/runtime.GOARCH usage — vars set purely by GoReleaser ldflags at build time; defaults serve local builds"
  - "SetVersionTemplate suppresses Cobra's default 'appname version X.Y.Z' prefix so --version outputs bare format string"
metrics:
  duration: ~5min
  completed: "2026-06-09"
requirements_satisfied: [CLI-10]
---

# Phase 4 Plan 01: Version Injection Summary

ldflags-injectable version package wiring `agentkit --version` to output `agentkit/dev (unknown/unknown)` locally and `agentkit/0.1.0 (darwin/arm64)` format when built by GoReleaser.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create internal/version package | d75c079 | internal/version/version.go |
| 2 | Wire version into cmd/root.go | a9ec554 | cmd/root.go |

## What Was Built

**`internal/version/version.go`**: New package declaring three ldflags-injectable vars (`Version="dev"`, `GOOS="unknown"`, `GOARCH="unknown"`) and a `String()` function returning `agentkit/<version> (<os>/<arch>)`. No runtime imports — values are set purely at build time.

**`cmd/root.go`**: Added import for the version package, set `rootCmd.Version = version.String()`, and added `rootCmd.SetVersionTemplate("{{.Version}}\n")` in `init()` to suppress Cobra's default version prefix.

## Verification Results

```
$ agentkit --version
agentkit/dev (unknown/unknown)

$ agentkit -v
agentkit/dev (unknown/unknown)

$ go vet ./...
(no warnings)
```

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — version vars default to "dev"/"unknown" intentionally for local builds. GoReleaser will inject real values via ldflags.

## Threat Flags

None — no new network endpoints, auth paths, or trust boundary changes introduced.

## Self-Check: PASSED

- internal/version/version.go: FOUND
- cmd/root.go: FOUND (modified)
- Commit d75c079: FOUND
- Commit a9ec554: FOUND
