# Phase 3: Bundled Skills - Pattern Map

**Mapped:** 2026-06-09
**Files analyzed:** 9 (code files to create/modify) + skills/agents content directories
**Analogs found:** 8 / 9

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/installer/github_release.go` | installer | file-I/O + request-response | `internal/installer/binary.go` | exact |
| `internal/installer/github_release_test.go` | test | request-response | `internal/installer/binary_test.go` | exact |
| `internal/bundle/bundles.go` | config loader | transform | `internal/registry/local.go` | role-match |
| `internal/bundle/bundles.json` | config data | — | `testdata/registry.json` | data analog |
| `internal/domain/package.go` (modify) | model | — | itself | self |
| `internal/installer/installer.go` (modify) | factory | — | itself | self |
| `internal/config/paths.go` (modify) | utility | — | itself | self |
| `cmd/install.go` (modify) | controller | request-response | itself | self |
| `internal/service/install.go` (modify) | service | request-response | itself | self |
| `skills/<name>/SKILL.md` + `references/` | content | — | none in codebase | no analog |
| `agents/auto-researcher/AGENT.md` | content | — | none in codebase | no analog |

---

## Pattern Assignments

### `internal/installer/github_release.go` (installer, file-I/O + request-response)

**Analog:** `internal/installer/binary.go`

**Imports pattern** (`binary.go` lines 1–14):
```go
package installer

import (
    "crypto/sha256"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "path/filepath"

    "github.com/ejyle/agentkit/internal/config"
    "github.com/ejyle/agentkit/internal/domain"
)
```

New installer will add: `archive/tar`, `compress/gzip`, `bytes`, `strings`, `sync`.

**Struct + constructor pattern** (`binary.go` lines 19–33):
```go
type BinaryInstaller struct {
    client  *http.Client
    binPath string // empty means use config.AgentBinPath() at runtime
}

func NewBinaryInstaller() *BinaryInstaller {
    return &BinaryInstaller{client: http.DefaultClient}
}

// Test-injectable constructor with injected client + dir:
func NewBinaryInstallerWithBinDir(client *http.Client, binDir string) *BinaryInstaller {
    return &BinaryInstaller{client: client, binPath: binDir}
}
```

Mirror for `GitHubReleaseInstaller`:
```go
type GitHubReleaseInstaller struct {
    client  *http.Client
    version string   // set via ldflags; empty falls back to "latest"
    cache   sync.Map // in-process: key "repo@version" -> []byte tarball
}

func NewGitHubReleaseInstaller() *GitHubReleaseInstaller {
    return &GitHubReleaseInstaller{client: http.DefaultClient, version: Version}
}

func NewGitHubReleaseInstallerWithClient(client *http.Client, version string) *GitHubReleaseInstaller {
    return &GitHubReleaseInstaller{client: client, version: version}
}
```

**Method() + IsAvailable() pattern** (`binary.go` lines 36–43):
```go
func (b *BinaryInstaller) Method() domain.InstallMethod {
    return domain.InstallMethodBinary
}

func (b *BinaryInstaller) IsAvailable() bool {
    return true // no runtime prerequisite
}
```

**HTTPS enforcement pattern** (`binary.go` lines 52–56):
```go
u, err := url.Parse(spec.URL)
if err != nil || u.Scheme != "https" {
    return ErrInsecureURL
}
```

**HTTP download to memory pattern** (`binary.go` lines 58–67):
```go
resp, err := b.client.Get(spec.URL)
if err != nil {
    return fmt.Errorf("binary download failed: %w", err)
}
defer resp.Body.Close()
data, err := io.ReadAll(resp.Body)
if err != nil {
    return fmt.Errorf("reading download body: %w", err)
}
```

**Atomic write to disk pattern** (`binary.go` lines 91–114):
```go
// Write to temp file in same dir, then rename (atomic).
tmpFile, err := os.CreateTemp(binDir, spec.Package+".*.tmp")
...
if err := os.Rename(tmpPath, outPath); err != nil {
    os.Remove(tmpPath)
    return fmt.Errorf("renaming binary: %w", err)
}
```

For the tarball cache write, use `renameio.WriteFile` (same as `registry/cache.go` line 46):
```go
// internal/registry/cache.go lines 38–47
func saveCache(path string, c CachedManifest) error {
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }
    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }
    return renameio.WriteFile(path, data, 0644)
}
```

**Cache path pattern** (`config/paths.go` lines 20–27 — follow this for `TarballCachePath`):
```go
func ManifestCachePath(registryID string) (string, error) {
    base, err := os.UserCacheDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(base, "agentkit", registryID, "manifest.json"), nil
}
```

New function to add to `internal/config/paths.go`:
```go
func TarballCachePath(repo, version string) (string, error) {
    base, err := os.UserCacheDir()
    if err != nil {
        return "", err
    }
    slug := strings.ReplaceAll(repo, "/", "-")
    return filepath.Join(base, "agentkit", "releases", slug, version, "tarball.tar.gz"), nil
}
```

**Sentinel error pattern** (`installer/installer.go` lines 11–23):
```go
var (
    ErrNodeNotFound     = errors.New("node not found on PATH; install Node.js ...")
    ErrChecksumMismatch = errors.New("SHA256 checksum mismatch: ...")
    ErrInsecureURL      = errors.New("insecure download URL: only https:// URLs are allowed")
)
```

Add to same file:
```go
ErrGitHubReleaseNotFound = errors.New("github-release: tarball not found for version")
```

---

### `internal/installer/github_release_test.go` (test, request-response)

**Analog:** `internal/installer/binary_test.go`

**Test package + httptest pattern** (`binary_test.go` lines 1–14):
```go
package installer_test

import (
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "testing"

    "github.com/ejyle/agentkit/internal/domain"
    "github.com/ejyle/agentkit/internal/installer"
)
```

**httptest.NewTLSServer pattern** (`binary_test.go` lines 22–27):
```go
ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write(content)
}))
defer ts.Close()

inst := installer.NewBinaryInstallerWithBinDir(ts.Client(), binDir)
```

For `github_release_test.go`, the test server must serve a real `.tar.gz` payload built in-test using `archive/tar` + `compress/gzip`. Tests verify:
1. Skill subdir extracted to correct path
2. Path traversal attack rejected (`../../etc/passwd` entry)
3. Correct tarball URL construction (archive root prefix stripped correctly)
4. Cache: second install of same version skips HTTP request

---

### `internal/bundle/bundles.go` (config loader, transform)

**Analog:** `internal/registry/local.go`

**JSON unmarshal from file pattern** (`registry/local.go` lines 30–39):
```go
func (r *LocalFileRegistry) load() (*domain.Manifest, error) {
    data, err := os.ReadFile(r.path)
    if err != nil {
        return nil, fmt.Errorf("reading registry file %q: %w", r.path, err)
    }
    var m domain.Manifest
    if err := json.Unmarshal(data, &m); err != nil {
        return nil, fmt.Errorf("parsing registry file %q: %w", r.path, err)
    }
    return &m, nil
}
```

For `bundles.go`, the data is embedded (no file path needed at runtime):
```go
package bundle

import (
    _ "embed"
    "encoding/json"
    "fmt"
)

//go:embed bundles.json
var bundlesData []byte

type BundleDef struct {
    Packages []string `json:"packages"`
}

type BundleManifest struct {
    Bundles map[string]BundleDef `json:"bundles"`
}

func LoadBundles() (*BundleManifest, error) {
    var m BundleManifest
    if err := json.Unmarshal(bundlesData, &m); err != nil {
        return nil, fmt.Errorf("parsing bundles.json: %w", err)
    }
    return &m, nil
}

func (m *BundleManifest) Resolve(name string) ([]string, error) {
    b, ok := m.Bundles[name]
    if !ok {
        return nil, fmt.Errorf("bundle %q not found; available: cloud, dev, context", name)
    }
    return b.Packages, nil
}
```

---

### `internal/bundle/bundles.json` (config data)

**Analog:** `testdata/registry.json` (project's JSON data format convention)

JSON must use the same indentation style (2-space, object keys lowercase):
```json
{
  "bundles": {
    "cloud":   { "packages": ["aws", "gcp", "azure"] },
    "dev":     { "packages": ["playwright", "github", "cicd"] },
    "context": { "packages": ["context-mode", "rtk", "serena"] }
  }
}
```

---

### `internal/domain/package.go` (modify — add constant + fields)

**Current file** (`domain/package.go` lines 7–39).

Add after line 18 (after `InstallMethodDocker`):
```go
// InstallMethodGitHubRelease fetches a skill from the GitHub release tarball.
InstallMethodGitHubRelease InstallMethod = "github-release"
```

Add to `InstallSpec` struct after `Args` field (line 38):
```go
// Repo and Path are used by the github-release install method only.
Repo string `json:"repo,omitempty"` // e.g. "ejyle/agentkit"
Path string `json:"path,omitempty"` // e.g. "skills/aws"
```

Both fields use `omitempty` — existing testdata/registry.json unmarshals unchanged.

---

### `internal/installer/installer.go` (modify — add factory case)

**Current file** (`installer/installer.go` lines 36–51).

Add case before the `default` (line 48):
```go
case domain.InstallMethodGitHubRelease:
    return NewGitHubReleaseInstaller(), nil
```

---

### `internal/config/paths.go` (modify — add TarballCachePath)

**Current file** (`config/paths.go`). Append new function following the `ManifestCachePath` pattern (lines 20–27). See excerpt above in the `github_release.go` section.

---

### `cmd/install.go` (modify — add --bundle flag + dispatch)

**Analog:** itself (the existing `runInstall` goroutine + channel pattern)

**Flag registration pattern** (add to `init()` — currently lines 33–35):
```go
func init() {
    rootCmd.AddCommand(installCmd)
    // Add to installCmd flags (not persistent — bundle is install-only):
    installCmd.Flags().StringP("bundle", "b", "", "Install a preset bundle (cloud, dev, context)")
}
```

**Args change** (line 29 — change `ExactArgs(1)` to):
```go
Args: cobra.RangeArgs(0, 1),
```

**Dispatch pattern at top of `runInstall`** (after line 39):
```go
bundleName, _ := cmd.Flags().GetString("bundle")
if bundleName != "" {
    return runBundleInstall(cmd, bundleName, target, svc)
}
// ... existing single-package path unchanged below
```

**Parallel bundle install pattern** (new `runBundleInstall` function, following D-04 best-effort semantics using `sync.WaitGroup` — NOT errgroup):
```go
type bundleResult struct {
    name string
    pkg  *domain.Package
    err  error
}

func runBundleInstall(cmd *cobra.Command, bundleName, target string, svc *service.InstallService) error {
    manifest, err := bundle.LoadBundles()
    if err != nil {
        return err
    }
    pkgNames, err := manifest.Resolve(bundleName)
    if err != nil {
        return err
    }

    results := make([]bundleResult, len(pkgNames))
    var wg sync.WaitGroup
    for i, name := range pkgNames {
        wg.Add(1)
        go func(idx int, n string) {
            defer wg.Done()
            pkg, err := svc.Install(n, target)
            results[idx] = bundleResult{name: n, pkg: pkg, err: err}
        }(i, name)
    }
    wg.Wait()

    // Print per-package result lines (D-04 best-effort output):
    failed := 0
    for _, r := range results {
        if r.err != nil {
            fmt.Fprintf(os.Stderr, "  %s ✗ %s\n", r.name, r.err)
            failed++
        } else {
            fmt.Printf("  %s ✓\n", r.name)
        }
    }
    fmt.Printf("%d/%d installed", len(pkgNames)-failed, len(pkgNames))
    if failed > 0 {
        fmt.Printf(" — %d failed\n", failed)
        os.Exit(1) // D-05: exit code 1 on any failure
    }
    fmt.Println()
    return nil
}
```

**Import to add:** `"sync"`, `"github.com/ejyle/agentkit/internal/bundle"`.

**Error output style** — copy from existing `handleInstallError` (lines 118–121):
```go
fmt.Fprintf(os.Stderr, "✗ Error: %s\n", err.Error())
```

---

### `internal/service/install.go` (modify — gate WriteSkill for github-release)

**Current file** (`service/install.go` lines 142–148) calls `WriteSkill` unconditionally for all skill types.

**Required change** (gate around the `WriteSkill` call, line 143):
```go
// Step 6b for skills: write skill placeholder UNLESS the installer already placed files on disk.
if pkg.Install.Method != domain.InstallMethodGitHubRelease {
    if err := s.adapter.WriteSkill(name, map[string][]byte{
        "SKILL.md": []byte(""),
    }); err != nil {
        return nil, fmt.Errorf("writing skill files for %q: %w", name, err)
    }
}
```

This prevents `WriteSkill` from clobbering the files already extracted by `GitHubReleaseInstaller.Install()` (RESEARCH.md Pitfall 5).

---

## Shared Patterns

### Sentinel Errors
**Source:** `internal/installer/installer.go` lines 11–23
**Apply to:** `github_release.go` (add `ErrGitHubReleaseNotFound` to the existing `var` block in `installer.go`)
```go
var (
    ErrNodeNotFound     = errors.New("node not found on PATH; ...")
    ErrChecksumMismatch = errors.New("SHA256 checksum mismatch: ...")
    ErrInsecureURL      = errors.New("insecure download URL: only https:// URLs are allowed")
    ErrUvxNotFound      = errors.New("uvx not found on PATH; ...")
    ErrDockerNotFound   = errors.New("docker not found on PATH; ...")
    // Add:
    ErrGitHubReleaseNotFound = errors.New("github-release: tarball not found; check version tag exists")
)
```

### HTTP Download to Memory
**Source:** `internal/installer/binary.go` lines 58–67
**Apply to:** `github_release.go` — same `client.Get` + `io.ReadAll` pattern for tarball download

### Atomic File Write
**Source:** `internal/registry/cache.go` lines 38–47 (`renameio.WriteFile`)
**Apply to:** `github_release.go` tarball disk cache write

### Path Resolution via `os.UserCacheDir`
**Source:** `internal/config/paths.go` lines 20–27 (`ManifestCachePath`)
**Apply to:** `TarballCachePath` addition in `config/paths.go`

### Test-Injectable Constructor
**Source:** `internal/installer/binary.go` lines 29–33 (`NewBinaryInstallerWithBinDir`)
**Apply to:** `github_release.go` — provide `NewGitHubReleaseInstallerWithClient(client, version)` for test injection

### httptest.NewTLSServer Test Pattern
**Source:** `internal/installer/binary_test.go` lines 22–29
**Apply to:** `github_release_test.go` — serves fake `.tar.gz` payload

### JSON Unmarshal from Embedded Bytes
**Source:** `internal/registry/local.go` lines 30–39 (pattern); `//go:embed` directive is new (no codebase analog — use standard Go 1.16+ embed pattern)
**Apply to:** `internal/bundle/bundles.go`

### `os.Exit(1)` on Fatal Install Error
**Source:** `cmd/install.go` lines 121, 149 — uses `os.Exit(1)` directly (not `return err`) for user-facing failures
**Apply to:** `runBundleInstall` partial failure case (D-05)

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `skills/<name>/SKILL.md` | content | — | No skill content files exist in repo yet; use agentskills.io spec frontmatter + CONTEXT.md §D-08 |
| `skills/<name>/references/*.md` | content | — | Same — no reference files exist yet |
| `skills/external/<name>/SKILL.md` | content | — | External skill adaptations; no prior format in repo |
| `agents/auto-researcher/AGENT.md` | content | — | No agent content files exist yet |
| `//go:embed` usage | config | — | No existing embed in codebase; use standard Go 1.16+ `_ "embed"` import |

---

## Metadata

**Analog search scope:** `internal/installer/`, `internal/registry/`, `internal/config/`, `internal/domain/`, `internal/service/`, `cmd/`, `testdata/`
**Files scanned:** 14 Go files + 1 JSON file
**Pattern extraction date:** 2026-06-09
