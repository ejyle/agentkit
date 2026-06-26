<!-- generated-by: gsd-doc-writer -->
# Development

This guide covers local setup, build commands, code style, branching, and the PR process for contributors to agentkit.

## Local Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/ejyle/agentkit.git
   cd agentkit
   ```

2. **Install Go**

   agentkit requires Go 1.26.3 or later. Verify your installation:

   ```bash
   go version
   ```

3. **Install dependencies**

   ```bash
   go mod tidy
   ```

4. **Build and run locally**

   ```bash
   go build -o agentkit .
   ./agentkit --version
   ```

   To run without building a binary:

   ```bash
   go run . --version
   ```

## Build Commands

| Command | Description |
|---------|-------------|
| `go build ./...` | Compile all packages (checks for errors) |
| `go build -o agentkit .` | Build the CLI binary into the current directory |
| `go run . <args>` | Run the CLI directly without a binary |
| `go test ./...` | Run the full test suite |
| `go test -v ./...` | Run tests with verbose output |
| `go test ./internal/...` | Run only internal package tests |
| `go mod tidy` | Sync `go.mod` and `go.sum` with actual imports |
| `go vet ./...` | Run the Go static analysis tool |
| `go fmt ./...` | Format all Go source files |

### Snapshot Build (GoReleaser)

To produce cross-platform binaries locally without publishing:

```bash
goreleaser release --snapshot --clean --skip=publish,sign,homebrew
```

This requires [GoReleaser](https://goreleaser.com/) installed locally. Output lands in `dist/`.

## Code Style

agentkit uses standard Go tooling for formatting and static analysis. There is no separate linter configuration file — the project relies on the Go standard tools.

- **Formatting**: `go fmt ./...` — always run before committing. CI will flag unformatted files.
- **Static analysis**: `go vet ./...` — run to catch common mistakes before opening a PR.
- **Naming**: Follow standard Go conventions — exported identifiers use `PascalCase`, unexported use `camelCase`, package names are lowercase single words.
- **Error handling**: Return typed errors at package boundaries; do not use `panic` for recoverable conditions.
- **Packages**: All non-entry-point code lives under `internal/` and is not importable by external consumers.

An `.editorconfig` file is not present. Use your editor's Go plugin (e.g., `gopls`) to enforce formatting on save.

## Branch Conventions

No branch naming convention is formally documented. The following pattern is used in practice:

| Branch | Purpose |
|--------|---------|
| `main` | Stable — CI runs a GoReleaser snapshot build on every push |
| `develop` | Integration branch for active development |
| `feat/<description>` | New features |
| `fix/<description>` | Bug fixes |
| `docs/<description>` | Documentation-only changes |
| `chore/<description>` | Dependency updates, tooling changes |

Target PRs at `develop` unless the change is an urgent hotfix that must go directly to `main`.

## PR Process

No formal pull request template is present. Follow these guidelines when opening a PR:

- **Scope**: Keep each PR focused on a single concern. Avoid combining feature work with refactors.
- **Tests**: Add or update `*_test.go` files for any changed behaviour. Run `go test ./...` locally before pushing.
- **Formatting**: Ensure `go fmt ./...` and `go vet ./...` pass with no output.
- **Commit messages**: Use conventional commit prefixes — `feat:`, `fix:`, `docs:`, `chore:`, `test:`. The GoReleaser changelog filters out `docs:`, `test:`, and `chore:` entries from release notes automatically.
- **Description**: Summarise *what* changed and *why*. Reference any related issue numbers.
- **CI**: The `release.yml` workflow runs on tag pushes (full release) and `main` pushes (snapshot). Verify the snapshot job passes after your PR merges.

## Project Layout Reference

```
agentkit/
├── main.go                  # Entry point — delegates to cmd.Execute()
├── cmd/                     # Cobra command definitions (install, uninstall, list, search, update, doctor)
├── internal/
│   ├── adapter/             # Per-assistant config writers (Claude, Copilot, Codex, Gemini, OpenCode, Pi)
│   ├── bundle/              # Preset package groups (--bundle flag)
│   ├── config/              # Config store and XDG/platform path resolution
│   ├── domain/              # Core domain types
│   ├── fileutil/            # Atomic file write helpers
│   ├── installer/           # MCP install strategies (npx, uvx, binary, Docker, GitHub release)
│   ├── registry/            # Registry client (GitHub manifest, builtin, local override)
│   ├── service/             # Orchestration layer (install, uninstall, search, update flows)
│   ├── skill/               # Skill validation logic
│   ├── ui/                  # Terminal UI helpers (spinner, table, TTY detection)
│   └── version/             # Build-time version injection
├── scripts/
│   ├── install.sh           # macOS/Linux install script
│   └── install.ps1          # Windows PowerShell install script
├── testdata/
│   └── registry.json        # Test fixture registry
├── .goreleaser.yaml         # Cross-platform release config
└── .github/workflows/
    └── release.yml          # CI: tag → full release; main push → snapshot
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for component diagrams and data flow details.
