# agentkit Registry Entry Format

This file defines the `domain.Package` JSON schema used in `testdata/registry.json` and
the live `ejyle/agentkit-registry`. Every entry you generate must conform to this schema.

## domain.Package Schema

```json
{
  "name": "string — lowercase, hyphen-separated, unique in registry",
  "version": "string — SemVer (e.g. '1.0.0') or 'latest' for always-current packages",
  "description": "string — one sentence, what it does and key use cases; under 120 chars",
  "type": "mcp | skill | agent",
  "source": "string — github.com/{owner}/{repo} (no https://, no trailing slash)",
  "install": { ... },        // InstallSpec — see below
  "mcp_entry": { ... },      // MCPServerEntry — required for type:mcp only
  "sha256": ""               // leave empty string; filled by release pipeline
}
```

## type field

| Value | Meaning |
|-------|---------|
| `mcp` | MCP server — configure in assistant's mcpServers config |
| `skill` | Skill with SKILL.md — installed to assistant's skills directory |
| `agent` | Coding agent — custom agent definition |

## InstallSpec (install field)

### method: "npx"
For npm-published MCP servers that run via Node.

```json
"install": {
  "method": "npx",
  "package": "@scope/package-name@latest",
  "args": ["--version"]
}
```

`args` are **install-time only** — used to verify the package runs cleanly during install.
Set `args: ["--version"]` for any package that errors without a subcommand (e.g. `@azure/mcp`,
`@azure-devops/mcp`). Leave `args: []` for packages that start cleanly with no args (e.g.
`@playwright/mcp`, `@upstash/context7-mcp`).

### method: "uvx"
For Python-published MCP servers that run via uv/uvx.

```json
"install": {
  "method": "uvx",
  "package": "package-name"
}
```

No `args` needed — uvx packages typically start cleanly.

### method: "github-release"
For skills bundled inside an agentkit release tarball. `repo` must be `ejyle/agentkit`.
`path` is the subdirectory inside the release archive.

```json
"install": {
  "method": "github-release",
  "repo": "ejyle/agentkit",
  "path": "skills/external/skill-name"
}
```

### method: "github-default-branch"
For skills or skill collections hosted in external GitHub repos. Always fetches the latest
`main` branch — best for repos that ship updates frequently (e.g. `microsoft/azure-skills`).

```json
"install": {
  "method": "github-default-branch",
  "repo": "owner/repo",
  "path": "path/to/skill/dir"
}
```

**`path: "."`** — extracts the entire repo root. Only use this if `SKILL.md` exists at
the repo root. **Verify first** (see Path Verification below).

**`multi_skill: true`** — set this when the `path` points to a directory containing
multiple skills as immediate subdirectories (each with their own `SKILL.md`). The installer
will extract each subdir as its own skill.

```json
"install": {
  "method": "github-default-branch",
  "repo": "microsoft/azure-skills",
  "path": "skills",
  "multi_skill": true
}
```

### method: "docker"
For MCP servers distributed as Docker images.

```json
"install": {
  "method": "docker",
  "package": "ghcr.io/owner/image-name:latest"
}
```

### method: "custom"
For packages with non-standard install flows (e.g. `@opengsd/gsd-core` which takes a
`--$TARGET` flag substituted at install time).

```json
"install": {
  "method": "custom",
  "package": "npx",
  "args": ["@opengsd/gsd-core@latest", "--$TARGET", "--global"]
}
```

`$TARGET` is replaced at runtime with `--claude`, `--gemini`, etc.

## MCPServerEntry (mcp_entry field)

Required for `type: mcp`. This is the entry written to the assistant's MCP config.
**These are runtime args, not install-time args.**

```json
"mcp_entry": {
  "name": "same as package name field",
  "command": "npx | uvx | docker",
  "args": ["runtime", "args", "here"],
  "env": {}
}
```

### Key distinction: install.args vs mcp_entry.args

| Field | Purpose | Example |
|-------|---------|---------|
| `install.args` | Passed during install to verify package runs cleanly | `["--version"]` |
| `mcp_entry.args` | Written to mcpServers config; used every time Claude starts the server | `["-y", "@azure/mcp@latest", "server", "start"]` |

These are independent. A package might need `--version` at install time but `server start`
at runtime (e.g. `@azure/mcp`), or might need an org argument at runtime that you can't
know at install time (e.g. `@azure-devops/mcp`).

### Examples by method

**npx (standard):**
```json
"mcp_entry": { "name": "context7", "command": "npx", "args": ["-y", "@upstash/context7-mcp@latest"], "env": {} }
```

**npx (needs subcommand):**
```json
"mcp_entry": { "name": "azure-mcp", "command": "npx", "args": ["-y", "@azure/mcp@latest", "server", "start"], "env": {} }
```

**uvx:**
```json
"mcp_entry": { "name": "aws-mcp", "command": "uvx", "args": ["awslabs.aws-api-mcp-server@latest"], "env": {} }
```

## Path Verification (before setting path)

Before setting `path` in any `github-default-branch` or `github-release` entry:

1. **Confirm SKILL.md exists at the path** using the GitHub raw URL:
   `https://raw.githubusercontent.com/{owner}/{repo}/HEAD/{path}/SKILL.md`
   - Returns 200 → path is correct
   - Returns 404 → wrong path; check repo structure

2. **Detect multi-skill repos**: Use the GitHub tree API to list `*/SKILL.md` paths:
   `https://api.github.com/repos/{owner}/{repo}/git/trees/HEAD?recursive=1`
   Filter for entries ending in `/SKILL.md`. If skills live under a subdirectory
   (e.g. `skills/azure-ai/SKILL.md`, `skills/azure-compute/SKILL.md`), set:
   - `path` to the parent dir (e.g. `"skills"`)
   - `multi_skill: true`

3. **Never use `path: "."` blindly.** Only use it if `SKILL.md` is at the repo root.
   Extracting the root of a multi-skill repo dumps CI configs, assets, and landing pages
   into a single skill directory — validation will fail.

## Complete Examples

### MCP server (npx, needs subcommand at runtime)
```json
{
  "name": "azure-mcp",
  "version": "latest",
  "description": "Official Azure MCP server — 200+ tools across 40+ Azure services",
  "type": "mcp",
  "source": "github.com/microsoft/mcp",
  "install": { "method": "npx", "package": "@azure/mcp@latest", "args": ["--version"] },
  "mcp_entry": { "name": "azure-mcp", "command": "npx", "args": ["-y", "@azure/mcp@latest", "server", "start"], "env": {} },
  "sha256": ""
}
```

### MCP server (uvx, Python)
```json
{
  "name": "aws-mcp",
  "version": "latest",
  "description": "Official AWS MCP server — 15,000+ AWS API operations",
  "type": "mcp",
  "source": "github.com/awslabs/mcp",
  "install": { "method": "uvx", "package": "awslabs.aws-api-mcp-server" },
  "mcp_entry": { "name": "aws-mcp", "command": "uvx", "args": ["awslabs.aws-api-mcp-server@latest"], "env": {} },
  "sha256": ""
}
```

### Multi-skill collection (github-default-branch)
```json
{
  "name": "azure-skills",
  "version": "latest",
  "description": "Official Microsoft Azure skills — installs all 27+ Azure skills from microsoft/azure-skills",
  "type": "skill",
  "source": "github.com/microsoft/azure-skills",
  "install": { "method": "github-default-branch", "repo": "microsoft/azure-skills", "path": "skills", "multi_skill": true },
  "sha256": ""
}
```

### Single external skill (github-default-branch)
```json
{
  "name": "azure-ai",
  "version": "latest",
  "description": "Official Azure AI skill — AI Foundry, OpenAI on Azure, AI Search patterns",
  "type": "skill",
  "source": "github.com/microsoft/azure-skills",
  "install": { "method": "github-default-branch", "repo": "microsoft/azure-skills", "path": "skills/azure-ai" },
  "sha256": ""
}
```

### Bundled skill (github-release, ships in agentkit releases)
```json
{
  "name": "claude-api",
  "version": "1.0.0",
  "description": "Integrate the Anthropic Claude API — messages, streaming, tool use, prompt caching",
  "type": "skill",
  "source": "github.com/anthropics/skills",
  "install": { "method": "github-release", "repo": "ejyle/agentkit", "path": "skills/external/claude-api" },
  "sha256": ""
}
```

## Where to Add Entries

- **Test/dev**: `testdata/registry.json` — loaded via `AGENTKIT_REGISTRY_FILE` env var
- **Production**: `ejyle/agentkit-registry` repo at `registry.json` — the live registry
  fetched by `NewRegistryManager()` in the CLI

Always add to `testdata/registry.json` first and verify with:
```bash
AGENTKIT_REGISTRY_FILE=testdata/registry.json ./agentkit search <name>
AGENTKIT_REGISTRY_FILE=testdata/registry.json ./agentkit install <name> --target claude
```
