---
status: deferred
phase: 04-distribution-hardening
source: [04-VERIFICATION.md]
started: 2026-06-13T00:00:00Z
updated: 2026-06-13T00:00:00Z
deferred_reason: Homebrew tap setup deferred by developer; tests 2-4 depend on tag push which requires test 1
code_approved: true
code_approved_by: developer
code_approved_at: 2026-06-13T00:00:00Z
code_approval_evidence: "134 tests pass; go build clean; doctor 7/9 checks pass (PATH+Docker expected in dev); goreleaser config verified; install.sh bash -n syntax OK; release workflow correct"
---

## Current Test

[deferred — Homebrew tap setup not yet done; tests 2-4 blocked until tap + tag push]

## Tests

### 1. Homebrew tap GitHub setup
expected: ejyle/homebrew-agentkit repo exists (public) and HOMEBREW_TAP_GITHUB_TOKEN secret is set in ejyle/agentkit Actions
result: [deferred — developer decision: not yet prioritized]

### 2. v0.1.0 tag push triggers release pipeline
expected: GitHub Actions "Release" workflow completes green on v0.1.0 tag; 5 binary archives + checksums.txt + checksums.txt.sigstore.json attached to GitHub Release
result: [deferred — blocked by test 1 (Homebrew tap pre-conditions)]

### 3. Homebrew install on macOS
expected: `brew install ejyle/agentkit/agentkit` succeeds; `agentkit --version` outputs `agentkit/0.1.0 (darwin/arm64)`
result: [deferred — blocked by test 2 (tag push)]

### 4. curl|sh install on Linux
expected: `sh scripts/install.sh` downloads, verifies checksum, and installs agentkit to ~/.local/bin; `agentkit --version` correct
result: [deferred — blocked by test 2 (tag push)]

## Summary

total: 4
passed: 0
issues: 0
pending: 0
skipped: 0
blocked: 0
deferred: 4

## Gaps
