# Architecture Research: agentkit

**Domain:** Go CLI for AI agent skill/MCP management
**Researched:** 2026-06-08
**Overall confidence:** HIGH (pattern-based) / MEDIUM (MCP config paths — some per-version variance)

---

## Major Components

```
┌─────────────────────────────────────────────────────────┐
│                    CLI Layer (Cobra)                    │
│  install / search / update / list / registry commands  │
└───────────────────┬─────────────────────────────────────┘
                    │ calls
┌───────────────────▼─────────────────────────────────────┐
│                  Service Layer                          │
│  InstallService / SearchService / UpdateService        │
│  (business logic — no I/O knowledge here)              │
└──────┬────────────┬──────────────┬──────────────────────┘
       │            │              │
┌──────▼──┐  ┌──────▼────┐  ┌─────▼─────────────────────┐
│Registry │  │ Assistant │  │   Config Store             │
│Manager  │  │ Adapter   │  │ (~/.agentkit/state.json,   │
│         │  │ Registry  │  │  .agent-utils/config.json) │
└──────┬──┘  └──────┬────┘  └────────────────────────────┘
       │            │
┌──────▼──┐  ┌──────▼──────────────┐
│Registry │  │ MCP Install         │
│Sources  │  │ Adapters            │
│(GitHub/ │  │ (npx/pip/docker/    │
│API/Git) │  │  binary)            │
└─────────┘  └─────────────────────┘
```

### Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| CLI Layer | Flag parsing, command routing, user output | Service Layer only |
| InstallService | Orchestrates resolve → fetch → install → configure | RegistryManager, AssistantAdapter, MCPInstaller, ConfigStore |
| SearchService | Fan-out across registries, rank, deduplicate | RegistryManager |
| RegistryManager | Resolves package names across ordered registry sources | Registry Source implementations |
| Registry Sources | Fetch manifest data from GitHub/API/custom | HTTP client, local cache |
| AssistantAdapter | Writes skill/agent files and MCP config to correct per-assistant path | Filesystem |
| MCPInstaller | Executes the correct install method for an MCP entry | OS exec, Docker |
| ConfigStore | Read/write agentkit state and per-project `.agent-utils/config.json` | Filesystem |
| Background Agent | Detects new infra facts at skill end, writes config diff | ConfigStore, LLM subprocess |

---

## Registry Abstraction

**Pattern:** Interface + ordered source list, same as Homebrew taps or npm scoped registries.

Define a single `Registry` interface. `RegistryManager` holds a priority-ordered slice of implementations. On `Resolve(name)`, it walks the list and returns the first match, enabling deduplication by name. This is the same mental model as Homebrew taps: primary tap first, third-party taps after.

```go
type Package struct {
    Name        string
    Version     string
    Type        string // "skill" | "agent" | "mcp"
    Source      string // registry ID that provided this
    Manifest    Manifest
}

type Registry interface {
    ID() string
    Resolve(name string) (*Package, error)
    Search(query string) ([]Package, error)
    List() ([]Package, error)
}
```

**Implementations to build:**

| Implementation | Source | Notes |
|---------------|--------|-------|
| `GitHubManifestRegistry` | GitHub repo with `registry.json` at root | Fetches manifest, caches with ETag |
| `MCPMarketRegistry` | mcpmarket.com REST API | Paginated search endpoint |
| `GSDCoreRegistry` | open-gsd/gsd-core as a source | Treated as a GitHubManifestRegistry with known URL |
| `LocalRegistry` | Local directory path | Dev/testing; also how custom sources work |

**RegistryManager behavior:**

- Ordered list in `~/.agentkit/config.json` under `"registries"` key
- `Resolve`: first-match wins (priority order)
- `Search`: fan-out all, merge, deduplicate by `name`, label each with `source`
- Cache layer: per-registry TTL (30 min default), stored in `~/.agentkit/cache/`

**Confidence:** HIGH — mirrors established Homebrew tap and cargo source patterns.

---

## Target Assistant Adapters

**Pattern:** Interface per concern (skill install, MCP config write), with one struct per assistant.

Each assistant has different: (a) skill/agent file paths, (b) MCP config file location, (c) MCP JSON structure.

```go
type AssistantAdapter interface {
    ID() string
    SkillDir() string               // where SKILL.md files land
    AgentDir() string               // where .md agent files land
    MCPConfigPath() string          // where to write MCP config JSON
    WriteMCPConfig(server MCPServerEntry) error
    ReadMCPConfig() (MCPConfig, error)
}
```

**Known config paths (confidence: MEDIUM — verify at build time):**

| Assistant | Skill/Agent Dir | MCP Config Path | Notes |
|-----------|----------------|-----------------|-------|
| Claude Code | `~/.claude/` (CLAUDE.md + skills/) | `~/.claude/settings.json` → `mcpServers` key | Also `~/.mcp.json` in newer versions |
| GitHub Copilot CLI | `~/.copilot/` | `~/.copilot/mcp-config.json` → `mcpServers` key | COPILOT_HOME env overrides |
| Gemini CLI | `~/.gemini/` | `~/.gemini/settings.json` → `mcpServers` key | |
| OpenCode | `~/.config/opencode/` | OpenCode config under `mcp` key | Verify at build time |
| Codex | `~/.codex/` | TBD — verify at build time | LOW confidence |

**MCP JSON structure variance:**

- Claude Code, Copilot, Gemini: root key `mcpServers`
- VS Code (Copilot in editor): root key `servers` — different from CLI
- All use `{ "command": "...", "args": [...], "env": {} }` for stdio servers

**Implementation note:** Each adapter's `WriteMCPConfig` must merge (not overwrite) the existing file — read-modify-write with file lock. Use `go.uber.org/atomic` or simple `sync.Mutex` in a single process; for cross-process safety use `github.com/gofrs/flock`.

**Confidence:** MEDIUM — MCP config path specifics have had documented instability (Claude Code issue #4976, #32398). Build with path resolution at runtime via env/XDG fallbacks, not hardcoded strings.

---

## MCP Install Adapters

**Pattern:** Same interface abstraction as AssistantAdapters, one struct per install method.

```go
type MCPInstaller interface {
    Method() string  // "npx" | "pip" | "docker" | "binary"
    Install(entry MCPManifestEntry) error
    IsAvailable() bool  // check runtime prereq (node, python, docker)
}
```

| Installer | Command | Prereq Check | Notes |
|-----------|---------|--------------|-------|
| `NpxInstaller` | `npx -y <package>` at runtime | `node --version` | No install step; npx fetches on first run |
| `PipInstaller` | `pip install <package>` | `pip --version` | Install once, then `python -m <module>` |
| `BinaryInstaller` | Download URL from manifest, chmod +x | None | Store in `~/.agentkit/bin/` |
| `DockerInstaller` | `docker pull <image>` | `docker info` | Run as `docker run -i --rm <image>` |

**Manifest-override pattern:** The registry manifest entry can specify `install.override` with custom shell steps. The `CustomInstaller` wraps this as a shell script executed in a sandboxed temp dir.

**Post-install:** Every installer calls `adapter.WriteMCPConfig(entry)` to register the server with the target assistant.

**Confidence:** HIGH — standard patterns, no novel approach needed.

---

## Manifest Format Design

Every package in any registry is described by a manifest entry. GitHub-based registries serve a `registry.json` at repo root listing all packages; each entry follows this schema.

```json
{
  "name": "playwright",
  "version": "1.2.0",
  "type": "mcp",
  "description": "Browser automation and E2E testing via Playwright MCP",
  "tags": ["browser", "testing", "e2e"],
  "source": {
    "registry": "github:microsoft/playwright-mcp",
    "ref": "v1.2.0"
  },
  "install": {
    "method": "npx",
    "package": "@playwright/mcp",
    "args": [],
    "env": {},
    "override": null
  },
  "targets": ["claude", "copilot", "gemini", "opencode", "codex"],
  "skill": {
    "entry": "SKILL.md",
    "references": ["references/"],
    "scripts": ["scripts/"]
  },
  "requires": [],
  "license": "MIT",
  "homepage": "https://github.com/microsoft/playwright-mcp"
}
```

**Field rationale:**

| Field | Purpose |
|-------|---------|
| `type` | `skill` / `agent` / `mcp` / `bundle` — drives install path |
| `install.method` | Selects MCPInstaller implementation |
| `install.override` | Escape hatch for non-standard install |
| `targets` | Which assistants this package supports — used to filter `--target` |
| `skill.*` | Present only when type=skill; describes progressive disclosure structure |
| `requires` | Dependency names resolved from same registry chain |

**Registry index format** (`registry.json` at repo root):

```json
{
  "version": "1",
  "updated": "2026-06-08T00:00:00Z",
  "packages": [ ...manifest entries... ],
  "bundles": {
    "cloud": ["aws", "gcp", "azure"],
    "testing": ["playwright", "serena"]
  }
}
```

**Confidence:** HIGH — synthesized from npm/brew/cargo patterns plus agentkit-specific needs.

---

## Project Config Schema

**Location:** `.agent-utils/config.json` (per-project, gitignore-able)

```json
{
  "$schema": "https://agentkit.dev/schemas/project-config/v1.json",
  "version": "1",
  "updated": "2026-06-08T12:34:56Z",
  "project": {
    "id": "auto-generated-uuid",
    "name": "inferred from git remote or dirname"
  },
  "facts": {
    "aws.region": {
      "value": "us-east-1",
      "source": "aws-skill",
      "confidence": "discovered",
      "updated": "2026-06-08T10:00:00Z"
    },
    "aws.account_id": {
      "value": "123456789012",
      "source": "aws-skill",
      "confidence": "confirmed",
      "updated": "2026-06-08T10:01:00Z"
    }
  },
  "installed_skills": ["aws", "gcp", "playwright"],
  "custom": {}
}
```

**Schema versioning strategy:**

- `"version"` field in JSON is the schema version integer
- Go embeds the JSON schema file via `//go:embed schemas/project-config/v1.json`
- On read: validate against embedded schema for current version
- On upgrade: migration functions keyed by `fromVersion → toVersion` in a registry map

**Validation library:** `kaptinlin/jsonschema` — supports Draft 2020-12, direct struct validation, default-aware unmarshaling. Prefer this over `gojsonschema` (older, less active).

**Fact entry fields:**

| Field | Values | Purpose |
|-------|--------|---------|
| `confidence` | `discovered` / `confirmed` / `manual` | `discovered` = auto-saved silently; `confirmed` = user approved; `manual` = user typed |
| `source` | skill name | Which skill wrote this fact |
| `updated` | ISO 8601 | For drift detection |

**Drift detection:** When a skill runs and discovers a value that differs from stored value by >threshold (configurable), the background agent prompts for confirmation before overwriting.

**Confidence:** HIGH — straightforward JSON config design.

---

## Background Agent Pattern

**Borrowed from:** GSD's subagent architecture — XML-prompted, focused scope, fresh context window per invocation.

**Pattern:** Skills do not embed an agent binary. Instead each SKILL.md includes a reference to a shared agent prompt file installed at `~/.agentkit/agents/config-writer.md`. This agent is invoked via `claude --agent config-writer` (or equivalent per-assistant) only when the skill detects new facts to persist.

**Agent invocation decision tree:**

```
skill_end_hook:
  if new_facts == empty → skip (no agent call)
  if token_cost(new_facts) < 200 tokens → auto-write, print one-liner
  else → invoke config-writer agent with fact diff, await confirmation
```

**Agent prompt structure** (XML-sectioned, following GSD convention):

```xml
<role>
You are the agentkit config-writer agent. Your only job is to update
.agent-utils/config.json with the provided fact diff. Do not reason about
the project. Do not execute commands. Only read the current config, apply
the diff, validate the schema, and write the result.
</role>

<constraints>
- Token budget: 1000 tokens max
- No tool calls except Read and Write
- No general reasoning
- Output: updated config JSON only
</constraints>

<fact_diff>
{{ INJECTED AT RUNTIME }}
</fact_diff>
```

**Key design decisions:**

- Agent is shared across all skills (single file, installed once)
- Invoked as a subagent / subprocess — not embedded in main CLI binary
- XML prompt format ensures the LLM parses section boundaries precisely (Claude is trained to treat XML tags as structural markers, not Markdown headers)
- `fact_diff` is injected inline at invocation time — no file reads needed by the agent
- The agent is stateless: it receives the full current config + diff, produces the new config

**Confidence:** MEDIUM — GSD background agent pattern is documented conceptually; exact invocation mechanism for each assistant (Claude Code has `--agent` flag; others may differ) needs verification per target.

---

## Skill Structure Convention

**Standard layout for every installed skill:**

```
~/.claude/skills/
  aws/
    SKILL.md           # entry point, <500 lines, loaded into context
    references/
      ec2.md           # domain doc, loaded on demand
      iam.md
      s3.md
    scripts/
      aws-health.sh    # executable, run without loading into context
      cost-report.py
```

**SKILL.md frontmatter (YAML block):**

```yaml
---
name: aws
version: 1.0.0
description: AWS CLI and SDK skill for EC2, S3, IAM, ECS/EKS
registry: github:open-gsd/gsd-core
targets: [claude, copilot, gemini, opencode, codex]
requires: []
agentkit_version: ">=1.0"
---
```

**Agentkit validation on install:**

1. Frontmatter parseable and all required fields present
2. `SKILL.md` body line count <= 500
3. All `references/` files reachable
4. All `scripts/` files present and executable bit set
5. Schema version in frontmatter matches installed agentkit capability

**Scaffolding on `agentkit scaffold skill <name>`:**

- Creates the directory tree
- Generates SKILL.md stub with frontmatter filled from flags
- Creates empty `references/` and `scripts/` dirs with `.gitkeep`
- Does not write content — author fills it in

**Confidence:** HIGH — directly derived from PROJECT.md requirements and existing skill conventions observed in the repo.

---

## Suggested Build Order

Dependencies flow bottom-up. Each layer can only be built after its dependencies are solid.

```
Layer 0 (Foundation — no external deps)
  ├── Domain types: Package, Manifest, Fact, MCPServerEntry
  ├── ConfigStore: read/write state.json and project config.json
  └── Schema: embed JSON schemas, validate on read

Layer 1 (I/O Adapters — depends on domain types)
  ├── Registry Sources: GitHubManifest, LocalRegistry
  ├── AssistantAdapters: ClaudeCode, Copilot (stub others)
  └── MCPInstallers: Npx, Binary (stub Pip, Docker)

Layer 2 (Orchestration — depends on Layer 1 interfaces)
  ├── RegistryManager: ordered resolution, search fan-out, cache
  └── InstallService: resolve → fetch → install → configure flow

Layer 3 (CLI — depends on Layer 2)
  ├── `agentkit install <name>`
  ├── `agentkit search <query>`
  ├── `agentkit list`
  ├── `agentkit registry add/remove/list`
  └── `agentkit update [name]`

Layer 4 (Skills + Background Agent)
  ├── Skill scaffolding and validation
  ├── Background config-writer agent prompt
  └── Skill install (type=skill) via AssistantAdapter

Layer 5 (Bundled Skills)
  └── AWS, GCP, Azure, Playwright, context-mode, RTK, Serena, GitHub, CI/CD skills
```

**Build implications for roadmap phases:**

- Phase 1 must produce Layers 0-2 as working code — without this, nothing installs
- Phase 2 should wire Layer 3 CLI with at least Claude Code + Npx path end-to-end
- Phase 3 adds remaining AssistantAdapters and MCPInstallers
- Phase 4 builds the background agent and per-project config system
- Phase 5 authors the bundled skills using the validated structure
- The registry abstraction (Layer 1) should have clean interfaces before any CLI commands are written — adding a new registry source must never require changing the service layer

**Risk: MCP config path instability.** Claude Code has had multiple issues with MCP config file location (GitHub issues #4976, #32398). Build path resolution as a runtime lookup with explicit fallback chain, not a compile-time constant. Write an integration test that verifies the correct path for each supported version.

---

## Data Flow

**Install flow:**

```
CLI: agentkit install playwright --target claude
  → InstallService.Install("playwright", target=claude)
    → RegistryManager.Resolve("playwright")
      → [GitHubManifestRegistry, GSDCoreRegistry, MCPMarketRegistry].Resolve()
      ← returns Package{manifest: ..., source: "github:microsoft/playwright-mcp"}
    → MCPInstaller(method=npx).Install(manifest.install)
      → exec: npx -y @playwright/mcp (no-op; npx fetches at first use)
    → ClaudeCodeAdapter.WriteMCPConfig(MCPServerEntry{name:"playwright", ...})
      → read ~/.claude/settings.json
      → merge mcpServers."playwright" entry
      → write back atomically
    → ConfigStore.RecordInstalled("playwright", version, source)
      → write ~/.agentkit/state.json
  ← print: "Installed playwright@1.2.0 from github:microsoft/playwright-mcp → Claude Code"
```

**Skill-end config-write flow:**

```
Skill runs in Claude Code session
  → skill detects new fact: aws.region = "eu-west-1"
  → compares against .agent-utils/config.json
  → delta is small (1 fact, ~50 tokens)
  → directly writes fact with confidence="discovered"
  → prints: "Saved aws.region=eu-west-1 to project config"
```

---

## Sources

- [GSD Core — GitHub](https://github.com/open-gsd/gsd-core)
- [Claude Code MCP Config](https://code.claude.com/docs/en/mcp)
- [GitHub Copilot CLI MCP Config](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers)
- [Gemini CLI MCP Setup](https://geminicli.com/docs/cli/tutorials/mcp-setup/)
- [Go Registry Pattern](https://github.com/Faheetah/registry-pattern)
- [Go Service/Repository Pattern](https://irahardianto.github.io/service-pattern-go/)
- [kaptinlin/jsonschema](https://github.com/kaptinlin/jsonschema)
- [Go Modules for Package Management Tooling](https://nesbitt.io/2026/02/19/go-modules-for-package-management-tooling.html)
- [GSD XML Plan Structure](https://docs.bswen.com/blog/2026-04-21-gsd-xml-plan-structure/)
- [MCP Config File Locations Guide](https://mcpplaygroundonline.com/blog/complete-guide-mcp-config-files-claude-desktop-cursor-lovable)
- Claude Code issues [#4976](https://github.com/anthropics/claude-code/issues/4976) and [#32398](https://github.com/anthropics/claude-code/issues/32398) on MCP config path instability
