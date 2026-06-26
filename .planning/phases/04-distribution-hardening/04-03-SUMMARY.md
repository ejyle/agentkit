---
phase: "04"
plan: "03"
subsystem: distribution
tags: [goreleaser, release-pipeline, homebrew, cosign, cross-platform]
dependency_graph:
  requires: [04-01]
  provides: [goreleaser-config, release-pipeline-config]
  affects: [04-04, 04-05, 04-06]
tech_stack:
  added: [GoReleaser v2, cosign keyless signing]
  patterns: [cross-platform build matrix, homebrew_casks tap update]
key_files:
  created:
    - .goreleaser.yaml
  modified: []
decisions:
  - "Use homebrew_casks (not brews) — brews is deprecated in GoReleaser v2.16+"
  - "CGO_ENABLED=0 for static binaries with no dynamic linking"
  - "cosign keyless signing via --bundle flag (not deprecated --output-certificate)"
  - "Defer Homebrew tap GitHub setup — user chose skip, will configure before v0.1.0 tag"
metrics:
  duration: "~5min"
  completed: "2026-06-09"
  tasks_completed: 1
  tasks_total: 2
  files_created: 1
  files_modified: 0
---

# Phase 04 Plan 03: GoReleaser Config Summary

GoReleaser v2 config for cross-platform build + release with cosign keyless signing and Homebrew tap update via homebrew_casks.

## Tasks Completed

| Task | Status | Commit | Notes |
|------|--------|--------|-------|
| 1: Create .goreleaser.yaml | Done | 1b66765 | Full v2 config with ldflags, cosign, homebrew_casks |
| 2: Checkpoint — Homebrew tap GitHub setup | Skipped by user | — | User chose "skip"; deferred to pre-release |

## What Was Built

`.goreleaser.yaml` at the repo root with:

- **5 platform targets**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 (windows/arm64 excluded)
- **ldflags version injection**: injects `Version`, `GOOS`, `GOARCH` into `internal/version` package vars
- **cosign keyless signing**: `--bundle=${signature}` produces `checksums.txt.sigstore.json` (v3 API, not deprecated v2 flags)
- **homebrew_casks block**: points to `ejyle/homebrew-agentkit` repo via `HOMEBREW_TAP_GITHUB_TOKEN` secret
- **Archive formats**: `tar.gz` for Linux/macOS, `zip` for Windows
- **Snapshot support**: version template `{{ .Version }}-SNAPSHOT-{{ .ShortCommit }}`
- **Changelog filter**: excludes `^docs:`, `^test:`, `^chore:` commits

## Deviations from Plan

### User Decision at Checkpoint

**Checkpoint:** Task 2 — Homebrew tap GitHub pre-conditions

**Decision:** User typed "skip" — both pre-conditions are deferred:
- `ejyle/homebrew-agentkit` GitHub repository: NOT created yet
- `HOMEBREW_TAP_GITHUB_TOKEN` Actions secret: NOT created yet

**Impact:** The release pipeline (build, sign, GitHub Release) will work correctly on the first `v0.1.0` tag push. The `homebrew_casks` push step will fail with a 403 until the tap repo and PAT secret are created. This is expected and non-blocking for binary distribution.

**Required before `brew install` works:** Complete the two manual steps from the checkpoint:
1. Create `ejyle/homebrew-agentkit` (public repo with README)
2. Add `HOMEBREW_TAP_GITHUB_TOKEN` classic PAT (repo scope) as Actions secret in `ejyle/agentkit`

## Known Stubs

None — `.goreleaser.yaml` is a complete, functional config. The `HOMEBREW_TAP_GITHUB_TOKEN` reference is intentional and will be resolved at release time.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: secret-dependency | .goreleaser.yaml | HOMEBREW_TAP_GITHUB_TOKEN PAT not yet created; tap push will 403 until secret exists |

## Self-Check: PASSED

- `.goreleaser.yaml` exists at repo root: FOUND
- Commit 1b66765 exists: FOUND (cherry-picked to main)
