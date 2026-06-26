---
phase: 03-bundled-skills
reviewed: 2026-06-09T00:00:00Z
depth: standard
files_reviewed: 32
files_reviewed_list:
  - cmd/install.go
  - cmd/list.go
  - cmd/root.go
  - cmd/search.go
  - cmd/uninstall.go
  - cmd/update.go
  - go.mod
  - internal/adapter/adapter.go
  - internal/adapter/claude.go
  - internal/adapter/codex.go
  - internal/adapter/factory.go
  - internal/adapter/gemini.go
  - internal/adapter/jsonbase.go
  - internal/adapter/opencode.go
  - internal/adapter/pi.go
  - internal/bundle/bundles.go
  - internal/bundle/bundles.json
  - internal/config/paths.go
  - internal/domain/package.go
  - internal/installer/github_release.go
  - internal/installer/installer.go
  - internal/registry/local.go
  - internal/registry/registry.go
  - internal/service/install.go
  - internal/service/search.go
  - internal/service/uninstall.go
  - internal/service/update.go
  - internal/skill/validate.go
  - internal/ui/spinner.go
  - internal/ui/table.go
  - internal/ui/tty.go
  - testdata/registry.json
findings:
  critical: 5
  warning: 6
  info: 3
  total: 14
status: issues_found
---

# Phase 03: Code Review Report

**Reviewed:** 2026-06-09T00:00:00Z
**Depth:** standard
**Files Reviewed:** 32
**Status:** issues_found

## Summary

Reviewed the full agentkit implementation including the bundle install flow, GitHub release
installer, adapter layer, service layer, registry, and CLI commands. The core architecture
is clean and the path-traversal guard in the tarball extractor is present. However, several
correctness bugs, one security gap, and several logic errors were found — the most critical
being an unbounded tarball download (no size cap), a broken validation gate that always
passes, a `cmd/update.go` hardcoded adapter, an `os.Exit(1)` inside `runBundleInstall` that
bypasses Cobra's error-return contract, and a missing nil-check after validation errors loop.

---

## Critical Issues

### CR-01: No size limit on tarball download — unbounded memory/disk allocation

**File:** `internal/installer/github_release.go:164`
**Issue:** `io.ReadAll(resp.Body)` reads the entire GitHub release tarball into memory with no
size cap. A malicious or misconfigured registry entry pointing at a multi-GB repo will exhaust
heap. There is no `http.MaxBytesReader` wrapper and no `Content-Length` pre-check. The bytes
are also written verbatim to disk cache, so a large tarball fills the user's `~/.cache` partition.
**Fix:**
```go
const maxTarballBytes = 500 << 20 // 500 MiB hard cap
limited := http.MaxBytesReader(nil, resp.Body, maxTarballBytes)
data, err := io.ReadAll(limited)
if err != nil {
    return nil, fmt.Errorf("github-release: reading response body (limit %d bytes): %w", maxTarballBytes, err)
}
```

---

### CR-02: Skill validation error loop never returns an error — gate is broken

**File:** `internal/service/install.go:152-156`
**Issue:** The validation failure gate iterates `result.Errors` and returns inside the loop on
the first error. That part is correct. However, `result.Valid` is checked (`!result.Valid`) at
line 152, but within the body the code does:

```go
if !result.Valid {
    for _, e := range result.Errors {
        return nil, fmt.Errorf("validating installed skill %q: %s", name, e)
    }
}
```

If `result.Valid == false` but `result.Errors` is empty (a reachable state — the validator in
`skill/validate.go` sets `Valid=false` and appends to `Errors` in most paths, but a caller
implementing the `SkillValidator` interface could return `Valid=false` with no Errors), the
`for` loop body never executes and install silently proceeds past the validation gate. The
logic should check `!result.Valid` after iterating all errors and return a fallback error.
**Fix:**
```go
if !result.Valid {
    if len(result.Errors) == 0 {
        return nil, fmt.Errorf("validating installed skill %q: validation failed (no error details)", name)
    }
    for _, e := range result.Errors {
        return nil, fmt.Errorf("validating installed skill %q: %s", name, e)
    }
}
```

---

### CR-03: `cmd/update.go` hardcodes `ClaudeCodeAdapter` regardless of `--target`

**File:** `cmd/update.go:50`
**Issue:** `runUpdate` calls `adapter.NewClaudeCodeAdapter(store)` unconditionally, ignoring
the `--target` flag value. If a user runs `agentkit update --target gemini`, agentkit silently
reads and writes to `~/.claude.json` (or `~/.gemini/settings.json`) via the wrong adapter,
potentially corrupting the wrong assistant's config.

Compare `cmd/install.go:53` where `adapter.NewAdapter(target, store)` is used correctly.
**Fix:**
```go
ad, err := adapter.NewAdapter(target, store)
if err != nil {
    return err
}
```
Replace line 50 (`ad := adapter.NewClaudeCodeAdapter(store)`) with the above two lines.

---

### CR-04: `os.Exit(1)` inside `runBundleInstall` breaks Cobra error propagation

**File:** `cmd/install.go:173`
**Issue:** `runBundleInstall` calls `os.Exit(1)` directly when any package fails instead of
returning an error. Cobra's `RunE` contract expects errors to be returned; callers (tests,
library users, shell scripts using `$?`) can intercept returned errors but cannot intercept
`os.Exit`. Deferred cleanup in the process (e.g., any `defer` in `runInstall` or wrapping
test harnesses) is skipped. The comment `// D-05: exit code 1 on any failure` treats this as
intentional but it contradicts Cobra's `RunE` pattern used in every other command in this
codebase.
**Fix:** Return a sentinel error instead and let Cobra set the exit code:
```go
if failed > 0 {
    fmt.Fprintf(os.Stderr, "%d/%d installed — %d failed\n", len(pkgNames)-failed, len(pkgNames), failed)
    return fmt.Errorf("%d of %d bundle packages failed to install", failed, len(pkgNames))
}
```
Cobra exits with code 1 when `RunE` returns a non-nil error, satisfying D-05.

---

### CR-05: Path traversal guard uses substring match on `rel` — bypassable on Windows

**File:** `internal/installer/github_release.go:220`
**Issue:** The traversal guard does:
```go
if strings.Contains(rel, "..") {
```
This rejects `../foo` but also rejects legitimate filenames containing `..` in the middle
(e.g., `file..txt`). More critically, the guard is a defense-in-depth layer but it is checked
on the **stripped relative path** (`rel`) not the raw tar header name. On case-insensitive
Windows filesystems or with NTFS alternate data streams, `filepath.Join` may not produce a
path starting with `destDirPrefix` even when `rel` appears clean.

The canonical guard already exists two lines below (the `strings.HasPrefix(resolved, destDirPrefix)` check at line 224). The `strings.Contains(rel, "..")` check at line 220 is redundant and subtly wrong: it blocks `file..txt` (a valid filename) while providing no additional security beyond the canonical check. If the canonical check is ever removed the substring guard would still be bypassable via `....//` on some tar implementations (double-dot avoidance).

The real issue is that the `strings.Contains` guard should use path component semantics, not substring match.
**Fix:** Replace the substring check with a proper component-level check, or remove it and rely solely on the canonical boundary check which is already correct:
```go
// Remove the strings.Contains(rel, "..") block entirely.
// The filepath.Join + HasPrefix check below is sufficient and correct on all platforms.
resolved := filepath.Join(destDirClean, rel)
if resolved != destDirClean && !strings.HasPrefix(resolved, destDirPrefix) {
    return fmt.Errorf("github-release: path traversal rejected: %q", hdr.Name)
}
```

---

## Warnings

### WR-01: `bundle.Resolve` error message hardcodes bundle names — stale after any bundle addition

**File:** `internal/bundle/bundles.go:39`
**Issue:** The error message is:
```go
return nil, fmt.Errorf("bundle %q not found; available: cloud, dev, context", name)
```
The list `cloud, dev, context` is hardcoded. If a bundle is added to `bundles.json` the error
message will not reflect it, misleading users. The manifest already contains the full list.
**Fix:**
```go
var names []string
for k := range m.Bundles {
    names = append(names, k)
}
sort.Strings(names)
return nil, fmt.Errorf("bundle %q not found; available: %s", name, strings.Join(names, ", "))
```

---

### WR-02: `cmd/uninstall.go` hardcodes `ClaudeCodeAdapter` regardless of `--target`

**File:** `cmd/uninstall.go:36`
**Issue:** Same class of bug as CR-03. `runUninstall` instantiates `adapter.NewClaudeCodeAdapter(store)`
regardless of the `--target` value. Uninstalling a gemini MCP server with `--target gemini`
will remove it from Claude Code's config instead.
**Fix:**
```go
ad, err := adapter.NewAdapter(target, store)
if err != nil {
    return err
}
svc := service.NewUninstallService(ad, store)
```

---

### WR-03: `extractSubdir` silently drops symlinks and hard links from tarballs

**File:** `internal/installer/github_release.go:228-247`
**Issue:** The `switch hdr.Typeflag` only handles `tar.TypeDir` and `tar.TypeReg`. Entries of
type `tar.TypeSymlink`, `tar.TypeLink`, or `tar.TypeGNUSparse` are silently skipped. A skill
that legitimately contains a symlink (e.g., `references/ -> ../shared/`) will install
silently with missing files, causing a confusing validation error later. The current behavior
is acceptable as a security stance against symlink attacks, but it is not documented and
produces no warning, making debugging very difficult.
**Fix:** Log a warning (to stderr) for any skipped non-directory, non-regular-file entry, or
add an explicit case that returns an error on symlinks:
```go
case tar.TypeSymlink, tar.TypeLink:
    return fmt.Errorf("github-release: symlinks not permitted in skill tarballs: %q", hdr.Name)
default:
    // Ignore other entry types (device files, fifos, etc.) silently.
}
```

---

### WR-04: `skill/validate.go` uses `manifest.Install.Args` as reference list — semantic overload

**File:** `internal/skill/validate.go:60`
**Issue:** `ValidateSkill` interprets `manifest.Install.Args` as the list of required reference
file names (SKL-03). `InstallSpec.Args` is documented as install arguments for npx/uvx
methods and is also used by the GitHub release installer for something else. Using the same
field for two semantically different purposes (install CLI arguments vs. required reference
file names) is a type-safety violation that will cause silent bugs: any skill whose registry
entry has `install.args: ["-y", "--no-telemetry"]` (for an npx variant) would be required to
have `references/-y.md` and `references/--no-telemetry.md`, which will fail validation.
**Fix:** Add a dedicated field to `domain.InstallSpec` or `domain.Package` for required
references, and update `ValidateSkill` to use it:
```go
// In domain/package.go
type Package struct {
    ...
    RequiredRefs []string `json:"required_refs,omitempty"` // SKL-03
}
// In skill/validate.go
for _, ref := range manifest.RequiredRefs {
    ...
}
```

---

### WR-05: `registry.RegistryManager.Resolve` silently swallows all registry errors

**File:** `internal/registry/registry.go:75-81`
**Issue:** When a registry returns an error (network failure, 500 from GitHub raw CDN, JSON
parse error), `Resolve` continues to the next registry and ultimately returns "not found in
any registry". The caller has no indication that one registry was unavailable vs. the package
genuinely not existing. This means a transient network error silently downgrades to a
confusing "package not found" error.
**Fix:** Collect errors separately and include them in the not-found error:
```go
var errs []string
for _, reg := range m.registries {
    pkg, err := reg.Resolve(name)
    if err != nil {
        errs = append(errs, fmt.Sprintf("%s: %v", reg.Name(), err))
        continue
    }
    if pkg != nil {
        return pkg, nil
    }
}
if len(errs) > 0 {
    return nil, fmt.Errorf("%q not found; registry errors: %s", name, strings.Join(errs, "; "))
}
return nil, fmt.Errorf("%q not found in any registry", name)
```

---

### WR-06: `service/update.go` `UpdateAll` returns only the first error — other errors are lost

**File:** `internal/service/update.go:89-97`
**Issue:** `UpdateAll` records `firstErr` and continues, but returns only `firstErr`. All
subsequent update errors are silently discarded. The caller (`cmd/update.go`) prints the
first error and exits, giving no indication of how many additional failures occurred. This
matches the comment "continues updating remaining packages even if one fails" but the
discarded errors are never surfaced.
**Fix:** Return a combined error or a slice of errors:
```go
var errs []string
for _, rec := range records {
    msg, updateErr := s.Update(rec.Name, target)
    if updateErr != nil {
        errs = append(errs, fmt.Sprintf("%s: %v", rec.Name, updateErr))
        continue
    }
    msgs = append(msgs, msg)
}
if len(errs) > 0 {
    return msgs, fmt.Errorf("update errors: %s", strings.Join(errs, "; "))
}
return msgs, nil
```

---

## Info

### IN-01: `go.mod` specifies `go 1.26.3` — a non-existent Go version

**File:** `go.mod:3`
**Issue:** The `go` directive is `go 1.26.3`. As of the knowledge cutoff (August 2025) the
latest stable Go release is 1.24.x. Go 1.26.3 does not exist. This will cause `go mod tidy`
and toolchain auto-download to fail on real systems. Likely a typo for `go 1.23.3` or
`go 1.22.3`.
**Fix:** Use the actual Go version:
```
go 1.23.3
```

---

### IN-02: `BurntSushi/toml` is declared `// indirect` but is directly used

**File:** `go.mod:15`
**Issue:** `github.com/BurntSushi/toml v1.6.0` is marked `// indirect` but is directly
imported by `internal/adapter/codex.go`. Running `go mod tidy` would correct this to a
direct dependency. While not a runtime bug, it indicates `go mod tidy` has not been run,
which may cause CI `go mod verify` checks to fail.
**Fix:** Run `go mod tidy` or remove the `// indirect` marker manually.

---

### IN-03: `handleForeignConflict` prints a `--force` flag hint for a flag that does not exist

**File:** `cmd/install.go:222`
**Issue:** After the user confirms overwrite, the code prints:
```
✗ To force-overwrite foreign config, use: agentkit install <name> --target <target> --force
```
and then calls `os.Exit(1)`. There is no `--force` flag defined anywhere in the codebase.
This message is misleading — users will attempt `--force` and get an "unknown flag" error.
**Fix:** Either implement the `--force` flag, or change the message to indicate the feature
is not yet available:
```go
fmt.Fprintf(os.Stderr, "Force-overwrite is not yet supported. Remove the existing entry manually from your config and re-run.\n")
```

---

_Reviewed: 2026-06-09T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
