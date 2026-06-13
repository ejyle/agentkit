# agentkit

> Zero-friction install of the right skills, agents, and MCP servers for any AI coding assistant.

**agentkit** is a production-grade CLI for discovering, installing, updating, and managing AI agent skills, MCP servers, and coding agents across all major AI coding assistants — Claude Code, GitHub Copilot, OpenAI Codex, Gemini CLI, and OpenCode.

Ships as a single cross-platform binary with no runtime dependency.

## Features

- Install skills, MCP servers, and agents from multiple registries
- Supports Claude Code, GitHub Copilot, Codex, Gemini CLI, OpenCode
- Single binary — no runtime dependency, no Node/Python required
- `agentkit doctor` — environment health checks with pass/warn/fail output
- Registry-driven — pull from GitHub manifest registries or custom sources

## Install

**macOS / Linux (curl | sh):**
```bash
curl -fsSL https://raw.githubusercontent.com/ejyle/agentkit/dev/scripts/install.sh | sh
```

**Homebrew** *(coming soon)*:
```bash
brew install ejyle/agentkit/agentkit
```

**Go:**
```bash
go install github.com/ejyle/agentkit@latest
```

## Quick start

```bash
agentkit --version
agentkit doctor
agentkit install gsd
agentkit update
```

## Development

All active development happens on the `dev` branch.

```bash
git clone https://github.com/ejyle/agentkit
git checkout dev
go build ./...
```

## License

MIT
