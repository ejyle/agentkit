---
status: partial
phase: 04-distribution-hardening
source: [04-VERIFICATION.md]
started: 2026-06-13T00:00:00Z
updated: 2026-06-13T00:00:00Z
---

## Current Test

[awaiting human testing]

## Tests

### 1. Homebrew tap GitHub setup
expected: ejyle/homebrew-agentkit repo exists (public) and HOMEBREW_TAP_GITHUB_TOKEN secret is set in ejyle/agentkit Actions
result: [pending]

### 2. v0.1.0 tag push triggers release pipeline
expected: GitHub Actions "Release" workflow completes green on v0.1.0 tag; 5 binary archives + checksums.txt + checksums.txt.sigstore.json attached to GitHub Release
result: [pending]

### 3. Homebrew install on macOS
expected: `brew install ejyle/agentkit/agentkit` succeeds; `agentkit --version` outputs `agentkit/0.1.0 (darwin/arm64)`
result: [pending]

### 4. curl|sh install on Linux
expected: `sh scripts/install.sh` downloads, verifies checksum, and installs agentkit to ~/.local/bin; `agentkit --version` correct
result: [pending]

## Summary

total: 4
passed: 0
issues: 0
pending: 4
skipped: 0
blocked: 0

## Gaps
