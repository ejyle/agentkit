# Requirements: agentkit

_Version: 0.1.0 (v1) | Generated: 2026-06-08_

---

## v1 Requirements (0.1.0)

### CLI Commands

- [ ] **CLI-01**: User can install a skill, agent, or MCP server by name (`agentkit install <name>`)
- [ ] **CLI-02**: User can specify the target coding assistant at install time (`--target <claude|copilot|codex|gemini|opencode|pi>`)
- [ ] **CLI-03**: User can install the full GSD workflow suite in one command (`agentkit install gsd`)
- [ ] **CLI-04**: User can bulk-install a preset bundle (`agentkit install --bundle <cloud|gsd|context|dev>`)
- [ ] **CLI-05**: User can cleanly uninstall a package with no leftover artifacts (`agentkit uninstall <name>`)
- [ ] **CLI-06**: User can search all registries and get ranked results (`agentkit search <query>`)
- [ ] **CLI-07**: User can update one or all installed packages to latest versions (`agentkit update [name]`)
- [ ] **CLI-08**: User can list all installed packages with version, source registry, and target assistant (`agentkit list`)
- [ ] **CLI-09**: User can add or remove registry sources (`agentkit registry add <url>` / `agentkit registry remove <url>`)
- [ ] **CLI-10**: CLI ships as a single binary with no runtime dependency, runs on Windows/Linux/macOS without root or sudo

### Registry System

- [ ] **REG-01**: GitHub manifest-driven registry — any GitHub repo with a `registry.json` manifest is a valid registry source
- [ ] **REG-02**: open-gsd/gsd-core registry integrated as a built-in default source
- [ ] **REG-03**: mcpmarket.com API registry integrated with graceful degradation (treated as optional; works offline if unavailable)
- [ ] **REG-04**: User can add custom registry sources via `agentkit registry add` (Homebrew tap model)
- [ ] **REG-05**: agentkit-registry GitHub repo ships as the default curated registry (lists 9 bundled skills, GSD suite, well-known MCP servers)
- [ ] **REG-06**: Registry manifests cached locally with ETag validation; stale cache used with warning when network unavailable

### Target Assistant Adapters

- [ ] **AST-01**: Claude Code adapter fully implemented — skills to `~/.claude/skills/`, agents to `~/.claude/agents/`, MCP config merged into `~/.claude/settings.json` (or runtime-detected path)
- [ ] **AST-02**: GitHub Copilot CLI adapter — runtime path detection, MCP config written to correct location (CLI vs VS Code divergence handled as separate sub-adapters)
- [ ] **AST-03**: OpenAI Codex adapter — TOML config at `~/.codex/config.toml`; MCP key names verified at build time
- [ ] **AST-04**: Gemini CLI adapter — `~/.gemini/settings.json`; SKILL.md mapped to GEMINI.md conventions
- [ ] **AST-05**: OpenCode adapter — `~/.config/opencode/opencode.json`; `mcp` key (not `mcpServers`), `type` field required
- [ ] **AST-06**: Pi (pi.dev) adapter — install path and skill mechanism researched and implemented; degraded gracefully if pi.dev has no CLI-level skill system

### MCP Install

- [ ] **MCP-01**: npx install adapter (`npx -y <package>`) — handles node-based MCP servers
- [ ] **MCP-02**: pip install adapter (`pip install <package>`) — handles Python MCP servers
- [ ] **MCP-03**: Binary download adapter — fetches pre-built binary from release URL, places in user PATH
- [ ] **MCP-04**: Docker adapter — pulls image, generates run command, writes to MCP config as `docker run` entry
- [ ] **MCP-05**: Provider manifest can override install method with custom steps
- [ ] **MCP-06**: Post-install verify step re-reads written MCP config to confirm it parses correctly; install fails loudly on invalid config
- [ ] **MCP-07**: Each adapter detects existing MCP config path at runtime (never hardcoded); merge-writes (read → merge → atomic write)

### Skill Structure

- [ ] **SKL-01**: agentkit validates installed skills follow agentskills.io spec: `SKILL.md` + optional `references/` + optional `scripts/`
- [ ] **SKL-02**: SKILL.md line count checked at install time; warning issued if >500 lines (install proceeds, validation is non-blocking)
- [ ] **SKL-03**: Multi-domain skills organized by variant under `references/` (e.g. `references/aws.md`, `references/gcp.md`) — enforced by registry manifest schema

### Bundled Skills (user scope, authored by agentkit project)

**Cloud bundle:**
- [ ] **BND-01**: AWS skill — EC2, S3, IAM, ECS/EKS management; SKILL.md entry + `references/ec2.md`, `references/s3.md`, `references/iam.md`; includes install detection script
- [ ] **BND-02**: GCP skill — Compute Engine, GKE, Cloud Run, IAM via gcloud; SKILL.md + domain references
- [ ] **BND-03**: Azure skill — VMs, AKS, App Service via az CLI; SKILL.md + domain references

**Dev bundle:**
- [ ] **BND-04**: Playwright skill — browser automation and E2E testing; SKILL.md + MCP server entry (Microsoft Playwright MCP)
- [ ] **BND-05**: GitHub skill — PR management, issues, Actions CI/CD via gh CLI; SKILL.md + references
- [ ] **BND-06**: CI/CD skill — GitHub Actions, build pipelines, deploy workflows; SKILL.md + references

**Context bundle:**
- [ ] **BND-07**: context-mode skill — sandbox routing for large outputs, context window protection
- [ ] **BND-08**: RTK skill — token-optimized CLI proxy for dev operations
- [ ] **BND-09**: Serena skill — LSP-powered code intelligence and symbol navigation

---

## v2 Requirements (0.2.0) — Deferred

- Per-project facts system: `.agent-utils/config.json` captures any project-specific facts discovered during skill use (infra details, deployed URLs, config values, custom settings)
- Background config agent: XML-prompted, 3-tool-call max, token-adaptive (auto-write cheap facts inline, prompt user for expensive ones), shared across all skills
- Drift detection: agent detects when cached project facts have changed and prompts update
- `agentkit config refresh` command: re-run discovery for current project
- OpenCode adapter (if 0.1.0 adapter is stub): full implementation with verified config paths

---

## Out of Scope

- GUI or web dashboard — CLI only
- Project-scope install (`./.claude/`) — user scope only; project scope is v3+
- Skill authoring wizard via agentkit CLI — v3+
- Private registry authentication beyond a token in user config — v2
- Windows package manager integration (winget/choco) — distribution only, post-GA
- Automatic skill update on schedule (cron-style) — v2

---

## Traceability

_Each requirement maps to exactly one phase._

| Requirement ID | Phase | Status |
|----------------|-------|--------|
| CLI-01 | Phase 1 | Pending |
| CLI-02 | Phase 1 | Pending |
| CLI-03 | Phase 3 | Pending |
| CLI-04 | Phase 3 | Pending |
| CLI-05 | Phase 1 | Pending |
| CLI-06 | Phase 1 | Pending |
| CLI-07 | Phase 1 | Pending |
| CLI-08 | Phase 1 | Pending |
| CLI-09 | Phase 1 | Pending |
| CLI-10 | Phase 4 | Pending |
| REG-01 | Phase 1 | Pending |
| REG-02 | Phase 1 | Pending |
| REG-03 | Phase 2 | Pending |
| REG-04 | Phase 2 | Pending |
| REG-05 | Phase 1 | Pending |
| REG-06 | Phase 1 | Pending |
| AST-01 | Phase 1 | Pending |
| AST-02 | Phase 2 | Pending |
| AST-03 | Phase 2 | Pending |
| AST-04 | Phase 2 | Pending |
| AST-05 | Phase 2 | Pending |
| AST-06 | Phase 2 | Pending |
| MCP-01 | Phase 1 | Pending |
| MCP-02 | Phase 2 | Pending |
| MCP-03 | Phase 1 | Pending |
| MCP-04 | Phase 2 | Pending |
| MCP-05 | Phase 1 | Pending |
| MCP-06 | Phase 1 | Pending |
| MCP-07 | Phase 1 | Pending |
| SKL-01 | Phase 1 | Pending |
| SKL-02 | Phase 1 | Pending |
| SKL-03 | Phase 1 | Pending |
| BND-01 | Phase 3 | Pending |
| BND-02 | Phase 3 | Pending |
| BND-03 | Phase 3 | Pending |
| BND-04 | Phase 3 | Pending |
| BND-05 | Phase 3 | Pending |
| BND-06 | Phase 3 | Pending |
| BND-07 | Phase 3 | Pending |
| BND-08 | Phase 3 | Pending |
| BND-09 | Phase 3 | Pending |
