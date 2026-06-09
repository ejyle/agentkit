---
phase: 04-distribution-hardening
plan: "04"
subsystem: ci-release
tags: [github-actions, goreleaser, cosign, homebrew, release-automation]
dependency_graph:
  requires: [04-01, 04-03]
  provides: [release-workflow]
  affects: [.github/workflows/release.yml]
tech_stack:
  added: [goreleaser-action@v7, cosign-installer@v3, actions/checkout@v4, actions/setup-go@v5]
  patterns: [dual-trigger-workflow, keyless-signing, snapshot-builds]
key_files:
  created: [.github/workflows/release.yml]
  modified: []
decisions:
  - "Snapshot job uses --skip=publish,sign,homebrew to prevent tap pollution on main push"
  - "id-token: write scoped only to release job — snapshot job uses contents: read"
  - "cosign-installer step omitted from snapshot job since signing is skipped"
metrics:
  duration: "5min"
  completed_date: "2026-06-09"
---

# Phase 04 Plan 04: GitHub Actions Release Workflow Summary

**One-liner:** Dual-trigger GitHub Actions workflow — v* tag push runs full goreleaser release with cosign OIDC signing; main push runs snapshot build with publish/sign/homebrew skipped.

---

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create .github/workflows/release.yml | dd60145 | .github/workflows/release.yml |

---

## What Was Built

`.github/workflows/release.yml` with two jobs:

**release job** (triggers on `v*` tag push):
- `actions/checkout@v4` with `fetch-depth: 0` (full history for GoReleaser changelog)
- `actions/setup-go@v5` with `go-version: stable`
- `sigstore/cosign-installer@v3` (installs cosign binary for OIDC keyless signing)
- `goreleaser/goreleaser-action@v7` with `args: release --clean`
- Permissions: `contents: write`, `id-token: write` (OIDC required for cosign)
- Env: `GITHUB_TOKEN` + `HOMEBREW_TAP_GITHUB_TOKEN` (for tap repo push)

**snapshot job** (triggers on `main` branch push):
- `actions/checkout@v4` with `fetch-depth: 0`
- `actions/setup-go@v5` with `go-version: stable`
- `goreleaser/goreleaser-action@v7` with `args: release --snapshot --clean --skip=publish,sign,homebrew`
- Permissions: `contents: read` only (no release upload)
- No cosign-installer (signing skipped)
- No `HOMEBREW_TAP_GITHUB_TOKEN` (homebrew skipped)

---

## Deviations from Plan

None — plan executed exactly as written.

---

## Threat Mitigations Applied

| Threat ID | Mitigation |
|-----------|------------|
| T-04-04-01 | `HOMEBREW_TAP_GITHUB_TOKEN` referenced via `${{ secrets.* }}` — GitHub masks in logs |
| T-04-04-02 | Release job gated by `startsWith(github.ref, 'refs/tags/')` — fork PRs cannot trigger |
| T-04-04-04 | Snapshot job uses `--skip=publish,sign,homebrew` — no tap pollution |

---

## Self-Check: PASSED

- `.github/workflows/release.yml` exists: FOUND
- Commit dd60145 exists: FOUND
- Contains `goreleaser-action@v7`: 2 occurrences (release + snapshot jobs)
- Contains `id-token: write`: release job only
- Contains `fetch-depth: 0`: both jobs
- Contains `cosign-installer@v3`: release job only
- `COSIGN_EXPERIMENTAL` absent: CONFIRMED
