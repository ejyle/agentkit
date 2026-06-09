# Phase 4: Distribution & Hardening - Research

**Researched:** 2026-06-09
**Domain:** GoReleaser v2, Cosign keyless signing, Homebrew tap (cask), GitHub Actions release workflow, Go version injection, curl|sh install scripts
**Confidence:** MEDIUM-HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Release workflow triggers on two events: `push` to tags matching `v*` (real GitHub Release) AND `push` to `main` (GoReleaser snapshot/pre-release). Two separate jobs or one with conditional `--snapshot`.
- **D-02:** Cosign keyless signing — GoReleaser signs release artifacts using GitHub OIDC token (no key management). Verification via `cosign verify-blob`.
- **D-03:** Version injected at build time via `ldflags -X`: `go build -ldflags "-X github.com/ejyle/agentkit/internal/version.Version={{.Version}} -X github.com/ejyle/agentkit/internal/version.GOOS={{.Os}} -X github.com/ejyle/agentkit/internal/version.GOARCH={{.Arch}}"`. Dedicated `internal/version/version.go` package.
- **D-04:** `--version` output format: `agentkit/0.1.0 (darwin/arm64)`.
- **D-05:** Homebrew formula in separate `ejyle/homebrew-agentkit` repo.
- **D-06:** Users install with `brew install ejyle/agentkit/agentkit`.
- **D-07:** GoReleaser authenticates to tap repo using classic PAT with `repo` scope, secret `HOMEBREW_TAP_GITHUB_TOKEN`.
- **D-08 through D-11:** doctor command checks, output format (✓/⚠/✗), diagnose-only, exit codes.
- **D-12:** Initial release tag `v0.1.0`.
- **D-13:** curl|sh install script at `scripts/install.sh`; installs to `~/.local/bin`; verifies SHA256 checksum.
- **D-14:** GoReleaser generates `checksums.txt`; install script verifies SHA256 before installing.

### Claude's Discretion
- GoReleaser archive naming convention
- Whether snapshot builds on `main` are pre-release or draft in GitHub
- Whether `doctor` checks all supported targets or only installed ones
- Exact GoReleaser `changelog` config (conventional commits or full list)
- `internal/version/version.go` structure (exported vars vs `String()` method)

### Deferred Ideas (OUT OF SCOPE)
- `agentkit doctor --fix` (v0.2.0)
- macOS notarization (v0.2.0)
- Windows winget/choco integration
- install.agentkit.dev domain
- Scoop manifest for Windows (v0.2.0)
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CLI-10 | CLI ships as a single binary with no runtime dependency, runs on Windows/Linux/macOS without root or sudo | GoReleaser cross-compile matrix, curl\|sh install to `~/.local/bin`, Homebrew cask distribution |
</phase_requirements>

---

## Summary

Phase 4 delivers agentkit v0.1.0 as a publicly installable tool via three distribution channels: Homebrew tap (macOS primary), GitHub Releases binary download (all platforms), and a curl|sh install script (Linux/macOS no-manager path). The release pipeline is driven by GoReleaser v2 with cosign keyless signing for supply-chain security.

**Critical discovery:** GoReleaser v2.10+ deprecated `brews` in favor of `homebrew_casks`. As of v2.16, `brews` is fully deprecated. The planner MUST use `homebrew_casks` in `.goreleaser.yaml`, not `brews`. This changes D-05/D-06 — the Homebrew distribution is now a cask (correct for pre-compiled binaries), and `brew install ejyle/agentkit/agentkit` still works the same way from the user's perspective.

**Cosign v3 update:** The cosign signing workflow changed in cosign v3 — the `--bundle` flag replaces separate `--output-certificate` + `--output-signature` flags. GoReleaser's `signs` block must use the new `--bundle` approach producing a single `.sigstore.json` bundle file. `COSIGN_EXPERIMENTAL=1` is no longer needed for keyless signing in cosign v3+.

**Primary recommendation:** Use `goreleaser/goreleaser-action@v7` with `version: "~> v2"`, `homebrew_casks` (not `brews`), cosign v3 bundle-based keyless signing, and a two-job workflow (release job on `v*` tags, snapshot job on `main` push).

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Version string (`agentkit --version`) | CLI binary | — | ldflags inject at build time; Cobra `Version` field renders it |
| Cross-platform build | GoReleaser CI | GitHub Actions | GoReleaser computes GOOS/GOARCH matrix; Actions provides runner |
| Homebrew distribution | GoReleaser CI | `ejyle/homebrew-agentkit` tap repo | GoReleaser generates and pushes cask formula; tap repo serves brew |
| Cosign artifact signing | GoReleaser CI | sigstore OIDC | GoReleaser invokes cosign; GitHub OIDC provides ephemeral key |
| Install script | Static repo file | GitHub Releases CDN | `scripts/install.sh` downloads release asset; no server needed |
| `doctor` command | CLI binary | OS/network | Pure Go stdlib checks; no external process except `exec.LookPath` |
| Release trigger | GitHub Actions | GoReleaser | Tag push starts real release; main push starts snapshot |

---

## Standard Stack

### Core
| Library / Tool | Version | Purpose | Why Standard |
|----------------|---------|---------|--------------|
| GoReleaser | `~> v2` (latest v2.16+) | Cross-platform build + release pipeline | De facto Go release tool; native homebrew_casks, cosign, GitHub integration | [ASSUMED - version from training; v2.10+ for homebrew_casks, v2.16 full deprecation of brews] |
| `goreleaser/goreleaser-action` | `v7` | GitHub Actions integration | Official action; v7 is current recommended version [ASSUMED - confirmed v7 from search] |
| `sigstore/cosign-installer` | latest | Install cosign binary in CI | Official sigstore action for CI keyless signing [ASSUMED] |
| `actions/setup-go` | `v5` | Go toolchain in CI | Official; v5 is current [ASSUMED] |
| `actions/checkout` | `v4` | Git checkout with full history | Full history required by GoReleaser for changelog |
| Go stdlib `runtime` | — | GOOS/GOARCH in version package | `runtime.GOOS`/`runtime.GOARCH` as fallback if ldflags not injected |
| Go stdlib `net/http` | — | `doctor` network check | Simple GET with `http.NewRequestWithContext` + 5s timeout |
| Go stdlib `os/exec` | — | `doctor` binary-in-PATH check | `exec.LookPath("agentkit")` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `charmbracelet/lipgloss` | v1.x (already in go.mod) | Colorized doctor output (✓/⚠/✗) | Already a dependency; use for styled output |
| `internal/config/paths.go` | — (existing) | Assistant config dir paths in doctor | Reuse `SkillInstallPath` home-dir logic for dir existence checks |
| `internal/registry/registry.go` | — (existing) | Network check in doctor | Extract or add a lightweight `Ping()` / HEAD request to registry URL |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `homebrew_casks` | `brews` (legacy) | `brews` fully deprecated in v2.16; casks are the correct mechanism for pre-compiled binaries |
| cosign `--bundle` | `--output-certificate` + `--output-signature` | Old approach split into 2 files; `--bundle` (cosign v3) is one `.sigstore.json` — simpler, correct for 2025+ |
| `goreleaser-action@v7` | `v6` | v7 is current; v6 still works but v7 supports Pro v2.7.0+ versioning scheme |

**Installation (CI, not local dev):**
```bash
# .goreleaser.yaml uses GoReleaser as a build tool, installed by goreleaser-action
# No local npm/pip install; Go dependencies already in go.mod
```

---

## Package Legitimacy Audit

No new Go module dependencies are introduced in this phase. All tools (GoReleaser, cosign) are installed in CI by official GitHub Actions — not via `go get`. The existing `go.mod` dependency set is unchanged.

| Tool / Action | Source | Age | Trust | Disposition |
|---------------|--------|-----|-------|-------------|
| `goreleaser/goreleaser-action@v7` | github.com/goreleaser/goreleaser-action | Active, official | Official GoReleaser org | Approved |
| `sigstore/cosign-installer` | github.com/sigstore/cosign-installer | Active, official | Sigstore/Linux Foundation | Approved |
| `actions/checkout@v4` | github.com/actions/checkout | Official GitHub | GitHub | Approved |
| `actions/setup-go@v5` | github.com/actions/setup-go | Official GitHub | GitHub | Approved |

*slopcheck not applicable — no npm/PyPI packages introduced. All tools are GitHub Actions from official organizations.*

---

## Architecture Patterns

### System Architecture Diagram

```
Tag push (v*)                    main push
     |                               |
     v                               v
GitHub Actions: release.yml     GitHub Actions: release.yml
     |                               |
     v                               v
goreleaser-action@v7            goreleaser-action@v7
  args: release --clean           args: release --snapshot --skip-publish
     |
     +-- go build (5 targets)
     |    GOOS/GOARCH matrix:
     |    darwin/amd64, darwin/arm64
     |    linux/amd64, linux/arm64
     |    windows/amd64
     |    ldflags: -X internal/version.Version -X .GOOS -X .GOARCH
     |
     +-- archives (.tar.gz linux/mac, .zip windows)
     |    naming: agentkit_{{.Version}}_{{.Os}}_{{.Arch}}
     |
     +-- checksums.txt (SHA256 of all archives)
     |
     +-- cosign sign-blob --bundle (keyless, OIDC)
     |    produces: checksums.txt.sigstore.json
     |
     +-- GitHub Release (upload archives + checksums + bundle)
     |
     +-- homebrew_casks -> ejyle/homebrew-agentkit
          (PAT: HOMEBREW_TAP_GITHUB_TOKEN)
          Cask: agentkit.rb in Casks/
```

### Recommended Project Structure
```
.
├── .github/
│   └── workflows/
│       └── release.yml          # Dual-trigger: v* tag + main push
├── .goreleaser.yaml             # GoReleaser v2 config
├── scripts/
│   └── install.sh               # curl|sh install script
├── internal/
│   └── version/
│       └── version.go           # New: ldflags injection target
└── cmd/
    ├── root.go                  # Add Version field (D-04 format)
    └── doctor.go                # New: doctor command
```

### Pattern 1: GoReleaser v2 Minimal Config

**What:** `.goreleaser.yaml` with version 2, multi-platform builds, ldflags, archives, checksums, cosign signing, and homebrew_casks.
**When to use:** Always — this is the complete release config for this project.

```yaml
# Source: goreleaser.com/customization/ [CITED]
version: 2

project_name: agentkit

before:
  hooks:
    - go mod tidy

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w
      - -X github.com/ejyle/agentkit/internal/version.Version={{.Version}}
      - -X github.com/ejyle/agentkit/internal/version.GOOS={{.Os}}
      - -X github.com/ejyle/agentkit/internal/version.GOARCH={{.Arch}}
    main: ./

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

snapshot:
  version_template: "{{ .Version }}-SNAPSHOT-{{ .ShortCommit }}"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"

signs:
  - cmd: cosign
    artifacts: checksum
    args:
      - sign-blob
      - "--bundle=${signature}"
      - "${artifact}"
      - "--yes"

homebrew_casks:
  - repository:
      owner: ejyle
      name: homebrew-agentkit
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    homepage: https://github.com/ejyle/agentkit
    description: "AI agent skill and MCP server manager"
    license: MIT
    commit_author:
      name: goreleaserbot
      email: bot@goreleaser.com
```

**Note on `signs` output field:** GoReleaser v2 uses `${signature}` as the bundle output path when using `--bundle`. The resulting file is `checksums.txt.sigstore.json`. [ASSUMED - based on cosign v3 migration docs and GoReleaser supply-chain blog]

### Pattern 2: GitHub Actions Release Workflow

**What:** Dual-trigger workflow — real release on `v*` tags, snapshot on `main` push.
**When to use:** Single workflow file with two jobs.

```yaml
# Source: goreleaser.com/ci/actions/ [CITED]
name: Release

on:
  push:
    tags:
      - "v*"
    branches:
      - main

permissions:
  contents: write
  id-token: write  # Required for cosign OIDC keyless signing

jobs:
  release:
    if: startsWith(github.ref, 'refs/tags/')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # GoReleaser needs full history for changelog
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - uses: sigstore/cosign-installer@v3
      - uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}

  snapshot:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --snapshot --clean --skip=publish,sign,homebrew
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Key points:**
- `id-token: write` permission is REQUIRED for cosign OIDC keyless signing [CITED: goreleaser.com/blog/supply-chain-security/]
- `fetch-depth: 0` is REQUIRED — GoReleaser reads full git history for changelog
- Snapshot job skips publish, sign, and homebrew to avoid polluting the tap
- `goreleaser-action@v7` is current recommended version [ASSUMED]

### Pattern 3: internal/version Package

**What:** Minimal version package with ldflags-injectable string vars and a `String()` method.
**When to use:** New file, `internal/version/version.go`.

```go
// Source: goreleaser.com/cookbooks/using-main.version/ [CITED]
package version

import "fmt"

// These vars are set at build time via ldflags.
// Defaults are used when building locally without GoReleaser.
var (
    Version = "dev"
    GOOS    = "unknown"
    GOARCH  = "unknown"
)

// String returns the formatted version string for --version output.
// Format: agentkit/0.1.0 (darwin/arm64)
func String() string {
    return fmt.Sprintf("agentkit/%s (%s/%s)", Version, GOOS, GOARCH)
}
```

**Usage in `cmd/root.go`:**
```go
import "github.com/ejyle/agentkit/internal/version"

var rootCmd = &cobra.Command{
    // ...existing fields...
    Version: version.String(),
}
```

Cobra renders `--version` / `-v` using the `Version` field automatically. The output will be:
`agentkit version agentkit/0.1.0 (darwin/arm64)` — note Cobra prepends `agentkit version `.

To match D-04 exactly (`agentkit/0.1.0 (darwin/arm64)` with no preamble), override `SetVersionTemplate`:
```go
rootCmd.SetVersionTemplate("{{.Version}}\n")
```

### Pattern 4: doctor Command Structure

**What:** Sequential checks with `[]CheckResult`, printed line-by-line.
**When to use:** `cmd/doctor.go`.

```go
// [ASSUMED pattern — standard Go CLI pattern, not from official docs]
type CheckResult struct {
    Label   string
    Status  string // "pass", "warn", "fail"
    Message string
    Hint    string
}

func runDoctor(cmd *cobra.Command, args []string) error {
    results := []CheckResult{}
    results = append(results, checkBinaryInPath())
    results = append(results, checkConfigDirWritable())
    results = append(results, checkRegistryReachable())
    results = append(results, checkAssistantDirs()...)
    results = append(results, checkRuntimeDeps()...)

    hasFail := false
    for _, r := range results {
        printCheckResult(r)
        if r.Status == "fail" {
            hasFail = true
        }
    }
    if hasFail {
        return fmt.Errorf("one or more checks failed")
    }
    return nil
}
```

**Network check with 5s timeout (stdlib only):**
```go
// [ASSUMED — standard Go http pattern]
func checkRegistryReachable() CheckResult {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    req, _ := http.NewRequestWithContext(ctx, http.MethodHead, registryURL, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil || resp.StatusCode >= 500 {
        return CheckResult{Label: "registry reachable", Status: "fail", Hint: "check network connectivity"}
    }
    return CheckResult{Label: "registry reachable", Status: "pass"}
}
```

**Config dir writable check:**
```go
// [ASSUMED]
func checkConfigDirWritable() CheckResult {
    dir := filepath.Join(os.UserHomeDir(), ".agentkit")  // or use config.AgentConfigDir()
    if err := os.MkdirAll(dir, 0755); err != nil {
        return CheckResult{Label: "~/.agentkit/ writable", Status: "fail", Hint: "Run: mkdir -p ~/.agentkit"}
    }
    // write test
    testFile := filepath.Join(dir, ".write-test")
    if err := os.WriteFile(testFile, []byte{}, 0600); err != nil {
        return CheckResult{Label: "~/.agentkit/ writable", Status: "fail"}
    }
    os.Remove(testFile)
    return CheckResult{Label: "~/.agentkit/ writable", Status: "pass"}
}
```

**Assistant dir check — reuse SkillInstallPath home dir logic:**

`internal/config/paths.go` has `SkillInstallPath(target, name)` which returns `~/.claude/skills/<name>` etc. For doctor, we only need the assistant root (`~/.claude/`, `~/.gemini/`, etc.). Extract the home + target mapping by calling `filepath.Dir(filepath.Dir(path))` or write a new helper `AssistantRootDir(target)` that mirrors the switch in `SkillInstallPath` without the `skills/<name>` suffix.

### Pattern 5: curl|sh Install Script

**What:** `scripts/install.sh` — detects OS/arch, downloads release asset, verifies SHA256, installs.
**Key mappings:**

```bash
# [ASSUMED — standard shell pattern for Go binary install scripts]
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map uname -m to Go GOARCH names
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)        echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Map uname -s to GoReleaser Os names
case "$OS" in
  linux)  EXT="tar.gz" ;;
  darwin) EXT="tar.gz" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

FILENAME="agentkit_${VERSION}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/ejyle/agentkit/releases/download/v${VERSION}/${FILENAME}"

# Download binary and checksums
curl -fsSL "$URL" -o "/tmp/${FILENAME}"
curl -fsSL "https://github.com/ejyle/agentkit/releases/download/v${VERSION}/checksums.txt" -o /tmp/checksums.txt

# Verify checksum
cd /tmp && sha256sum --check --ignore-missing checksums.txt

# Install
INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "$INSTALL_DIR"
tar -xzf "/tmp/${FILENAME}" -C /tmp agentkit
mv /tmp/agentkit "$INSTALL_DIR/agentkit"
chmod +x "$INSTALL_DIR/agentkit"
echo "Installed to $INSTALL_DIR/agentkit"
echo "Add to PATH: export PATH=\"\$HOME/.local/bin:\$PATH\""
```

**macOS SHA256 note:** macOS uses `shasum -a 256` not `sha256sum`. Script must detect and use the right command:
```bash
if command -v sha256sum >/dev/null 2>&1; then
    SHA_CMD="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
    SHA_CMD="shasum -a 256"
else
    echo "Cannot verify checksum — sha256sum/shasum not found"; exit 1
fi
```

### Anti-Patterns to Avoid

- **Using `brews` in .goreleaser.yaml:** Fully deprecated in v2.16. Use `homebrew_casks`. [CITED: goreleaser.com/deprecations/]
- **Missing `fetch-depth: 0`:** GoReleaser fails to generate changelog without full git history. [CITED: goreleaser.com/ci/actions/]
- **Missing `id-token: write` permission:** Cosign OIDC keyless signing fails silently or with a cryptic error. [CITED: goreleaser.com/blog/supply-chain-security/]
- **Using `runtime.GOOS` / `runtime.GOARCH` for version string:** These reflect the host that ran the build, not the target. Use ldflags `{{.Os}}` / `{{.Arch}}` from GoReleaser templates which reflect the cross-compiled target.
- **Not using `CGO_ENABLED=0`:** CGO introduces dynamic libc dependency; breaks "no runtime dependency" requirement (CLI-10).
- **Snapshot publishing to Homebrew tap:** Skip `publish,sign,homebrew` on snapshot builds to avoid polluting the tap with dev versions.
- **Cobra `Version` field raw string:** Cobra prepends `<appname> version `. Call `rootCmd.SetVersionTemplate("{{.Version}}\n")` to get clean D-04 output.
- **Old cosign signing args:** Using `--output-certificate` + `--output-signature` (cosign v2 style) instead of `--bundle` (cosign v3). The old approach still works but produces 2 files and is not the current standard.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cross-platform build matrix | Custom Makefile per platform | GoReleaser build matrix | GoReleaser handles Windows .exe suffix, archive format switching, archive naming, ldflags interpolation automatically |
| Homebrew tap cask generation | Manual Ruby cask file | GoReleaser `homebrew_casks` | GoReleaser generates correct cask Ruby with version, SHA256, download URL from release metadata |
| SHA256 checksum generation | Manual sha256sum loop | GoReleaser `checksum` block | GoReleaser creates `checksums.txt` for all archives; format expected by `cosign verify-blob` |
| Artifact signing | Custom GPG/key management | Cosign keyless (OIDC) | No key rotation, no secret storage; GitHub OIDC provides ephemeral identity tied to workflow run |
| Binary PATH detection | Custom which/where implementation | `os/exec.LookPath` | stdlib; handles cross-platform PATH search correctly |

**Key insight:** GoReleaser's template system (`{{.Version}}`, `{{.Os}}`, `{{.Arch}}`) combined with `homebrew_casks` eliminates the need for any scripted release automation — one YAML replaces hundreds of lines of CI shell scripting.

---

## Common Pitfalls

### Pitfall 1: `brews` vs `homebrew_casks` Confusion
**What goes wrong:** Using deprecated `brews` key in `.goreleaser.yaml`. GoReleaser v2.16+ emits a hard error or ignored the section.
**Why it happens:** Most tutorials and examples pre-date v2.10 and show `brews`.
**How to avoid:** Always use `homebrew_casks` in the config. Check GoReleaser version in goreleaser-action (`version: "~> v2"` pinned to latest v2).
**Warning signs:** GoReleaser log says "brews is deprecated" or tap repo receives no update after release.

### Pitfall 2: Missing Full Git History in CI
**What goes wrong:** GoReleaser exits with "current tag is not a release tag" or generates empty changelog.
**Why it happens:** `actions/checkout@v4` default `fetch-depth: 1` (shallow clone).
**How to avoid:** Always set `fetch-depth: 0` in checkout step.
**Warning signs:** GoReleaser error mentioning "couldn't find previous tag" or empty release notes.

### Pitfall 3: Cosign OIDC Permission Missing
**What goes wrong:** Cosign keyless signing fails with "OIDC token not available" or a 403 error.
**Why it happens:** Workflow missing `id-token: write` in `permissions`.
**How to avoid:** Add at job level: `permissions: contents: write / id-token: write`.
**Warning signs:** Error from cosign referencing OIDC or token during the `signs` step.

### Pitfall 4: ldflags GOOS/GOARCH Reflecting Build Host
**What goes wrong:** `agentkit --version` on Linux shows `agentkit/0.1.0 (linux/amd64)` even though the binary is a darwin build.
**Why it happens:** Using `runtime.GOOS` / `runtime.GOARCH` (evaluated at run time) instead of ldflags from GoReleaser's `{{.Os}}` / `{{.Arch}}` (injected at cross-compile time for the target).
**How to avoid:** In `.goreleaser.yaml` ldflags, use `{{.Os}}` and `{{.Arch}}` — these are GoReleaser template vars bound to the target GOOS/GOARCH for each matrix combination.
**Warning signs:** `--version` output shows build machine's OS/arch, not binary's target platform.

### Pitfall 5: Doctor Network Check Hanging
**What goes wrong:** `agentkit doctor` hangs for 30+ seconds on a slow or blocked network.
**Why it happens:** `http.DefaultClient` has no timeout.
**How to avoid:** Always use `context.WithTimeout(ctx, 5*time.Second)` and `http.NewRequestWithContext`. Never use `http.Get()` in doctor.
**Warning signs:** Doctor appears to hang at "registry reachable" line.

### Pitfall 6: Homebrew Tap PAT Scope
**What goes wrong:** GoReleaser fails to push cask to `ejyle/homebrew-agentkit` with a 403.
**Why it happens:** Using a fine-grained PAT (limited to the source repo) or a PAT without `repo` scope on the tap repo.
**How to avoid:** Classic PAT with `repo` scope, or fine-grained PAT with `Contents: write` on `ejyle/homebrew-agentkit`. Store as `HOMEBREW_TAP_GITHUB_TOKEN` secret in `ejyle/agentkit` repo settings.
**Warning signs:** GoReleaser log shows "remote: Permission to ejyle/homebrew-agentkit.git denied".

### Pitfall 7: Windows Binary in tar.gz
**What goes wrong:** Windows users can't easily open `.tar.gz` — expected is `.zip`.
**How to avoid:** Use `format_overrides` in archives section: `- goos: windows / format: zip`. GoReleaser automatically adds `.exe` suffix to Windows binary — no manual config needed.

---

## Code Examples

### internal/version/version.go (complete)
```go
// Source: goreleaser.com/cookbooks/using-main.version/ [CITED pattern, adapted]
package version

import "fmt"

// Version, GOOS, GOARCH are set at build time via ldflags.
// When building locally without GoReleaser, these default to "dev" / "unknown".
var (
    Version = "dev"
    GOOS    = "unknown"
    GOARCH  = "unknown"
)

// String returns the version string in D-04 format: agentkit/0.1.0 (darwin/arm64)
func String() string {
    return fmt.Sprintf("agentkit/%s (%s/%s)", Version, GOOS, GOARCH)
}
```

### cmd/root.go addition (version wiring)
```go
// [ASSUMED — Cobra Version field pattern]
import "github.com/ejyle/agentkit/internal/version"

var rootCmd = &cobra.Command{
    Use:     "agentkit",
    Short:   "AI agent skill and MCP server manager",
    Version: version.String(),
    // ...
}

func init() {
    rootCmd.SetVersionTemplate("{{.Version}}\n")
    // ... existing flags ...
}
```

### GoReleaser signs block (cosign v3 bundle approach)
```yaml
# Source: goreleaser.com/blog/cosign-v3/ [CITED]
signs:
  - cmd: cosign
    artifacts: checksum
    args:
      - sign-blob
      - "--bundle=${signature}"
      - "${artifact}"
      - "--yes"
```

The `${signature}` variable in GoReleaser resolves to `checksums.txt.sigstore.json`. [ASSUMED — GoReleaser template variable name for bundle output]

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| GoReleaser `brews` (Homebrew formulas) | `homebrew_casks` (Homebrew casks) | GoReleaser v2.10 (deprecated), v2.16 (removed) | Must use `homebrew_casks` — formulas are for source-compiled tools, casks for pre-built binaries |
| Cosign `--output-certificate` + `--output-signature` (2 files) | `--bundle` (single `.sigstore.json`) | cosign v3 | Simpler signing; one bundle file; GoReleaser `signs` block uses `--bundle` |
| `COSIGN_EXPERIMENTAL=1` env var | Not needed (GA since cosign v1.13+) | cosign v1.13 | Remove `COSIGN_EXPERIMENTAL=1` from workflow env |
| `goreleaser-action@v6` | `goreleaser-action@v7` | 2024-2025 | v7 required for Pro v2.7.0+ versioning scheme; use v7 |
| `brews[].tap` field | `homebrew_casks[].repository` block | GoReleaser v2 | New structure uses `repository.owner/name/token` |

**Deprecated/outdated:**
- `brews` key in `.goreleaser.yaml`: replaced by `homebrew_casks` — do not use.
- `COSIGN_EXPERIMENTAL=1`: no longer needed in cosign v1.13+.
- Separate cosign `--output-certificate` / `--output-signature` flags: replaced by `--bundle` in cosign v3.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `goreleaser-action@v7` is the current recommended version | Standard Stack, GH Actions pattern | If v7 has breaking changes vs v6, use v6; impact is CI-only |
| A2 | `${signature}` in GoReleaser signs block resolves to the bundle output path when using `--bundle` | Code Examples, goreleaser.yaml pattern | If wrong, cosign step fails; easy to debug in CI logs |
| A3 | cosign `--bundle` flag produces `checksums.txt.sigstore.json` as the output filename | Signs block pattern | If naming differs, install script or verification docs need updating |
| A4 | `brew install ejyle/agentkit/agentkit` works with casks (not just formulas) | Homebrew tap section | casks in third-party taps are supported; confirmed by GoReleaser discussions |
| A5 | `homebrew_casks` is fully supported in GoReleaser v2.16+ for third-party taps | Standard Stack | Core to D-05/D-06; if wrong, need to investigate formula workaround |
| A6 | GoReleaser automatically adds `.exe` to Windows binary filenames | Anti-patterns | If not, Windows binary won't run without manual rename |
| A7 | `COSIGN_EXPERIMENTAL=1` is not needed for keyless signing | GitHub Actions workflow | If cosign version in CI is older, add env var; harmless to add defensively |

---

## Open Questions

1. **`homebrew_casks` in third-party taps: formula vs cask install command**
   - What we know: `brew install owner/tap/name` works for both formulas and casks; GoReleaser generates the cask Ruby file.
   - What's unclear: Whether `brew install ejyle/agentkit/agentkit` auto-taps and installs a cask correctly vs requiring `brew tap` first.
   - Recommendation: Verify by actually running the first release; the tap + cask install in one command is standard behavior.

2. **Snapshot pre-release vs draft in GitHub**
   - What we know: GoReleaser `--snapshot --skip=publish` means no GitHub release created at all on main push.
   - What's unclear: Whether user wants a draft/pre-release GitHub release visible from main builds, or just local artifacts.
   - Recommendation (discretion): Skip publishing entirely on snapshot — keep it local to CI for build verification. Avoids cluttering GitHub releases.

3. **doctor checks all targets vs only installed**
   - What we know: D-08 says "each installed target assistant's config directory exists".
   - What's unclear: "Installed" = has skills/MCP via agentkit, or just "assistant app is present on machine".
   - Recommendation (discretion): Check all 5 supported targets with `⚠` (warn, not fail) if directory missing — consistent with brew doctor style that warns about uninstalled optional tools.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| GitHub Actions runner | release.yml | ✓ (CI) | ubuntu-latest | — |
| Go toolchain | goreleaser build | ✓ (via setup-go@v5) | stable | — |
| GoReleaser v2 | .goreleaser.yaml | ✓ (via goreleaser-action@v7) | ~> v2 | — |
| cosign | keyless signing | ✓ (via cosign-installer@v3) | latest | Skip signing (degrade gracefully) |
| `ejyle/homebrew-agentkit` repo | homebrew_casks push | Must be created | — | Create before first release |
| `HOMEBREW_TAP_GITHUB_TOKEN` secret | homebrew_casks auth | Must be configured | — | Release works; tap not updated |
| Homebrew (macOS local) | `brew install` install path | macOS only | — | curl\|sh for Linux |

**Missing dependencies with no fallback:**
- `ejyle/homebrew-agentkit` GitHub repo must exist before the first release (GoReleaser cannot create repos, only push to them).
- `HOMEBREW_TAP_GITHUB_TOKEN` secret must be added to `ejyle/agentkit` repo secrets before tag push.

**Missing dependencies with fallback:**
- cosign not available: `--skip=sign` flag can be added to GoReleaser args; sign step skipped.

---

## Security Domain

`security_enforcement` not explicitly set to false in config — including this section.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | N/A (distribution tooling, no user auth) |
| V3 Session Management | no | N/A |
| V4 Access Control | partial | PAT scoped to `repo` on tap repo only; `GITHUB_TOKEN` auto-scoped |
| V5 Input Validation | yes | install.sh validates checksum before installing; doctor fails fast on bad inputs |
| V6 Cryptography | yes | Cosign keyless signing via sigstore; SHA256 checksums; never hand-roll |

### Known Threat Patterns for Distribution Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Tampered binary download | Tampering | SHA256 checksum verification in install.sh before binary execution |
| Supply chain via dependency | Tampering | `CGO_ENABLED=0` + go.sum pinning; GoReleaser build from clean checkout |
| PAT token leakage | Information Disclosure | PAT stored as GitHub Actions secret; never logged; scoped to tap repo only |
| Malicious tap push | Tampering | PAT has only `Contents: write` on tap repo; cannot modify source repo |
| install.sh MITM | Spoofing | SHA256 verification catches tampered downloads; use `curl -fsSL` (fail on error) |

---

## Sources

### Primary (HIGH confidence)
- [GoReleaser Customization](https://goreleaser.com/customization/) — overall v2 config structure
- [GoReleaser GitHub Actions](https://goreleaser.com/ci/actions/) — workflow structure, permissions, fetch-depth
- [GoReleaser Homebrew Casks](https://goreleaser.com/customization/publish/homebrew_casks/) — homebrew_casks config
- [GoReleaser Homebrew Formulas (deprecated)](https://goreleaser.com/customization/publish/homebrew_formulas/) — confirmed deprecation
- [GoReleaser Deprecations](https://goreleaser.com/deprecations/) — brews deprecation timeline
- [GoReleaser Supply Chain Security](https://goreleaser.com/blog/supply-chain-security/) — cosign + OIDC workflow
- [GoReleaser Cosign v3 migration](https://goreleaser.com/blog/cosign-v3/) — bundle flag, new signing approach
- [GoReleaser Snapshots](https://goreleaser.com/customization/publish/snapshots/) — snapshot version template
- [GoReleaser Templates](https://goreleaser.com/customization/templates/) — {{.Version}}, {{.Os}}, {{.Arch}} template vars
- [GoReleaser ldflags cookbook](https://goreleaser.com/cookbooks/using-main.version/) — ldflags -X pattern
- [goreleaser/example-supply-chain](https://github.com/goreleaser/example-supply-chain) — reference config with keyless signing
- [Homebrew Tap docs](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap) — tap creation, cask naming

### Secondary (MEDIUM confidence)
- [GoReleaser v2.16 announcement](https://goreleaser.com/blog/goreleaser-v2.16/) — confirmed brews fully deprecated
- [goreleaser-action releases](https://github.com/goreleaser/goreleaser-action/releases) — v7 is current
- [sigstore/cosign-installer](https://github.com/sigstore/cosign-installer) — official CI installer action
- [DigitalOcean ldflags tutorial](https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications) — ldflags pattern confirmation

### Tertiary (LOW confidence)
- WebSearch results on uname -m GOARCH mapping — standard pattern, needs integration testing on real machines

---

## Metadata

**Confidence breakdown:**
- GoReleaser v2 config: MEDIUM — `homebrew_casks` confirmed deprecated from `brews`, exact field names need verification against current docs
- GitHub Actions workflow: HIGH — official docs pattern, well-established
- Cosign keyless signing: MEDIUM — cosign v3 bundle approach confirmed; exact GoReleaser `${signature}` var name is ASSUMED
- doctor command: HIGH — pure Go stdlib patterns, straightforward
- install.sh: MEDIUM — uname mapping is standard; macOS `shasum` vs Linux `sha256sum` requires test

**Research date:** 2026-06-09
**Valid until:** 2026-07-09 (GoReleaser moves fast; verify homebrew_casks field names against current docs before implementation)

---

## RESEARCH COMPLETE
