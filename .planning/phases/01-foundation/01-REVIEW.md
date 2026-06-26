---
phase: 01-foundation
reviewed: 2026-06-08T00:00:00Z
depth: standard
files_reviewed: 29
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
  - internal/config/paths.go
  - internal/config/store.go
  - internal/domain/installed.go
  - internal/domain/package.go
  - internal/installer/binary.go
  - internal/installer/installer.go
  - internal/installer/npx.go
  - internal/registry/cache.go
  - internal/registry/github.go
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
  - main.go
  - testdata/registry.json
findings:
  critical: 7
  warning: 8
  info: 5
  total: 20
status: issues_found
---

# Phase 01: Code Review Report

**Reviewed:** 2026-06-08T00:00:00Z
**Depth:** standard
**Files Reviewed:** 29
**Status:** issues_found

## Summary

Reviewed the complete foundation layer of `agentkit`: CLI commands, registry client, installer adapters, config store, and UI. The code is generally well-structured with good separation of concerns and consistent use of atomic writes. However, several serious correctness and security defects were found spanning every layer.

Critical issues include: a checksum verification bypass (binary installer skips verification when `spec.Args` is empty), a deadlock in `WriteMCPConfig` that calls a method which re-acquires the store's mutex while already holding it (the store's `GetRecord` is called from within an already-locked `RecordInstalled` path), skill install writing an empty placeholder SKILL.md destroying real content, the `errWriter` struct writing to stdout instead of stderr, and the `handleForeignConflict` function advertising a `--force` flag that does not exist. Warnings include the `go.mod` declaring `go 1.26.3` (a non-existent Go version), unprotected `os.Exit` calls in command handlers that bypass deferred cleanup, and race conditions in the spinner goroutine pattern.

---

## Critical Issues

### CR-01: Binary installer skips SHA256 verification when `spec.Args` is empty

**File:** `internal/installer/binary.go:70-77`
**Issue:** The checksum guard is wrapped in `if len(spec.Args) > 0 && spec.Args[0] != ""`. When a registry manifest entry omits `install.args` (as both entries in `testdata/registry.json` do — `"args": []`), the downloaded binary is installed with **no integrity check at all**. An attacker who compromises the CDN or performs a MITM gets silent arbitrary code execution. The architecture doc calls out SHA256 verification as a mandatory security requirement (T-03-02).

**Fix:** Make a populated `spec.Package` SHA256 (better: a dedicated `Checksum` field in `InstallSpec`) mandatory for binary installs. Return an error rather than skipping:
```go
// In Install():
expected := ""
if len(spec.Args) > 0 {
    expected = spec.Args[0]
}
if expected == "" {
    return fmt.Errorf("binary install for %q has no checksum: refusing install (T-03-02)", spec.Package)
}
sum := sha256.Sum256(data)
actual := fmt.Sprintf("sha256:%x", sum)
if actual != expected {
    return ErrChecksumMismatch
}
```

---

### CR-02: `WriteMCPConfig` calls `store.GetRecord` which acquires the store's mutex — but `RecordInstalled` already holds it, causing a deadlock

**File:** `internal/adapter/claude.go:114`, `internal/config/store.go:36-51`, `internal/config/store.go:90-99`
**Issue:** `WriteMCPConfig` calls `a.store.GetRecord(entry.Name)` (line 114 of claude.go). `GetRecord` acquires `s.mu` (store.go:91). In the install flow (`service/install.go:160`), `WriteMCPConfig` is called before `RecordInstalled`, so the lock is not yet held — the flow works today. However, if any caller invokes `WriteMCPConfig` inside a lock scope (e.g. in a future upgrade path or test), the result is a deadlock. More concretely: `WriteMCPConfig` receives a `*domain.InstalledRecord` ownership parameter (adapter interface line 17) but the concrete implementation ignores it (signature uses `_ *domain.InstalledRecord`) and instead calls `store.GetRecord` independently. This is a layering violation: the adapter reaches back into the store, creating hidden coupling and a future deadlock trap. The ownership parameter exists precisely to pass this information in — it should be used instead.

**Fix:** Use the passed `ownership` parameter instead of re-querying the store:
```go
func (a *ClaudeCodeAdapter) WriteMCPConfig(entry domain.MCPServerEntry, ownership *domain.InstalledRecord) error {
    raw, err := a.readRawConfig()
    if err != nil {
        return err
    }
    // ... extract mcpServers map ...
    if _, exists := mcpServers[entry.Name]; exists {
        if ownership == nil {
            // Foreign conflict
            old := extractEntry(mcpServers[entry.Name])
            old.Name = entry.Name
            return &ErrForeignConflict{OldEntry: old, NewEntry: entry}
        }
        // ownership != nil → agentkit-owned, auto-overwrite
    }
    // ... rest unchanged
}
```
Then pass the ownership record from the install service (do a `GetRecord` lookup before calling `WriteMCPConfig`).

---

### CR-03: Skill install writes an empty `SKILL.md`, destroying real file content

**File:** `internal/service/install.go:143-147`
**Issue:** When installing a skill-type package, the service unconditionally writes `map[string][]byte{"SKILL.md": []byte("")}` — an **empty byte slice** — as the skill content. If a skill already exists (upgrade path), this silently overwrites the user's `SKILL.md` with an empty file. Even on first install, the resulting `SKILL.md` is empty and useless. The comment "For mock/unit tests, dir is empty — validator handles gracefully" suggests this is a stub that was never completed.

**Fix:** The skill installer should download and write the real skill files from the registry package source, not hardcode an empty placeholder. At minimum, guard against overwriting with empty content:
```go
// Do not write empty skill files — this destroys real content.
// Skill file fetching from pkg.Source must be implemented here.
return nil, fmt.Errorf("skill file download not yet implemented for %q", name)
```

---

### CR-04: `errWriter.Write` writes to stdout, not stderr

**File:** `internal/service/install.go:186-189`
**Issue:** The `errWriter` struct is intended as an `io.Writer` that directs to stderr (it is used to emit skill validation warnings and errors). However its `Write` method calls `fmt.Print(string(p))` which writes to **stdout**. This means validation errors and warnings are silently swallowed into stdout rather than surfaced on stderr, and could corrupt machine-parseable output (e.g., when piped).

**Fix:**
```go
func (errWriter) Write(p []byte) (int, error) {
    n, err := os.Stderr.Write(p)
    return n, err
}
```
Also add `import "os"` if not already present.

---

### CR-05: `handleInstallError` and `handleForeignConflict` advertise `--force` flag that does not exist

**File:** `cmd/install.go:145`
**Issue:** `handleForeignConflict` prints `"To force-overwrite foreign config, use: agentkit install %s --target %s --force"` then calls `os.Exit(1)`. The `--force` flag is never defined on `installCmd` (or anywhere in the codebase). Following this instruction produces `"unknown flag: --force"`. This is misleading error output to the user — they cannot resolve the conflict. Moreover, the entire user confirmation prompt at lines 127–148 collects a "y/yes" response and then _ignores it_ (exits with the --force suggestion instead of proceeding), making the prompt a dead interaction.

**Fix:** Either define and implement `--force`, or remove the prompt entirely and display a clear message that force-overwrite is not yet supported:
```go
fmt.Fprintf(os.Stderr, "Overwrite of foreign MCP config entries is not yet supported.\n")
fmt.Fprintf(os.Stderr, "Manually remove mcpServers.%s from your config file and retry.\n", name)
os.Exit(1)
```

---

### CR-06: `go.mod` declares `go 1.26.3` — a non-existent Go version

**File:** `go.mod:3`
**Issue:** The module requires `go 1.26.3`. As of the knowledge cutoff (August 2025), Go's latest release is in the 1.24.x line. Go 1.26 does not exist. This will cause `go build` and `go mod tidy` to fail with `note: module requires Go >= 1.26.3` on any real developer machine, and will break CI on any published toolchain. This is a blocker for anyone trying to build the binary.

**Fix:** Change to the actual minimum required version. Based on the dependencies (bubbletea v1.3.10 requires Go 1.21+), use:
```
go 1.21
```
or the actual version in use on the team's machines.

---

### CR-07: `ConfigStore.NewConfigStore` silently swallows the `InstalledStatePath` error

**File:** `internal/config/store.go:23-26`
**Issue:** `NewConfigStore` calls `InstalledStatePath(target)` and discards the error with `_`. If `os.UserConfigDir()` fails (e.g., `$HOME` is unset in certain container/CI environments), `basePath` will be an empty string `""`. All subsequent read/write operations on the store will operate on a file named `""` or fail with confusing OS errors. The user sees no indication that home directory resolution failed.

**Fix:** Return an error from `NewConfigStore`, or panic-fail fast with a clear message:
```go
func NewConfigStore(target string) (*ConfigStore, error) {
    path, err := InstalledStatePath(target)
    if err != nil {
        return nil, fmt.Errorf("resolving config path for target %q: %w", target, err)
    }
    return &ConfigStore{target: target, basePath: path}, nil
}
```
All callers will need updating to handle the error. Alternatively, defer the error to first use within `loadStateUnlocked`.

---

## Warnings

### WR-01: Race condition between spinner goroutines — `doneCh` may block after `p.Run()` returns on spinner error

**File:** `cmd/install.go:85-89`, `cmd/search.go:75-84`
**Issue:** If `p.Run()` returns an error (the spinner itself errors), execution falls through to `installResult := <-doneCh`. The relay goroutine (lines 75–83) is blocked on `<-resultCh`. If the background install goroutine is still running, the main goroutine blocks on `<-doneCh` indefinitely. This is a goroutine leak and potential hang. The same pattern appears in `cmd/search.go`.

**Fix:** Use a `context.Context` to cancel the install goroutine on spinner error, or use a `select` with a timeout:
```go
if _, err := p.Run(); err != nil {
    fmt.Fprintf(os.Stderr, "spinner error: %v\n", err)
    // drain or timeout rather than block forever
}
select {
case installResult = <-doneCh:
case <-time.After(30 * time.Second):
    fmt.Fprintf(os.Stderr, "timed out waiting for install to complete\n")
    os.Exit(1)
}
```

---

### WR-02: `os.Exit` in command handlers bypasses all deferred cleanup

**File:** `cmd/install.go:118`, `cmd/list.go:38`, `cmd/search.go:49`, `cmd/uninstall.go:44`, `cmd/uninstall.go:49`, `cmd/update.go:68`, `cmd/update.go:72`, `cmd/update.go:97`
**Issue:** Multiple command handlers call `os.Exit(1)` directly after printing error messages. This is problematic because: (1) any `defer` statements in the call stack (e.g., file handles, temp file cleanup) are skipped; (2) Cobra's own cleanup is bypassed; (3) it makes the functions un-testable (tests cannot catch `os.Exit`). The idiomatic Go/Cobra pattern is to return the error and let the root command handle exit.

**Fix:** Replace `os.Exit(1)` with `return fmt.Errorf(...)` in RunE functions. Cobra will print the error and exit. For D-04 formatted messages (custom stderr + suggested command), use a custom error type or write to stderr before returning a sentinel error that suppresses Cobra's own error print:
```go
// Use cobra's SilenceErrors + SilenceUsage on rootCmd, write custom message,
// then return a pre-formatted error.
```

---

### WR-03: `LocalFileRegistry.Search` ignores the query parameter — always returns all packages

**File:** `internal/registry/local.go:56-62`
**Issue:** `Search(_ string)` discards its query argument and always returns all packages. When a `LocalFileRegistry` is prepended as the highest-priority registry (via `AGENTKIT_REGISTRY_FILE`), all search calls against it return the full manifest regardless of the query. This means the local registry effectively defeats search filtering for any developer using this environment variable.

**Fix:** Apply the same filtering logic used in `GitHubManifestRegistry.Search`:
```go
func (r *LocalFileRegistry) Search(query string) ([]domain.Package, error) {
    m, err := r.load()
    if err != nil {
        return nil, err
    }
    if query == "" {
        return m.Packages, nil
    }
    needle := strings.ToLower(query)
    var results []domain.Package
    for _, p := range m.Packages {
        if strings.Contains(strings.ToLower(p.Name), needle) ||
            strings.Contains(strings.ToLower(p.Description), needle) {
            results = append(results, p)
        }
    }
    return results, nil
}
```

---

### WR-04: `WriteMCPConfig` calls `mcpConfigPath()` twice — TOCTOU window between path resolution and write

**File:** `internal/adapter/claude.go:49-63`, `internal/adapter/claude.go:95-163`
**Issue:** `WriteMCPConfig` calls `mcpConfigPath()` once at line 70 (via `readRawConfig`) and again at line 139. Between these two calls, the presence of `~/.claude.json` vs `~/.claude/settings.json` could theoretically change (e.g., a concurrent Claude Code install creates the primary path). The read and the write may target different files. The result is a silent config split where the read comes from one file and the write goes to another.

**Fix:** Resolve the path once at the start of `WriteMCPConfig` and pass it down to both `readRawConfig` (via a new `readRawConfigFromPath(path string)` variant) and the write step:
```go
func (a *ClaudeCodeAdapter) WriteMCPConfig(entry domain.MCPServerEntry, ownership *domain.InstalledRecord) error {
    path, err := a.mcpConfigPath()
    if err != nil {
        return err
    }
    raw, err := a.readRawConfigFromPath(path)
    // ... use path for the write below
}
```

---

### WR-05: `UpdateAll` returns only the first error, silently drops subsequent errors

**File:** `internal/service/update.go:82-101`
**Issue:** When multiple packages fail to update, `UpdateAll` captures only `firstErr` and discards all subsequent errors. The caller (cmd/update.go) receives a single error message and has no visibility into which packages failed. On a machine with 10 installed packages where 3 fail, 2 failure messages are silently lost.

**Fix:** Accumulate all errors using `errors.Join` (Go 1.20+) or a `[]error` slice:
```go
var errs []error
for _, rec := range records {
    msg, updateErr := s.Update(rec.Name, target)
    if updateErr != nil {
        errs = append(errs, fmt.Errorf("%s: %w", rec.Name, updateErr))
        continue
    }
    msgs = append(msgs, msg)
}
return msgs, errors.Join(errs...)
```

---

### WR-06: `registryNameFromURL` extracts wrong component for the actual URL format used

**File:** `internal/ui/table.go:44-57`
**Issue:** The comment says "parts[i-2] is the repo name" for the path pattern `owner/repo/ref/registry.json`. The parts array for `https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json` splits as: `["https:", "", "raw.githubusercontent.com", "ejyle", "agentkit-registry", "main", "registry.json"]`. When `p == "registry.json"`, `i=6`, so `parts[i-2] = parts[4] = "agentkit-registry"` — this is actually correct. However, for `InstalledRecord.SourceURL` set to `"github.com/ejyle/agentkit-registry"` (as seen in `service/install.go:172` and `testdata/registry.json`), the format is NOT a raw GitHub URL and does not contain `registry.json` at all, so the function falls back to returning the full string `"github.com/ejyle/agentkit-registry"` in the table. The `source` field and the expected display format are inconsistent.

**Fix:** Normalise `SourceURL` at record time to use the raw manifest URL (which is known at resolve time), or update `registryNameFromURL` to also handle the `github.com/owner/repo` shorthand format.

---

### WR-07: Binary installer uses `http.DefaultClient` — inherits process-wide transport with no timeout

**File:** `internal/installer/binary.go:26-27`
**Issue:** `NewBinaryInstaller` sets `client: http.DefaultClient`. `http.DefaultClient` has no timeout. A binary download from a slow or stalled server will hang the process indefinitely. The registry client correctly sets timeouts via `go-retryablehttp`, but the binary installer does not.

**Fix:**
```go
func NewBinaryInstaller() *BinaryInstaller {
    return &BinaryInstaller{
        client: &http.Client{Timeout: 5 * time.Minute}, // generous for large binaries
    }
}
```

---

### WR-08: `skill.ValidateSkill` called with empty `dir` in production installs — always fails SKILL.md check

**File:** `internal/service/install.go:133`
**Issue:** The validator is invoked as `s.validator("", pkg)`. An empty `dir` means `filepath.Join("", "SKILL.md")` = `"SKILL.md"` — a relative path in the current working directory. On a production run, this file does not exist, so `ValidateSkill` always returns `Valid=false` with error "SKILL.md missing". The service code checks `if !result.Valid` but does not return an error in that branch (lines 136–140) — it logs the errors and then continues anyway. This means: skill validation always reports errors in production, those errors go to stdout (see CR-04), and install continues regardless.

**Fix:** Either pass the actual install directory path (computed via `config.SkillInstallPath`), or skip filesystem validation until after `WriteSkill` places the files. Validate post-write using the real path.

---

## Info

### IN-01: `handleForeignConflict` reads stdin for confirmation but the interactive check was already passed

**File:** `cmd/install.go:123-148`
**Issue:** `handleForeignConflict` is called from `handleInstallError`, which is called in both the terminal and non-terminal code paths. The stdin prompt at line 129 will block in non-interactive (piped) contexts, reading from a closed pipe and immediately getting an EOF error. The `ui.IsTerminal()` check is only on the spinner path, not on the error handler dispatch.

**Fix:** Guard the stdin prompt with `ui.IsTerminal()` before prompting, or pass the interactive flag as a parameter.

---

### IN-02: `IsErrNodeNotFound`, `IsErrChecksumMismatch`, `IsErrInsecureURL` compare by string instead of using `errors.Is`

**File:** `internal/installer/npx.go:79-81`, `internal/installer/binary.go:118-125`
**Issue:** All three helper functions use `err.Error() == sentinel.Error()` for comparison. This breaks if the error is wrapped (e.g., `fmt.Errorf("install failed: %w", ErrNodeNotFound)`). The standard Go idiom is `errors.Is(err, Sentinel)`.

**Fix:**
```go
func IsErrNodeNotFound(err error) bool {
    return errors.Is(err, ErrNodeNotFound)
}
```

---

### IN-03: `testdata/registry.json` has empty `sha256` field for both packages

**File:** `testdata/registry.json:9`, `testdata/registry.json:39`
**Issue:** Both test packages have `"sha256": ""`. Combined with CR-01 (binary installer skips checksum when args are empty), acceptance tests using this test data never exercise the checksum verification path. The test data should include a valid checksum for at least the binary install test path, and a deliberately wrong checksum to test rejection.

**Fix:** Add a binary test fixture with a real SHA256, or document that testdata is npx-only and binary tests use a separate fixture.

---

### IN-04: `cmd/update.go:87-91` reports generic "already up to date" without the package name for `UpdateAll`

**File:** `cmd/update.go:87-91`
**Issue:** When updating all packages, the message printed for an already-up-to-date package is `"✓ already up to date\n"` with no package name. With multiple packages installed, the output is a series of identical lines that cannot be correlated to specific packages.

**Fix:**
```go
// UpdateService.UpdateAll should return "already up to date: <name>" or
// embed the name in the message, then cmd/update.go can print it directly.
fmt.Printf("✓ %s\n", m) // if Update() returns "already up to date" or "updated name: old → new"
```

---

### IN-05: `cmd/list.go:24` defines `--target` flag locally, shadowing the persistent root flag

**File:** `cmd/list.go:24`
**Issue:** `listCmd` defines its own local `--target` flag (`listCmd.Flags().String("target", "claude", ...)`), while all other commands use the root-level persistent `--target` flag (`rootCmd.PersistentFlags().StringP`). The `PersistentPreRunE` validator in `root.go` reads `cmd.Flags().GetString("target")`. For `listCmd`, this reads the local flag correctly, but the two flag declarations create inconsistency: `list` accepts `-t` only if it inherits from root, but the local override does not define the `-t` shorthand. Documentation and help text will show two different `--target` descriptions for `list` vs other commands.

**Fix:** Remove the local flag definition from `cmd/list.go:24` and rely solely on the inherited persistent flag from `rootCmd`.

---

_Reviewed: 2026-06-08T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
