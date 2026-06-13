# Phase 3: Bundled Skills - Context

**Gathered:** 2026-06-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 3 delivers:
1. **10 bundled skills** authored to the agentskills.io spec and stored in this repo under `skills/` (9 BND skills + 1 skill-author meta-skill), including adapted content from Anthropic, Vercel Labs, and skills.sh top picks (10–12 external skills curated, content owned by this repo).
2. **`--bundle` CLI command** — `agentkit install --bundle <cloud|dev|context>` installs a preset group of skills in parallel with best-effort semantics.
3. **`agentkit install gsd`** — delegates to the gsd-core registry (REG-02, already integrated); agentkit does not bundle GSD content in this repo.
4. **`github-release` install method** — new installer that fetches skills from the GitHub release tarball matching the current binary version and extracts a named subdirectory.

**Requirements in scope:** CLI-03, CLI-04, BND-01, BND-02, BND-03, BND-04, BND-05, BND-06, BND-07, BND-08, BND-09

**⚠ SCOPE CHANGES from original ROADMAP.md:**
- Phase 3 now includes a 10th bundled skill: **skill-author meta-skill** (not in original BND list).
- Phase 3 now includes **10–12 external skills** adapted from Anthropic/Vercel Labs/skills.sh — content copied into this repo under `skills/external/` with agentskills.io spec compliance.
- Update mechanism (re-sync from upstream) and skill publishing pipeline are **deferred to v0.2.0**.

</domain>

<decisions>
## Implementation Decisions

### Skill File Delivery
- **D-01:** Skills live in this repo under `skills/` (not embedded in the binary, not in a separate repo). At install time, agentkit fetches the GitHub release tarball for the current binary version and extracts the relevant skill subdirectory.
- **D-02:** New install method in registry.json: `"method": "github-release"`. Registry entry format:
  ```json
  { "install": { "method": "github-release", "repo": "ejyle/agentkit", "path": "skills/aws" } }
  ```
  The `github-release` installer resolves the tarball URL from the current binary version tag, downloads once per install batch, and extracts `path` to the target skill directory.
- **D-03:** The `github-release` installer works for any GitHub repo — third-party skills can use it with a different `repo` value. The installer is not agentkit-specific.

### Bundle Command Mechanics
- **D-04:** Bundle install is **best-effort and parallel**. All packages in a bundle start concurrently. Failures are collected; the command reports successes and failures at completion. No rollback of already-installed packages.
- **D-05:** Exit code: **1 if any package in the bundle failed**, 0 if all succeeded. Scripts and CI get a reliable signal.
- **D-06:** Bundle definitions live in a configuration file (e.g., `internal/bundle/bundles.json`) — not hardcoded in the CLI. Each bundle entry lists package names. Adding a new bundle does not require a CLI code change.
- **D-07:** Bundles defined for v0.1.0: `cloud` (aws, gcp, azure), `dev` (playwright, github, cicd), `context` (context-mode, rtk, serena). GSD is a standalone install, not a bundle member.

### The 10 Bundled Skills (Authored in this repo)
- **D-08:** All 10 skills are **full reference-quality** — real SKILL.md (<500 lines each) + domain reference files + optional scripts. No stubs.
- **D-09:** **context-mode, RTK, Serena** (BND-07, BND-08, BND-09) are adapted from existing `~/.claude/skills/` installs. Content reviewed and updated to meet agentskills.io spec (frontmatter, references/ structure, line limit).
- **D-10:** **10th skill: skill-author meta-skill** — helps Claude evaluate, improve, and author agentkit skills. Structure: `SKILL.md` (entrypoint) + `references/` (progressive discovery docs) + `scripts/` (helper scripts) + bundled **auto-researcher agent** (from existing install, included in this repo under `agents/`). This meta-skill is NOT deferred.
- **D-11:** The **auto-researcher agent** is bundled in this repo (not just referenced externally). It lives in `agents/auto-researcher/` and is referenced from the skill-author meta-skill.

### External Skill Curation (10–12 skills)
- **D-12:** External skill content is **copied into this repo** under `skills/external/` (we own it). Sources: Anthropic skills repo (frontend-design), Vercel Labs (vercel-react-best-practices, agent-browser), skills.sh top list, Azure top skills.
- **D-13:** Researcher evaluates each candidate external skill for: (a) quality, (b) agentskills.io spec compliance (needs SKILL.md + references/ structure), (c) agentkit license compatibility. Researcher adapts content to spec before including.
- **D-14:** External skill update mechanism (re-sync from upstream) is **deferred to v0.2.0**. Phase 3 does the initial copy only.
- **D-15:** Researcher to discover and recommend the final list of 10–12 external skills. Sources: `https://www.skills.sh`, `https://github.com/anthropics/skills`, `https://github.com/vercel-labs/agent-skills`, `https://github.com/vercel-labs/agent-browser`.

### GSD Package
- **D-16:** `agentkit install gsd` does **not** bundle GSD content in this repo. It resolves the `gsd` package from the gsd-core registry (REG-02, already integrated as `NewGitHubManifestRegistry("gsd-core", ...)`).
- **D-17:** Researcher MUST check `https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json` to verify what install method the `gsd` entry uses before implementing. Do NOT assume `custom`.

### Claude's Discretion
- Exact `skills/` directory structure layout within this repo (flat vs nested by category)
- Whether the `github-release` installer caches the tarball across multiple skill installs in the same session
- Spinner output format for parallel bundle installs (per-package lines vs combined progress)
- Exact agentkit-registry entries for the external skills (names, descriptions, version tags)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Roadmap
- `.planning/REQUIREMENTS.md` — Full v1 requirement list; CLI-03, CLI-04, BND-01 through BND-09 are Phase 3 items. SKL-01, SKL-02, SKL-03 already implemented in `internal/skill/validate.go`.
- `.planning/ROADMAP.md` — Phase 3 success criteria (note: criteria now include 10 skills not 9, and external curation)
- `.planning/PROJECT.md` — Core constraints and technology stack rationale
- `.planning/phases/01-foundation/01-CONTEXT.md` — D-01/D-02 curated-only model; D-03/D-04 output style; D-07/D-09 merge/conflict patterns; architecture constraints
- `.planning/phases/02-multi-assistant-full-install/02-CONTEXT.md` — D-19 adapter patterns; ErrNotSupported pattern

### Existing Skill Infrastructure (already implemented)
- `internal/skill/validate.go` — SKL-01/02/03 validator: SKILL.md check, line-count warning, references/ check. Phase 3 uses this at bundle install time.
- `internal/config/paths.go` — `SkillInstallPath(target, name)` — where skills land per assistant
- `internal/installer/installer.go` — `MCPInstaller` interface + `NewInstaller` factory — add `InstallMethodGitHubRelease` case here

### Existing Installer Patterns (read before authoring new installer)
- `internal/installer/npx.go` — Reference: `IsAvailable()`, error types (`ErrNodeNotFound`), `Install()` signature
- `internal/installer/binary.go` — Reference: binary download pattern (HTTP fetch + write to disk)
- `internal/installer/custom.go` — Reference: custom script execution pattern

### External Sources (researcher must verify before including)
- `https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json` — GSD registry entries; MUST check before implementing D-17
- `https://www.skills.sh` — Top skills directory; evaluate candidates for D-15
- `https://github.com/anthropics/skills` — Anthropic skills (frontend-design candidate)
- `https://github.com/vercel-labs/agent-skills` — Vercel agent skills candidates
- `https://github.com/vercel-labs/agent-browser` — Vercel agent-browser candidate

### Technology Stack
- `CLAUDE.md` §Technology Stack — existing dependencies (Cobra, Bubbletea, Lipgloss, go-retryablehttp)
- `CLAUDE.md` §Skill Structure — SKILL.md frontmatter spec, progressive disclosure structure

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/skill/validate.go` (`ValidateSkill`) — run this at bundle install time for every skill extracted from the GitHub release tarball; non-blocking warnings for line-count, blocking errors for missing SKILL.md
- `internal/installer/installer.go` (`MCPInstaller` interface + `NewInstaller` factory) — add `InstallMethodGitHubRelease = "github-release"` case; new `GitHubReleaseInstaller` struct implements the interface
- `internal/installer/binary.go` — HTTP download pattern reusable for fetching the release tarball
- `internal/config/paths.go` (`SkillInstallPath`) — already handles claude/gemini/pi targets; no changes needed for skill delivery
- `cmd/install.go` (`runInstall`) — add `--bundle` flag handling; parallel installs via goroutines + collect results
- `registry.NewRegistryManager()` — gsd-core registry already registered; `agentkit install gsd` routes through this naturally

### Established Patterns
- Parallel goroutine installs: already done in `cmd/install.go` for the spinner — extend this pattern for `--bundle` parallelism
- Error collection: collect `[]error` across goroutines, print summary at end (D-04 style)
- ETag-based manifest cache (`internal/registry/registry.go`) — tarball cache for `github-release` installer should follow same pattern

### Integration Points
- `cmd/install.go` — add `--bundle` flag; when set, resolve bundle → list of package names → parallel `svc.Install()` calls
- `internal/installer/installer.go` `NewInstaller()` — add `case InstallMethodGitHubRelease: return NewGitHubReleaseInstaller(), nil`
- `internal/domain/package.go` — add `InstallMethodGitHubRelease InstallMethod = "github-release"` constant; add `Repo` and `Path` fields to `InstallSpec`
- `internal/registry/registry.go` `NewRegistryManager()` — no changes; gsd-core already pre-registered

</code_context>

<specifics>
## Specific Ideas

- `agentkit install --bundle cloud` should show per-package spinner lines (aws ✓, gcp ✓, azure ✗) then a summary line: `2/3 installed — azure failed: <reason>`
- External skills under `skills/external/` should have a clear header in their SKILL.md: `# [Skill Name] (via [source-org]/[source-repo])`
- The skill-author meta-skill should include the full evaluation rubric in `references/evaluation-rubric.md` so any skill can be self-assessed by Claude before submission
- The `github-release` installer should cache the downloaded tarball in `~/.cache/agentkit/releases/<version>/` for the session to avoid re-downloading when a bundle installs multiple skills from the same release

</specifics>

<deferred>
## Deferred Ideas

- **External skill update mechanism** — `agentkit skill sync <name>` to re-fetch from upstream source. Deferred to v0.2.0. Phase 3 does initial copy only.
- **Skill publishing pipeline** — meta-skill + auto-researcher + approval + publish to registry. The meta-skill IS in Phase 3; the full automated pipeline (including approval gating and registry PR automation) is v0.2.0.
- **Skill-author framework as a CLI command** — `agentkit skill validate <path>` and `agentkit skill improve <path>` as first-class CLI subcommands. Deferred to v0.2.0; Phase 3 uses the meta-skill (skill-author) for this workflow manually.
- **Additional bundle definitions** — user hinted at Azure-specific bundles and more. v0.2.0 with the skill-author pipeline will make bundle expansion low-friction.
- **Background config agent** — per-project facts system. Deferred to v0.2.0 per REQUIREMENTS.md.

</deferred>

---

*Phase: 3-Bundled Skills*
*Context gathered: 2026-06-09*
