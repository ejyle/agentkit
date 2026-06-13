# Phase 2: Multi-Assistant & Full Install - Research

**Researched:** 2026-06-09
**Domain:** Multi-assistant adapter implementations (Copilot, Codex, Gemini, OpenCode, Pi) + uvx and Docker MCP installers
**Confidence:** MEDIUM-HIGH (core config formats verified; Pi and some edge cases remain LOW due to rapidly evolving docs)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01/D-02:** Curated-only registry model is permanent for v1. No mcpmarket.com, no custom registry add. REG-03 and REG-04 are OUT OF SCOPE entirely.
- **D-03:** Phase 2 has no new registry work.
- **D-04:** Pi adapter is best-effort partial: implement only what pi.dev actually supports, return ErrNotSupported for unimplemented operations.
- **D-05:** Implement BOTH WriteMCPConfig AND WriteSkill for Pi if pi.dev has both mechanisms.
- **D-06:** Unsupported Pi operations return ErrNotSupported with a clear message (never silent no-op).
- **D-07:** Runtime config path detection using file-existence order at runtime, fall back to primary on first write (same pattern as ClaudeCodeAdapter).
- **D-08:** Docker config entry format: command="docker", args=["run", "-i", "--rm", "image:tag", ...manifest_extra_args].
- **D-09:** Docker install step: run `docker pull <image>` at install time (not lazy). Fail loudly on pull failure.
- **D-10:** DockerInstaller.IsAvailable() checks for `docker` on PATH. Missing → ErrDockerNotFound with install URL + exit code 1.
- **D-11:** Use `uvx` (not pip) for Python-based MCP servers.
- **D-12:** uvx config entry default: command="uvx", args=["mcp-package", ...manifest_args]. Researcher verifies exact pattern.
- **D-13:** UvxInstaller.IsAvailable() checks for `uvx` on PATH. Missing → ErrUvxNotFound with install URL.
- **D-14:** Researcher must verify Gemini CLI exact skill directory and SKILL.md vs GEMINI.md entrypoint.
- **D-15:** Researcher must verify Gemini settings.json format divergence from Claude before deciding on shared base struct.
- **D-16:** OpenCode full implementation required (not stub). Uses "mcp" key, requires "type" field.
- **D-17:** Runtime path detection for ~/.config/opencode/opencode.json. Atomic write + post-verify pattern.
- **D-18:** All adapters fully implement AssistantAdapter interface. Pi returns ErrNotSupported for unimplemented ops.
- **D-19:** All adapters carry forward Phase 1 patterns: homeDir injection, renameio atomic writes, post-install verify, merge behavior (ErrForeignConflict vs auto-overwrite for agentkit-owned).

### Claude's Discretion

- Whether to extract a shared JSON merge base struct for JSON-based adapters (Claude/Gemini/Codex) — depends on format divergence findings.
- Exact uvx invocation args format for common Python MCP packages — researcher to verify (done below).
- Whether Copilot CLI vs VS Code Copilot requires separate sub-adapters or single adapter covers both — researcher verifies (done below).

### Deferred Ideas (OUT OF SCOPE)

- REG-03 (mcpmarket.com API registry)
- REG-04 (agentkit registry add/remove)
- Background config agent (v2)
- Per-project install scope (v3+)
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| AST-02 | GitHub Copilot CLI adapter — runtime path detection, MCP config to correct location, CLI vs VS Code divergence handled as separate sub-adapters | Copilot CLI uses ~/.copilot/mcp-config.json with mcpServers key + type:"local" field. VS Code uses separate ~/.config/Code/User/mcp.json with "servers" key. Two distinct sub-adapters required. |
| AST-03 | OpenAI Codex adapter — TOML config at ~/.codex/config.toml; MCP key names verified | Uses [mcp_servers.<name>] TOML table sections. Fields: command, args (array), env (subtable), env_vars (array of names to forward). TOML writer: BurntSushi/toml already in stack. |
| AST-04 | Gemini CLI adapter — ~/.gemini/settings.json; SKILL.md mapped to GEMINI.md conventions | Gemini uses mcpServers key (same as Claude). Skills live at ~/.gemini/skills/<name>/SKILL.md — entrypoint IS SKILL.md (not GEMINI.md). Formats largely compatible; shared base struct viable. |
| AST-05 | OpenCode adapter — ~/.config/opencode/opencode.json; mcp key (not mcpServers), type field required | Uses "mcp" key. Each entry: type="local" (for stdio), command=[array], environment={} object. Command is an array (not split string+args). Distinct schema from Claude. |
| AST-06 | Pi adapter — install path and skill mechanism researched; degrade gracefully if no CLI skill system | Pi supports BOTH MCP config (~/.pi/agent/mcp.json, format: mcpServers) AND skills (~/.agents/skills/<name>/SKILL.md). Both WriteMCPConfig and WriteSkill implementable per D-05. |
| MCP-02 | Python MCP server install via uvx | Verified pattern: command="uvx", args=["package-name", ...extra-args]. Installer checks `uvx` on PATH. |
| MCP-04 | Docker adapter — pulls image, generates run command, writes docker run entry to MCP config | Verified pattern: command="docker", args=["run", "-i", "--rm", "image:tag", ...extra]. docker pull at install time. |
</phase_requirements>

---

## Summary

Phase 2 expands agentkit from Claude-only to all 5 supported coding assistants, plus two new install methods (uvx and Docker). The core adapter pattern from Phase 1 (ClaudeCodeAdapter) is well-designed for reuse: homeDir injection, renameio atomic writes, runtime path detection, merge-write via map[string]interface{}, and post-install verify all carry forward unchanged.

The five new assistant adapters divide into three groups by config format. Group 1 (Gemini) is near-identical to Claude: same `mcpServers` JSON key, same command/args/env structure — a shared base struct is viable and recommended. Group 2 (Copilot CLI) uses `mcpServers` but adds a required `"type": "local"` field. Group 3 (OpenCode) uses a fundamentally different schema: `mcp` key, `command` is a JSON array (not split command+args), and a required `type` field. Codex is entirely separate: TOML format with `[mcp_servers.<name>]` table syntax.

The Copilot CLI vs VS Code divergence is confirmed significant: different file paths, different top-level keys (`mcpServers` vs `servers`), different locations. Two separate sub-adapters (`copilot-cli` and `copilot-vscode`) are required per the Phase 1 CONTEXT.md note. Pi is confirmed to support both MCP config and skills, enabling full D-05 implementation.

**Primary recommendation:** Implement adapters in this order: (1) CopilotCLIAdapter (closest to Claude pattern + type field), (2) GeminiAdapter (shared base struct candidate), (3) CodexAdapter (TOML — isolated concern), (4) OpenCodeAdapter (distinct JSON schema), (5) PiAdapter (partially shared patterns, two config locations). For installers: UvxInstaller then DockerInstaller, both modeled on NpxInstaller.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| MCP config write (all assistants) | Adapter layer (`internal/adapter/`) | config.paths.go (path resolution) | Adapter owns assistant-specific format; paths.go owns location |
| Skill file write (all assistants) | Adapter layer (`internal/adapter/`) | config.paths.go | Same pattern as ClaudeCodeAdapter.WriteSkill |
| Python MCP server install | Installer layer (`internal/installer/`) | — | UvxInstaller mirrors NpxInstaller |
| Docker image pull + config entry | Installer layer (`internal/installer/`) | — | DockerInstaller: pull at install time, entry at write time |
| Install method dispatch | `NewInstaller()` factory (`internal/installer/installer.go`) | — | Add InstallMethodUvx, InstallMethodDocker cases |
| Target assistant dispatch | `cmd/root.go` + `NewAdapter()` factory | — | --target flag validation + adapter selection |
| Cross-platform path resolution | `config.SkillInstallPath()` | `os.UserHomeDir()` / `os.UserConfigDir()` | Add explicit cases for copilot, codex, gemini, opencode, pi |

---

## Standard Stack

### Core (all already in stack — no new dependencies)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `BurntSushi/toml` | v1.x | Read/write Codex `config.toml` | Already in stack; only TOML dep needed |
| `google/renameio/v2` | v2.x | Atomic file writes | Already in use by ClaudeCodeAdapter |
| `encoding/json` | stdlib | All JSON config R/W | Already in use; no dep |
| `os/exec` | stdlib | Docker/uvx subprocess invocation | Already in use by NpxInstaller |

### No New Dependencies Required
Phase 2 requires zero new Go module dependencies. All needed libraries are already pulled in by Phase 1.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `BurntSushi/toml` | `pelletier/go-toml` | go-toml v2 has better roundtrip but BurntSushi already in stack |
| `os/exec` for Docker | Docker SDK | SDK is massive (~10MB); exec.Command is sufficient for `docker pull` + config entry generation |

---

## Package Legitimacy Audit

> Phase 2 adds NO new Go module dependencies. All packages used are from Phase 1's existing stack. No new slopcheck run required.

| Package | Registry | Age | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|
| `BurntSushi/toml` | go modules | ~12 yrs | N/A (Phase 1) | Approved (Phase 1 vetted) |
| `google/renameio/v2` | go modules | ~5 yrs | N/A (Phase 1) | Approved (Phase 1 vetted) |

**Packages removed due to slopcheck [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

---

## Architecture Patterns

### System Architecture Diagram

```
agentkit install <name> --target <assistant> --method <method>
          │
          ▼
  cmd/root.go → validates --target flag (add copilot-cli, copilot-vscode, codex, gemini, opencode, pi)
          │
          ├─► NewInstaller(method) ──────────────────────────────────────┐
          │      ├── InstallMethodNpx   → NpxInstaller (existing)        │
          │      ├── InstallMethodUvx   → UvxInstaller (NEW)            │
          │      ├── InstallMethodDocker → DockerInstaller (NEW)         │
          │      └── InstallMethodBinary → BinaryInstaller (existing)    │
          │                                                               │
          │      installer.IsAvailable() → fail fast with typed error    │
          │      installer.Install(spec) → run subprocess                │
          │                                                               ▼
          └─► NewAdapter(target) ────────────────────────────────────────┘
                 ├── "claude"         → ClaudeCodeAdapter (existing)
                 ├── "copilot-cli"    → CopilotCLIAdapter (NEW)
                 ├── "copilot-vscode" → CopilotVSCodeAdapter (NEW)
                 ├── "codex"          → CodexAdapter (NEW, TOML)
                 ├── "gemini"         → GeminiAdapter (NEW)
                 ├── "opencode"       → OpenCodeAdapter (NEW)
                 └── "pi"             → PiAdapter (NEW, partial)
                          │
                          ▼
                 adapter.WriteMCPConfig(entry, ownership)
                   → readRawConfig() → merge → renameio.WriteFile() → post-verify
                 adapter.WriteSkill(name, files)
                   → config.SkillInstallPath(target, name) → write files
```

### Recommended Project Structure
```
internal/
├── adapter/
│   ├── adapter.go          # AssistantAdapter interface, ErrForeignConflict (existing)
│   ├── claude.go           # ClaudeCodeAdapter (existing)
│   ├── jsonbase.go         # NEW: shared JSON merge base for Claude/Gemini/CopilotCLI
│   ├── copilot_cli.go      # NEW: CopilotCLIAdapter
│   ├── copilot_vscode.go   # NEW: CopilotVSCodeAdapter
│   ├── codex.go            # NEW: CodexAdapter (TOML)
│   ├── gemini.go           # NEW: GeminiAdapter
│   ├── opencode.go         # NEW: OpenCodeAdapter
│   └── pi.go               # NEW: PiAdapter
├── installer/
│   ├── installer.go        # MCPInstaller interface + NewInstaller factory (extend)
│   ├── npx.go              # NpxInstaller (existing)
│   ├── uvx.go              # NEW: UvxInstaller
│   ├── docker.go           # NEW: DockerInstaller
│   ├── binary.go           # BinaryInstaller (existing)
│   └── custom.go           # CustomInstaller (existing)
└── config/
    └── paths.go            # SkillInstallPath — add explicit cases (extend)
```

### Pattern 1: Shared JSON Base Struct (Claude / Gemini / CopilotCLI)

Three adapters share the same `mcpServers` JSON key and command/args/env structure. Extract a `jsonMCPAdapter` base that handles readRawConfig, merge, atomic write, and post-verify. Adapters embed it and override only: config path detection and any assistant-specific extra fields (e.g., CopilotCLI adds `"type": "local"`).

```go
// Source: derived from ClaudeCodeAdapter pattern (internal/adapter/claude.go)
type jsonMCPAdapter struct {
    homeDir    string
    store      *config.ConfigStore
    configPath func(home string) (string, error) // injected by each adapter
    mcpKey     string                            // "mcpServers" for all three
    extraFields func() map[string]interface{}    // Copilot injects {"type":"local"}; others nil
}
```

**When to use:** GeminiAdapter, CopilotCLIAdapter can embed this. ClaudeCodeAdapter may be refactored to embed it in a follow-up, or left standalone to avoid disrupting Phase 1 work.

### Pattern 2: TOML Adapter (Codex)

Codex is the only TOML-based adapter. Use `BurntSushi/toml` for both read and write. The config schema differs enough that it cannot share the JSON base struct.

```toml
# Source: developers.openai.com/codex/mcp [CITED]
[mcp_servers.my-server]
command = "uvx"
args = ["mcp-server-fetch"]

[mcp_servers.my-server.env]
MY_KEY = "MY_VALUE"
```

Read strategy: `toml.DecodeFile()` into a `map[string]interface{}` (preserves unknown keys). Write: re-encode with `toml.NewEncoder`. Atomic write via renameio after encoding to a bytes.Buffer.

### Pattern 3: OpenCode Array-Command Schema

OpenCode's `command` field is a JSON array (executable + args combined), not a split command+args. The agentkit domain.MCPServerEntry has separate Command and Args — combine them at write time:

```go
// Source: opencode.ai/docs/mcp-servers/ [CITED]
entryMap := map[string]interface{}{
    "type":    "local",
    "command": append([]string{entry.Command}, entry.Args...),
}
if len(entry.Env) > 0 {
    entryMap["environment"] = entry.Env
}
mcp[entry.Name] = entryMap
raw["mcp"] = mcp
```

On read (ReadMCPConfig), split the array back: first element is Command, rest are Args.

### Pattern 4: UvxInstaller (mirrors NpxInstaller)

```go
// Source: derived from NpxInstaller pattern (internal/installer/npx.go)
// Verified invocation: command="uvx", args=["package-name", ...extra-args]
// Example: uvx mcp-server-fetch
// Example: uvx mcp-server-sqlite --db-path /path/to/db.sqlite

var ErrUvxNotFound = errors.New("uvx not found on PATH; install uv to use Python-based MCP servers: https://docs.astral.sh/uv/")

func (u *UvxInstaller) Install(spec domain.InstallSpec) error {
    if !u.IsAvailable() { return ErrUvxNotFound }
    return u.run("uvx", []string{spec.Package}) // extra args from spec.Args appended
}
```

### Pattern 5: DockerInstaller

```go
// Source: verified from Docker Hub mcp/* images and docker.com/blog MCP guide [CITED]
// Invocation: docker run -i --rm image:tag [extra-args...]
// -i keeps stdin open (required for stdio transport)
// --rm cleans up container on exit

var ErrDockerNotFound = errors.New("docker not found on PATH; install Docker: https://docs.docker.com/get-docker/")

func (d *DockerInstaller) Install(spec domain.InstallSpec) error {
    if !d.IsAvailable() { return ErrDockerNotFound }
    return d.run("docker", []string{"pull", spec.Package}) // spec.Package = image:tag
}
// WriteMCPConfig entry built by adapter:
// command="docker", args=["run", "-i", "--rm", "image:tag", ...spec.ExtraArgs]
```

### Anti-Patterns to Avoid
- **Copilot CLI and VS Code sharing one adapter:** Config paths, top-level keys, and field names differ. Single adapter creates a confused mess. Two sub-adapters are clean.
- **Hardcoding OpenCode command as a string:** OpenCode's schema requires a JSON array for `command`. Passing a string will cause parse errors in the OpenCode CLI.
- **Lazy docker pull:** D-09 requires eager pull at install time. Lazy pull means failures surface only when the user's AI assistant first calls the tool — confusing UX.
- **Using `env` key for OpenCode:** OpenCode uses `environment` (not `env`) as the key for MCP server environment variables. Using `env` writes a key OpenCode ignores silently.
- **Assuming Pi skill path is ~/.pi/skills:** Pi loads from `~/.agents/skills/` (user-global) and `.agents/skills/` (project). The `~/.pi/agent/skills/` path is Pi-internal, not the shared skills convention.

---

## Assistant-Specific Config Details

### AST-02: GitHub Copilot — Two Sub-Adapters Required

**CopilotCLIAdapter:**
- Config path: `~/.copilot/mcp-config.json` (or `$COPILOT_HOME/mcp-config.json` if env set)
- Top-level key: `mcpServers`
- Entry format: `{"type": "local", "command": "...", "args": [...], "env": {}, "tools": ["*"]}`
- The `"type": "local"` and `"tools": ["*"]` fields are required by Copilot CLI
- Skills: Copilot CLI has no user-global skill directory analogous to Claude's `~/.claude/skills/`. WriteSkill should return ErrNotSupported with message "copilot-cli adapter: WriteSkill not supported — Copilot CLI has no CLI-level skill directory" [ASSUMED — needs confirmation if Copilot CLI gains skill support]

**CopilotVSCodeAdapter:**
- Config path: Platform-dependent VS Code user settings directory:
  - macOS: `~/Library/Application Support/Code/User/mcp.json`
  - Linux: `~/.config/Code/User/mcp.json`
  - Windows: `%APPDATA%\Code\User\mcp.json`
- Top-level key: `servers` (NOT `mcpServers`)
- Entry format: same command/args/env structure, but under `servers` key
- Skills: Same ErrNotSupported as CopilotCLI

**Confidence:** MEDIUM — Copilot CLI config path and mcpServers format verified via GitHub Docs. VS Code mcp.json path verified via VS Code docs. The `"tools": ["*"]` field is documented as having default `*` — may be optional to write. [CITED: docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers]

### AST-03: OpenAI Codex — TOML

- Config path: `~/.codex/config.toml`
- Section syntax: `[mcp_servers.<server-name>]`
- Fields:
  - `command` (string, required): executable name
  - `args` (array of strings, optional): arguments
  - `env_vars` (array of strings, optional): environment variable names to forward from parent process
  - `[mcp_servers.<name>.env]` (TOML subtable, optional): static key=value env vars
- Read strategy: `toml.DecodeFile` into `map[string]interface{}` to preserve all non-MCP keys
- Write strategy: modify only `mcp_servers` subtable, re-encode with `toml.NewEncoder`
- Skills: No user-global skill directory for Codex. WriteSkill returns ErrNotSupported [ASSUMED — Codex CLI README does not document a skill directory mechanism]

**Confidence:** HIGH for config path and table syntax [CITED: developers.openai.com/codex/mcp, developers.openai.com/codex/config-reference]

### AST-04: Gemini CLI — Near-Claude Format

- Config path: `~/.gemini/settings.json` (runtime detection: check existence, fall back to this path on first write)
- Top-level key: `mcpServers` (same as Claude)
- Entry format: `{"command": "...", "args": [...], "env": {...}}` — no required `type` field
- Skills path: `~/.gemini/skills/<name>/` with `SKILL.md` as entrypoint (NOT GEMINI.md)
  - SKILL.md frontmatter required: `name:` and `description:` fields
  - Capitalization matters: must be `SKILL.md` exactly (case-sensitive on Linux/macOS)
- **Shared base struct verdict:** YES — Gemini and Claude share the same mcpServers format. GeminiAdapter can embed `jsonMCPAdapter` with a different `configPath` function.

**Confidence:** HIGH for mcpServers format [CITED: geminicli.com/docs/tools/mcp-server/]. HIGH for SKILL.md entrypoint [CITED: github.com/google-gemini/gemini-cli/blob/main/docs/cli/skills.md and geminicli.com/docs/cli/skills/]

### AST-05: OpenCode — Distinct Schema

- Config path: `~/.config/opencode/opencode.json`
  - Uses `os.UserConfigDir()` as base on all platforms (correct XDG handling)
- Top-level key: `mcp` (NOT `mcpServers`)
- Entry format (local/stdio):
  ```json
  {
    "type": "local",
    "command": ["executable", "arg1", "arg2"],
    "environment": {"KEY": "value"},
    "enabled": true
  }
  ```
  - `command` is a JSON array (command + args combined) — NOT split into command/args
  - `environment` key (NOT `env`)
  - `type` must be `"local"` for stdio servers
- Skills: OpenCode has no user-global skill directory. WriteSkill returns ErrNotSupported [ASSUMED — opencode.ai docs do not document a skill directory]
- **Cannot share JSON base struct** due to different key names and command array format

**Confidence:** HIGH for mcp key and type field [CITED: opencode.ai/docs/mcp-servers/, github.com/opencode-ai/opencode/blob/main/opencode-schema.json]. HIGH for command-as-array [CITED: multiple opencode.ai examples confirm array format]

### AST-06: Pi — Both MCP and Skills Supported

- MCP config path (user-global): `~/.pi/agent/mcp.json`
  - Alternative shared path `~/.config/mcp/mcp.json` is a cross-tool standard; write to Pi-specific path
  - Top-level key: `mcpServers`
  - Entry format: `{"command": "...", "args": [...], "env": {...}}` (same structure as Claude/Gemini)
- Skills path: `~/.agents/skills/<name>/SKILL.md`
  - Pi discovers skills from `~/.agents/skills/` (user-global) — this is the correct user-scope path
  - `SKILL.md` is the entrypoint (same convention as Claude and Gemini)
- Both WriteMCPConfig and WriteSkill are implementable (satisfies D-05)
- ErrNotSupported not needed for either operation — Pi supports both

**Confidence:** MEDIUM for MCP path [CITED: pi.dev/docs/latest/skills, github.com/nicobailon/pi-mcp-adapter]. LOW for exact ~/.pi/agent/mcp.json path — multiple sources reference both ~/.config/mcp/mcp.json and ~/.pi/agent/mcp.json; recommend writing to ~/.pi/agent/mcp.json (Pi-specific, not shared). Flag for human verify.

### MCP-02: uvx Invocation Pattern (VERIFIED)

```
command = "uvx"
args    = ["package-name", ...additional-args]
```

Real examples:
- `uvx mcp-server-fetch` → args: `["mcp-server-fetch"]`
- `uvx mcp-server-sqlite --db-path /path/db.sqlite` → args: `["mcp-server-sqlite", "--db-path", "/path/db.sqlite"]`

uvx creates an isolated Python virtualenv per-run, analogous to npx. No separate install step beyond running it — but agentkit's UvxInstaller.Install() should still run `uvx <package>` once to pre-cache the virtualenv (warm-up). [CITED: pypi.org/project/mcp-server-fetch/, dev.to/leomarsh/mcp-server-executables-explained]

### MCP-04: Docker Invocation Pattern (VERIFIED)

```
command = "docker"
args    = ["run", "-i", "--rm", "image:tag", ...extra-args]
```

Real examples:
- `docker run -i --rm ghcr.io/github/github-mcp-server` (with env vars via -e flags in extra-args)
- `docker run -i --rm -v /local:/local mcp/filesystem /local`

The `-i` flag is mandatory for stdio transport — without it, stdin closes and the MCP server exits immediately. The `--rm` flag is required to prevent container accumulation. Volume mounts and env vars are supplied as additional elements in the args array from the registry manifest. [CITED: hub.docker.com/r/mcp/filesystem, docker.com/blog/simplify-ai-development-with-the-model-context-protocol-and-docker]

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Atomic file writes | Custom temp+rename logic | `google/renameio/v2` | Already in stack; handles Windows NTFS rename semantics |
| TOML serialization | Manual TOML string formatting | `BurntSushi/toml` encoder | TOML escaping, inline tables, multiline strings are complex |
| JSON merge with unknown keys | Full struct deserialization | `map[string]interface{}` + `json.Unmarshal` | Struct deserialization drops unmanaged keys; map preserves them |
| Docker availability check | Parsing `docker version` output | `exec.LookPath("docker")` | LookPath is sufficient; only need PATH presence, not version |
| VS Code config path | Platform switch with hardcoded strings | `os.UserConfigDir()` + path join for Linux/macOS; `os.Getenv("APPDATA")` for Windows | XDG-aware stdlib handles Linux/macOS; Windows needs APPDATA |

---

## Common Pitfalls

### Pitfall 1: OpenCode Command Field is an Array
**What goes wrong:** Writing `"command": "uvx"` and `"args": ["mcp-server-fetch"]` as separate fields. OpenCode ignores the entry or fails to parse it.
**Why it happens:** OpenCode's schema deviates from the MCP stdio convention used by Claude/Gemini/Copilot.
**How to avoid:** At write time, combine: `command: append([]string{entry.Command}, entry.Args...)`. At read time, split: `Command = arr[0]`, `Args = arr[1:]`.
**Warning signs:** OpenCode shows MCP server as "not configured" despite file containing the entry.

### Pitfall 2: Missing `-i` Flag in Docker MCP Invocation
**What goes wrong:** Docker container starts, then immediately exits. MCP connection times out.
**Why it happens:** Without `-i`, stdin is not connected. stdio-transport MCP servers read from stdin and exit when it closes.
**How to avoid:** D-08 locks this: args always start with `["run", "-i", "--rm", "image:tag"]`.
**Warning signs:** Docker container runs with exit code 0 immediately after start.

### Pitfall 3: Copilot CLI Missing `type: local` Field
**What goes wrong:** MCP server entry is written but Copilot CLI silently ignores it or reports "invalid server config."
**Why it happens:** Copilot CLI requires the `"type"` field; without it the entry is malformed for Copilot's parser.
**How to avoid:** CopilotCLIAdapter always injects `"type": "local"` in the entry map.
**Warning signs:** `gh copilot` command shows no MCP tools despite the config file being correctly updated.

### Pitfall 4: TOML Merge Drops Unknown Keys
**What goes wrong:** Codex config.toml has user-written settings (model, theme, etc.) that disappear after agentkit writes MCP config.
**Why it happens:** Decoding TOML into a typed struct and re-encoding drops fields not in the struct.
**How to avoid:** Decode into `map[string]interface{}`, modify only `mcp_servers` subtable, re-encode the map.
**Warning signs:** User reports losing their Codex settings after running agentkit.

### Pitfall 5: Pi Skill Path vs MCP Path Confusion
**What goes wrong:** Skills written to `~/.pi/skills/` or MCP config written to `~/.config/mcp/mcp.json` may not be read by Pi depending on version and config.
**Why it happens:** Pi has multiple discovery paths with precedence; the shared `~/.config/mcp/mcp.json` is a cross-tool standard (not Pi-specific).
**How to avoid:** Write skills to `~/.agents/skills/<name>/` (user-global, Pi-confirmed path). Write MCP to `~/.pi/agent/mcp.json` (Pi-specific override).
**Warning signs:** Pi agent does not list installed skills or MCP server when using `/skill:` or listing tools.

### Pitfall 6: VS Code Copilot Config Dir Varies by Edition
**What goes wrong:** Writing to `~/.config/Code/User/mcp.json` but user has VS Code Insiders or a code-server variant; file is in a different directory.
**Why it happens:** VS Code Insiders uses `Code - Insiders` directory; code-server uses `code-server` directory.
**How to avoid:** CopilotVSCodeAdapter runtime detection: check `Code`, then `Code - Insiders`, then `code-server` directories in order. Document this in the adapter's godoc.
**Warning signs:** File is written successfully but VS Code Copilot doesn't pick up the MCP server.

---

## Shared Base Struct Decision (Claude's Discretion Resolved)

**Recommendation: YES — extract `jsonMCPAdapter` base struct.**

Rationale: Three adapters (Claude, Gemini, CopilotCLI) share the same `mcpServers` JSON format. The only differences are:
1. Config file path (each adapter provides its own path detection function)
2. Extra fields in the entry map (CopilotCLI adds `"type": "local"` and optionally `"tools": ["*"]`)

The refactor is low-risk: ClaudeCodeAdapter already implements the canonical pattern; extracting it to `jsonMCPAdapter` and having all three embed it reduces ~300 lines of duplicated merge/write/verify logic to ~30 lines per adapter (path detection + optional entry extras).

OpenCode and Pi do NOT use the shared base (different key names, array command format). CodexAdapter does NOT use it (TOML format).

---

## Code Examples

### Gemini mcpServers write (shares Claude pattern)
```go
// Source: geminicli.com/docs/tools/mcp-server/ [CITED]
// settings.json format:
{
  "mcpServers": {
    "my-server": {
      "command": "uvx",
      "args": ["mcp-server-fetch"],
      "env": {
        "MY_KEY": "value"
      }
    }
  }
}
```

### Codex TOML write
```go
// Source: developers.openai.com/codex/mcp [CITED]
// config.toml format:
[mcp_servers.my-server]
command = "uvx"
args = ["mcp-server-fetch"]

[mcp_servers.my-server.env]
MY_KEY = "value"
```

### OpenCode mcp write
```go
// Source: opencode.ai/docs/mcp-servers/ [CITED]
// opencode.json format:
{
  "mcp": {
    "my-server": {
      "type": "local",
      "command": ["uvx", "mcp-server-fetch"],
      "environment": {
        "MY_KEY": "value"
      }
    }
  }
}
```

### Pi MCP write
```go
// Source: pi.dev/docs/latest/skills, github.com/nicobailon/pi-mcp-adapter [CITED]
// ~/.pi/agent/mcp.json format:
{
  "mcpServers": {
    "my-server": {
      "command": "uvx",
      "args": ["mcp-server-fetch"],
      "env": {}
    }
  }
}
```

### Copilot CLI mcp-config.json write
```go
// Source: docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers [CITED]
{
  "mcpServers": {
    "my-server": {
      "type": "local",
      "command": "uvx",
      "args": ["mcp-server-fetch"],
      "env": {},
      "tools": ["*"]
    }
  }
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| pip install for Python MCP servers | uvx (uv toolchain) | 2024-2025 | No virtualenv management, isolated envs per run |
| Claude Desktop settings.json | ~/.claude.json (primary path) | 2025 | ClaudeCodeAdapter already handles both via runtime detection |
| OpenCode `mcpServers` key | `mcp` key | OpenCode v1 | Must not assume mcpServers works for OpenCode |
| Copilot unified config | Split CLI/VS Code configs | Current (open issue #187954) | Two sub-adapters required; no convergence yet |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Copilot CLI WriteSkill returns ErrNotSupported — no CLI-level skill directory | AST-02 Copilot | If Copilot gains skill support, adapter needs WriteSkill implementation |
| A2 | Codex CLI has no skill directory; WriteSkill returns ErrNotSupported | AST-03 | If Codex gains skill support, adapter needs WriteSkill implementation |
| A3 | OpenCode has no skill directory; WriteSkill returns ErrNotSupported | AST-05 | If OpenCode gains skill support, adapter needs WriteSkill implementation |
| A4 | Pi MCP config primary path is ~/.pi/agent/mcp.json (not ~/.config/mcp/mcp.json) | AST-06 Pi | Wrong path = Pi doesn't read the MCP config; user experience failure |
| A5 | Copilot CLI `"tools": ["*"]` field is optional (default is all tools) | AST-02 | If required and omitted, Copilot CLI may not expose tools from the server |
| A6 | VS Code Copilot uses `Code` subdirectory name (not `Code - Insiders` or variant) on the user's machine | AST-02 VS Code | Wrong path = config written but not read by VS Code |

---

## Open Questions

1. **Pi MCP config path: ~/.pi/agent/mcp.json vs ~/.config/mcp/mcp.json**
   - What we know: Pi reads both; ~/.config/mcp/mcp.json is a cross-tool shared convention; ~/.pi/agent/mcp.json is Pi-specific
   - What's unclear: Which path Pi treats as authoritative for user-global installs; whether writing to the shared path risks conflicts with other tools
   - Recommendation: Write to ~/.pi/agent/mcp.json (Pi-specific, no cross-tool collision). Document the shared path as a fallback read source.

2. **Copilot CLI `tools` field: required or optional?**
   - What we know: GitHub Docs show examples with `"tools": ["*"]`; the field description says default is `*`
   - What's unclear: Whether Copilot CLI's parser requires the field to be present explicitly
   - Recommendation: Always write `"tools": ["*"]` to be safe. The field is idempotent.

3. **VS Code Copilot variant detection**
   - What we know: Standard VS Code is `Code/`; Insiders is `Code - Insiders/`; code-server may differ
   - What's unclear: How common non-standard editions are in agentkit's target user base
   - Recommendation: Detect in order: Code → Code - Insiders → code-server. Emit a warning if none found rather than failing.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `uvx` (uv toolchain) | UvxInstaller | Runtime check via IsAvailable() | — | ErrUvxNotFound with install URL |
| `docker` CLI | DockerInstaller | Runtime check via IsAvailable() | — | ErrDockerNotFound with install URL |
| `node` / `npx` | NpxInstaller (Phase 1) | Runtime check (existing) | — | ErrNodeNotFound (existing) |

All dependencies are user-installed at runtime; agentkit does not vendor them. IsAvailable() gates the install path cleanly.

---

## Validation Architecture

> workflow.nyquist_validation not explicitly set to false in .planning/config.json — include section.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify (existing from Phase 1) |
| Config file | none (standard `go test ./...`) |
| Quick run command | `go test ./internal/adapter/... ./internal/installer/...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| AST-02 | CopilotCLI writes mcpServers with type:local | unit | `go test ./internal/adapter/... -run TestCopilotCLI` | ❌ Wave 0 |
| AST-02 | CopilotVSCode writes to servers key | unit | `go test ./internal/adapter/... -run TestCopilotVSCode` | ❌ Wave 0 |
| AST-03 | Codex writes [mcp_servers.x] TOML; preserves other keys | unit | `go test ./internal/adapter/... -run TestCodex` | ❌ Wave 0 |
| AST-04 | Gemini writes mcpServers; WriteSkill creates ~/.gemini/skills/name/SKILL.md | unit | `go test ./internal/adapter/... -run TestGemini` | ❌ Wave 0 |
| AST-05 | OpenCode writes mcp key with command array | unit | `go test ./internal/adapter/... -run TestOpenCode` | ❌ Wave 0 |
| AST-06 | Pi writes ~/.pi/agent/mcp.json and ~/.agents/skills/name/SKILL.md | unit | `go test ./internal/adapter/... -run TestPi` | ❌ Wave 0 |
| MCP-02 | UvxInstaller.Install calls uvx with package name | unit | `go test ./internal/installer/... -run TestUvx` | ❌ Wave 0 |
| MCP-04 | DockerInstaller.Install calls docker pull; entry has -i --rm | unit | `go test ./internal/installer/... -run TestDocker` | ❌ Wave 0 |

All test files follow the existing pattern: homeDir injection for adapters, injected lookPath/run functions for installers (see NpxInstaller pattern).

### Wave 0 Gaps
- [ ] `internal/adapter/copilot_cli_test.go` — covers AST-02 (CLI sub-adapter)
- [ ] `internal/adapter/copilot_vscode_test.go` — covers AST-02 (VS Code sub-adapter)
- [ ] `internal/adapter/codex_test.go` — covers AST-03
- [ ] `internal/adapter/gemini_test.go` — covers AST-04
- [ ] `internal/adapter/opencode_test.go` — covers AST-05
- [ ] `internal/adapter/pi_test.go` — covers AST-06
- [ ] `internal/installer/uvx_test.go` — covers MCP-02
- [ ] `internal/installer/docker_test.go` — covers MCP-04

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | N/A — local file writes only |
| V3 Session Management | no | N/A |
| V4 Access Control | no | File permissions set to 0644 (readable by user only via OS) |
| V5 Input Validation | yes | Package name, image tag, and args validated via domain.InstallSpec; no shell interpolation |
| V6 Cryptography | no | N/A |

### Known Threat Patterns for This Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Shell injection via package name or image tag | Tampering | exec.Command arg-array form only (existing pattern from NpxInstaller — never shell string interpolation) |
| Docker image pull from untrusted registry | Spoofing | Manifest specifies full image:tag; registry manifest is SHA256-verified (Phase 1 pattern carries forward) |
| Malicious postinstall via uvx package | Tampering | uvx runs in isolated env; agentkit does not install into the user's global Python env |
| Config file clobber (foreign keys wiped) | Tampering | map[string]interface{} merge pattern preserves all unmanaged keys; tested per Wave 0 gap items |

---

## Sources

### Primary (HIGH confidence)
- [docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers) — Copilot CLI mcp-config.json format, mcpServers key, type:local field
- [developers.openai.com/codex/mcp](https://developers.openai.com/codex/mcp) — Codex [mcp_servers.<name>] TOML table syntax, command/args/env fields
- [developers.openai.com/codex/config-reference](https://developers.openai.com/codex/config-reference) — Full Codex config reference
- [geminicli.com/docs/tools/mcp-server/](https://geminicli.com/docs/tools/mcp-server/) — Gemini settings.json mcpServers format
- [geminicli.com/docs/cli/skills/](https://geminicli.com/docs/cli/skills/) — Gemini skills directory (~/.gemini/skills/), SKILL.md entrypoint
- [opencode.ai/docs/mcp-servers/](https://opencode.ai/docs/mcp-servers/) — OpenCode mcp key, type:local, command array format, environment key
- [github.com/opencode-ai/opencode/blob/main/opencode-schema.json](https://github.com/opencode-ai/opencode/blob/main/opencode-schema.json) — OpenCode JSON schema (command is array type)
- [pi.dev/docs/latest/skills](https://pi.dev/docs/latest/skills) — Pi skills path (.agents/skills/), SKILL.md format
- [hub.docker.com/r/mcp/filesystem](https://hub.docker.com/r/mcp/filesystem) — Docker MCP image invocation pattern (docker run -i --rm)
- [pypi.org/project/mcp-server-fetch/](https://pypi.org/project/mcp-server-fetch/) — uvx invocation pattern for Python MCP servers

### Secondary (MEDIUM confidence)
- [deepwiki.com/github/copilot-cli/5.3-mcp-server-configuration](https://deepwiki.com/github/copilot-cli/5.3-mcp-server-configuration) — Copilot CLI config structure details
- [github.com/nicobailon/pi-mcp-adapter](https://github.com/nicobailon/pi-mcp-adapter) — Pi MCP config path (~/.pi/agent/mcp.json) and mcpServers format
- [github.com/orgs/community/discussions/187954](https://github.com/orgs/community/discussions/187954) — Confirmed Copilot CLI vs VS Code config divergence (community discussion)
- [dev.to/leomarsh/mcp-server-executables-explained](https://dev.to/leomarsh/mcp-server-executables-explained-npx-uvx-docker-and-beyond-1i1n) — uvx and Docker invocation patterns
- [code.visualstudio.com/docs/agent-customization/mcp-servers](https://code.visualstudio.com/docs/agent-customization/mcp-servers) — VS Code mcp.json path and `servers` key

### Tertiary (LOW confidence)
- Various community blog posts confirming Pi ~/.agents/skills/ path — corroborate official pi.dev docs

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies; all confirmed in Phase 1
- Copilot CLI adapter: HIGH — GitHub Docs authoritative
- Codex adapter: HIGH — OpenAI Developers docs authoritative
- Gemini adapter: HIGH — Google Gemini CLI official docs and GitHub repo
- OpenCode adapter: HIGH — opencode.ai official docs + schema.json
- Pi adapter (MCP path): MEDIUM — multiple sources agree but slight ambiguity on primary path
- Pi adapter (skills path): MEDIUM — pi.dev docs confirm ~/.agents/skills/
- uvx invocation: HIGH — verified against PyPI packages and community documentation
- Docker invocation: HIGH — verified against Docker Hub official MCP images

**Research date:** 2026-06-09
**Valid until:** 2026-07-09 (30 days; Copilot CLI and OpenCode docs evolve faster — re-verify if planning is delayed beyond 2 weeks)
