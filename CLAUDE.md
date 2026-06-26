# context-mode — MANDATORY routing rules

You have context-mode MCP tools available. These rules are NOT optional — they protect your context window from flooding. A single unrouted command can dump 56 KB into context and waste the entire session.

## BLOCKED commands — do NOT attempt these

### curl / wget — BLOCKED
Any Bash command containing `curl` or `wget` is intercepted and replaced with an error message. Do NOT retry.
Instead use:
- `ctx_fetch_and_index(url, source)` to fetch and index web pages
- `ctx_execute(language: "javascript", code: "const r = await fetch(...)")` to run HTTP calls in sandbox

### Inline HTTP — BLOCKED
Any Bash command containing `fetch('http`, `requests.get(`, `requests.post(`, `http.get(`, or `http.request(` is intercepted and replaced with an error message. Do NOT retry with Bash.
Instead use:
- `ctx_execute(language, code)` to run HTTP calls in sandbox — only stdout enters context

### WebFetch — BLOCKED
WebFetch calls are denied entirely. The URL is extracted and you are told to use `ctx_fetch_and_index` instead.
Instead use:
- `ctx_fetch_and_index(url, source)` then `ctx_search(queries)` to query the indexed content

## REDIRECTED tools — use sandbox equivalents

### Bash (>20 lines output)
Bash is ONLY for: `git`, `mkdir`, `rm`, `mv`, `cd`, `ls`, `npm install`, `pip install`, and other short-output commands.
For everything else, use:
- `ctx_batch_execute(commands, queries)` — run multiple commands + search in ONE call
- `ctx_execute(language: "shell", code: "...")` — run in sandbox, only stdout enters context

### Read (for analysis)
If you are reading a file to **Edit** it → Read is correct (Edit needs content in context).
If you are reading to **analyze, explore, or summarize** → use `ctx_execute_file(path, language, code)` instead. Only your printed summary enters context. The raw file content stays in the sandbox.

### Grep (large results)
Grep results can flood context. Use `ctx_execute(language: "shell", code: "grep ...")` to run searches in sandbox. Only your printed summary enters context.

## Tool selection hierarchy

1. **GATHER**: `ctx_batch_execute(commands, queries)` — Primary tool. Runs all commands, auto-indexes output, returns search results. ONE call replaces 30+ individual calls.
2. **FOLLOW-UP**: `ctx_search(queries: ["q1", "q2", ...])` — Query indexed content. Pass ALL questions as array in ONE call.
3. **PROCESSING**: `ctx_execute(language, code)` | `ctx_execute_file(path, language, code)` — Sandbox execution. Only stdout enters context.
4. **WEB**: `ctx_fetch_and_index(url, source)` then `ctx_search(queries)` — Fetch, chunk, index, query. Raw HTML never enters context.
5. **INDEX**: `ctx_index(content, source)` — Store content in FTS5 knowledge base for later search.

## Subagent routing

When spawning subagents (Agent/Task tool), the routing block is automatically injected into their prompt. Bash-type subagents are upgraded to general-purpose so they have access to MCP tools. You do NOT need to manually instruct subagents about context-mode.

## Output constraints

- Keep responses under 500 words.
- Write artifacts (code, configs, PRDs) to FILES — never return them as inline text. Return only: file path + 1-line description.
- When indexing content, use descriptive source labels so others can `ctx_search(source: "label")` later.

## ctx commands

| Command | Action |
|---------|--------|
| `ctx stats` | Call the `ctx_stats` MCP tool and display the full output verbatim |
| `ctx doctor` | Call the `ctx_doctor` MCP tool, run the returned shell command, display as checklist |
| `ctx upgrade` | Call the `ctx_upgrade` MCP tool, run the returned shell command, display as checklist |

<!-- GSD:project-start source:PROJECT.md -->
## Project

**agentkit**

agentkit is a production-grade Go CLI for discovering, installing, updating, and managing AI agent skills, MCP servers, and coding agents across all major AI coding assistants (Claude Code, GitHub Copilot, Codex, Gemini CLI, OpenCode). It ships as a single cross-platform binary with no runtime dependency, pulling from multiple registries (GitHub manifest-driven, mcpmarket.com, open-gsd/gsd-core, user-defined custom sources). Each skill it installs follows a progressive disclosure structure (SKILL.md + references/ + scripts/) so skills are modular, token-efficient, and context-aware. A companion background agent embedded in every skill detects new project infrastructure on skill use and updates a per-project `.agent-utils/config.json` — eliminating repeated discovery runs and reducing token cost on every future skill invocation.

**Core Value:** Zero-friction install of the right skills, agents, and MCP servers for any AI coding assistant — one command gets you from bare machine to fully instrumented dev environment, and every skill gets smarter about your project over time without wasting tokens.

### Constraints

- **Language**: Go — single binary, no runtime dependency, cross-compile to Windows/Linux/macOS
- **Install scope**: User scope default (~/.claude/, ~/.config/github-copilot/, ~/.codex/, ~/.gemini/) — no root/admin required
- **Skill token budget**: SKILL.md body must stay under 500 lines; heavy content goes in references/
- **Background agent**: Must not use more than ~1,000 tokens per invocation; no general reasoning, config diff only
- **MCP compatibility**: Must write MCP config in the format expected by each target assistant
<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->
## Technology Stack

## CLI Framework
| Framework | Stars | Verdict |
|-----------|-------|---------|
| spf13/cobra | 44K | **Use this.** Proven, composable, widely understood |
| urfave/cli | 22K | Older API; positional-arg style doesn't fit subcommand-heavy tools |
| alecthomas/kong | 7K | Struct-tag based, elegant for smaller tools, but less ecosystem support for shell completion and plugin extension |
| charmbracelet/bubbletea | 30K | TUI framework, not a CLI router. Use alongside Cobra for interactive confirm prompts |
## Cross-Platform Path Handling
- Linux: `$XDG_CONFIG_HOME` or `~/.config`
- macOS: `~/Library/Application Support`
- Windows: `%APPDATA%`
| Assistant | Skills Path | MCP Config Path |
|-----------|-------------|-----------------|
| Claude Code | `~/.claude/skills/` | `~/.claude/settings.json` (mcpServers key) |
| GitHub Copilot | `~/.copilot/skills/` | `~/.copilot/mcp-config.json` |
| OpenAI Codex | `~/.codex/skills/` | `~/.codex/config.toml` (mcp section) |
| Gemini CLI | `~/.gemini/skills/` | `~/.gemini/settings.json` (mcpServers key) |
| OpenCode | `~/.config/opencode/skills/` | `~/.config/opencode/opencode.json` (mcp key) |
## Registry Client
- Automatic retries with exponential backoff
- HTTP 429 / 503 handling with `Retry-After` header support (rate limiting)
- Wraps `net/http` — same interface, zero learning curve
- Each registry is a GitHub repo with a `registry.json` (or `manifest.json`) at the root
- The manifest is a flat JSON index: `{ "packages": [ { "name": "gsd", "version": "1.2.0", "description": "...", "type": "skill|mcp|agent", "source": "github.com/open-gsd/gsd-core", "path": "skills/gsd/" } ] }`
- agentkit fetches `https://raw.githubusercontent.com/{owner}/{repo}/{ref}/registry.json`
- No registry server needed — raw GitHub serves as the CDN
## MCP Config Formats (per assistant)
### Claude Code — `~/.claude/settings.json`
### GitHub Copilot CLI — `~/.copilot/mcp-config.json`
### Gemini CLI — `~/.gemini/settings.json`
### OpenAI Codex — `~/.codex/config.toml`
### OpenCode — `~/.config/opencode/opencode.json`
## Distribution / Release Tooling
- Cross-compile for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- GitHub Releases with checksums and SBOMs
- Homebrew tap auto-update (publish `agentkit` to a `homebrew-tap` repo)
- Scoop manifest for Windows (generated automatically)
- `.goreleaser.yaml` in repo root
- macOS/Linux: `brew install yourorg/tap/agentkit` or `curl | sh` script
- Windows: `scoop install agentkit` or direct binary download
- Any: `go install github.com/yourorg/agentkit@latest`
## Config File Format
- `.agent-utils/config.json` is already specified in PROJECT.md — match the ecosystem
- Claude Code (`settings.json`), Gemini CLI (`settings.json`), Copilot (`mcp-config.json`), OpenCode (`opencode.json`) all use JSON
- Go's `encoding/json` is stdlib — zero dependencies
- Grep/jq-friendly for power users
- Human-readable enough for the developer audience
- JSON — consistent with the rest of the ecosystem
- Schema: `{ "registries": [...], "defaults": { "target": "claude" }, "auth": { ... } }`
## Version Management
- Each registry manifest entry includes `version` (SemVer) and `sha256` of the installable artifact
- On install, agentkit writes an entry to `~/.config/agentkit/installed.json`:
- This is the "lock file" for user installs — analogous to `package-lock.json`
- `agentkit update` checks current version against registry and upgrades if newer
- No project-scope lock file in v1 (user scope only per PROJECT.md)
## Supporting Libraries
| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| `spf13/cobra` | v1.10.x | CLI command routing | De facto standard, 44K stars |
| `charmbracelet/bubbletea` | v1.x | Interactive TUI prompts | Best-in-class Go TUI, pairs with Cobra |
| `charmbracelet/lipgloss` | v1.x | Terminal styling | Same ecosystem as Bubbletea |
| `hashicorp/go-retryablehttp` | v0.7.x | Resilient HTTP for registry | Auto-retry, rate-limit handling, HashiCorp battle-tested |
| `BurntSushi/toml` | v1.x | Read/write Codex config.toml | Only TOML dep; high quality, stable API |
| stdlib `encoding/json` | — | All JSON config R/W | No dep needed |
| stdlib `os`, `path/filepath` | — | Cross-platform path handling | Sufficient, avoids adrg/xdg overhead |
- `spf13/viper` — overkill for a tool with one config file; adds 10+ transitive deps
- `adrg/xdg` — stdlib is sufficient for the paths needed
- `gopkg.in/yaml.v3` — no YAML needed in this stack
- `urfave/cli` — inferior subcommand ergonomics vs Cobra for this use case
## Recommended Stack Summary
## Sources
- [Cobra GitHub](https://github.com/spf13/cobra) — v1.10.2, Dec 2025
- [Bubbletea GitHub](https://github.com/charmbracelet/bubbletea)
- [hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp)
- [GoReleaser](https://goreleaser.com/)
- [adrg/xdg Go Packages](https://pkg.go.dev/github.com/adrg/xdg) — evaluated and rejected
- [Codex CLI MCP config](https://developers.openai.com/codex/mcp)
- [Gemini CLI MCP setup](https://geminicli.com/docs/tools/mcp-server/)
- [OpenCode MCP servers](https://opencode.ai/docs/mcp-servers/)
- [GitHub Copilot skills](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills)
- [Mise registry system](https://deepwiki.com/jdx/mise/6.1-registry-system)
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
