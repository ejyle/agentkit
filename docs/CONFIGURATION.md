<!-- generated-by: gsd-doc-writer -->
# Configuration

agentkit is a single binary with no runtime config file of its own. Configuration is expressed through environment variables, CLI flags, and the per-assistant state files it writes during operation.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AGENTKIT_REGISTRY_FILE` | Optional | — | Path to a local `registry.json` file. When set, this registry is prepended ahead of all remote registries and takes highest priority for package resolution. Used for development, testing, and private registries. |
| `COPILOT_HOME` | Optional | `~/.copilot` | Override the GitHub Copilot CLI home directory. When set, agentkit writes the Copilot MCP config to `$COPILOT_HOME/mcp-config.json` instead of the default `~/.copilot/mcp-config.json`. |

## CLI Flags

The `--target` flag is the primary runtime configuration for every command.

| Flag | Short | Default | Valid Values | Description |
|------|-------|---------|--------------|-------------|
| `--target` | `-t` | `claude` | `claude`, `copilot-cli`, `copilot-vscode`, `codex`, `gemini`, `opencode`, `pi` | Target AI coding assistant. Determines where skills and MCP server config are written. |

Example:

```bash
agentkit install playwright --target gemini
agentkit list --target codex
```

## State Files

agentkit writes two categories of state files at runtime — an install record and a registry manifest cache. These are created automatically; you do not configure them manually.

### Installed Packages State (`installed.json`)

Tracks every package agentkit has installed. One file per target assistant.

**Paths by platform:**

| Platform | Path |
|----------|------|
| macOS | `~/Library/Application Support/agentkit/<target>/installed.json` |
| Linux | `~/.config/agentkit/<target>/installed.json` |
| Windows | `%APPDATA%\agentkit\<target>\installed.json` |

**Schema:**

```json
{
  "packages": {
    "playwright": {
      "name": "playwright",
      "version": "latest",
      "type": "mcp",
      "install_path": "/Users/you/.claude/settings.json#mcpServers.playwright",
      "installed_at": "2026-01-15T10:30:00Z",
      "source_url": "https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json",
      "checksum": ""
    }
  }
}
```

### Registry Manifest Cache

Caches the remote registry manifests locally using ETag-based conditional requests. Stale cache is used automatically when the registry is unreachable.

**Paths by platform:**

| Platform | Path |
|----------|------|
| macOS | `~/Library/Caches/agentkit/<registryID>/manifest.json` |
| Linux | `~/.cache/agentkit/<registryID>/manifest.json` |
| Windows | `%LOCALAPPDATA%\agentkit\<registryID>\manifest.json` |

Registry IDs: `agentkit-registry`, `gsd-core`.

## Skills and MCP Config Locations (Per Assistant)

agentkit writes to these locations when installing packages. The paths are fixed per target assistant.

| Target | Skills Path | MCP Config Path |
|--------|-------------|-----------------|
| `claude` | `~/.claude/skills/<name>/` | `~/.claude/settings.json` or `~/.claude.json` |
| `gemini` | `~/.gemini/skills/<name>/` | `~/.gemini/settings.json` |
| `pi` | `~/.agents/skills/<name>/` | `~/.pi/agent/mcp.json` |
| `copilot-cli` | Not supported | `~/.copilot/mcp-config.json` (or `$COPILOT_HOME/mcp-config.json`) |
| `copilot-vscode` | Not supported | <!-- VERIFY: copilot-vscode MCP config path --> |
| `codex` | Not supported | `~/.codex/config.toml` |
| `opencode` | Not supported | `~/.config/opencode/opencode.json` |

## Registry Format

Registries are GitHub repositories with a `registry.json` at the repo root. The two built-in registries are:

- `agentkit-registry`: `https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json`
- `gsd-core`: `https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json`

**Manifest schema:**

```json
{
  "packages": [
    {
      "name": "playwright",
      "version": "latest",
      "description": "Playwright MCP server for browser automation",
      "type": "mcp",
      "source": "github.com/microsoft/playwright-mcp",
      "install": {
        "method": "npx",
        "package": "@playwright/mcp@latest",
        "args": []
      },
      "mcp_entry": {
        "name": "playwright",
        "command": "npx",
        "args": ["-y", "@playwright/mcp@latest"],
        "env": {}
      },
      "sha256": ""
    }
  ]
}
```

**Supported install methods:** `npx`, `uvx`, `docker`, `binary`, `github-release`, `github-default-branch`, `custom`.

**Package types:** `mcp`, `skill`, `agent`.

### Using a Local Registry

Point agentkit at a local registry file for development or private packages:

```bash
AGENTKIT_REGISTRY_FILE=/path/to/my-registry.json agentkit install my-package
```

The local registry takes priority over all remote registries. Resolution falls through to remote registries for packages not found locally.

## Network Behavior

agentkit makes outbound HTTPS requests only. No credentials are required for the default registries. Timeouts are fixed at 3s dial + 10s response header, with up to 3 automatic retries on transient failures.

For MCP servers that require authentication (such as `github-mcp`), credentials are passed as environment variables in the MCP config entry — for example, `GITHUB_PERSONAL_ACCESS_TOKEN`. These values are set by the user in their shell environment or the assistant's env config, not in agentkit itself.

## Defaults Summary

| Setting | Default |
|---------|---------|
| Target assistant | `claude` |
| Copilot home | `~/.copilot` |
| Local registry override | None (remote registries only) |
| HTTP retry count | 3 |
| HTTP dial timeout | 3 seconds |
| HTTP response timeout | 10 seconds |
