# Phase 3: Bundled Skills - Research

**Researched:** 2026-06-09
**Domain:** Go tarball installer, bundle parallelism, skill content authoring, registry integration
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Skills live in this repo under `skills/` (not embedded in binary). At install time, agentkit fetches the GitHub release tarball for the current binary version and extracts the relevant skill subdirectory.
- **D-02:** New install method `"github-release"` in registry.json. Format: `{ "install": { "method": "github-release", "repo": "ejyle/agentkit", "path": "skills/aws" } }`. Installer resolves tarball URL from current binary version tag, downloads once per batch, extracts `path` to target skill directory.
- **D-03:** `github-release` installer is generic — works for any GitHub repo, not agentkit-specific.
- **D-04:** Bundle install is best-effort and parallel. All packages in a bundle start concurrently. Failures collected; command reports at completion. No rollback.
- **D-05:** Exit code 1 if any package failed; 0 if all succeeded.
- **D-06:** Bundle definitions live in `internal/bundle/bundles.json` — not hardcoded. Adding a bundle does not require CLI code change.
- **D-07:** Bundles: `cloud` (aws, gcp, azure), `dev` (playwright, github, cicd), `context` (context-mode, rtk, serena). GSD is standalone, not a bundle member.
- **D-08:** All 10 skills are full reference-quality — real SKILL.md (<500 lines) + domain reference files + optional scripts. No stubs.
- **D-09:** context-mode, RTK, Serena (BND-07, BND-08, BND-09) adapted from existing `~/.claude/skills/` installs. Reviewed and updated to agentskills.io spec.
- **D-10:** 10th skill: skill-author meta-skill — SKILL.md entrypoint + references/ + scripts/ + bundled auto-researcher agent. NOT deferred.
- **D-11:** Auto-researcher agent lives in `agents/auto-researcher/` in this repo; referenced from skill-author meta-skill.
- **D-12:** External skill content copied into `skills/external/`. Sources: anthropics/skills, vercel-labs/agent-skills, vercel-labs/agent-browser, skills.sh.
- **D-13:** Researcher evaluates each candidate for quality, agentskills.io spec compliance, license compatibility.
- **D-14:** External skill update mechanism deferred to v0.2.0. Phase 3 does initial copy only.
- **D-15:** Researcher to discover and recommend final list of 10–12 external skills.
- **D-16:** `agentkit install gsd` does NOT bundle GSD content here; routes through gsd-core registry (REG-02, already integrated).
- **D-17:** Researcher MUST verify actual `gsd` entry install method from `open-gsd/gsd-core/registry.json` before implementing.

### Claude's Discretion

- Exact `skills/` directory structure layout (flat vs nested by category)
- Whether `github-release` installer caches tarball across multiple skill installs in same session
- Spinner output format for parallel bundle installs (per-package lines vs combined progress)
- Exact agentkit-registry entries for external skills (names, descriptions, version tags)

### Deferred Ideas (OUT OF SCOPE)

- External skill update mechanism (`agentkit skill sync <name>`) — v0.2.0
- Full automated skill publishing pipeline (approval gating + registry PR automation) — v0.2.0
- `agentkit skill validate <path>` and `agentkit skill improve <path>` as CLI subcommands — v0.2.0
- Additional bundle definitions — v0.2.0
- Background config agent — v0.2.0

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CLI-03 | User can install full GSD workflow suite in one command (`agentkit install gsd`) | GSD routes through gsd-core registry (already integrated); D-17 research needed for install method verification |
| CLI-04 | User can bulk-install a preset bundle (`agentkit install --bundle <cloud\|gsd\|context\|dev>`) | Bundle command mechanics, parallel goroutine pattern, bundles.json config |
| BND-01 | AWS skill: EC2, S3, IAM, ECS/EKS; SKILL.md + references/ec2.md, s3.md, iam.md | Content plan: 3 reference files, detect-aws-env script |
| BND-02 | GCP skill: Compute Engine, GKE, Cloud Run, IAM via gcloud; SKILL.md + domain references | Content plan: 4 reference files (compute.md, gke.md, cloudrun.md, iam.md) |
| BND-03 | Azure skill: VMs, AKS, App Service via az CLI; SKILL.md + domain references | Content plan: 3 reference files (vms.md, aks.md, appservice.md) |
| BND-04 | Playwright skill: browser automation + E2E testing; SKILL.md + MCP server entry | Playwright MCP already in registry (npx @playwright/mcp); skill wraps it + adds testing patterns |
| BND-05 | GitHub skill: PR management, issues, Actions via gh CLI; SKILL.md + references | Content plan: references/prs.md, issues.md, actions.md |
| BND-06 | CI/CD skill: GitHub Actions, build pipelines, deploy workflows; SKILL.md + references | Content plan: references/github-actions.md, pipelines.md, deployments.md |
| BND-07 | context-mode skill: sandbox routing, context window protection | Adapt from existing `~/.claude/skills/context-mode/` install (confirmed exists via CLAUDE.md) |
| BND-08 | RTK skill: token-optimized CLI proxy for dev operations | Adapt from `~/.claude/skills/rtk/` (RTK.md in global CLAUDE.md references it) |
| BND-09 | Serena skill: LSP-powered code intelligence and symbol navigation | Adapt from existing Serena skill install |

</phase_requirements>

---

## Summary

Phase 3 has three parallel workstreams: (1) a new `GitHubReleaseInstaller` that fetches release tarballs and extracts skill subdirectories using Go stdlib `archive/tar` + `compress/gzip`, (2) a `--bundle` command that runs parallel `svc.Install()` calls and collects all results before printing a summary, and (3) 10+ skill files authored into `skills/` and `skills/external/` in this repo.

The codebase is clean and well-factored. All three installer patterns (`NpxInstaller`, `BinaryInstaller`, `CustomInstaller`) follow identical interfaces. The new `GitHubReleaseInstaller` slots in as a 6th case in `NewInstaller()`. The bundle command is a new `--bundle` flag on `installCmd` that bypasses the single-package path entirely and drives parallel goroutines with error collection.

The critical D-17 finding: the gsd-core `registry.json` is not directly accessible via web search in raw form. Based on what's known from GSD install documentation — GSD uses `npx @opengsd/gsd-core@latest --antigravity --global` — the install method in gsd-core's registry.json is almost certainly `"custom"` or `"npx"`. This is tagged `[ASSUMED]` and must be verified before the `github-release` installer is built, because `agentkit install gsd` must match whatever the gsd-core registry entry specifies. If gsd-core's entry uses `"custom"`, agentkit routes it through `CustomInstaller` — no new installer needed for the GSD path.

**Primary recommendation:** Implement `GitHubReleaseInstaller` first (tarball cache + subdirectory extraction), then `--bundle` command, then author all 10 skills. The GSD path (CLI-03) will work automatically through the existing registry manager if gsd-core's entry uses `"custom"` or `"npx"` — verify this before writing any code for CLI-03.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| GitHub release tarball fetch | installer layer (`internal/installer/github_release.go`) | config layer (cache path) | Same layer as BinaryInstaller; cache path uses existing `config.ManifestCachePath` pattern |
| Bundle definitions | config file (`internal/bundle/bundles.json`) | cmd layer (reads file) | D-06: data-driven, not hardcoded |
| Parallel bundle installs | cmd layer (`cmd/install.go`) | service layer | cmd already owns install orchestration; service.Install is already concurrency-safe |
| Skill validation at install | service layer (`internal/service/install.go`) | skill layer | Already done in `Install()` for skill-type packages; github-release installer triggers it |
| Tarball caching | installer layer | `os.UserCacheDir()` | Session-scope in-memory map OR disk cache under `~/.cache/agentkit/releases/<version>/` |
| Skill file extraction to disk | installer layer | adapter layer | GitHubReleaseInstaller extracts to temp dir; adapter's WriteSkill copies files to final path |
| Bundle UI (per-package status) | cmd layer | ui layer | New multi-line model or simple sequential print; bubbletea model optional |

---

## 1. GitHub Release Installer

### Tarball URL Pattern [ASSUMED]

GitHub release tarball URLs follow this pattern:
```
https://github.com/<owner>/<repo>/archive/refs/tags/<version>.tar.gz
```
Example: `https://github.com/ejyle/agentkit/archive/refs/tags/v0.1.0.tar.gz`

The extracted archive root will be `<repo>-<version-without-v>/` (e.g., `agentkit-0.1.0/`). When the registry entry specifies `"path": "skills/aws"`, the installer strips the archive root prefix and extracts only entries whose path starts with `agentkit-0.1.0/skills/aws/`.

**Alternative asset URL** (for explicit release assets, not source archives):
```
https://github.com/<owner>/<repo>/releases/download/<version>/<asset-name>.tar.gz
```
This is used when the release includes explicit build artifacts. For skill delivery, using the source archive tarball is simpler — no explicit asset upload needed.

**Recommendation (Claude's Discretion):** Use source archive URL. Skills are source content, not build artifacts. No goreleaser changes needed to publish them as separate assets — they're included naturally in the source archive. [ASSUMED]

### Go stdlib tarball extraction [VERIFIED: pkg.go.dev]

Use `archive/tar` + `compress/gzip` from stdlib — zero new dependencies.

```go
// Source: https://pkg.go.dev/archive/tar
func extractSkillSubdir(tarball []byte, archivePrefix, skillPath, destDir string) error {
    gr, err := gzip.NewReader(bytes.NewReader(tarball))
    if err != nil {
        return fmt.Errorf("gzip open: %w", err)
    }
    defer gr.Close()
    tr := tar.NewReader(gr)
    
    // archivePrefix = "agentkit-0.1.0/"
    // skillPath = "skills/aws/"
    fullPrefix := archivePrefix + skillPath
    
    for {
        hdr, err := tr.Next()
        if err == io.EOF { break }
        if err != nil { return fmt.Errorf("tar next: %w", err) }
        
        if !strings.HasPrefix(hdr.Name, fullPrefix) { continue }
        
        rel := strings.TrimPrefix(hdr.Name, fullPrefix)
        if rel == "" { continue }
        
        // Path traversal guard
        if strings.Contains(rel, "..") {
            return fmt.Errorf("path traversal in archive: %q", hdr.Name)
        }
        
        outPath := filepath.Join(destDir, filepath.FromSlash(rel))
        switch hdr.Typeflag {
        case tar.TypeDir:
            os.MkdirAll(outPath, 0755)
        case tar.TypeReg:
            os.MkdirAll(filepath.Dir(outPath), 0755)
            f, _ := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0755)
            io.Copy(f, tr)
            f.Close()
        }
    }
    return nil
}
```

### Tarball Cache Strategy (Claude's Discretion)

The CONTEXT.md specifies: `~/.cache/agentkit/releases/<version>/tarball.tar.gz`

**Recommended:** In-process cache (sync.Map keyed by `repo@version`) for the duration of a bundle install. Also write to disk at `os.UserCacheDir()/agentkit/releases/<repo-slug>/<version>/tarball.tar.gz` so repeated standalone installs don't re-download.

```go
// Cache path: ~/.cache/agentkit/releases/ejyle-agentkit/v0.1.0/tarball.tar.gz
func tarballCachePath(repo, version string) (string, error) {
    base, err := os.UserCacheDir()
    if err != nil { return "", err }
    slug := strings.ReplaceAll(repo, "/", "-")
    return filepath.Join(base, "agentkit", "releases", slug, version, "tarball.tar.gz"), nil
}
```

### Version Resolution

The `github-release` installer needs to know the current binary version. Two options:
- **Option A:** Pass version in `InstallSpec` via a new `Version` field. Registry entry specifies `"version": "latest"` and installer resolves to binary version at runtime.
- **Option B:** Installer reads a package-level `var Version string` set by goreleaser ldflags.

**Recommendation:** Option B — version as ldflags var is the standard Go release pattern and avoids adding a field to `InstallSpec`. [ASSUMED]

---

## 2. Bundle Command

### Parallelism Pattern

`errgroup` from `golang.org/x/sync/errgroup` provides cleaner API than `sync.WaitGroup + []error` for cases where you want to stop early. However, D-04 requires **best-effort** (no cancellation on first error) — so `errgroup` with context cancellation is NOT the right choice here.

**Correct pattern for D-04 (best-effort, collect all errors):** [VERIFIED: pkg.go.dev]

```go
// Source: stdlib sync + golang.org/x/sync — but NOT errgroup (errgroup cancels on first error)
// Use sync.WaitGroup + mutex-protected []BundleResult
type BundleResult struct {
    Name string
    Err  error
    Pkg  *domain.Package
}

func runBundleInstall(pkgNames []string, target string, svc *service.InstallService) []BundleResult {
    results := make([]BundleResult, len(pkgNames))
    var wg sync.WaitGroup
    for i, name := range pkgNames {
        wg.Add(1)
        go func(idx int, n string) {
            defer wg.Done()
            pkg, err := svc.Install(n, target)
            results[idx] = BundleResult{Name: n, Err: err, Pkg: pkg}
        }(i, name)
    }
    wg.Wait()
    return results
}
```

Note: `results[idx]` write is safe because each goroutine writes to a unique index — no mutex needed.

**Do NOT use `golang.org/x/sync/errgroup`** for this phase. errgroup cancels the context on first error — the opposite of D-04 semantics. `sync.WaitGroup` is in stdlib (no new dependency). [ASSUMED based on D-04 semantics]

### UI / Spinner Approach

The existing `SpinnerModel` is single-phase (one phase string at a time). For bundle installs showing per-package status, two options:

**Option A (simple, recommended):** Bypass bubbletea entirely for bundle mode. Print package names as they complete using a simple channel, output `  aws ✓`, `  gcp ✓`, `  azure ✗ <reason>` lines sequentially to stdout/stderr. Final line: `2/3 installed — azure failed: <reason>`.

**Option B (richer):** New `BundleProgressModel` bubbletea model that maintains a `[]BundleRow` slice and renders per-package spinner lines. More complex but visually cleaner.

**Recommendation:** Option A for Phase 3. Bubbletea multi-line progress models have lifecycle complexity (goroutine synchronization via `tea.Program.Send`) — Option A ships faster and satisfies the CONTEXT.md UI hint verbatim. [ASSUMED]

### bundles.json Schema

```json
{
  "bundles": {
    "cloud":   { "packages": ["aws", "gcp", "azure"] },
    "dev":     { "packages": ["playwright", "github", "cicd"] },
    "context": { "packages": ["context-mode", "rtk", "serena"] }
  }
}
```

Location: `internal/bundle/bundles.json` (embedded via `//go:embed` for zero-dependency access). [ASSUMED — embed is stdlib since Go 1.16]

### cmd/install.go changes

The `--bundle` flag changes `Args: cobra.ExactArgs(1)` to `cobra.RangeArgs(0, 1)` — when `--bundle` is set, the positional arg is not required.

```go
// In init():
installCmd.Flags().StringP("bundle", "b", "", "Install a preset bundle (cloud, dev, context)")

// In runInstall():
bundleName, _ := cmd.Flags().GetString("bundle")
if bundleName != "" {
    return runBundleInstall(cmd, bundleName, target)
}
// ... existing single-package path unchanged
```

---

## 3. GSD Registry Entry (D-17)

### Finding

The raw gsd-core registry.json at `https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json` could not be fetched directly (WebFetch is blocked per CLAUDE.md). Web search reveals that GSD installs via `npx @opengsd/gsd-core@latest --antigravity --global` [CITED: github.com/open-gsd/gsd-core install docs].

**Most likely registry entry format [ASSUMED — must be verified at implementation time]:**

```json
{
  "name": "gsd",
  "version": "1.2.0",
  "description": "Git. Ship. Done — full GSD workflow suite",
  "type": "skill",
  "install": {
    "method": "custom",
    "package": "npx",
    "args": ["@opengsd/gsd-core@latest", "--antigravity", "--global"]
  }
}
```

OR it may use method `"npx"` with the package string including flags.

**Action required before implementing CLI-03:** The implementer MUST run:
```bash
# In a network-enabled environment (not via WebFetch):
curl -s https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json | jq '.packages[] | select(.name == "gsd")'
```
If the entry uses `"custom"` or `"npx"` — no new installer is needed; `agentkit install gsd` works today. If it uses `"github-release"` — the new installer is needed for GSD too.

**Implication:** CLI-03 (`agentkit install gsd`) may be a no-op implementation task — the gsd-core registry is already integrated in `NewRegistryManager()`. It may simply work once the gsd-core registry.json has the correct entry.

---

## 4. Skills Directory Layout

### Recommended Structure (Claude's Discretion)

```
skills/
├── aws/
│   ├── SKILL.md
│   └── references/
│       ├── ec2.md
│       ├── s3.md
│       └── iam.md
├── gcp/
│   ├── SKILL.md
│   └── references/
│       ├── compute.md
│       ├── gke.md
│       ├── cloudrun.md
│       └── iam.md
├── azure/
│   ├── SKILL.md
│   └── references/
│       ├── vms.md
│       ├── aks.md
│       └── appservice.md
├── playwright/
│   ├── SKILL.md
│   └── references/
│       └── e2e-patterns.md
├── github/
│   ├── SKILL.md
│   └── references/
│       ├── prs.md
│       ├── issues.md
│       └── actions.md
├── cicd/
│   ├── SKILL.md
│   └── references/
│       ├── github-actions.md
│       ├── pipelines.md
│       └── deployments.md
├── context-mode/
│   ├── SKILL.md
│   └── references/
│       └── routing-rules.md
├── rtk/
│   ├── SKILL.md
│   └── references/
│       └── commands.md
├── serena/
│   ├── SKILL.md
│   └── references/
│       └── lsp-usage.md
└── skill-author/
    ├── SKILL.md
    ├── references/
    │   ├── evaluation-rubric.md
    │   ├── spec-compliance.md
    │   └── authoring-guide.md
    └── scripts/
        └── validate-skill.sh

skills/external/
├── frontend-design/          # from anthropics/skills
│   ├── SKILL.md
│   └── references/
├── vercel-react/              # from vercel-labs/agent-skills
│   ├── SKILL.md
│   └── references/
├── agent-browser/             # from vercel-labs/agent-browser
│   └── SKILL.md
└── ...

agents/
└── auto-researcher/
    └── AGENT.md (or similar entry point)
```

**Flat structure per category** (no nested `bundled/` vs `external/` at the top level within each — that's already handled by `skills/` vs `skills/external/`).

### GoReleaser Archive Configuration

Phase 4 owns GoReleaser setup, but Phase 3 must ensure `skills/` is included in the release tarball. If using a source archive (recommended), skills are included automatically — no config change. If Phase 4 uses asset-based archives, add to `.goreleaser.yaml`: [CITED: goreleaser.com/customization/archive/]

```yaml
archives:
  - files:
      - skills/**
      - agents/**
```

**Recommendation:** Document this as a Phase 4 concern. Phase 3 delivers the `skills/` content; Phase 4 wires the release pipeline. For development/testing, the installer reads from a local file path (via `AGENTKIT_REGISTRY_FILE` pattern) or from a test server. [ASSUMED]

---

## 5. External Skill Candidates (D-12 to D-15)

### Source Assessment

| Source | Quality | License | Spec Compliance | Notes |
|--------|---------|---------|-----------------|-------|
| anthropics/skills | HIGH | MIT [ASSUMED] | Partial — uses SKILL.md but may lack references/ structure | skill-creator skill present; spec defined in `spec/agent-skills-spec.md` |
| vercel-labs/agent-skills | HIGH | MIT [ASSUMED] | Partial — AGENTS.md focus, may need SKILL.md adaptation | Contains react-best-practices, ui-review, writing-handbook |
| vercel-labs/agent-browser | HIGH | MIT [ASSUMED] | Partial — discovery stub pattern, thin SKILL.md | agent-browser skill: CDP-based, excellent for AI agents |
| skills.sh | MIXED | Varies | Many lack references/ | 200+ skills; top picks by install count; security audit found 6.3 issues/skill avg — curate carefully |

**Security note on skills.sh:** A 2026 audit found prompt injection in 36% of tested skills [CITED: toolworthy.ai/tool/skills-sh]. Any skill copied from skills.sh must be manually reviewed for injected instructions.

### Recommended 10–12 External Skills

| Skill Name | Source | Source Path | Rationale | Compliance Work Needed |
|------------|--------|-------------|-----------|----------------------|
| `frontend-design` | anthropics/skills | `skills/frontend-design/` | Official Anthropic, high quality, UI/CSS expertise | Add references/ structure |
| `skill-creator` | anthropics/skills | `skills/skill-creator/` | Meta-skill; complements skill-author; official source | Merge/adapt with our skill-author |
| `vercel-react` | vercel-labs/agent-skills | `react-best-practices/` | 40+ React/Next.js rules from Vercel engineering; high value | Add SKILL.md frontmatter, split into references/ |
| `ui-review` | vercel-labs/agent-skills | skill in repo | 100+ UI/accessibility rules; high value | Rename, add frontmatter |
| `agent-browser` | vercel-labs/agent-browser | `skills/agent-browser/` | CDP browser automation; complements Playwright | Thin SKILL.md already exists; verify references/ |
| `vercel-deploy` | vercel-labs/agent-skills | vercel-deploy skill | Vercel-specific deployment; high usage | Adapt |
| `react-native` | vercel-labs/agent-skills | react-native skill | Mobile dev; 16 rules across 7 sections | Adapt |
| `docs-writing` | vercel-labs/agent-skills | writing-handbook skill | Technical writing; high value for doc tasks | Adapt |
| `docker` | skills.sh top list | top-installed skill | Container skills; high demand | Review for injection; add references/ |
| `terraform` | skills.sh top list | top-installed skill | IaC; complements cloud bundle | Review for injection; add references/ |
| `python-best-practices` | skills.sh top list | top-installed skill | Language-specific; high demand | Review for injection |
| `sql-query` | skills.sh top list | top-installed skill | Database queries; universal need | Review for injection |

**Total: 12 external skills.** All content copied into `skills/external/` with header `# [Name] (via [source-org]/[source-repo])`.

---

## 6. The 10 Bundled Skills — Content Plan

### agentskills.io Spec (SKILL.md frontmatter) [CITED: agentskills.io/specification]

```yaml
---
name: aws          # matches folder name; lowercase + hyphens; max 64 chars
description: >     # tells agent WHEN to activate; required
  Use when working with AWS infrastructure — EC2, S3, IAM, ECS, or EKS.
license: Apache-2.0  # optional
---
```
Body: Markdown instructions < 500 lines. Heavy content in `references/`.

### BND-01: AWS Skill

- **SKILL.md body:** Overview, activation triggers, quick commands by service, when to load which reference
- **references/ec2.md:** Instance lifecycle, AMIs, security groups, userdata, SSM Session Manager
- **references/s3.md:** Bucket ops, presigned URLs, lifecycle policies, cross-account access, encryption
- **references/iam.md:** Policy authoring, roles, assume-role, permission boundaries, AWS SSO
- **Install.Args in registry:** `["ec2", "s3", "iam"]` — ValidateSkill checks these exist
- **Optional scripts/detect-aws-env.sh:** Checks `aws sts get-caller-identity` to confirm credentials

### BND-02: GCP Skill

- **SKILL.md body:** gcloud commands, activation triggers, project/region context
- **references/compute.md:** VM lifecycle, images, networking, startup scripts
- **references/gke.md:** Cluster create/update, node pools, workload identity, kubectl integration
- **references/cloudrun.md:** Service deploy, traffic splits, secrets, VPC connector
- **references/iam.md:** Service accounts, IAM bindings, Workload Identity Federation

### BND-03: Azure Skill

- **SKILL.md body:** az CLI overview, subscription context, activation triggers
- **references/vms.md:** VM create/manage, extensions, scale sets, bastion
- **references/aks.md:** Cluster ops, node pools, RBAC, Azure AD integration
- **references/appservice.md:** Web apps, slots, autoscale, deployment pipelines

### BND-04: Playwright Skill

- **SKILL.md body:** When to use (E2E testing, browser automation), how to launch via MCP
- **references/e2e-patterns.md:** POM patterns, assertion strategies, async/await, fixtures
- **MCP entry in registry:** `{ "command": "npx", "args": ["@playwright/mcp@latest"] }` — already in testdata registry
- **Note:** Playwright MCP server is an npx install; the skill teaches usage patterns, not re-installs the server

### BND-05: GitHub Skill

- **SKILL.md body:** `gh` CLI activation triggers, PR review workflow, issue triage
- **references/prs.md:** Create, review, merge, rebase, conflict resolution via gh
- **references/issues.md:** Labels, milestones, filters, bulk operations
- **references/actions.md:** Workflow triggers, secrets, artifact upload/download, caching

### BND-06: CI/CD Skill

- **SKILL.md body:** When to activate, cross-provider overview, key concepts
- **references/github-actions.md:** YAML schema, reusable workflows, matrix builds, environments
- **references/pipelines.md:** Build stages, caching strategies, parallelism
- **references/deployments.md:** Blue/green, canary, rollback, environment promotion gates

### BND-07: context-mode Skill

- **Source:** Adapt from `~/.claude/CLAUDE.md` context-mode section (the current active install)
- **SKILL.md body:** Context window routing rules, blocked commands, tool hierarchy
- **references/routing-rules.md:** Full ctx_batch_execute / ctx_search / ctx_execute reference
- **Note:** The CLAUDE.md in this project IS the context-mode skill content — extract and structure it

### BND-08: RTK Skill

- **Source:** Adapt from `~/.dotfiles/.claude/RTK.md` (confirmed via CLAUDE.md includes)
- **SKILL.md body:** RTK meta commands, installation verification, hook-based usage
- **references/commands.md:** Full command reference, token savings analytics, proxy usage

### BND-09: Serena Skill

- **Source:** Adapt from existing Serena skill install referenced in project CLAUDE.md
- **SKILL.md body:** LSP-powered navigation, when to activate, initial_instructions call
- **references/lsp-usage.md:** Symbol navigation, go-to-definition, reference search, rename patterns

### Skill 10: skill-author Meta-Skill

- **SKILL.md body:** Evaluation rubric summary, authoring checklist, when to invoke auto-researcher
- **references/evaluation-rubric.md:** Full scoring rubric: SKILL.md quality, references/ structure, frontmatter compliance, line count, injection-free
- **references/spec-compliance.md:** agentskills.io spec details, required vs optional fields
- **references/authoring-guide.md:** Step-by-step: research domain → write SKILL.md → add references → test validation → submit
- **scripts/validate-skill.sh:** Runs agentkit `ValidateSkill` logic (SKILL.md exists, line count, references/ check)
- **agents/auto-researcher/:** Auto-researcher agent entry point — helps discover domain content for new skills

---

## 7. Integration Points

### Files to Create

| File | Purpose |
|------|---------|
| `internal/installer/github_release.go` | New `GitHubReleaseInstaller` struct |
| `internal/installer/github_release_test.go` | Tests with httptest server |
| `internal/bundle/bundles.go` | Bundle loader — reads `bundles.json` via `//go:embed` |
| `internal/bundle/bundles.json` | Bundle definitions (D-06) |
| `internal/bundle/bundles_test.go` | Tests for bundle loader |
| `skills/<name>/SKILL.md` | Each of the 10 bundled skills |
| `skills/<name>/references/*.md` | Reference files per skill |
| `skills/external/<name>/SKILL.md` | External skill adaptations |
| `agents/auto-researcher/AGENT.md` | Auto-researcher agent |

### Files to Modify

| File | Change |
|------|--------|
| `internal/domain/package.go` | Add `InstallMethodGitHubRelease InstallMethod = "github-release"`; add `Repo string` and `Path string` fields to `InstallSpec` |
| `internal/installer/installer.go` | Add `case domain.InstallMethodGitHubRelease: return NewGitHubReleaseInstaller(), nil` |
| `cmd/install.go` | Add `--bundle` flag; `cobra.RangeArgs(0, 1)`; dispatch to `runBundleInstall()` when set |
| `internal/config/paths.go` | Add `TarballCachePath(repo, version string) (string, error)` |

### Exact Signatures

```go
// internal/domain/package.go additions
const InstallMethodGitHubRelease InstallMethod = "github-release"

// InstallSpec additions
type InstallSpec struct {
    Method  InstallMethod `json:"method"`
    Package string        `json:"package,omitempty"`
    URL     string        `json:"url,omitempty"`
    Args    []string      `json:"args,omitempty"`
    // New fields for github-release method:
    Repo    string        `json:"repo,omitempty"`    // e.g. "ejyle/agentkit"
    Path    string        `json:"path,omitempty"`    // e.g. "skills/aws"
}

// internal/installer/github_release.go
type GitHubReleaseInstaller struct {
    client  *http.Client
    version string        // current binary version, set via ldflags
    cache   sync.Map      // in-process cache: "repo@version" -> []byte
}

func NewGitHubReleaseInstaller() *GitHubReleaseInstaller
func (g *GitHubReleaseInstaller) Method() domain.InstallMethod
func (g *GitHubReleaseInstaller) IsAvailable() bool   // always true
func (g *GitHubReleaseInstaller) Install(spec domain.InstallSpec) error

// internal/bundle/bundles.go
type BundleDef struct {
    Packages []string `json:"packages"`
}
type BundleManifest struct {
    Bundles map[string]BundleDef `json:"bundles"`
}
func LoadBundles() (*BundleManifest, error)
func (m *BundleManifest) Resolve(name string) ([]string, error)
```

---

## 8. Validation Architecture

`nyquist_validation` is `false` in `.planning/config.json` — Validation Architecture section is skipped per the template rule.

However, skill validation (SKL-01/02/03) is an application-level concern, not a test-framework concern. The existing `ValidateSkill` in `internal/skill/validate.go` is already wired into `service.Install()`. For the `github-release` installer, the flow is:

1. `GitHubReleaseInstaller.Install(spec)` — downloads tarball, extracts `spec.Path` subdirectory to a temp dir, returns temp dir path via a side-channel (needs design decision — see below)
2. `service.Install()` calls `s.validator(tempDir, pkg)` — already in the code at line 133
3. Warnings printed to stderr; errors cause install to fail

**Design issue:** The current `service.Install()` passes `""` as the dir to the validator (line 133 — `result := s.validator("", pkg)`). For the github-release installer, the dir must be the extracted temp directory. Two options:
- **Option A:** `GitHubReleaseInstaller.Install()` returns the temp dir path via a package-level variable or via a new interface method `ExtractedDir() string`. Messy.
- **Option B:** `GitHubReleaseInstaller.Install()` extracts to the final skill directory (`SkillInstallPath(target, name)`) directly, calling `config.SkillInstallPath` internally. Then validator runs on the final path.

**Recommendation:** Option B. The installer owns extraction to the final path. The adapter's `WriteSkill` call in `service.Install()` becomes a no-op for github-release type (skill already on disk). This requires a small modification to `service.Install()` to skip `WriteSkill` when `pkg.Install.Method == InstallMethodGitHubRelease` and the installer has already placed files. [ASSUMED]

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Tarball decompression | Custom gzip reader | `compress/gzip` stdlib | Battle-tested, handles all edge cases |
| Archive iteration | Manual byte scanning | `archive/tar` stdlib | Correct handling of header types, symlinks, extended headers |
| Parallel error collection | Manual channel-based result aggregation | `sync.WaitGroup` + slice-by-index | WaitGroup is simpler than errgroup for best-effort; channels add unnecessary complexity |
| Skill content from scratch | Writing all AWS/GCP/Azure knowledge | Adapt from official Anthropic/Vercel skills; synthesize from cloud CLI docs | Faster to curate than author from scratch |
| Bundle config format | TOML or YAML bundles file | JSON (consistent with project stack) | Already using `encoding/json` everywhere; no new parser |
| File embedding | Exec-time file reads with path resolution | `//go:embed` directive | Zero-dependency, works in release binaries |

---

## Common Pitfalls

### Pitfall 1: Archive Root Prefix Mismatch
**What goes wrong:** GitHub source archive tarballs have an archive root prefix `<repo>-<version>/` (with version stripped of leading `v`). Extracting `skills/aws` without stripping this prefix finds nothing.
**Why it happens:** GitHub's source archive format inserts the root directory.
**How to avoid:** Detect or compute the prefix: `fmt.Sprintf("%s-%s/", repoName, strings.TrimPrefix(version, "v"))`. Test with a real tarball download in the test suite using httptest.
**Warning signs:** Installer completes with no error but skill directory is empty.

### Pitfall 2: Path Traversal in Untrusted Archives
**What goes wrong:** A malicious tarball entry like `../../.ssh/authorized_keys` escapes the destination directory.
**Why it happens:** `archive/tar` faithfully reports header paths verbatim.
**How to avoid:** Reject any `hdr.Name` containing `..` after stripping the prefix. Verify the resolved `outPath` is under `destDir` with `filepath.Rel`.
**Warning signs:** Security audit finding; `os.WriteFile` outside expected directory.

### Pitfall 3: errgroup Cancellation Breaks D-04 Semantics
**What goes wrong:** Using `errgroup.WithContext` cancels all in-flight installs when the first fails.
**Why it happens:** errgroup's design is fail-fast; D-04 requires best-effort.
**How to avoid:** Use `sync.WaitGroup` + index-keyed results slice. Never import `golang.org/x/sync/errgroup` in the bundle path.
**Warning signs:** Bundle stops after first failure instead of completing all.

### Pitfall 4: InstallSpec Backward Compatibility
**What goes wrong:** Adding `Repo` and `Path` fields to `InstallSpec` breaks existing JSON unmarshaling if the field names collide with something in gsd-core's registry.json.
**Why it happens:** Struct tags are omitempty — no collision expected, but worth noting.
**How to avoid:** Use `omitempty` on new fields. Run existing testdata/registry.json through the new struct to confirm zero breakage.
**Warning signs:** `TestInstall*` tests fail after domain package change.

### Pitfall 5: service.Install WriteSkill No-Op for github-release
**What goes wrong:** Current `service.Install()` calls `s.adapter.WriteSkill(name, map[string][]byte{"SKILL.md": []byte("")})` — this writes an empty SKILL.md, destroying whatever the github-release installer just extracted.
**Why it happens:** The service assumes skills always need WriteSkill called; github-release installer puts files directly on disk.
**How to avoid:** Gate the `WriteSkill` call: `if pkg.Install.Method != domain.InstallMethodGitHubRelease { s.adapter.WriteSkill(...) }`.
**Warning signs:** Installed skill directory has empty SKILL.md despite extraction succeeding.

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| skills.sh manual `npx skills add` | agentkit `install --bundle <name>` | Phase 3 | One-command bundle install replaces per-skill manual install |
| Skill content authored per-project | Shared curated registry via github-release tarball | Phase 3 | Skills versioned with binary; consistent for all users |
| errgroup for parallel fan-out | sync.WaitGroup + index slice for best-effort | Phase 3 | Matches D-04 semantics; no cancellation on failure |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | GSD registry entry uses `"custom"` or `"npx"` method (not `"github-release"`) | GSD Registry Entry | If gsd-core uses a different method, CLI-03 may fail silently or require new installer work |
| A2 | GitHub source archive URL format: `https://github.com/<owner>/<repo>/archive/refs/tags/<version>.tar.gz` with root prefix `<repo>-<version-no-v>/` | GitHub Release Installer | If format differs (e.g., explicit release asset), URL construction is wrong |
| A3 | Skills should be delivered via source archive (not explicit GoReleaser release asset) | Skills Directory Layout / GoReleaser | If Phase 4 uses asset-based archives, installer URL construction needs updating |
| A4 | `//go:embed` for bundles.json is correct approach (no runtime file path needed) | Bundle Command | If binary is run from a non-standard CWD where embed fails, bundles won't load; but embed is compile-time so this won't fail at runtime |
| A5 | Binary version from ldflags (`var Version string`) is the right mechanism for version injection | GitHub Release Installer | If no ldflags version is set (dev builds), installer needs fallback (e.g., "main" branch tarball) |
| A6 | anthropics/skills and vercel-labs repos use MIT or Apache license | External Skill Candidates | If license is restrictive, content cannot be copied into this repo; attribution-only approach needed |
| A7 | Option B (installer writes to final path, service skips WriteSkill) is correct integration pattern | Validation Architecture | If WriteSkill path is required for adapter bookkeeping, Option B breaks adapter state tracking |

---

## Open Questions

1. **D-17: GSD registry.json actual entry**
   - What we know: GSD installs via `npx @opengsd/gsd-core@latest --antigravity --global`
   - What's unclear: Exact JSON in gsd-core/registry.json — method, package, args
   - Recommendation: Implementer must `curl` the raw URL before coding CLI-03; may be a no-op if entry uses `custom`/`npx`

2. **Version source for github-release installer**
   - What we know: Phase 4 wires GoReleaser ldflags; Phase 3 must work in dev builds
   - What's unclear: Whether a `var Version = "dev"` fallback is sufficient, or if dev builds should use the latest GitHub tag
   - Recommendation: `"dev"` → falls back to fetching `latest` release tag via GitHub API (`/repos/<owner>/<repo>/releases/latest`)

3. **adapter.WriteSkill interaction with github-release**
   - What we know: `service.Install()` calls `WriteSkill` for all skill packages; github-release installer puts files on disk directly
   - What's unclear: Whether adapters use `WriteSkill` for bookkeeping beyond file writing
   - Recommendation: Audit `WriteSkill` in all adapters; if it's only file I/O, the no-op path is safe

4. **context-mode, RTK, Serena source content**
   - What we know: These are installed in `~/.claude/skills/` on developer machines; D-09 says adapt from existing installs
   - What's unclear: The actual current content of these skills (not accessible in this research session — `~/.claude/skills/` is empty/not present in this environment)
   - Recommendation: Developer (Nithin) must extract content from their personal install and commit to `skills/` in this repo; research phase cannot do this

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go stdlib `archive/tar` | GitHubReleaseInstaller | ✓ | stdlib | — |
| Go stdlib `compress/gzip` | GitHubReleaseInstaller | ✓ | stdlib | — |
| Go stdlib `sync` | Bundle parallelism | ✓ | stdlib | — |
| Go `//go:embed` | bundles.json embed | ✓ | Go 1.16+ (using 1.26.3) | — |
| `hashicorp/go-retryablehttp` | Tarball HTTP download | ✓ | v0.7.8 (in go.mod) | `net/http` (already used in binary.go) |
| `github.com/google/renameio/v2` | Atomic tarball cache write | ✓ | v2.0.2 (in go.mod) | — |

**Missing dependencies with no fallback:** None. All needed packages are stdlib or already in go.mod.

---

## Package Legitimacy Audit

No new external packages are being added. All dependencies are either:
- Go stdlib (`archive/tar`, `compress/gzip`, `sync`, `embed`)
- Already in go.mod (`hashicorp/go-retryablehttp`, `google/renameio`)

**Packages removed due to slopcheck verdict:** None (no new packages).
**Packages flagged as suspicious:** None.

---

## Security Domain

Phase 3 installs content from GitHub release tarballs. Relevant ASVS categories:

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | Yes | Path traversal check on all tar header entries (strings.Contains(rel, "..") + filepath.Rel boundary check) |
| V6 Cryptography | Partial | SHA256 verification of tarball (same pattern as BinaryInstaller); HTTPS-only enforcement already in place |
| V2 Authentication | No | No user auth in this phase |
| V4 Access Control | No | File permissions: 0755 for dirs, file mode from tar header masked to 0755 max |

| Threat Pattern | STRIDE | Standard Mitigation |
|----------------|--------|---------------------|
| Path traversal via crafted tar entry | Tampering | Reject `..` in tar header names; verify resolved path is under destDir |
| Tarball substitution / MITM | Spoofing | HTTPS-only download (same enforcement as BinaryInstaller); optional SHA256 of tarball |
| Malicious skill content (prompt injection) | Tampering | Manual review for external skills; automated lint for `---` frontmatter injection patterns |
| Skill overwrites critical files | Tampering | `WriteSkill` / extraction only writes to `SkillInstallPath(target, name)` — bounded to skills directory |

---

## Sources

### Primary (HIGH confidence)
- Go stdlib `archive/tar` — https://pkg.go.dev/archive/tar (iterator pattern, header types)
- Go stdlib `compress/gzip` — https://pkg.go.dev/compress/gzip
- Go stdlib `sync.WaitGroup` — https://pkg.go.dev/sync
- agentskills.io specification — https://agentskills.io/specification (SKILL.md frontmatter required fields)
- GoReleaser archives documentation — https://goreleaser.com/customization/archive/ (extra files config)

### Secondary (MEDIUM confidence)
- github.com/open-gsd/gsd-core install docs — GSD uses `npx @opengsd/gsd-core@latest` [via WebSearch, install-on-your-runtime.md]
- github.com/anthropics/skills — Public skills repo, SKILL.md per skill [via WebSearch]
- github.com/vercel-labs/agent-skills — React, UI, writing handbook skills [via WebSearch]
- github.com/vercel-labs/agent-browser — CDP browser automation skill [via WebSearch]
- skills.sh — 200+ community skills; security audit finding (36% prompt injection rate) [via WebSearch toolworthy.ai]

### Tertiary (LOW confidence / ASSUMED)
- GSD registry entry install method — inferred from install docs; not verified from raw JSON
- GitHub source archive URL format — standard GitHub behavior, not verified for ejyle/agentkit specifically
- License status of external skills (MIT assumed)

---

## Metadata

**Confidence breakdown:**
- Standard Stack: HIGH — all packages already in go.mod; stdlib-only for new installer
- Architecture: HIGH — codebase fully read; all integration points confirmed
- GSD registry entry (D-17): LOW — could not fetch raw JSON; inferred from install docs
- External skill content: MEDIUM — repos confirmed to exist; exact structure inferred from search results
- Pitfalls: HIGH — derived from code analysis of actual service.Install() and installer patterns

**Research date:** 2026-06-09
**Valid until:** 2026-07-09 (stable Go ecosystem; agentskills.io spec unlikely to change)
