<!-- generated-by: gsd-doc-writer -->
# agentkit

> Zero-friction install of the right skills, agents, and MCP servers for any AI coding assistant.

**agentkit** is a production-grade CLI for discovering, installing, updating, and managing AI agent skills, MCP servers, and coding agents across all major AI coding assistants — Claude Code, GitHub Copilot, OpenAI Codex, Gemini CLI, and OpenCode.

Ships as a single cross-platform binary with no runtime dependency.

## Features

- Install skills, MCP servers, and agents from multiple registries
- Supports Claude Code, GitHub Copilot (CLI + VS Code), Codex, Gemini CLI, OpenCode
- Single binary — no runtime dependency, no Node/Python required
- `agentkit doctor` — environment health checks with pass/warn/fail output
- Registry-driven — pull from GitHub manifest registries or custom sources
- Bundle installs — install preset groups with `--bundle cloud`, `--bundle dev`, or `--bundle context`
- `--all` flag — install every available package from all registries in one command

## Install

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/ejyle/agentkit/main/scripts/install.sh | sh
```

Installs to `~/.local/bin/agentkit` (no root or sudo required).

**Windows (PowerShell — no admin required):**

```powershell
irm https://raw.githubusercontent.com/ejyle/agentkit/main/scripts/install.ps1 | iex
```

Installs to `%LOCALAPPDATA%\Programs\agentkit` and adds it to your user PATH automatically.

**Homebrew** *(coming soon)*:

```bash
brew install ejyle/agentkit/agentkit
```

**Scoop** *(coming soon)*:

```powershell
scoop bucket add ejyle https://github.com/ejyle/scoop-agentkit
scoop install agentkit
```

**Go:**

```bash
go install github.com/ejyle/agentkit@latest
```

## Quick start

```bash
agentkit --version
agentkit doctor
agentkit search playwright
agentkit install playwright
agentkit list
```

## Usage

### Install a package

```bash
# Install a skill or MCP server for Claude Code (default target)
agentkit install gsd

# Install for a different assistant
agentkit install playwright --target copilot-cli

# Install a preset bundle
agentkit install --bundle dev

# Install everything available
agentkit install --all
```

Supported targets: `claude`, `copilot-cli`, `copilot-vscode`, `codex`, `gemini`, `opencode`, `pi`

### Search the registry

```bash
agentkit search playwright
```

### List installed packages

```bash
agentkit list
agentkit list --target gemini
```

### Uninstall a package

```bash
agentkit uninstall playwright
agentkit uninstall --all --target claude
```

### Update installed packages

```bash
agentkit update
```

### Run health checks

```bash
agentkit doctor
```

## Development

All active development happens on the `develop` branch.

```bash
git clone https://github.com/ejyle/agentkit
cd agentkit
git checkout develop
go build ./...
go test ./...
```

## License

MIT
