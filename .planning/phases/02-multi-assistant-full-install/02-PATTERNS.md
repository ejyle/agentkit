# Phase 2: Multi-Assistant & Full Install - Pattern Map

**Mapped:** 2026-06-09
**Files analyzed:** 14 new/modified files
**Analogs found:** 14 / 14

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/adapter/jsonbase.go` | utility | CRUD | `internal/adapter/claude.go` | exact |
| `internal/adapter/copilot_cli.go` | adapter | CRUD | `internal/adapter/claude.go` | exact |
| `internal/adapter/copilot_vscode.go` | adapter | CRUD | `internal/adapter/claude.go` | exact |
| `internal/adapter/gemini.go` | adapter | CRUD | `internal/adapter/claude.go` | exact |
| `internal/adapter/opencode.go` | adapter | CRUD | `internal/adapter/claude.go` | role-match |
| `internal/adapter/codex.go` | adapter | CRUD | `internal/adapter/claude.go` | role-match |
| `internal/adapter/pi.go` | adapter | CRUD | `internal/adapter/claude.go` | role-match |
| `internal/installer/uvx.go` | installer | request-response | `internal/installer/npx.go` | exact |
| `internal/installer/docker.go` | installer | request-response | `internal/installer/npx.go` | exact |
| `internal/installer/installer.go` | config | — | `internal/installer/installer.go` | modify existing |
| `internal/config/paths.go` | config | — | `internal/config/paths.go` | modify existing |
| `cmd/root.go` | config | — | `cmd/root.go` | modify existing |
| `internal/adapter/copilot_cli_test.go` | test | — | `internal/adapter/claude_test.go` | exact |
| `internal/installer/uvx_test.go` | test | — | `internal/installer/npx_test.go` | exact |

---

## Pattern Assignments

### `internal/adapter/jsonbase.go` (utility, shared JSON base)

**Analog:** `internal/adapter/claude.go`

**Purpose:** Extract the shared read/merge/write/verify loop used by Claude, Gemini, CopilotCLI, CopilotVSCode, and Pi (all use `mcpServers` JSON key with command/args/env structure). Each embedding adapter injects its own path-detection function and optional extra-fields function.

**Imports pattern** (claude.go lines 1-12):
```go
import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "github.com/ejyle/agentkit/internal/config"
    "github.com/ejyle/agentkit/internal/domain"
    "github.com/google/renameio/v2"
)
```

**Struct pattern** (new — derive from claude.go structure):
```go
// jsonMCPAdapter is the shared base for all mcpServers-keyed JSON config adapters.
type jsonMCPAdapter struct {
    store      *config.ConfigStore
    homeDir    string
    configPath func(home string) (string, error) // injected by each adapter
    mcpKey     string                            // always "mcpServers" for this group
    extraFields func(entry domain.MCPServerEntry) map[string]interface{} // nil = no extras
}
```

**homeDir injection pattern** (claude.go lines 38-43):
```go
func (a *ClaudeCodeAdapter) home() (string, error) {
    if a.homeDir != "" {
        return a.homeDir, nil
    }
    return os.UserHomeDir()
}
```

**readRawConfig pattern** (claude.go lines 67-84):
```go
func (a *ClaudeCodeAdapter) readRawConfig() (map[string]interface{}, error) {
    path, err := a.mcpConfigPath()
    // ...
    data, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        return map[string]interface{}{}, nil
    }
    var raw map[string]interface{}
    if err := json.Unmarshal(data, &raw); err != nil {
        return nil, fmt.Errorf("parsing MCP config at %s: %w", path, err)
    }
    return raw, nil
}
```

**Merge + atomic write + post-verify pattern** (claude.go lines 95-163):
```go
// 1. readRawConfig()
// 2. Extract/init mcpServers map
// 3. Conflict check via a.store.GetRecord(entry.Name)
//    - not owned → return &ErrForeignConflict{OldEntry, NewEntry}
//    - owned     → auto-overwrite
// 4. Build entryMap: {"command": entry.Command, "args": entry.Args}
//    if len(entry.Env) > 0 { entryMap["env"] = entry.Env }
//    // inject extra fields here (e.g., "type":"local" for Copilot)
// 5. mcpServers[entry.Name] = entryMap; raw["mcpServers"] = mcpServers
// 6. os.MkdirAll(filepath.Dir(path), 0755)
// 7. json.MarshalIndent(raw, "", "  ")
// 8. renameio.WriteFile(path, data, 0644)
// 9. Post-verify: a.ReadMCPConfig() → check entry.Name present
```

**extractEntry helper** (claude.go lines 243-268):
```go
func extractEntry(val interface{}) domain.MCPServerEntry {
    m, ok := val.(map[string]interface{})
    // ... extracts Command (string), Args ([]interface{}→[]string), Env (map)
}
```

---

### `internal/adapter/copilot_cli.go` (adapter, CRUD)

**Analog:** `internal/adapter/claude.go` (embed `jsonMCPAdapter`)

**Name():** `"copilot-cli"`

**Config path detection** (runtime, no hardcoding — same pattern as claude.go lines 49-63):
```go
// Primary: ~/.copilot/mcp-config.json
// No legacy fallback needed (Copilot CLI has single canonical path)
func (a *CopilotCLIAdapter) mcpConfigPath() (string, error) {
    home, err := a.home()
    if err != nil { return "", err }
    if env := os.Getenv("COPILOT_HOME"); env != "" {
        return filepath.Join(env, "mcp-config.json"), nil
    }
    return filepath.Join(home, ".copilot", "mcp-config.json"), nil
}
```

**Extra fields injection** (Copilot CLI requires `"type"` and `"tools"` fields):
```go
// In entryMap construction, after command/args/env:
entryMap["type"] = "local"
entryMap["tools"] = []string{"*"}
```

**WriteSkill — ErrNotSupported** (D-18, A1):
```go
var ErrNotSupported = errors.New("operation not supported")

func (a *CopilotCLIAdapter) WriteSkill(name string, files map[string][]byte) error {
    return fmt.Errorf("copilot-cli adapter: WriteSkill not supported — Copilot CLI has no CLI-level skill directory: %w", ErrNotSupported)
}
func (a *CopilotCLIAdapter) RemoveSkill(name string) error {
    return fmt.Errorf("copilot-cli adapter: RemoveSkill not supported — Copilot CLI has no CLI-level skill directory: %w", ErrNotSupported)
}
```

---

### `internal/adapter/copilot_vscode.go` (adapter, CRUD)

**Analog:** `internal/adapter/claude.go` (embed `jsonMCPAdapter`)

**Name():** `"copilot-vscode"`

**Config path detection** (platform-aware, runtime detection of VS Code edition):
```go
// Uses os.UserConfigDir() as base on Linux/macOS; os.Getenv("APPDATA") on Windows.
// Checks in order: "Code" → "Code - Insiders" → "code-server"
// Path: <configDir>/<edition>/User/mcp.json
// Top-level key: "servers" (NOT "mcpServers") — overrides the base mcpKey
```

**Differs from CopilotCLI:** Top-level key is `"servers"` not `"mcpServers"`. Same ErrNotSupported for WriteSkill/RemoveSkill.

---

### `internal/adapter/gemini.go` (adapter, CRUD)

**Analog:** `internal/adapter/claude.go` (embed `jsonMCPAdapter`)

**Name():** `"gemini"`

**Config path** (single path, no legacy fallback needed):
```go
// ~/.gemini/settings.json
func (a *GeminiAdapter) mcpConfigPath() (string, error) {
    home, err := a.home()
    if err != nil { return "", err }
    return filepath.Join(home, ".gemini", "settings.json"), nil
}
```

**No extra fields** — Gemini mcpServers entry format is identical to Claude (command/args/env, no `type` field).

**WriteSkill** — fully implemented (SKILL.md at `~/.gemini/skills/<name>/`):
```go
func (a *GeminiAdapter) WriteSkill(name string, files map[string][]byte) error {
    skillPath, err := config.SkillInstallPath("gemini", name)
    // ... same pattern as claude.go lines 216-231
}
```

---

### `internal/adapter/opencode.go` (adapter, CRUD)

**Analog:** `internal/adapter/claude.go` (role-match; different JSON schema — cannot embed `jsonMCPAdapter`)

**Name():** `"opencode"`

**Config path** (uses `os.UserConfigDir()`, not `os.UserHomeDir()`):
```go
func (a *OpenCodeAdapter) mcpConfigPath() (string, error) {
    base, err := os.UserConfigDir()
    if err != nil { return "", err }
    return filepath.Join(base, "opencode", "opencode.json"), nil
}
```

**Key differences from Claude pattern:**
- Top-level key: `"mcp"` (not `"mcpServers"`)
- `command` field is a JSON array (command + args combined)
- Env key is `"environment"` (not `"env"`)
- Required `"type": "local"` field

**entryMap construction** (write):
```go
entryMap := map[string]interface{}{
    "type":    "local",
    "command": append([]string{entry.Command}, entry.Args...),
    "enabled": true,
}
if len(entry.Env) > 0 {
    entryMap["environment"] = entry.Env
}
mcp[entry.Name] = entryMap
raw["mcp"] = mcp
```

**extractEntry for OpenCode** (read — split array back):
```go
// command array: arr[0] → entry.Command, arr[1:] → entry.Args
// env key: "environment" not "env"
```

**WriteSkill** — ErrNotSupported (A3):
```go
func (a *OpenCodeAdapter) WriteSkill(name string, files map[string][]byte) error {
    return fmt.Errorf("opencode adapter: WriteSkill not supported — OpenCode has no user-global skill directory: %w", ErrNotSupported)
}
```

---

### `internal/adapter/codex.go` (adapter, CRUD)

**Analog:** `internal/adapter/claude.go` (role-match; TOML format — cannot embed `jsonMCPAdapter`)

**Name():** `"codex"`

**Config path:**
```go
// ~/.codex/config.toml
func (a *CodexAdapter) mcpConfigPath() (string, error) {
    home, err := a.home()
    if err != nil { return "", err }
    return filepath.Join(home, ".codex", "config.toml"), nil
}
```

**TOML read strategy** (decode into map to preserve all non-MCP keys):
```go
import "github.com/BurntSushi/toml"

func (a *CodexAdapter) readRawConfig() (map[string]interface{}, error) {
    path, err := a.mcpConfigPath()
    // ...
    var raw map[string]interface{}
    if _, err := toml.DecodeFile(path, &raw); err != nil {
        // handle os.IsNotExist → return empty map
    }
    return raw, nil
}
```

**TOML write strategy** (encode via bytes.Buffer, then renameio — same atomic pattern):
```go
import (
    "bytes"
    "github.com/BurntSushi/toml"
    "github.com/google/renameio/v2"
)

var buf bytes.Buffer
if err := toml.NewEncoder(&buf).Encode(raw); err != nil { ... }
if err := renameio.WriteFile(path, buf.Bytes(), 0644); err != nil { ... }
```

**TOML entry schema:**
```toml
[mcp_servers.my-server]
command = "uvx"
args = ["mcp-server-fetch"]

[mcp_servers.my-server.env]
MY_KEY = "value"
```

**WriteSkill** — ErrNotSupported (A2):
```go
func (a *CodexAdapter) WriteSkill(name string, files map[string][]byte) error {
    return fmt.Errorf("codex adapter: WriteSkill not supported — Codex CLI has no user-global skill directory: %w", ErrNotSupported)
}
```

---

### `internal/adapter/pi.go` (adapter, CRUD)

**Analog:** `internal/adapter/claude.go` (embed `jsonMCPAdapter` for MCP config; claude.go WriteSkill for skill write)

**Name():** `"pi"`

**MCP config path:**
```go
// ~/.pi/agent/mcp.json
// mcpServers key (same as Claude/Gemini format — no extra fields needed)
func (a *PiAdapter) mcpConfigPath() (string, error) {
    home, err := a.home()
    if err != nil { return "", err }
    return filepath.Join(home, ".pi", "agent", "mcp.json"), nil
}
```

**WriteSkill** — fully implemented (skills at `~/.agents/skills/<name>/`):
```go
func (a *PiAdapter) WriteSkill(name string, files map[string][]byte) error {
    skillPath, err := config.SkillInstallPath("pi", name)
    // SkillInstallPath("pi", name) must resolve to ~/.agents/skills/<name>
    // (overrides the default ~/.pi/skills/<name> fallback — see paths.go section)
    // ... same atomic write loop as claude.go lines 216-231
}
```

Both WriteMCPConfig and WriteSkill are fully implemented — D-05 satisfied. No ErrNotSupported needed.

---

### `internal/installer/uvx.go` (installer, request-response)

**Analog:** `internal/installer/npx.go` (exact mirror)

**Imports pattern** (npx.go lines 1-9):
```go
import (
    "bytes"
    "fmt"
    "os/exec"

    "github.com/ejyle/agentkit/internal/domain"
)
```

**Struct + constructor pattern** (npx.go lines 18-56):
```go
type UvxInstaller struct {
    lookPath lookPathFunc
    run      runFunc
}

func NewUvxInstaller() *UvxInstaller {
    u := &UvxInstaller{}
    u.lookPath = exec.LookPath
    u.run = func(name string, args []string) error {
        cmd := exec.Command(name, args...)
        var out bytes.Buffer
        cmd.Stdout = &out
        cmd.Stderr = &out
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("uvx exited non-zero: %w\noutput: %s", err, out.String())
        }
        return nil
    }
    return u
}

func NewUvxInstallerWithLookPath(lp lookPathFunc) *UvxInstaller { ... }
func NewUvxInstallerWithRunner(run runFunc) *UvxInstaller { ... }
```

**Sentinel error** (from installer.go pattern):
```go
var ErrUvxNotFound = errors.New("uvx not found on PATH; install uv to use Python-based MCP servers: https://docs.astral.sh/uv/")
```

**IsAvailable + Method + Install** (npx.go lines 59-76):
```go
func (u *UvxInstaller) IsAvailable() bool {
    _, err := u.lookPath("uvx")
    return err == nil
}

func (u *UvxInstaller) Method() domain.InstallMethod {
    return domain.InstallMethodUvx
}

// Install runs "uvx <spec.Package> [spec.Args...]"
// command="uvx", args=["package-name", ...extra-args]
func (u *UvxInstaller) Install(spec domain.InstallSpec) error {
    if _, err := u.lookPath("uvx"); err != nil {
        return ErrUvxNotFound
    }
    args := append([]string{spec.Package}, spec.Args...)
    return u.run("uvx", args)
}
```

---

### `internal/installer/docker.go` (installer, request-response)

**Analog:** `internal/installer/npx.go` (exact mirror, different subprocess)

**Sentinel error:**
```go
var ErrDockerNotFound = errors.New("docker not found on PATH; install Docker: https://docs.docker.com/get-docker/")
```

**Install** — runs `docker pull` at install time (D-09, eager pull):
```go
func (d *DockerInstaller) Install(spec domain.InstallSpec) error {
    if _, err := d.lookPath("docker"); err != nil {
        return ErrDockerNotFound
    }
    // spec.Package = image:tag (e.g., "ghcr.io/github/github-mcp-server")
    return d.run("docker", []string{"pull", spec.Package})
}
```

**Config entry is built by the adapter** (not the installer) using:
```go
// command="docker", args=["run", "-i", "--rm", "image:tag", ...spec.ExtraArgs]
entryMap := map[string]interface{}{
    "command": "docker",
    "args":    append([]string{"run", "-i", "--rm", spec.Package}, spec.ExtraArgs...),
}
```

**Method:**
```go
func (d *DockerInstaller) Method() domain.InstallMethod {
    return domain.InstallMethodDocker
}
```

---

### `internal/installer/installer.go` (modify existing — add new cases)

**File:** `internal/installer/installer.go` lines 14-18 (sentinel errors block) and lines 31-43 (NewInstaller switch)

**Add to sentinel errors block:**
```go
var (
    ErrNodeNotFound   = errors.New("node not found on PATH; install Node.js to use npx-based MCP servers")
    ErrChecksumMismatch = errors.New("SHA256 checksum mismatch: ...")
    ErrInsecureURL    = errors.New("insecure download URL: only https:// URLs are allowed")
    // ADD:
    ErrUvxNotFound    = errors.New("uvx not found on PATH; install uv to use Python-based MCP servers: https://docs.astral.sh/uv/")
    ErrDockerNotFound = errors.New("docker not found on PATH; install Docker: https://docs.docker.com/get-docker/")
)
```

**Add to NewInstaller switch** (installer.go lines 32-43):
```go
case domain.InstallMethodUvx:
    return NewUvxInstaller(), nil
case domain.InstallMethodDocker:
    return NewDockerInstaller(), nil
```

Also add `InstallMethodUvx` and `InstallMethodDocker` constants to `internal/domain/package.go` (wherever `InstallMethodNpx` is defined).

---

### `internal/config/paths.go` (modify existing — add explicit target cases)

**File:** `internal/config/paths.go` lines 40-51 (`SkillInstallPath` switch)

**Current default** (line 49): `filepath.Join(home, "."+target, "skills", name)` — would incorrectly resolve `pi` as `~/.pi/skills/<name>`.

**Add explicit cases:**
```go
switch target {
case "claude":
    return filepath.Join(home, ".claude", "skills", name), nil
case "gemini":
    return filepath.Join(home, ".gemini", "skills", name), nil
case "pi":
    // Pi discovers from ~/.agents/skills/ (user-global, NOT ~/.pi/skills/)
    return filepath.Join(home, ".agents", "skills", name), nil
case "copilot-cli", "copilot-vscode", "codex", "opencode":
    // These assistants have no user-global skill directory.
    // Callers should check ErrNotSupported before calling SkillInstallPath.
    return "", fmt.Errorf("SkillInstallPath: %q has no user-global skill directory", target)
default:
    return filepath.Join(home, "."+target, "skills", name), nil
}
```

---

### `cmd/root.go` (modify existing — add new targets to validation)

**File:** `cmd/root.go` lines 16-43

**Current validTargets** (line 16):
```go
var validTargets = []string{"claude", "copilot", "codex", "gemini", "opencode"}
```

**Replace with:**
```go
var validTargets = []string{
    "claude",
    "copilot-cli",
    "copilot-vscode",
    "codex",
    "gemini",
    "opencode",
    "pi",
}
```

**Update error message** (line 39-42):
```go
return fmt.Errorf(
    "invalid target %q: must be one of claude, copilot-cli, copilot-vscode, codex, gemini, opencode, pi",
    target,
)
```

**Update flag description** (line 24-27):
```go
rootCmd.PersistentFlags().StringP(
    "target", "t", "claude",
    "Target coding assistant (claude|copilot-cli|copilot-vscode|codex|gemini|opencode|pi)",
)
```

---

### Test files (adapter + installer)

**Analog:** `internal/adapter/claude_test.go` and `internal/installer/npx_test.go`

**Test package pattern** (claude_test.go line 1, npx_test.go line 1):
```go
package adapter_test  // external test package
package installer_test
```

**Adapter test helper pattern** (claude_test.go lines 15-25):
```go
func makeAdapter(t *testing.T, tmpHome string) *adapter.XxxAdapter {
    t.Helper()
    storeDir := filepath.Join(tmpHome, "config", "agentkit", "<target>")
    os.MkdirAll(storeDir, 0755)
    storePath := filepath.Join(storeDir, "installed.json")
    store := config.NewConfigStoreWithPath("<target>", storePath)
    return adapter.NewXxxAdapterWithHome(store, tmpHome)
}
```

**Installer test pattern** — injected lookPath + runner (npx_test.go lines 29-43):
```go
// Test ErrXxxNotFound via injected LookPath:
n := installer.NewUvxInstallerWithLookPath(func(file string) (string, error) {
    return "", &notFoundError{file}
})

// Test arg-array form via injected runner:
n := installer.NewUvxInstallerWithRunner(func(name string, args []string) error {
    // assert name == "uvx" and args == expected
    return nil
})
```

**Key adapter test cases to cover** (derived from claude_test.go structure):
1. WriteMCPConfig creates config file if absent
2. WriteMCPConfig preserves all pre-existing foreign keys
3. WriteMCPConfig returns ErrForeignConflict when key exists and not owned
4. WriteMCPConfig auto-overwrites when agentkit-owned
5. ReadMCPConfig after Write returns written entry (post-install verify)
6. RemoveMCPConfig removes only target key; preserves others
7. For adapters with ErrNotSupported WriteSkill: verify correct error type returned

---

## Shared Patterns

### homeDir Injection (testability)
**Source:** `internal/adapter/claude.go` lines 18-43
**Apply to:** All 5 new adapter files and `jsonbase.go`

Every adapter struct carries `homeDir string`. Constructor variants:
- `NewXxxAdapter(store)` — uses `os.UserHomeDir()` at runtime
- `NewXxxAdapterWithHome(store, homeDir)` — used in tests to point at `t.TempDir()`

### Atomic Write
**Source:** `internal/adapter/claude.go` lines 138-151
**Apply to:** All adapter Write methods
```go
// Never os.WriteFile — always renameio:
if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil { ... }
data, _ := json.MarshalIndent(raw, "", "  ")
renameio.WriteFile(path, data, 0644)
```

### Post-Install Verify
**Source:** `internal/adapter/claude.go` lines 154-162
**Apply to:** All WriteMCPConfig implementations
```go
result, err := a.ReadMCPConfig()
if _, ok := result[entry.Name]; !ok {
    return fmt.Errorf("post-install verify failed: %s.%s not found after write", mcpKey, entry.Name)
}
```

### Conflict Check (ErrForeignConflict)
**Source:** `internal/adapter/claude.go` lines 113-125 and `internal/adapter/adapter.go` lines 31-45
**Apply to:** All WriteMCPConfig implementations except Codex (TOML — same logic, different key access)
```go
if _, exists := mcpServers[entry.Name]; exists {
    _, owned, err := a.store.GetRecord(entry.Name)
    if !owned {
        old := extractEntry(mcpServers[entry.Name])
        old.Name = entry.Name
        return &ErrForeignConflict{OldEntry: old, NewEntry: entry}
    }
    // agentkit-owned: auto-overwrite
}
```

### ErrNotSupported (for partial adapters)
**Source:** pattern established in D-06; `errors.New` in `internal/adapter/adapter.go` style
**Apply to:** CopilotCLI WriteSkill/RemoveSkill, CopilotVSCode WriteSkill/RemoveSkill, Codex WriteSkill/RemoveSkill, OpenCode WriteSkill/RemoveSkill
```go
// Define once in adapter.go or a new errors.go:
var ErrNotSupported = errors.New("operation not supported")

// In each adapter's WriteSkill:
return fmt.Errorf("<adapter-name> adapter: WriteSkill not supported — <reason>: %w", ErrNotSupported)
```

### Installer lookPath + run Injection
**Source:** `internal/installer/npx.go` lines 12-56
**Apply to:** `uvx.go`, `docker.go`
```go
type lookPathFunc func(file string) (string, error)
type runFunc func(name string, args []string) error
// Constructor provides real exec.LookPath + exec.Command; test constructors inject mocks
```

### No Shell Interpolation
**Source:** `internal/installer/npx.go` lines 27-36
**Apply to:** All installer Install() methods
```go
// Always exec.Command(name, args...) — never exec.Command("sh", "-c", combinedString)
cmd := exec.Command(name, args...)
```

---

## No Analog Found

All files have close analogs. No entries in this section.

---

## Metadata

**Analog search scope:** `internal/adapter/`, `internal/installer/`, `internal/config/`, `cmd/`
**Files scanned:** 8 source files read
**Pattern extraction date:** 2026-06-09
