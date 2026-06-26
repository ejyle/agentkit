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

### Agent-Managed Binaries

Binary packages installed via the `binary` or `github-release` methods are extracted to:

| Platform | Path |
|----------|------|
| macOS | `~/Library/Application Support/agentkit/bin/` |
| Linux | `~/.config/agentkit/bin/` |
| Windows | `%APPDATA%\agentkit\bin\` |

### GitHub Release Tarball Cache

Tarballs downloaded during `github-release` installs are cached to avoid re-downloading on reinstall:

| Platform | Path |
|----------|------|
| macOS | `~/Library/Caches/agentkit/releases/<repo-slug>/<version>/tarball.tar.gz` |
| Linux | `~/.cache/agentkit/releases/<repo-slug>/<version>/tarball.tar.gz` |
| Windows | `%LOCALAPPDATA%\agentkit\releases\<repo-slug>\<version>\tarball.tar.gz` |

## Skills and MCP Config Locations (Per Assistant)

agentkit writes to these locations when installing packages. The paths are fixed per target assistant.

| Target | Skills Path | MCP Config Path | MCP Config Key |
|--------|-------------|-----------------|----------------|
| `claude` | `~/.claude/skills/<name>/` | `~/.claude/settings.json` or `~/.claude.json` | `mcpServers` |
| `gemini` | `~/.gemini/skills/<name>/` | `~/.gemini/settings.json` | `mcpServers` |
| `pi` | `~/.agents/skills/<name>/` | `~/.pi/agent/mcp.json` | `mcpServers` |
| `copilot-cli` | Not supported | `~/.copilot/mcp-config.json` (or `$COPILOT_HOME/mcp-config.json`) | `mcpServers` |
| `copilot-vscode` | Not supported | `{UserConfigDir}/{edition}/User/mcp.json` (see note below) | `servers` |
| `codex` | Not supported | `~/.codex/config.toml` | — |
| `opencode` | Not supported | `~/.config/opencode/opencode.json` | — |

**copilot-vscode MCP config path note:** agentkit detects the installed VS Code edition by checking for an existing `User/` directory under `{UserConfigDir}/Code`, then `{UserConfigDir}/Code - Insiders`, then `{UserConfigDir}/code-server`, in that order. If none is found, it defaults to `{UserConfigDir}/Code/User/mcp.json`. On macOS `UserConfigDir` is `~/Library/Application Support`; on Linux it is `~/.config`; on Windows it is `%APPDATA%`.

**copilot-cli extra fields note:** Each MCP server entry written for `copilot-cli` includes `"type": "local"` and `"tools": ["*"]` fields required by the Copilot CLI MCP format.

## Registry Format

Registries are GitHub repositories with a `registry.json` at the repo root. agentkit ships with three registries queried in priority order:

1. **Local override** (`local`) — active only when `AGENTKIT_REGISTRY_FILE` is set; queried first.
2. **agentkit-registry** — `https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json`
3. **gsd-core** — `https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json`
4. **builtin** — hardcoded packages that ship with agentkit regardless of external registries (currently includes `gsd`); queried last.

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

For `github-release` and `github-default-branch` methods, the `install` object uses `repo` and `path` instead of `package`:

```json
{
  "install": {
    "method": "github-release",
    "repo": "ejyle/agentkit",
    "path": "skills/aws",
    "multi_skill": false
  }
}
```

`multi_skill: true` instructs the installer to extract each immediate subdirectory of `path` as a separate skill.

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
