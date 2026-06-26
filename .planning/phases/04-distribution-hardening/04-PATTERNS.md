# Phase 4: Distribution & Hardening - Pattern Map

**Mapped:** 2026-06-09
**Files analyzed:** 7 (5 new, 2 modified)
**Analogs found:** 4 / 7

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `cmd/doctor.go` | command/controller | request-response | `cmd/list.go` | role-match |
| `internal/version/version.go` | utility/package | transform | `internal/config/paths.go` | role-match |
| `.goreleaser.yaml` | config | batch | none | no analog |
| `.github/workflows/release.yml` | config/CI | event-driven | none | no analog |
| `scripts/install.sh` | utility/script | request-response | none | no analog |
| `cmd/root.go` (modify) | command/controller | request-response | `cmd/root.go` (self) | exact |
| `go.mod` (no changes expected) | config | — | `go.mod` (self) | exact |

---

## Pattern Assignments

### `cmd/doctor.go` (command, request-response)

**Analog:** `cmd/list.go` (simplest command with no spinner) and `cmd/install.go` (error formatting + exit-code logic)

**Imports pattern** — copy from `cmd/list.go` lines 1-11, extend with `os/exec`, `net/http`, `context`, `time`, `path/filepath`:
```go
package cmd

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    "github.com/spf13/cobra"

    "github.com/ejyle/agentkit/internal/config"
    "github.com/ejyle/agentkit/internal/version"
)
```

**Command registration pattern** — copy from `cmd/list.go` lines 13-26 and `cmd/search.go` lines 15-29:
```go
var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Diagnose agentkit installation and dependencies",
    Long: `Check that agentkit is correctly installed and that all dependencies are
reachable. Prints a line per check with ✓/⚠/✗ icons.`,
    RunE: runDoctor,
}

func init() {
    rootCmd.AddCommand(doctorCmd)
}
```

**Error output pattern** — copy from `cmd/list.go` lines 33-39 and `cmd/install.go` lines 192-196:
```go
// D-04 format used everywhere in this codebase:
fmt.Fprintf(os.Stderr, "✗ Error: %s\n", err.Error())
fmt.Fprintf(os.Stderr, "Run: agentkit install <name>\n")
os.Exit(1)
```

**Result collection / exit-code pattern** — copy from `cmd/install.go` lines 159-175 (`runBundleInstall` failure tracking):
```go
// collect results, count failures, exit(1) if any fail
failed := 0
for _, r := range results {
    if r.err != nil {
        fmt.Fprintf(os.Stderr, "  %s ✗ %s\n", r.name, r.err)
        failed++
    } else {
        fmt.Printf("  %s ✓\n", r.name)
    }
}
if failed > 0 {
    os.Exit(1)
}
```

**Config path helper reuse** — `internal/config/paths.go` `SkillInstallPath` (lines 53-71) exposes the `~/.claude/skills/<name>` path. For doctor, derive the assistant root dir by stripping the `skills/<name>` suffix:
```go
// assistantRootDir returns ~/.claude, ~/.gemini, etc. for a target.
// Reuses SkillInstallPath("claude", "_probe") and strips /skills/_probe.
func assistantRootDir(target string) (string, error) {
    probe, err := config.SkillInstallPath(target, "_probe")
    if err != nil {
        return "", err
    }
    // probe = ~/.claude/skills/_probe  →  parent parent = ~/.claude
    return filepath.Dir(filepath.Dir(probe)), nil
}
```

**Network check with timeout** — pure stdlib; `http.DefaultClient` has no timeout so always use `context.WithTimeout`:
```go
func checkRegistryReachable(url string) CheckResult {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    req, _ := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil || resp.StatusCode >= 500 {
        return CheckResult{Label: "registry reachable", Status: "fail",
            Hint: "check network connectivity"}
    }
    return CheckResult{Label: "registry reachable", Status: "pass"}
}
```

**Binary-in-PATH check** — stdlib `os/exec.LookPath`:
```go
path, err := exec.LookPath("agentkit")
if err != nil {
    return CheckResult{Label: "agentkit in PATH", Status: "fail",
        Hint: "Add ~/.local/bin to PATH: export PATH=\"$HOME/.local/bin:$PATH\""}
}
return CheckResult{Label: fmt.Sprintf("agentkit in PATH (%s)", version.Version),
    Status: "pass", Message: path}
```

---

### `internal/version/version.go` (utility, transform)

**Analog:** `internal/config/paths.go` (small standalone package, pure stdlib, no external deps, os package usage)

**Package structure pattern** — copy the single-file package structure from `internal/config/paths.go` lines 1-8:
```go
package version

import "fmt"
```

**Complete file** (per D-03 and RESEARCH.md Pattern 3):
```go
package version

import "fmt"

// Version, GOOS, GOARCH are injected at build time via ldflags.
// Defaults apply when building locally without GoReleaser.
var (
    Version = "dev"
    GOOS    = "unknown"
    GOARCH  = "unknown"
)

// String returns the D-04 version string: agentkit/0.1.0 (darwin/arm64)
func String() string {
    return fmt.Sprintf("agentkit/%s (%s/%s)", Version, GOOS, GOARCH)
}
```

**No test file needed** — pure string formatting, no logic branches to test.

---

### `cmd/root.go` (modify — add Version field + doctor registration)

**Self-analog:** `cmd/root.go` lines 1-44 (full file, already read)

**Change 1 — import version package** (after existing `"fmt"` import):
```go
import (
    "fmt"

    "github.com/spf13/cobra"

    "github.com/ejyle/agentkit/internal/version"
)
```

**Change 2 — add Version field to rootCmd struct** (after `Long:` field, line 13):
```go
var rootCmd = &cobra.Command{
    Use:   "agentkit",
    Short: "AI agent skill and MCP server manager",
    Long:  `agentkit installs, updates, and manages AI agent skills, MCP servers, and
coding agents across all major AI coding assistants.`,
    Version: version.String(),
}
```

**Change 3 — override Cobra version template in init()** (to match D-04 exactly, no "agentkit version " preamble):
```go
func init() {
    rootCmd.SetVersionTemplate("{{.Version}}\n")
    // ... existing flags unchanged ...
}
```

Doctor subcommand auto-registers via `cmd/doctor.go`'s own `init()` (same pattern as all other commands).

---

### `.goreleaser.yaml` (config, batch)

**No codebase analog** — use RESEARCH.md Pattern 1 (lines 187-258) verbatim. Key points for planner:
- `version: 2` at top (required for GoReleaser v2)
- `CGO_ENABLED=0` in builds env (required for CLI-10 single binary, no libc dep)
- ldflags reference `github.com/ejyle/agentkit/internal/version` package (from `go.mod` module path, line 1)
- `homebrew_casks` (NOT `brews` — fully deprecated in v2.16)
- `format_overrides` for Windows `.zip`
- `snapshot.version_template` using `{{ .Version }}-SNAPSHOT-{{ .ShortCommit }}`

---

### `.github/workflows/release.yml` (CI config, event-driven)

**No codebase analog** — use RESEARCH.md Pattern 2 (lines 267-322) verbatim. Key points for planner:
- `permissions: id-token: write` is required for cosign OIDC (missing this causes silent failure)
- `fetch-depth: 0` on checkout (shallow clone breaks changelog generation)
- Two jobs: `release` (on `v*` tag), `snapshot` (on `main` push, `--skip=publish,sign,homebrew`)
- `goreleaser/goreleaser-action@v7` with `version: "~> v2"`
- `sigstore/cosign-installer@v3` step before GoReleaser in release job only

---

### `scripts/install.sh` (utility/script, request-response)

**No codebase analog** — use RESEARCH.md Pattern 5 (lines 451-497) verbatim. Key points for planner:
- `uname -m` → GOARCH mapping: `x86_64`→`amd64`, `aarch64|arm64`→`arm64`
- `uname -s | tr '[:upper:]' '[:lower:]'` → OS mapping
- macOS uses `shasum -a 256`, Linux uses `sha256sum` — script must detect and use correct command
- Default install dir: `~/.local/bin` (no root required, matches CLI-10)
- Print PATH hint after install: `export PATH="$HOME/.local/bin:$PATH"`
- `curl -fsSL` (fail on error, follow redirects, silent)

---

## Shared Patterns

### Error Output Format
**Source:** `cmd/install.go` lines 192-195, `cmd/list.go` lines 33-38, `cmd/search.go` lines 46-50
**Apply to:** `cmd/doctor.go`
```go
fmt.Fprintf(os.Stderr, "✗ Error: %s\n", err.Error())
fmt.Fprintf(os.Stderr, "Run: <suggested command>\n")
os.Exit(1)
```

### Exit Code on Failure
**Source:** `cmd/install.go` lines 172-175 (bundle install), `cmd/list.go` line 38
**Apply to:** `cmd/doctor.go` (exit 1 if any `✗` check, exit 0 for `✓` and `⚠`)
```go
if failed > 0 {
    os.Exit(1)
}
```

### Cobra Subcommand Registration
**Source:** All `cmd/*.go` files — each uses an `init()` calling `rootCmd.AddCommand(xyzCmd)`
**Apply to:** `cmd/doctor.go` — same pattern, no deviation
```go
func init() {
    rootCmd.AddCommand(doctorCmd)
}
```

### Config Path Derivation
**Source:** `internal/config/paths.go` lines 53-71 (`SkillInstallPath`)
**Apply to:** `cmd/doctor.go` — derive assistant root dirs without duplicating the target switch
```go
// Use SkillInstallPath(target, "_probe") then filepath.Dir(filepath.Dir(result))
// to get ~/.claude, ~/.gemini, etc.
```

### Unicode Status Icons
**Source:** `cmd/install.go` lines 121, 163-165 (uses `✓`, `✗`, `⚠` as plain string literals)
**Apply to:** `cmd/doctor.go` — consistent icon usage; no lipgloss color required unless desired
```go
const (
    iconPass = "✓"
    iconWarn = "⚠"
    iconFail = "✗"
)
```

---

## No Analog Found

Files with no close match in the codebase (planner should use RESEARCH.md patterns instead):

| File | Role | Data Flow | Reason |
|---|---|---|---|
| `.goreleaser.yaml` | release config | batch | No GoReleaser config exists yet; RESEARCH.md Pattern 1 is the canonical reference |
| `.github/workflows/release.yml` | CI workflow | event-driven | No GitHub Actions workflows exist yet; RESEARCH.md Pattern 2 is the canonical reference |
| `scripts/install.sh` | install script | request-response | No shell scripts exist in this repo; RESEARCH.md Pattern 5 is the canonical reference |

---

## Metadata

**Analog search scope:** `/Users/nithin/Ejyle/coding-agent-utils/cmd/`, `/Users/nithin/Ejyle/coding-agent-utils/internal/`
**Files scanned:** `cmd/root.go`, `cmd/install.go`, `cmd/search.go`, `cmd/list.go`, `internal/config/paths.go`, `internal/ui/spinner.go`, `go.mod`
**Pattern extraction date:** 2026-06-09
