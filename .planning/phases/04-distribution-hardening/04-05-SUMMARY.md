---
phase: 04-distribution-hardening
plan: "05"
subsystem: infra
tags: [shell, install-script, sha256, curl, distribution]

# Dependency graph
requires:
  - phase: 04-03
    provides: GoReleaser pipeline producing versioned tar.gz archives and checksums.txt
  - phase: 04-04
    provides: Homebrew tap formula for macOS users
provides:
  - curl|sh install script for Linux and macOS users without Homebrew
  - SHA256 checksum-verified install flow before binary execution
  - User-scope install to ~/.local/bin with no root/sudo required
affects: [distribution, cli-install, linux-users, macos-users]

# Tech tracking
tech-stack:
  added: [POSIX sh, sha256sum, shasum]
  patterns: [mktemp + trap cleanup, OS/ARCH detection via uname, checksum-before-install]

key-files:
  created:
    - scripts/install.sh
  modified: []

key-decisions:
  - "Install to ~/.local/bin only — no sudo required per CLI-10"
  - "sha256sum verified against GoReleaser checksums.txt before binary is moved to install dir"
  - "VERSION overridable via AGENTKIT_VERSION env var; defaults to 0.1.0"
  - "mktemp -d plus trap 'rm -rf' EXIT used for secure temp dir (no hardcoded /tmp)"

patterns-established:
  - "Checksum verification runs before binary install — set -e causes immediate exit on mismatch"
  - "macOS/Linux checksum tool auto-detected: sha256sum first, shasum -a 256 fallback"

requirements-completed:
  - CLI-10

# Metrics
duration: 12min
completed: "2026-06-09"
---

# Phase 4 Plan 5: Install Script Summary

**curl|sh install script with SHA256 checksum verification, OS/ARCH detection, and user-scope install to ~/.local/bin**

## Performance

- **Duration:** 12 min
- **Started:** 2026-06-09T18:28:52Z
- **Completed:** 2026-06-09T18:40:00Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments

- Created `scripts/install.sh` — the primary install path for Linux users and macOS users without Homebrew
- SHA256 checksum verified from GoReleaser `checksums.txt` before binary is moved to install directory
- Handles both `sha256sum` (Linux) and `shasum -a 256` (macOS) automatically with graceful fallback
- Uses `mktemp -d` and `trap 'rm -rf'` for secure temp dir cleanup on success or failure
- `AGENTKIT_VERSION` env var override mechanism for installing specific versions

## Task Commits

1. **Task 1: Create scripts/install.sh** - `30be028` (feat)

**Plan metadata:** (docs commit — see below)

## Files Created/Modified

- `scripts/install.sh` — curl|sh install script with SHA256 verification, OS/ARCH detection, installs to ~/.local/bin

## Decisions Made

- Install to `~/.local/bin` only (no `/usr/local/bin`), satisfying CLI-10's no-root requirement
- SHA256 checksum verification runs before `mv` — `set -e` exits immediately on mismatch
- `AGENTKIT_VERSION` env override defaults to `0.1.0`; consistent with goreleaser version
- `mktemp -d` over hardcoded `/tmp` for security and cross-platform compatibility

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Known Stubs

None.

## Threat Flags

None — threats T-04-05-01 through T-04-05-05 all addressed per the threat model:
- T-04-05-01 (Tampered binary): SHA256 verified before install
- T-04-05-02 (MITM): curl -fsSL enforces TLS + SHA256 post-verification
- T-04-05-05 (Temp dir pollution): trap cleanup on EXIT

## Self-Check

- [x] `scripts/install.sh` exists: FOUND
- [x] Commit `30be028` exists
- [x] File passes `sh -n` syntax check
- [x] Contains `sha256` (4 occurrences)
- [x] Contains `checksums.txt` (2 occurrences)
- [x] Contains `mktemp -d` (1 occurrence)
- [x] Contains `trap` with rm -rf (1 occurrence)
- [x] Contains `$HOME/.local/bin` install path
- [x] Contains `AGENTKIT_VERSION` override
- [x] Does NOT install to `/usr/local/bin`
- [x] Checksum runs before `mv` to install dir

## Self-Check: PASSED

## Next Phase Readiness

- Install script complete — Linux and macOS users can install without Homebrew
- Combined with Phase 04-03 (GoReleaser CI/CD) and 04-04 (Homebrew tap), all three install paths are covered
- Ready for Phase 04-06 (integration testing / hardening completion)

---
*Phase: 04-distribution-hardening*
*Completed: 2026-06-09*
