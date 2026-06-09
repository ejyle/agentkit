---
phase: 03-bundled-skills
plan: "01"
subsystem: installer
tags: [github-release, installer, skill-extraction, path-traversal, caching]
dependency_graph:
  requires: []
  provides:
    - GitHubReleaseInstaller (internal/installer/github_release.go)
    - InstallMethodGitHubRelease constant (internal/domain/package.go)
    - TarballCachePath function (internal/config/paths.go)
    - ErrGitHubReleaseNotFound sentinel (internal/installer/installer.go)
  affects:
    - internal/service/install.go (WriteSkill guard, ValidateSkill wiring, SkillDir population)
    - internal/installer/installer.go (factory case)
tech_stack:
  added:
    - github.com/BurntSushi/toml v1.4.0 (pre-existing missing dep in adapter/codex.go)
  patterns:
    - sync.Map for in-process tarball cache
    - Disk cache with renameio.WriteFile (atomic write)
    - httptest.NewTLSServer + redirectTransport pattern for unit tests
    - WithDiskCacheDir() injection to isolate tests from real UserCacheDir
key_files:
  created:
    - internal/installer/github_release.go
    - internal/installer/github_release_test.go
    - internal/installer/github_release_integration_test.go
  modified:
    - internal/domain/package.go
    - internal/config/paths.go
    - internal/installer/installer.go
    - internal/service/install.go
    - internal/service/install_test.go
    - go.mod / go.sum
decisions:
  - "WithDiskCacheDir() added to GitHubReleaseInstaller for test isolation — prevents test artifacts from polluting real UserCacheDir/agentkit/releases/"
  - "errWriter type removed from service/install.go — replaced with direct os.Stderr write for clarity"
  - "BurntSushi/toml v1.4.0 added as Rule-3 auto-fix (pre-existing missing dependency causing service test build failure)"
metrics:
  duration: ~15min
  completed: "2026-06-09"
  tasks_completed: 3
  files_modified: 8
---

# Phase 3 Plan 01: GitHub Release Installer Summary

**One-liner:** GitHubReleaseInstaller fetches GitHub source-archive tarballs, extracts named skill subdirectories to spec.SkillDir with zip-slip guard and in-process/disk tarball caching.

---

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Extend domain types and add TarballCachePath | 62594a0 | domain/package.go, config/paths.go, installer/installer.go |
| 2 | Implement GitHubReleaseInstaller with path-traversal guard, caching, validator wiring | a097836 | installer/github_release.go, github_release_test.go, installer.go, service/install.go |
| 3 | Integration test — end-to-end extraction to SkillInstallPath | cb245b6 | installer/github_release_integration_test.go |

---

## What Was Built

**GitHubReleaseInstaller** (`internal/installer/github_release.go`):
- Fetches `https://github.com/{repo}/archive/refs/tags/{version}.tar.gz`
- Extracts `{repoName}-{version-without-v}/{path}/` subdirectory to `spec.SkillDir`
- In-process `sync.Map` cache keyed by `repo@version` — second install in same process is free
- Disk cache at `UserCacheDir/agentkit/releases/{repo-slug}/{version}/tarball.tar.gz` via `renameio.WriteFile`
- Path traversal guard (T-03-01): rejects entries with `..` in relative path AND verifies `filepath.Join(destDir, rel)` stays within `destDir`
- HTTPS-only URL construction (T-03-02)
- `WithDiskCacheDir(dir)` for test isolation

**Domain updates** (`internal/domain/package.go`):
- `InstallMethodGitHubRelease = "github-release"` constant
- `Repo`, `Path` fields on `InstallSpec` (serialised as `json:"repo,omitempty"`, `json:"path,omitempty"`)
- `SkillDir` field on `InstallSpec` with `json:"-"` — runtime-only, not serialised

**Config** (`internal/config/paths.go`):
- `TarballCachePath(repo, version string)` following `ManifestCachePath` pattern

**Service wiring** (`internal/service/install.go`):
- Populates `pkg.Install.SkillDir = config.SkillInstallPath(target, name)` before calling installer for github-release packages
- Guards `adapter.WriteSkill` call: skipped for `InstallMethodGitHubRelease` (files already extracted)
- Calls `s.validator(pkg.Install.SkillDir, pkg)` with real extracted dir for github-release; blocking errors returned

---

## Test Results

| Test | Result |
|------|--------|
| TestGitHubReleaseInstaller_Extract | PASS |
| TestGitHubReleaseInstaller_PathTraversalRejected | PASS |
| TestGitHubReleaseInstaller_CacheHit | PASS |
| TestGitHubReleaseInstaller_NotFound | PASS |
| TestGitHubReleaseInstaller_ExtractToSkillPath | PASS |
| TestInstallService_Install_GitHubRelease_NoWriteSkill | PASS |
| All pre-existing tests | PASS (132 total) |

---

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added WithDiskCacheDir() for test isolation**
- **Found during:** Task 2 test run
- **Issue:** Tests writing to real `UserCacheDir/agentkit/releases/ejyle-agentkit/` caused subsequent test runs to hit disk cache instead of mock HTTP server, making `TestGitHubReleaseInstaller_CacheHit`, `PathTraversalRejected`, and `NotFound` tests unreliable
- **Fix:** Added `diskCacheDir string` field to `GitHubReleaseInstaller` and exported `WithDiskCacheDir()` method; tests inject `t.TempDir()` to isolate from real cache
- **Files modified:** `internal/installer/github_release.go`, `internal/installer/github_release_test.go`, `internal/installer/github_release_integration_test.go`
- **Commit:** a097836

**2. [Rule 3 - Blocking] Added BurntSushi/toml v1.4.0 dependency**
- **Found during:** Task 2 service test run
- **Issue:** `internal/adapter/codex.go` imported `github.com/BurntSushi/toml` but it was absent from `go.mod`, causing all `service` package tests to fail with build error
- **Fix:** `go get github.com/BurntSushi/toml@v1.4.0`
- **Files modified:** `go.mod`, `go.sum`
- **Commit:** a097836

**3. [Rule 1 - Cleanup] Removed errWriter type from service/install.go**
- **Found during:** Task 2 service update
- **Issue:** `errWriter` was no longer used after warnings were changed to write to `os.Stderr` directly (cleaner, no indirection)
- **Fix:** Removed the dead `errWriter` type and its `Write` method
- **Files modified:** `internal/service/install.go`
- **Commit:** a097836

---

## Known Stubs

None — all acceptance criteria met, no placeholder data.

---

## Threat Surface Scan

All security mitigations from the plan's threat model are implemented:

| Threat ID | Mitigation | Location |
|-----------|-----------|---------|
| T-03-01 | Path traversal guard: `strings.Contains(rel, "..")` + `filepath.Join` boundary check | `extractSubdir()` in github_release.go |
| T-03-02 | HTTPS-only URL construction: URL hardcoded as `https://github.com/...` | `Install()` in github_release.go |
| T-03-04 | Repo field validated non-empty before use | `Install()` in github_release.go |
| T-03-05 | Files extracted to `spec.SkillDir` only (set by `SkillInstallPath`) | service/install.go SkillDir population |

No new security-relevant surface introduced beyond the planned threat model.

---

## Self-Check

### Files Created/Modified
- [x] `internal/installer/github_release.go` — FOUND
- [x] `internal/installer/github_release_test.go` — FOUND
- [x] `internal/installer/github_release_integration_test.go` — FOUND
- [x] `internal/domain/package.go` — contains `InstallMethodGitHubRelease`, `json:"-"` on SkillDir
- [x] `internal/config/paths.go` — contains `func TarballCachePath`
- [x] `internal/installer/installer.go` — contains `ErrGitHubReleaseNotFound`, `InstallMethodGitHubRelease` case

### Commits
- [x] 62594a0 — feat(03-01): extend domain types and add TarballCachePath
- [x] a097836 — feat(03-01): implement GitHubReleaseInstaller with path-traversal guard, caching, validator wiring
- [x] cb245b6 — test(03-01): integration test for end-to-end extraction to SkillInstallPath

## Self-Check: PASSED
