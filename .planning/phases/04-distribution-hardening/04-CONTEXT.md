# Phase 4: Distribution & Hardening - Context

**Gathered:** 2026-06-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 4 delivers:
1. **GoReleaser v2 pipeline** — cross-platform binaries (macOS arm64/amd64, Linux amd64/arm64, Windows amd64) built and published on `v*` tag push and as a snapshot on `main` push.
2. **Homebrew tap** — `brew install ejyle/agentkit/agentkit` installs the correct binary on macOS; formula auto-updated by GoReleaser via PAT.
3. **`agentkit doctor` command** — diagnose-only validator that checks: binary in PATH, config dir writable, network reaches default registry, target assistant config dirs exist, and MCP runtime dependencies (node/npx, docker, uvx) available.
4. **Version injection** — `agentkit --version` outputs `agentkit/0.1.0 (darwin/arm64)` via ldflags at build time.
5. **curl|sh install script** — hosted in repo (and optionally at install.agentkit.dev), detects OS/arch, downloads the right binary from GitHub releases.
6. **Initial v0.1.0 release** — tagged as `v0.1.0`, producing the first public release of agentkit.

**Requirements in scope:** CLI-10

**Not in scope:**
- Auto-fix capability in `agentkit doctor` (v0.2.0)
- macOS notarization / code-signing (v0.2.0)
- Windows package manager integration (winget/choco) — deferred per REQUIREMENTS.md
- Background config agent / project-scope facts — deferred to v0.2.0

</domain>

<decisions>
## Implementation Decisions

### Release Pipeline
- **D-01:** Release workflow triggers on **two events**: `push` to tags matching `v*` (creates a real GitHub Release) AND `push` to `main` (creates a GoReleaser snapshot/pre-release for testing). Two separate workflow jobs or a single job with conditional `snapshot: true`.
- **D-02:** **Cosign keyless signing** — GoReleaser signs release artifacts using GitHub OIDC token (no key management). Verification via `cosign verify-blob`. Adds `--annotations` referencing the release.
- **D-03:** Version injected at build time via **`ldflags -X`**: `go build -ldflags "-X github.com/ejyle/agentkit/internal/version.Version={{.Version}} -X github.com/ejyle/agentkit/internal/version.GOOS={{.Os}} -X github.com/ejyle/agentkit/internal/version.GOARCH={{.Arch}}"`. A dedicated `internal/version/version.go` package exposes `Version`, `GOOS`, `GOARCH` vars.
- **D-04:** `--version` output format: `agentkit/0.1.0 (darwin/arm64)` — version string includes OS/arch for easy bug reporting.

### Homebrew Tap
- **D-05:** Formula lives in a **separate `ejyle/homebrew-agentkit` repository**. GoReleaser's `brews` config points to this repo and commits the updated formula on each release.
- **D-06:** Users install with `brew install ejyle/agentkit/agentkit` (single command — taps and installs). No two-step required.
- **D-07:** GoReleaser authenticates to the homebrew-agentkit repo using a **classic GitHub PAT with `repo` scope**, stored as `HOMEBREW_TAP_GITHUB_TOKEN` secret in `ejyle/agentkit`.

### `agentkit doctor` Command
- **D-08:** Checks performed (in order):
  1. Binary in PATH (`which agentkit`)
  2. `~/.agentkit/` config directory is writable (creates if missing, then checks write access)
  3. Network can reach the default registry (HTTP GET to the registry manifest URL, timeout 5s)
  4. Each installed target assistant's config directory exists (`~/.claude/`, `~/.gemini/`, etc.)
  5. **Runtime deps (informational/warn, not blocking):** `node`/`npx` available, `docker` available, `uvx` available
- **D-09:** Output format: **line-by-line with `✓`/`⚠`/`✗` icons**. Each check on its own line. On failure/warn, print a hint line: `  → Run: mkdir -p ~/.agentkit`. Matches `brew doctor` style.
- **D-10:** **Diagnose-only** — no `--fix` flag. Prints fix hints for user to run manually. v0.2.0 for auto-fix.
- **D-11:** Exit code: 0 if all checks pass or warn, non-zero (1) if any check fails (`✗`). Warnings do not fail the exit code. Scripts can test `agentkit doctor` in CI.

### Version & Distribution
- **D-12:** Initial release tag: **`v0.1.0`** (SemVer with `v` prefix, as expected by GoReleaser and Homebrew).
- **D-13:** **curl|sh install script** (`scripts/install.sh` in repo): detects `$(uname -s)` and `$(uname -m)`, constructs the GitHub release asset URL for the detected OS/arch, downloads via `curl -fsSL`, verifies checksum against GoReleaser's `checksums.txt`, installs to `~/.local/bin` (or `/usr/local/bin` if writable). Also accessible at `install.agentkit.dev` (configured later; script in repo is the canonical source).
- **D-14:** GoReleaser generates `checksums.txt` for all release assets. Install script verifies SHA256 before installing.

### Claude's Discretion
- GoReleaser archive naming convention (e.g., `agentkit_{{.Version}}_{{.Os}}_{{.Arch}}.tar.gz`)
- Whether snapshot builds on `main` are marked as pre-release in GitHub or just draft
- Whether `agentkit doctor` checks assistant config dirs for ALL supported targets or only those with installed packages
- Exact GoReleaser `changelog` config (conventional commits? full commit list?)
- `internal/version/version.go` package structure (exported vars vs `String()` method)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Roadmap
- `.planning/REQUIREMENTS.md` — CLI-10 is the Phase 4 requirement (single binary, no runtime dep, cross-platform)
- `.planning/ROADMAP.md` — Phase 4 success criteria (4 criteria mapped to D-01 through D-13 above)
- `.planning/PROJECT.md` — Core constraints: Go single binary, user-scope installs, cross-platform
- `.planning/phases/01-foundation/01-CONTEXT.md` — D-01/D-02: curated-only model; architecture constraints; no root/admin
- `.planning/phases/03-bundled-skills/03-CONTEXT.md` — D-01/D-02: `github-release` installer; version tag must match binary version exactly (skills fetched from release tarball at current binary version tag)

### Existing Code (read before implementing)
- `cmd/root.go` — Cobra root command; `--version` flag goes here; `Use`, `Version` fields
- `go.mod` — Module path `github.com/ejyle/agentkit`; current dependency versions; go version

### External References (researcher must check)
- GoReleaser v2 docs: https://goreleaser.com/customization/ — ldflags, brews, cosign, archives, changelog config
- Cosign keyless signing in GoReleaser: https://goreleaser.com/customization/sign/ — OIDC-based signing config
- Homebrew tap conventions: https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap — `homebrew-agentkit` repo setup
- GitHub Actions cosign action: https://github.com/sigstore/cosign-installer — workflow step for cosign binary

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `cmd/root.go` — Root Cobra command; add `Version` field (Cobra renders it for `--version`/`-v` flags); or override with a `versionCmd` for custom formatting
- `internal/config/paths.go` — Config directory paths per assistant; `doctor` can reuse `AssistantConfigDir(target)` to check directory existence
- `internal/registry/registry.go` — Registry HTTP client; `doctor` can reuse the registry manager's `FetchManifest` to test network connectivity (or extract a `Ping()` method)

### Established Patterns
- Error collection pattern (from `runBundleInstall`) — collect `[]error` across parallel operations; `doctor` can collect `[]CheckResult` and print summary
- Atomic writes (`github.com/google/renameio/v2`) — already a dependency; install script can use same pattern for temp-file-then-rename installs
- `internal/ui/spinner.go` — Spinner for visual feedback; `doctor` may use it while running checks in sequence

### Integration Points
- `cmd/root.go` → add `doctor` subcommand (`cmd/doctor.go`)
- `main.go` → no changes needed
- `internal/version/version.go` → new package; vars set via ldflags; used in `cmd/root.go`
- `.goreleaser.yaml` → new file at repo root; references `internal/version` package for ldflags
- `.github/workflows/release.yml` → new file; triggers on `v*` tags and `main` push

</code_context>

<specifics>
## Specific Ideas

- `agentkit doctor` output should look like:
  ```
  ✓ agentkit in PATH (v0.1.0)
  ✓ ~/.agentkit/ writable
  ✓ registry reachable (agentkit-registry)
  ✓ ~/.claude/ exists
  ✓ ~/.gemini/ exists
  ⚠ ~/.copilot/ not found — Copilot CLI not installed
  ⚠ node not found — npx-based MCPs won't install
     → Install: https://nodejs.org
  ✓ docker available
  ⚠ uvx not found — Python MCPs won't install
     → Install: pip install uv
  ```
- The `github-release` installer (Phase 3) uses the binary's embedded version to construct the tarball URL — `internal/version.Version` must be the source of truth for this (no separate config file)
- The install script at `scripts/install.sh` should default to installing to `~/.local/bin` and print: `Add ~/.local/bin to your PATH: export PATH="$HOME/.local/bin:$PATH"`
- The snapshot build (on `main` push) should produce a version string like `0.1.0-SNAPSHOT-abc1234` so it's clearly distinguishable from a release build

</specifics>

<deferred>
## Deferred Ideas

- **`agentkit doctor --fix`** — auto-repair fixable issues (create missing dirs, add to PATH config). Deferred to v0.2.0.
- **macOS notarization** — code-sign the macOS binary so Gatekeeper allows it without quarantine prompt. Requires Apple Developer account. Deferred to v0.2.0.
- **Windows package manager integration** — winget/choco manifests. Deferred per REQUIREMENTS.md.
- **install.agentkit.dev domain** — custom domain for the install script URL. Phase 4 ships the script in-repo; domain redirect configured later.
- **Scoop manifest for Windows** — GoReleaser can generate this; defer to v0.2.0 alongside Windows-specific testing.

</deferred>

---

*Phase: 4-Distribution & Hardening*
*Context gathered: 2026-06-09*
