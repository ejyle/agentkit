# Phase 4: Distribution & Hardening - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-09
**Phase:** 4-distribution-hardening
**Areas discussed:** Release pipeline & signing, Homebrew tap setup, agentkit doctor checks, Version strategy & install script

---

## Release Pipeline & Signing

| Option | Description | Selected |
|--------|-------------|----------|
| Tag push only (v*) | Only runs on v* tag push | |
| Tag push + main snapshot | Tag push → real release; main push → snapshot/pre-release | ✓ |
| Manual trigger only | workflow_dispatch only | |

**User's choice:** Tag push + main snapshot

---

| Option | Description | Selected |
|--------|-------------|----------|
| No signing for v0.1.0 | Unsigned binaries, add signing in v0.2.0 | |
| Cosign keyless signing | GoReleaser signs via GitHub OIDC token, no key management | ✓ |
| macOS notarization only | Apple Developer account + codesign cert required | |

**User's choice:** Cosign keyless signing

---

| Option | Description | Selected |
|--------|-------------|----------|
| ldflags -X at build time | GoReleaser injects tag via ldflags, zero runtime overhead | ✓ |
| Embedded via go:generate | Generated version.go at release time | |
| Read from binary metadata at runtime | debug.ReadBuildInfo() — uses VCS metadata not release tag | |

**User's choice:** ldflags -X at build time (Recommended)

---

## Homebrew Tap Setup

| Option | Description | Selected |
|--------|-------------|----------|
| Separate homebrew-agentkit repo | Standard convention: ejyle/homebrew-agentkit | ✓ |
| Formula in main repo (Formula/ dir) | formula in ejyle/agentkit under Formula/ | |
| You decide | Claude picks standard convention | |

**User's choice:** Separate homebrew-agentkit repo (Recommended)

---

| Option | Description | Selected |
|--------|-------------|----------|
| brew install ejyle/agentkit/agentkit | Single command: taps and installs | ✓ |
| brew tap ejyle/agentkit && brew install agentkit | Two-step explicit | |
| You decide | Claude picks single command | |

**User's choice:** brew install ejyle/agentkit/agentkit

---

| Option | Description | Selected |
|--------|-------------|----------|
| GitHub PAT stored as repo secret | Classic PAT with repo scope, HOMEBREW_TAP_GITHUB_TOKEN secret | ✓ |
| GitHub Actions fine-grained token | Fine-grained PAT scoped to homebrew-agentkit repo | |
| Deploy key on homebrew-agentkit repo | SSH deploy key with write access | |

**User's choice:** GitHub PAT stored as repo secret (Recommended)

---

## agentkit doctor checks

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — check node/npx, docker, uvx availability | Extended checks for MCP runtime deps | ✓ |
| No — only 4 roadmap checks | Minimal v0.1.0 approach | |
| You decide | Claude recommends extended checks | |

**User's choice:** Yes — check node/npx, docker, uvx availability

---

| Option | Description | Selected |
|--------|-------------|----------|
| Line-by-line with pass/warn/fail icons | ✓/⚠/✗ per check, hint lines for failures | ✓ |
| Summary table | Check / Status / Notes columns | |
| JSON output with --json flag | Machine-readable overlay | |

**User's choice:** Line-by-line with pass/warn/fail icons (Recommended)

---

| Option | Description | Selected |
|--------|-------------|----------|
| No — diagnose only, print fix hints | Reports issue + what to run to fix it | ✓ |
| Yes — agentkit doctor --fix auto-repairs | Creates dirs, adds to PATH config | |
| You decide | Claude recommends diagnose-only | |

**User's choice:** No — diagnose only, print fix hints (Recommended)

---

## Version Strategy & Install Script

| Option | Description | Selected |
|--------|-------------|----------|
| v0.1.0 | Standard SemVer with v prefix | ✓ |
| 0.1.0 (no v prefix) | Less common for Go CLIs | |
| v1.0.0 | Skip pre-1.0 phase | |

**User's choice:** v0.1.0 (Recommended)

---

| Option | Description | Selected |
|--------|-------------|----------|
| agentkit version 0.1.0 | Clean human-friendly format | |
| 0.1.0 | Just the version number | |
| agentkit/0.1.0 (darwin/arm64) | Version + OS/arch for bug reports | ✓ |

**User's choice:** agentkit/0.1.0 (darwin/arm64)

---

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — hosted at install.agentkit.dev or in repo | Detects OS/arch, downloads binary, verifies checksum | ✓ |
| No — Homebrew + go install only | Power users covered by go install | |
| You decide | Claude recommends including script | |

**User's choice:** Yes — hosted in repo (install.agentkit.dev for later). User noted: "pass the hosted script for now."

---

## Claude's Discretion

- GoReleaser archive naming convention
- Whether snapshot builds on main are pre-release or draft in GitHub
- Whether `agentkit doctor` checks all assistant config dirs or only those with installed packages
- Exact GoReleaser `changelog` config (conventional commits or full list)
- `internal/version/version.go` package structure

## Deferred Ideas

- `agentkit doctor --fix` auto-repair capability → v0.2.0
- macOS notarization → v0.2.0
- Windows package manager (winget/choco/Scoop) → v0.2.0 per REQUIREMENTS.md
- install.agentkit.dev custom domain redirect → post-v0.1.0
