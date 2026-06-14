---
phase: 03-bundled-skills
verified: 2026-06-09T00:00:00Z
status: passed
score: 9/10 must-haves verified (1 deferred: gsd-core external registry)
overrides_applied: 0
passed_at: 2026-06-14T00:00:00Z
uat_evidence: "bundle cloud/dev/context all 3/3; aws skill dir SKILL.md+3 refs confirmed; all 9 skills pass validator 0 blocking errors; install.sh works end-to-end"
human_verification:
  - test: "Run `agentkit install --bundle cloud` (requires published GitHub release tag)"
    expected: "aws, gcp, azure skills installed to ~/.claude/skills/; each passes validator; success lines printed; exit 0"
    why_human: "D-17 gate — gsd-core registry 404; cloud bundle installs from agentkit-registry whose packages use github-release method requiring a published tarball; can't verify without a real release tag"
  - test: "Run `agentkit install gsd` (requires gsd-core registry to publish registry.json)"
    expected: "GSD suite installed in one command; open-gsd/gsd-core registry resolves the gsd package"
    why_human: "D-17 gate confirmed in cmd/install.go comment: open-gsd/gsd-core returns HTTP 404 at time of verification; code is pre-wired but registry not yet published"
  - test: "Verify installed aws skill: ls ~/.claude/skills/aws/ shows SKILL.md, references/ec2.md, references/s3.md, references/iam.md"
    expected: "All 4 paths exist after install"
    why_human: "Requires live tarball extraction from a published GitHub release"
  - test: "Run `agentkit install --bundle cloud` then `agentkit install --bundle dev` then `agentkit install --bundle context`; confirm all 9 bundled skills pass the install-time validator"
    expected: "All 9 skills install without errors; warnings (if any) logged to stderr; installs succeed"
    why_human: "Requires live github-release tarball download"
---

# Phase 3: Bundled Skills Verification Report

**Phase Goal:** Users can install curated skill bundles with one command, and each of the 10 initial skills is authored to the agentskills.io spec and validated at install time.
**Verified:** 2026-06-09T00:00:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `agentkit install --bundle cloud/dev/context` installs packages in parallel and reports per-package results | ✓ VERIFIED | `cmd/install.go` runBundleInstall() uses sync.WaitGroup, iterates bundle packages, prints `name ✓` or `name ✗ err`, exits 1 on failure |
| 2 | bundles.json defines cloud (aws/gcp/azure), dev (playwright/github/cicd), context (context-mode/rtk/serena) | ✓ VERIFIED | `internal/bundle/bundles.json` lines 1-7 match exactly; LoadBundles/Resolve wired in bundles.go |
| 3 | Each of the 9 bundled skills has SKILL.md with name+description frontmatter, under 500 lines, no placeholder text | ✓ VERIFIED | All 9 SKILL.md files confirmed present; line counts: aws=122, gcp=130, azure=120, playwright=144, github=139, cicd=135, context-mode=99, rtk=77, serena=103; all have `name:` and `description:` fields; no TODO/stub/placeholder found |
| 4 | ValidateSkill called during github-release install | ✓ VERIFIED | `internal/service/install.go` lines 141-156: validator called with pkg.Install.SkillDir for github-release; errors block install; warnings log to stderr |
| 5 | WriteSkill guard prevents overwriting extracted skill files | ✓ VERIFIED | `internal/service/install.go` line 161: `if pkg.Install.Method != domain.InstallMethodGitHubRelease { ... WriteSkill ... }` |
| 6 | GitHubReleaseInstaller implements extraction, path-traversal guard, caching | ✓ VERIFIED | `internal/installer/github_release.go`: path traversal guard at lines 220-226 (strings.Contains("..") + filepath boundary check); sync.Map cache; disk cache via TarballCachePath; ErrGitHubReleaseNotFound on 404 |
| 7 | SkillDir populated before Install() call in service.Install() | ✓ VERIFIED | `internal/service/install.go` lines 127-133: SkillDir set via config.SkillInstallPath(target, name) before inst.Install() |
| 8 | External skills: 11 skills under skills/external/ with attribution headers and name frontmatter | ✓ VERIFIED | 11 skills found; all have `# [Name] (via [org/repo])` H1 header; all have `name:` frontmatter; none exceed 500 lines (max: react-native-skills 240) |
| 9 | Live bundle install succeeds end-to-end with tarball extraction | ? UNCERTAIN | D-17 gate: no published GitHub release tarball exists; integration tests pass against httptest server (5 tests pass), but live install not verifiable |
| 10 | `agentkit install gsd` routes through gsd-core registry | ? UNCERTAIN | D-17 gate: gsd-core registry.json at open-gsd/gsd-core returns HTTP 404; code pre-wired (comment documented in cmd/install.go line 1-6); will auto-resolve when repo published |

**Score:** 8/10 truths verified (2 blocked by D-17 gate — human verification required after release)

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/installer/github_release.go` | GitHubReleaseInstaller with extraction, path-traversal guard, caching | ✓ VERIFIED | 252 lines; NewGitHubReleaseInstaller + NewGitHubReleaseInstallerWithClient; extractSubdir with traversal guard |
| `internal/installer/github_release_test.go` | Unit tests: extract, path traversal, cache hit, 404 | ✓ VERIFIED | File exists; `go test ./internal/installer/... -run TestGitHubRelease` — 5 passed |
| `internal/installer/github_release_integration_test.go` | Integration test: end-to-end extraction to SkillInstallPath | ✓ VERIFIED | File exists; TestGitHubReleaseInstaller_ExtractToSkillPath passes |
| `internal/domain/package.go` | InstallMethodGitHubRelease constant, Repo/Path fields, SkillDir with json:"-" | ✓ VERIFIED | line 20: `InstallMethodGitHubRelease = "github-release"`; fields Repo/Path with omitempty; SkillDir with `json:"-"` |
| `internal/config/paths.go` | TarballCachePath function | ✓ VERIFIED | Line 42: `func TarballCachePath(repo, version string)` |
| `internal/installer/installer.go` | ErrGitHubReleaseNotFound sentinel + factory case | ✓ VERIFIED | Line 23-24: ErrGitHubReleaseNotFound; line 50: InstallMethodGitHubRelease case in switch |
| `internal/bundle/bundles.go` | LoadBundles(), Resolve(), BundleDef/BundleManifest | ✓ VERIFIED | 42 lines; all exports present |
| `internal/bundle/bundles.json` | cloud/dev/context bundles | ✓ VERIFIED | 7 lines; exact match to spec |
| `cmd/install.go` | --bundle flag, runBundleInstall(), cobra.RangeArgs(0,1) | ✓ VERIFIED | Line 44: --bundle flag; runBundleInstall() at line 136; RangeArgs(0,1) line 38 |
| `skills/aws/SKILL.md` + `references/ec2.md`, `s3.md`, `iam.md` | AWS skill with 3 references | ✓ VERIFIED | All 4 files present |
| `skills/gcp/SKILL.md` + references (compute/gke/cloudrun/iam) | GCP skill with 4 references | ✓ VERIFIED | All 5 files present |
| `skills/azure/SKILL.md` + references (vms/aks/appservice) | Azure skill with 3 references | ✓ VERIFIED | All 4 files present |
| `skills/playwright/SKILL.md` + `references/e2e-patterns.md` | Playwright skill | ✓ VERIFIED | Both files present |
| `skills/github/SKILL.md` + references (prs/issues/actions) | GitHub skill | ✓ VERIFIED | All 4 files present |
| `skills/cicd/SKILL.md` + references (github-actions/pipelines/deployments) | CI/CD skill | ✓ VERIFIED | All 4 files present |
| `skills/context-mode/SKILL.md` + `references/routing-rules.md` | context-mode skill | ✓ VERIFIED | Both files present |
| `skills/rtk/SKILL.md` + `references/commands.md` | RTK skill | ✓ VERIFIED | Both files present |
| `skills/serena/SKILL.md` + `references/lsp-usage.md` | Serena skill | ✓ VERIFIED | Both files present |
| `skills/skill-author/SKILL.md` + 3 refs + scripts/ | skill-author meta-skill | ✓ VERIFIED | All files present including validate-skill.sh |
| `agents/auto-researcher/AGENT.md` | auto-researcher agent | ✓ VERIFIED | File exists (159 lines) |
| `skills/external/` (11 skills) | 11 external skills with attribution | ✓ VERIFIED | 11 skills found under skills/external/ |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `cmd/install.go` | `internal/bundle/bundles.go` | `bundle.LoadBundles()` call in runBundleInstall | ✓ WIRED | Line 137: `bundle.LoadBundles()` |
| `cmd/install.go` | `internal/service/install.go` | `svc.Install()` per package in goroutine | ✓ WIRED | Line 153: `svc.Install(n, target)` |
| `internal/installer/installer.go` | `internal/installer/github_release.go` | NewInstaller factory case | ✓ WIRED | Line 50: `domain.InstallMethodGitHubRelease -> NewGitHubReleaseInstaller()` |
| `internal/service/install.go` | `internal/domain/package.go` | WriteSkill guard | ✓ WIRED | Line 161: `if pkg.Install.Method != domain.InstallMethodGitHubRelease` |
| `internal/service/install.go` | `internal/skill/validate.go` | ValidateSkill call after extraction | ✓ WIRED | Line 148: `result := s.validator(validationDir, pkg)` — validator is `skill.ValidateSkill` |
| `skills/skill-author/SKILL.md` | `agents/auto-researcher/AGENT.md` | Reference in SKILL.md body | ✓ WIRED | Lines 24,77: `auto-researcher` mentioned in SKILL.md |

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full test suite | `go test ./...` | 134 passed, 0 failed | ✓ PASS |
| GitHub release unit tests | `go test ./internal/installer/... -run TestGitHubRelease` | 5 passed | ✓ PASS |
| Bundle loader tests | `go test ./internal/bundle/...` | 3 passed | ✓ PASS |
| Full build | `go build ./...` | Exit 0, no errors | ✓ PASS |
| Live bundle install | Requires published tarball | Blocked by D-17 gate | ? SKIP |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|---------|
| CLI-03 | 03-02 | `agentkit install gsd` one-command install | ? NEEDS HUMAN | D-17 gate: gsd-core registry returns 404; code pre-wired, auto-resolves when published |
| CLI-04 | 03-02 | `agentkit install --bundle <name>` | ✓ SATISFIED | --bundle flag, runBundleInstall(), bundles.json all verified |
| BND-01 | 03-01, 03-06 | AWS skill + references | ✓ SATISFIED | aws/SKILL.md + ec2/s3/iam.md all present; 122 lines; valid frontmatter |
| BND-02 | 03-03 | GCP skill + references | ✓ SATISFIED | gcp/SKILL.md + compute/gke/cloudrun/iam.md all present |
| BND-03 | 03-03 | Azure skill + references | ✓ SATISFIED | azure/SKILL.md + vms/aks/appservice.md all present |
| BND-04 | 03-03 | Playwright skill + MCP entry | ✓ SATISFIED | playwright/SKILL.md + e2e-patterns.md present |
| BND-05 | 03-03 | GitHub skill + references | ✓ SATISFIED | github/SKILL.md + prs/issues/actions.md present |
| BND-06 | 03-03 | CI/CD skill + references | ✓ SATISFIED | cicd/SKILL.md + github-actions/pipelines/deployments.md present |
| BND-07 | 03-04 | context-mode skill | ✓ SATISFIED | context-mode/SKILL.md + routing-rules.md present; 99 lines |
| BND-08 | 03-04 | RTK skill | ✓ SATISFIED | rtk/SKILL.md + commands.md present; 77 lines |
| BND-09 | 03-04 | Serena skill | ✓ SATISFIED | serena/SKILL.md + lsp-usage.md present; 103 lines |

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None found | — | — | — | No debt markers, no stubs, no placeholder text in phase-modified files |

---

### Human Verification Required

#### 1. Live Bundle Install (D-17 gate)

**Test:** Run `agentkit install --bundle cloud --target claude` with a published v0.1.0 release tag
**Expected:** aws, gcp, azure skills extracted to `~/.claude/skills/`; each passes ValidateSkill; exit code 0; 3/3 installed printed
**Why human:** The github-release installer requires a real tarball at `https://github.com/ejyle/agentkit/archive/refs/tags/v0.1.0.tar.gz`. The integration tests pass against a local httptest server, but the D-17 gate confirms no such release exists yet.

#### 2. GSD One-Command Install (D-17 gate)

**Test:** Run `agentkit install gsd` once gsd-core registry.json is published at `https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json`
**Expected:** GSD suite installs from gsd-core registry in one command; success line printed
**Why human:** The code comment in cmd/install.go explicitly documents D-17 verification result: open-gsd/gsd-core returns HTTP 404.

#### 3. AWS Skill Directory Contents After Install

**Test:** After a successful `agentkit install aws --target claude`, run `ls ~/.claude/skills/aws/`
**Expected:** `SKILL.md`, `references/ec2.md`, `references/s3.md`, `references/iam.md` all present (ROADMAP Success Criterion 4)
**Why human:** Requires live tarball extraction; integration test validates the extraction logic with a mock tarball.

#### 4. Install-time Validator for All 9 Bundled Skills

**Test:** Install all 9 bundled skills and confirm each passes ValidateSkill without blocking errors
**Expected:** 0 blocking errors; any line-count warnings logged to stderr; installs succeed
**Why human:** Requires live tarball extraction from published release.

---

### Gaps Summary

No blockers found. All code-verifiable must-haves pass. The 2 uncertain truths (CLI-03 gsd install, SC-1/2/4 live install) are gated by D-17 — no published GitHub release exists yet. This is a documented, expected gate: the code is fully wired and tested against mock servers; the only missing piece is a published `v0.1.0` tag, which is Phase 4's responsibility (GoReleaser pipeline).

---

_Verified: 2026-06-09T00:00:00Z_
_Verifier: Claude (gsd-verifier)_
