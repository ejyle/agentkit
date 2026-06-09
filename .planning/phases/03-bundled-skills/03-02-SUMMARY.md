---
phase: "03-bundled-skills"
plan: "02"
subsystem: "bundle-command"
tags: ["bundle", "parallel-install", "cobra", "go-embed"]
dependency_graph:
  requires:
    - "03-01 (GitHubReleaseInstaller, domain types)"
  provides:
    - "internal/bundle package (LoadBundles, Resolve)"
    - "--bundle flag on installCmd"
    - "runBundleInstall with sync.WaitGroup parallel goroutines"
  affects:
    - "cmd/install.go (extended)"
tech_stack:
  added:
    - "go:embed directive for bundles.json"
    - "sync.WaitGroup for parallel bundle installs"
  patterns:
    - "Embedded config via //go:embed"
    - "Best-effort parallel goroutines (not errgroup)"
    - "D-04 best-effort semantics / D-05 exit code 1 on any failure"
key_files:
  created:
    - "internal/bundle/bundles.go"
    - "internal/bundle/bundles.json"
    - "internal/bundle/bundles_test.go"
  modified:
    - "cmd/install.go"
decisions:
  - "sync.WaitGroup used instead of errgroup — errgroup cancels on first error, violating D-04 best-effort semantics"
  - "D-17 gate: gsd-core registry.json returns HTTP 404 (repo open-gsd/gsd-core not yet created); URL is pre-wired in NewRegistryManager(); no code change needed for CLI-03"
metrics:
  duration: "~15min"
  completed: "2026-06-09T08:20:39Z"
  tasks_completed: 2
  files_created: 3
  files_modified: 1
---

# Phase 3 Plan 2: Bundle Command Summary

**One-liner:** Embedded bundle manifest with LoadBundles/Resolve, --bundle flag on installCmd dispatching parallel sync.WaitGroup goroutines with D-04 best-effort semantics and D-05 exit code.

---

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create internal/bundle package with embedded bundles.json | a77cac1 | internal/bundle/bundles.go, bundles.json, bundles_test.go |
| 2 | Add --bundle flag to installCmd and implement runBundleInstall | 240ef32 | cmd/install.go |

---

## Verification Results

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./internal/bundle/... -v` (3 tests) | PASS |
| `go vet ./cmd/...` | PASS |
| `cobra.RangeArgs(0, 1)` present | PASS |
| `func runBundleInstall` present | PASS |
| `sync.WaitGroup` present (not errgroup) | PASS |
| `internal/bundle` import in cmd/install.go | PASS |
| D-17 gsd install method documented as comment | PASS |

---

## CLI-03 / D-17 Verification (Required Pre-Condition)

**D-17 Gate Result:** Verified before implementation.

- URL checked: `https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json`
- HTTP response: 404 — repository `open-gsd/gsd-core` does not yet exist
- Finding: The gsd-core registry URL is already pre-wired in `NewRegistryManager()` (registry.go line 48). When the repo is created, `agentkit install gsd` will automatically route through gsd-core without any code change.
- Until the repo exists: `agentkit install gsd` will attempt gsd-core (fail gracefully), then fall back to `agentkit-registry`.
- **No additional code change required for CLI-03.**
- Documented as comment at top of cmd/install.go.

---

## Design Decisions

### sync.WaitGroup (not errgroup)
`errgroup` cancels all goroutines on first error, violating D-04 (best-effort: all packages attempted regardless of individual failures). `sync.WaitGroup` with an index-keyed results slice gives each goroutine an isolated write slot — no mutex needed.

### cobra.RangeArgs(0, 1)
Changed from `ExactArgs(1)` to allow `agentkit install --bundle cloud` with zero positional args. Added explicit error when neither `--bundle` nor a positional name is provided.

### Embedded bundles.json
Using `//go:embed bundles.json` bakes the bundle definitions into the binary at compile time. No runtime file I/O, no path resolution needed, no possibility of file substitution (T-03-06 accepted in threat model).

---

## Deviations from Plan

None — plan executed exactly as written. D-17 pre-condition verified before implementation as required.

---

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns, or schema changes beyond the plan's threat model. Threat register entries T-03-06, T-03-07, T-03-08 all apply as documented in the plan.

---

## Known Stubs

None. Bundle definitions are complete (cloud/dev/context) and wired to real svc.Install() calls.

---

## Self-Check: PASSED

| Item | Status |
|------|--------|
| internal/bundle/bundles.go | FOUND |
| internal/bundle/bundles.json | FOUND |
| internal/bundle/bundles_test.go | FOUND |
| cmd/install.go | FOUND |
| .planning/phases/03-bundled-skills/03-02-SUMMARY.md | FOUND |
| commit a77cac1 | FOUND |
| commit 240ef32 | FOUND |
