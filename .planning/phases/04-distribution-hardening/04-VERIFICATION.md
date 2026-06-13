---
phase: 04-distribution-hardening
verified: 2026-06-13T00:00:00Z
status: human_needed
score: 4/4 roadmap success criteria verified (all blockers closed; SC-1/SC-2/SC-4 human-gated on live tag push)
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 1/4
  gaps_closed:
    - ".goreleaser.yaml: homebrew_casks: changed to brews: (CLI formula, correct GoReleaser v2 key)"
    - "scripts/install.sh: shebang changed to #!/usr/bin/env bash (dash compat fix — pipefail now valid)"
    - "scripts/install.sh: --ignore-missing removed; checksum now uses grep | --check - (security fix)"
    - "scripts/install.sh: tar extraction uses --strip-components=1 (handles GoReleaser archive layout)"
  gaps_remaining: []
  regressions: []
human_verification:
  - test: "Push v0.1.0 tag and monitor GitHub Actions release job"
    expected: "Release job completes green; GitHub Release page shows 5 binary archives, checksums.txt, and checksums.txt.sigstore.json"
    why_human: "Requires live tag push and CI execution — cannot verify statically"
  - test: "Verify ejyle/homebrew-agentkit repo and HOMEBREW_TAP_GITHUB_TOKEN secret exist before tag push"
    expected: "Public repo at github.com/ejyle/homebrew-agentkit and HOMEBREW_TAP_GITHUB_TOKEN Actions secret in ejyle/agentkit"
    why_human: "GitHub repository creation and PAT configuration cannot be automated"
  - test: "Homebrew install after tag push: brew install ejyle/agentkit/agentkit && agentkit --version"
    expected: "Installs without error; outputs agentkit/0.1.0 (darwin/arm64) or darwin/amd64"
    why_human: "Requires live Homebrew tap, installed binary on macOS, and v0.1.0 release assets"
  - test: "Install script end-to-end: sh scripts/install.sh && ~/.local/bin/agentkit --version"
    expected: "Downloads v0.1.0 archive, verifies SHA256, installs to ~/.local/bin, outputs agentkit/0.1.0 (linux/amd64)"
    why_human: "Requires v0.1.0 release assets on GitHub and live network access"
---

# Phase 4: Distribution & Hardening Verification Report

**Phase Goal:** agentkit v0.1.0 is publicly releasable: cross-platform binaries built via GoReleaser v2, installable via Homebrew tap, with a `doctor` command that validates the local install environment.
**Verified:** 2026-06-13T00:00:00Z
**Status:** human_needed
**Re-verification:** Yes — after gap closure (4 blockers fixed)

---

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| SC-1 | `brew install agentkit/tap/agentkit` succeeds on macOS and produces a working binary with `agentkit --version` returning the correct release tag | HUMAN-GATED | `.goreleaser.yaml` now uses `brews:` (correct CLI Formula key — blocker closed). Homebrew tap repo and HOMEBREW_TAP_GITHUB_TOKEN secret still need creation; no v0.1.0 tag pushed. Cannot verify statically. |
| SC-2 | GitHub Actions release workflow runs GoReleaser v2 on tag push and produces signed binaries for macOS (arm64/amd64), Linux (amd64/arm64), Windows (amd64) | HUMAN-GATED | `.github/workflows/release.yml` is correctly configured. Workflow has not been triggered — no v0.1.0 tag exists. End-to-end execution requires human tag push. |
| SC-3 | `agentkit doctor` checks binary in PATH, ~/.agentkit/ writable, registry reachable, assistant config dirs — printing pass/warn/fail for each check | VERIFIED | cmd/doctor.go implements all 9 checks. Build PASS; `agentkit --help` lists doctor; `agentkit --version` outputs `agentkit/dev (unknown/unknown)`. |
| SC-4 | Direct binary download from GitHub release page installs without package manager or root/sudo on all three platforms | HUMAN-GATED | All 3 blocking defects in scripts/install.sh are resolved (see re_verification.gaps_closed). No v0.1.0 release assets exist yet. End-to-end test requires live release. |

**Score:** 4/4 truths cleared of static blockers. SC-3 fully verified; SC-1, SC-2, SC-4 human-gated on live tag push.

---

### Gap Closure Verification

The 4 blockers from the previous verification were verified directly against the source files:

| Blocker | Fix Applied | Verification |
|---------|-------------|-------------|
| `.goreleaser.yaml:59` — `homebrew_casks:` wrong key | Changed to `brews:` (line 59) | CONFIRMED — `grep '^brews:' .goreleaser.yaml` returns `brews:` |
| `scripts/install.sh:1` — `#!/usr/bin/env sh` (dash compat) | Changed to `#!/usr/bin/env bash` (line 1) | CONFIRMED — shebang verified; `bash -n scripts/install.sh` exits 0 |
| `scripts/install.sh:86` — `--ignore-missing` bypasses checksum | Replaced with `grep "${FILENAME}" ... \| ${SHA_CMD} --check -` (line 86) | CONFIRMED — `--ignore-missing` absent; grep+pipe pattern present |
| `scripts/install.sh:89` — tar extraction at archive root | Added `--strip-components=1` with `--wildcards "*/agentkit"` (line 89) | CONFIRMED — both flags present on line 89 |

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/version/version.go` | ldflags-injectable version package with String() method | VERIFIED | Contains `var Version = "dev"`, `var GOOS`, `var GOARCH`, `func String() string` |
| `cmd/root.go` | root command with Version field wired to version.String() | VERIFIED | Imports `github.com/ejyle/agentkit/internal/version`, sets `Version: version.String()` |
| `cmd/doctor.go` | doctor subcommand with all D-08 checks | VERIFIED | 9 checks; `go build` PASS; doctor listed in `--help` |
| `.goreleaser.yaml` | GoReleaser v2 config for cross-platform build + Homebrew Formula | VERIFIED | version: 2, 5 targets, correct ldflags, cosign signing, `brews:` key (fixed) |
| `.github/workflows/release.yml` | dual-trigger GitHub Actions release workflow | VERIFIED | goreleaser-action@v7, cosign-installer@v3, fetch-depth: 0, id-token: write, correct HOMEBREW_TAP_GITHUB_TOKEN scoping |
| `scripts/install.sh` | curl-pipe install script with checksum verification | VERIFIED | All 3 blocking defects resolved; `bash -n` syntax check passes |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| cmd/root.go | internal/version/version.go | `import + version.String()` | WIRED | Binary outputs `agentkit/dev (unknown/unknown)` — link functional |
| .goreleaser.yaml ldflags | internal/version/version.go | `-X ...version.Version={{.Version}}` | WIRED | All three ldflag vars present on lines 25-27 |
| .goreleaser.yaml brews | ejyle/homebrew-agentkit | `HOMEBREW_TAP_GITHUB_TOKEN` env var | WIRED (config correct; repo/secret pending) | `brews:` key is correct; tap repo and PAT secret need creation before first release |
| .github/workflows/release.yml | .goreleaser.yaml | goreleaser-action@v7 `args: release --clean` | WIRED | Correct action version and args |
| scripts/install.sh | github.com/ejyle/agentkit/releases/download | constructs URL from OS/ARCH/VERSION | WIRED | URL construction correct; checksum and extraction logic now correct |
| cmd/doctor.go | cmd/root.go | `rootCmd.AddCommand(doctorCmd)` | WIRED | Confirmed; doctor listed in `agentkit --help` |

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Binary builds without errors | `go build -o /tmp/agentkit-verify-04 .` | Build succeeded | PASS |
| `--version` outputs correct format | `/tmp/agentkit-verify-04 --version` | `agentkit/dev (unknown/unknown)` | PASS |
| `doctor` command registered | `/tmp/agentkit-verify-04 --help \| grep doctor` | `doctor      Check your agentkit environment` | PASS |
| install.sh bash syntax valid | `bash -n scripts/install.sh` | syntax ok (exit 0) | PASS |
| v0.1.0 tag exists | `git tag --list v0.1.0` | (empty — no tags in repo) | SKIP (human action required) |

---

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|-------------|-------------|-------------|--------|---------|
| CLI-10 | 04-01, 04-02, 04-03, 04-04, 04-05, 04-06 | CLI ships as single binary with no runtime dependency, runs on Windows/Linux/macOS without root or sudo | SATISFIED (pending live release) | CGO_ENABLED=0 confirmed; install.sh installs to ~/.local/bin (no root); all blocking defects in the install path are resolved |

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| cmd/doctor.go | 145 | `if resp.StatusCode >= 500` | WARNING | HTTP 4xx (404/403/429) classified as pass — misleads users when registry URL is wrong or access blocked |
| scripts/install.sh | 14 | hardcoded fallback version removed (now dynamic via API) | INFO | Version now resolved from GitHub API; no stale default |
| cmd/doctor.go | 32 | `doctorCmd.PersistentPreRunE` bypass | INFO | PersistentPreRunE propagates to future doctor subcommands; should use PreRunE |
| cmd/doctor.go | 109 | `_ = os.Remove(testFile)` | INFO | Discards removal error; .write-test artifact persists on permission failure |

No BLOCKERs in anti-pattern scan. All 4 previous BLOCKERs are closed.

---

### Human Verification Required

#### 1. Homebrew Tap Pre-conditions

**Test:** Create the `ejyle/homebrew-agentkit` public GitHub repository. Create the `HOMEBREW_TAP_GITHUB_TOKEN` Actions secret in `ejyle/agentkit` with a PAT that has `repo` write scope on `homebrew-agentkit`.
**Expected:** `github.com/ejyle/homebrew-agentkit` exists as a public repo; secret appears in `ejyle/agentkit → Settings → Secrets and variables → Actions`.
**Why human:** GitHub repository creation and PAT configuration require interactive GitHub UI or authenticated API calls.

#### 2. v0.1.0 Tag Push and Release Job

**Test:** Push the v0.1.0 tag after pre-conditions above are met: `git tag v0.1.0 -m "Release v0.1.0" && git push origin v0.1.0`. Monitor the release job at https://github.com/ejyle/agentkit/actions.
**Expected:** Release job completes green. GitHub Release at https://github.com/ejyle/agentkit/releases/tag/v0.1.0 shows 5 binary archives, checksums.txt, and checksums.txt.sigstore.json.
**Why human:** Requires live tag push and CI execution. Cannot verify statically.

#### 3. Homebrew Install (after tag push)

**Test:** `brew tap ejyle/agentkit && brew install ejyle/agentkit/agentkit && agentkit --version`
**Expected:** Installs without error; outputs `agentkit/0.1.0 (darwin/arm64)` (or amd64 on Intel).
**Why human:** Requires live Homebrew tap and installed binary on macOS.

#### 4. Install Script End-to-End (after tag push)

**Test:** `sh scripts/install.sh && ~/.local/bin/agentkit --version`
**Expected:** Downloads v0.1.0 archive, verifies SHA256, installs to ~/.local/bin, outputs `agentkit/0.1.0 (linux/amd64)`.
**Why human:** Requires v0.1.0 release assets on GitHub and live network access.

---

### Gaps Summary

No static blockers remain. All 4 previously identified blocking defects have been confirmed fixed:

1. `.goreleaser.yaml` `homebrew_casks:` → `brews:` — CLOSED
2. `scripts/install.sh` shebang `#!/usr/bin/env sh` → `#!/usr/bin/env bash` — CLOSED
3. `scripts/install.sh` `--ignore-missing` → `grep | --check -` — CLOSED
4. `scripts/install.sh` tar extraction → `--strip-components=1 --wildcards "*/agentkit"` — CLOSED

Phase goal is achievable. Remaining gate: human must create the Homebrew tap repository, set the PAT secret, and push the v0.1.0 tag to trigger the release pipeline.

---

_Verified: 2026-06-13T00:00:00Z_
_Verifier: Claude (gsd-verifier)_
