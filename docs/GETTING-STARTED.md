<!-- generated-by: gsd-doc-writer -->
# Getting Started with agentkit

This guide walks you from installation to your first working install of a skill or MCP server.

## Prerequisites

**To use the pre-built binary (recommended):** No runtime dependency. agentkit ships as a single static binary — no Go, Node, or Python installation required.

**To build from source:**

| Tool | Minimum Version |
|------|----------------|
| Go   | `>= 1.26.3`    |
| Git  | any recent version |

Check your Go version:

```bash
go version
```

## Installation Steps

### Option 1: macOS / Linux (curl installer)

```bash
curl -fsSL https://raw.githubusercontent.com/ejyle/agentkit/main/scripts/install.sh | sh
```

Installs to `~/.local/bin/agentkit`. No root or sudo required.

### Option 2: Windows (PowerShell — no admin required)

```powershell
irm https://raw.githubusercontent.com/ejyle/agentkit/main/scripts/install.ps1 | iex
```

Installs to `%LOCALAPPDATA%\Programs\agentkit` and adds it to your user PATH automatically.

### Option 3: Homebrew (coming soon)

```bash
brew install ejyle/agentkit/agentkit
```

### Option 4: Scoop — Windows (coming soon)

```powershell
scoop bucket add ejyle https://github.com/ejyle/scoop-agentkit
scoop install agentkit
```

### Option 5: Go install

```bash
go install github.com/ejyle/agentkit@latest
```

Installs the binary to `$(go env GOPATH)/bin/agentkit`.

### Option 6: Build from source

```bash
git clone https://github.com/ejyle/agentkit
cd agentkit
git checkout develop
go build -o agentkit .
```

Move the resulting binary to a directory on your PATH (e.g. `~/.local/bin/`).

## First Run

Verify the install and run the environment health check:

```bash
agentkit --version
agentkit doctor
```

`agentkit doctor` reports which target assistants are configured on your machine and whether their config paths are writable.

Install your first package:

```bash
agentkit search playwright
agentkit install playwright
```

By default, packages are installed for Claude Code. To target a different assistant:

```bash
agentkit install playwright --target gemini
```

Supported targets: `claude`, `copilot-cli`, `copilot-vscode`, `codex`, `gemini`, `opencode`, `pi`

## Common Setup Issues

**`agentkit: command not found` after curl install (macOS/Linux)**

The installer writes to `~/.local/bin`. If that directory is not on your PATH, add it:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc   # or ~/.zshrc
source ~/.bashrc
```

Then verify:

```bash
agentkit --version
```

**`go install` binary not found**

Ensure `$(go env GOPATH)/bin` is on your PATH:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

**`go build` fails with version error**

The module requires Go `>= 1.26.3`. Upgrade Go if your version is older:

```bash
go version    # must be >= 1.26.3
```

**Doctor reports a target assistant is not configured**

agentkit installs to the assistant's standard config directory. The target assistant must be installed on your machine first. For example, Claude Code must be installed before `--target claude` works. Run `agentkit doctor` to see which targets are available.

## Next Steps

- **[ARCHITECTURE.md](ARCHITECTURE.md)** — How agentkit is structured internally (registry, installer, adapter layers)
- **[CONFIGURATION.md](CONFIGURATION.md)** — Environment variables, CLI flags, and state file locations
- **[DEVELOPMENT.md](DEVELOPMENT.md)** — Local development setup, build commands, and PR process
- **[TESTING.md](TESTING.md)** — Running the test suite and writing new tests
