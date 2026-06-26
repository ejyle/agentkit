---
status: passed
phase: 04-distribution-hardening
source: [04-VERIFICATION.md]
started: 2026-06-13T00:00:00Z
updated: 2026-06-14T00:00:00Z
passed_at: 2026-06-14T00:00:00Z
---

## Current Test

[complete — tests 2 and 4 passed 2026-06-14; test 1 (Homebrew tap) deferred; test 3 (Homebrew install) deferred]

## Tests

### 1. Homebrew tap GitHub setup
expected: ejyle/homebrew-agentkit repo exists (public) and HOMEBREW_TAP_GITHUB_TOKEN secret is set in ejyle/agentkit Actions
result: [deferred — developer decision: not yet prioritized]

### 2. v0.1.0 tag push triggers release pipeline
expected: GitHub Actions "Release" workflow completes green on v0.1.0 tag; 5 binary archives + checksums.txt + checksums.txt.sigstore.json attached to GitHub Release
result: [PASSED — release job green 2026-06-14; 5 binary archives + checksums.txt + checksums.txt.sigstore.json published at github.com/ejyle/agentkit/releases/tag/v0.1.0]

### 3. Homebrew install on macOS
expected: `brew install ejyle/agentkit/agentkit` succeeds; `agentkit --version` outputs `agentkit/0.1.0 (darwin/arm64)`
result: [deferred — blocked by test 1 (Homebrew tap not yet set up)]

### 4. curl|sh install on macOS/Linux
expected: `sh scripts/install.sh` downloads, verifies checksum, and installs agentkit to ~/.local/bin; `agentkit --version` correct
result: [PASSED — installs agentkit/0.1.0 (darwin/arm64) to ~/.local/bin; checksum OK; two bugs fixed during test: --wildcards (GNU tar only) and --strip-components=1 (flat archive)]

## Summary

total: 4
passed: 2
issues: 0
pending: 0
skipped: 0
blocked: 0
deferred: 2

## Gaps

- Test 1 (Homebrew tap): deferred — requires creating ejyle/homebrew-agentkit repo and PAT secret
- Test 3 (Homebrew install): deferred — blocked by test 1
- Two bugs found and fixed in install.sh during test 4: removed --wildcards (GNU tar only, fails on macOS BSD tar) and --strip-components=1 (GoReleaser archive is flat, not nested)
