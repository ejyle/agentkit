---
phase: 02-multi-assistant-full-install
reviewed: 2026-06-09T00:00:00Z
depth: standard
files_reviewed: 15
files_reviewed_list:
  - cmd/root.go
  - internal/adapter/adapter.go
  - internal/adapter/jsonbase.go
  - internal/adapter/factory.go
  - internal/adapter/copilot_cli.go
  - internal/adapter/copilot_vscode.go
  - internal/adapter/gemini.go
  - internal/adapter/pi.go
  - internal/adapter/codex.go
  - internal/adapter/opencode.go
  - internal/domain/package.go
  - internal/config/paths.go
  - internal/installer/installer.go
  - internal/installer/uvx.go
  - internal/installer/docker.go
findings:
  critical: 5
  warning: 5
  info: 3
  total: 13
status: issues_found
---

# Code Review: Phase 02 — Multi-Assistant Full Install

**Reviewed:** 2026-06-09
**Depth:** standard
**Files Reviewed:** 15
**Status:** issues_found

---

## Summary

Phase 2 extends agentkit with uvx/docker installers, a shared `jsonMCPAdapter` base, and five new assistant adapters (Copilot CLI, Copilot VSCode, Gemini, Pi, Codex, OpenCode). The structural approach — injected `configPath` closures, atomic renameio writes, and a conflict-check ownership gate — is sound. However, there are five critical issues that must be addressed before this ships: two path traversal vulnerabilities in `WriteSkill`, a TOCTOU race in `WriteMCPConfig`'s conflict check, a silent error discard in `NewConfigStore`, and a deprecated `os.IsNotExist` check that can silently swallow wrapped errors. The warnings cover a dropped `ownership` parameter, a `COPILOT_HOME` directory traversal vector, inconsistent `args` handling in the OpenCode read path, the `codex.go` `os.IsNotExist` usage, and the `CopilotVSCodeAdapter` constructor accepting a home dir it does not use.

---

## Critical Issues

### CR-01: Path Traversal in WriteSkill — Gemini and Pi Adapters

**File:** `internal/adapter/gemini.go:54`, `internal/adapter/pi.go:59`

**Issue:** The `filename` key from the caller-supplied `files map[string][]byte` is joined directly onto the skill install path with no validation. A filename such as `"../../.claude/settings.json"` will traverse out of the intended skill directory and overwrite arbitrary files in the user's home tree. Because `renameio.WriteFile` creates parent directories implicitly on some platforms, the attack surface extends beyond existing directories.

```go
// current — no validation
dest := filepath.Join(skillPath, filename)
if err := renameio.WriteFile(dest, content, 0644); err != nil { ... }
```

**Fix:** Validate that the resolved destination path is still rooted under `skillPath` before writing:

```go
dest := filepath.Join(skillPath, filename)
// Reject any filename that escapes the skill directory.
rel, err := filepath.Rel(skillPath, dest)
if err != nil || strings.HasPrefix(rel, "..") {
    return fmt.Errorf("skill file %q escapes skill directory — rejected", filename)
}
if err := renameio.WriteFile(dest, content, 0644); err != nil { ... }
```

This same fix applies identically to `pi.go:59` and to `claude.go`'s `WriteSkill` if it shares the same pattern (confirm before shipping).

---

### CR-02: TOCTOU Race in WriteMCPConfig — All JSON Adapters

**File:** `internal/adapter/jsonbase.go:90–101`, `internal/adapter/codex.go:120–131`, `internal/adapter/opencode.go:109–119`

**Issue:** The conflict check (read config → look up key → check ownership → write) is not atomic. There is a window between the `os.ReadFile` on line 49 and the `renameio.WriteFile` on line 135 where a concurrent process (another `agentkit` invocation, or the assistant itself) can write the same key. If it does, the ownership check will not detect the race and agentkit will silently overwrite a key it does not own — bypassing the foreign-conflict guard (D-07).

The `ConfigStore` correctly uses `sync.Mutex` for its own state, but the MCP config file (the assistant's settings file) is not held under any lock during the read-modify-write cycle.

**Fix:** The minimal fix is a file-level advisory lock using `golang.org/x/sys/unix.Flock` (Linux/macOS) or `github.com/gofrs/flock` (cross-platform) around the read-merge-write triple. Alternatively, document this as a known limitation with a user-visible warning when conflict detection matters more than throughput.

---

### CR-03: Silent Error Discard in NewConfigStore

**File:** `internal/config/store.go:23–25`

**Issue:** `NewConfigStore` calls `InstalledStatePath(target)` and discards the error with `_`. If `os.UserConfigDir()` fails (headless CI, sandboxed environment, Docker containers without a home), `basePath` is set to the empty string `""`. All subsequent reads and writes then operate against the current working directory, silently corrupting or reading state from the wrong location with no diagnostic.

```go
func NewConfigStore(target string) *ConfigStore {
    path, _ := InstalledStatePath(target)   // error silently discarded
    return &ConfigStore{target: target, basePath: path}
}
```

**Fix:** Return the error to the caller, or at minimum panic with a clear message if the path cannot be determined:

```go
func NewConfigStore(target string) (*ConfigStore, error) {
    path, err := InstalledStatePath(target)
    if err != nil {
        return nil, fmt.Errorf("cannot determine agentkit state path: %w", err)
    }
    return &ConfigStore{target: target, basePath: path}, nil
}
```

All callers of `NewConfigStore` must be updated to handle the error. This is a ripple-effect refactor but is necessary before the silent-corruption behaviour can ship.

---

### CR-04: Deprecated os.IsNotExist Misses Wrapped Errors — jsonbase.go and opencode.go

**File:** `internal/adapter/jsonbase.go:50`, `internal/adapter/opencode.go:65`

**Issue:** `os.IsNotExist(err)` was deprecated in Go 1.13 in favour of `errors.Is(err, os.ErrNotExist)` because the old form does not unwrap error chains. If `os.ReadFile` returns a wrapped `*fs.PathError` (which it does on all modern Go versions when the file is absent), the unwrapping still happens to work — but only because `os.IsNotExist` has a special-case path that inspects `*PathError` directly. The real risk is when the file is on a network filesystem or inside a container where the error is wrapped further (e.g., through `fmt.Errorf("...: %w", pathErr)`). In those cases `os.IsNotExist` returns `false` and the function returns an error instead of treating the absent file as a first-install case, breaking all first-run setups on those environments.

```go
// current — broken for wrapped errors
if os.IsNotExist(err) {
    return map[string]interface{}{}, nil
}
```

**Fix:**

```go
if errors.Is(err, os.ErrNotExist) {
    return map[string]interface{}{}, nil
}
```

Apply to both `jsonbase.go:50` and `opencode.go:65`. Note: `codex.go:65` has the same bug (see WR-01).

---

### CR-05: COPILOT_HOME Env Var Is an Unvalidated Path — Potential Directory Traversal

**File:** `internal/adapter/copilot_cli.go:39–41`

**Issue:** The `COPILOT_HOME` environment variable is read and used as a directory path without any validation. An attacker who controls the environment (e.g., malicious `.env` loaded by a parent process, or a CI workflow with injected env vars) can point this to any path — including sensitive directories like `~/.ssh/`, `/etc/`, or `~/.aws/`. The resulting `mcp-config.json` write would land in that directory. Because `os.MkdirAll` is also called via the write path, this can create new directories in attacker-controlled locations.

```go
if copilotHome := os.Getenv("COPILOT_HOME"); copilotHome != "" {
    return filepath.Join(copilotHome, "mcp-config.json"), nil
}
```

**Fix:** Validate that `COPILOT_HOME` is an absolute path and, if possible, that it is rooted under the user's home directory. At minimum, reject relative paths:

```go
if copilotHome := os.Getenv("COPILOT_HOME"); copilotHome != "" {
    if !filepath.IsAbs(copilotHome) {
        return "", fmt.Errorf("COPILOT_HOME must be an absolute path, got %q", copilotHome)
    }
    // Optionally: ensure it is under home dir.
    return filepath.Join(copilotHome, "mcp-config.json"), nil
}
```

---

## Warnings

### WR-01: os.IsNotExist Used in codex.go — Same Deprecated Pattern as CR-04

**File:** `internal/adapter/codex.go:65`

**Issue:** `codex.go` uses the same deprecated `os.IsNotExist(err)` pattern described in CR-04. The BurntSushi TOML library wraps path errors differently from `os.ReadFile`, meaning this is slightly more likely to break on unusual filesystems than the JSON adapters. Should be fixed alongside CR-04 for consistency.

**Fix:** Replace `os.IsNotExist(err)` with `errors.Is(err, os.ErrNotExist)` at line 65.

---

### WR-02: WriteMCPConfig Silently Ignores the ownership Parameter

**File:** `internal/adapter/jsonbase.go:72`, `internal/adapter/codex.go:102`, `internal/adapter/opencode.go:91`

**Issue:** The `AssistantAdapter` interface declares `WriteMCPConfig(entry domain.MCPServerEntry, ownership *domain.InstalledRecord) error` with documented semantics: `ownership` is non-nil on upgrade (D-08), meaning the caller has already confirmed agentkit owns the key and the write should skip the foreign-conflict check. All three implementations accept `ownership` but immediately ignore it (parameter named `_`), then re-derive ownership by calling `a.store.GetRecord(entry.Name)`. This is redundant I/O on the hot path and violates the stated contract — if the caller passes a non-nil `ownership`, the store lookup should be skipped.

More critically, the parameter is semantically misleading: callers may pass a non-nil record expecting it to suppress the foreign-conflict error, but it has no effect — the conflict check runs unconditionally against the store, and a foreign key will still return `ErrForeignConflict` even when the caller signals intent to upgrade.

**Fix:** Use the passed `ownership` parameter to decide whether to skip the store lookup:

```go
func (a *jsonMCPAdapter) WriteMCPConfig(entry domain.MCPServerEntry, ownership *domain.InstalledRecord) error {
    ...
    if _, exists := mcpServers[entry.Name]; exists {
        if ownership == nil {
            // Caller has not confirmed ownership — check store.
            _, owned, err := a.store.GetRecord(entry.Name)
            if err != nil { return fmt.Errorf(...) }
            if !owned {
                ...return &ErrForeignConflict{...}
            }
        }
        // ownership non-nil or store confirms agentkit-owned: auto-overwrite.
    }
    ...
}
```

---

### WR-03: OpenCode ReadMCPConfig Discards the "args" Field if Present

**File:** `internal/adapter/opencode.go:244–254` (in `extractOpenCodeEntry`)

**Issue:** OpenCode stores the full command as a combined `["command", "arg1", ...]` array (T-02-11), which `extractOpenCodeEntry` correctly handles. However, the shared `extractEntryFromRaw` helper in `jsonbase.go` (line 233–239) also tries to read a separate `"args"` field and appends it to `entry.Args` *after* parsing the array. In the OpenCode read path (via `jsonMCPAdapter.ReadMCPConfig`), if an OpenCode-written entry ever reaches that helper, its args would be doubled. `OpenCodeAdapter` does not embed `jsonMCPAdapter`, so `extractEntryFromRaw` is never called directly for OpenCode — but `jsonbase.go`'s `extractEntryFromRaw` handling of `[]interface{}` command arrays (lines 219–230) was designed for OpenCode and is now dead code in the JSON base, since OpenCode has its own adapter. This creates maintenance confusion about which path handles the array-command format.

**Fix:** Remove the `[]interface{}` branch from `extractEntryFromRaw` in `jsonbase.go` (lines 219–230) since no JSON-base adapter produces that format. If the format is needed in future, restore it explicitly with a test. Add a comment to `extractEntryFromRaw` clarifying that it is not used for OpenCode.

---

### WR-04: DockerInstaller and UvxInstaller Accept Empty spec.Package Without Validation

**File:** `internal/installer/docker.go:70`, `internal/installer/uvx.go:68`

**Issue:** Both installers pass `spec.Package` directly to `exec.Command` as an argument without checking whether it is non-empty. If a registry manifest entry is malformed and `spec.Package` is `""`, the resulting command is `docker pull ""` or `uvx ""`. Docker will return an error, but the error message ("invalid reference format: repository name must be lowercase") gives no indication of the root cause (empty package field). UVX's behaviour for an empty string is platform-dependent and can silently do unexpected things.

**Fix:** Add a guard at the top of `Install`:

```go
if spec.Package == "" {
    return fmt.Errorf("install spec has empty package name")
}
```

---

### WR-05: NewCopilotVSCodeAdapterWithHome Silently Ignores Its homeDir Argument

**File:** `internal/adapter/copilot_vscode.go:32–34`

**Issue:** `NewCopilotVSCodeAdapterWithHome` accepts a `homeDir string` parameter (matching the pattern of all other adapters) but immediately discards it and calls `newCopilotVSCodeAdapter(store, "")`. The docstring acknowledges this but frames it as acceptable ("homeDir parameter is unused for VS Code"). The problem is that callers using this constructor by analogy with other adapters will get no compiler error, no runtime error, and silently incorrect test isolation — their injected `homeDir` is ignored and the adapter will fall through to `os.UserConfigDir()`, potentially touching real user config in tests.

The correct constructor for tests is `NewCopilotVSCodeAdapterWithConfigDir`, but nothing prevents a caller from using the wrong one.

**Fix:** Either remove `NewCopilotVSCodeAdapterWithHome` entirely (breaking change for any test that uses it) or have it return an error or panic with a clear message:

```go
// NewCopilotVSCodeAdapterWithHome is intentionally a no-op; use NewCopilotVSCodeAdapterWithConfigDir.
// Deprecated: use NewCopilotVSCodeAdapterWithConfigDir for test injection.
func NewCopilotVSCodeAdapterWithHome(store *config.ConfigStore, _ string) *CopilotVSCodeAdapter {
    panic("NewCopilotVSCodeAdapterWithHome: homeDir has no effect for VS Code; use NewCopilotVSCodeAdapterWithConfigDir")
}
```

Or simply delete the function and update any tests.

---

## Info

### IN-01: InstalledRecord Missing Field in domain.package.go (InstallMethod Not Recorded)

**File:** `internal/domain/installed.go:7–15`

**Issue:** `InstalledRecord` records `InstallPath`, `SourceURL`, and `Checksum` but not the `InstallMethod` used. When `agentkit update` needs to re-run the installer for an existing package, it will have to infer the method from the `SourceURL` or re-parse the registry manifest — an unnecessary round-trip. Storing `Method domain.InstallMethod` in the record would make update logic simpler and more robust.

**Fix:** Add `Method InstallMethod \`json:"method"\`` to `InstalledRecord` and populate it in the install service.

---

### IN-02: SkillInstallPath Default Branch Is Undocumented and Potentially Incorrect

**File:** `internal/config/paths.go:58`

**Issue:** The `default` branch of `SkillInstallPath` constructs `~/.{target}/skills/{name}`, applying a heuristic for unknown assistants. This path shape may be wrong for any future assistant — and because the switch already lists every valid target from `cmd/root.go`, the default can only be reached by programmatic callers who pass an unrecognised string. The same function returns an error for known unsupported targets (copilot-cli, codex, opencode) but silently succeeds with a guessed path for unknown ones.

**Fix:** Change the default to return an error, consistent with the other unsupported cases:

```go
default:
    return "", fmt.Errorf("SkillInstallPath: unknown target %q", target)
```

---

### IN-03: Magic String "mcpServers" Repeated Across All Adapters

**File:** `internal/adapter/copilot_cli.go:36`, `internal/adapter/gemini.go:33`, `internal/adapter/pi.go:37`

**Issue:** The string literal `"mcpServers"` is hardcoded independently in every adapter that uses it. If the key name ever changes (or an assistant uses a different casing), each file must be updated individually. A single package-level constant would make this a single-point-of-truth change.

**Fix:** Define a constant in `adapter.go` or a new `keys.go`:

```go
const mcpServersKey = "mcpServers"
```

Reference it in all adapters rather than repeating the literal.

---

_Reviewed: 2026-06-09_
_Reviewer: Claude (adversarial code review)_
_Depth: standard_
