# agentkit

## What This Is

agentkit is a production-grade Go CLI for discovering, installing, updating, and managing AI agent skills, MCP servers, and coding agents across all major AI coding assistants (Claude Code, GitHub Copilot, Codex, Gemini CLI, OpenCode). It ships as a single cross-platform binary with no runtime dependency, pulling from multiple registries (GitHub manifest-driven, mcpmarket.com, open-gsd/gsd-core, user-defined custom sources). Each skill it installs follows a progressive disclosure structure (SKILL.md + references/ + scripts/) so skills are modular, token-efficient, and context-aware. A companion background agent embedded in every skill detects new project infrastructure on skill use and updates a per-project `.agent-utils/config.json` — eliminating repeated discovery runs and reducing token cost on every future skill invocation.

## Core Value

Zero-friction install of the right skills, agents, and MCP servers for any AI coding assistant — one command gets you from bare machine to fully instrumented dev environment, and every skill gets smarter about your project over time without wasting tokens.

## Requirements

### Validated

- `agentkit install <name>` installs MCP server from registry, writes config atomically — Validated in Phase 01: foundation
- `agentkit list` shows installed packages with version/type/target/registry columns — Validated in Phase 01: foundation
- `agentkit search <query>` returns ranked registry results — Validated in Phase 01: foundation
- `agentkit uninstall <name>` removes MCP config and installed.json entry without affecting other keys — Validated in Phase 01: foundation
- `agentkit update [name]` checks version and reports up-to-date or upgrades — Validated in Phase 01: foundation
- `--target claude|copilot|codex|gemini|opencode` flag validated and enforced — Validated in Phase 01: foundation

### Active

**CLI Core**
- [ ] `agentkit install <name>` installs a skill, agent, or MCP by name from any registered registry
- [ ] `agentkit install gsd` installs the full GSD suite (skills + agents + config) in one command
- [ ] `agentkit install --bundle <bundle>` bulk-installs a named bundle (e.g. `--bundle cloud`, `--bundle ci`)
- [ ] `agentkit search <query>` searches all registries and returns ranked results with source labels
- [ ] `agentkit update [name]` updates one or all installed packages to latest registry versions
- [ ] `agentkit list` shows all installed skills/agents/MCP with version, source, and install scope
- [ ] Install targets user scope by default; `--target <claude|copilot|codex|gemini|opencode>` installs for a specific assistant

**Registry System**
- [ ] GitHub manifest-driven registry: each source repo has a manifest listing available packages
- [ ] mcpmarket.com API registry for skills and MCP
- [ ] open-gsd/gsd-core registry for GSD skills and agents
- [ ] Custom registry support: user can add/remove registry sources (like npm source config)
- [ ] Registry resolution: multiple registries, priority order, deduplication by package name

**MCP Install**
- [ ] Built-in adapters: npx, pip, binary download, Docker
- [ ] Manifest override: provider manifest can specify custom install steps
- [ ] MCP config written to the correct location per target assistant after install

**Skill Structure (built skills produced by agentkit)**
- [ ] Every produced skill follows: `SKILL.md` (entry, <500 lines) + `references/` (domain docs, loaded on demand) + `scripts/` (bundled executables, run without loading into context)
- [ ] Multi-domain skills organized by variant (e.g. cloud/references/aws.md, cloud/references/gcp.md)
- [ ] SKILL.md frontmatter includes name, description, version, source registry

**Project Config System**
- [ ] Skills write discovered infra facts to `.agent-utils/config.json` on first use in a project
- [ ] Obvious facts (region, account ID, cluster name) auto-saved silently with a one-liner notify
- [ ] Ambiguous facts (custom endpoint, role ARN) confirmed before saving
- [ ] Skills read `.agent-utils/config.json` on load to skip re-discovery
- [ ] Config entries have timestamps and are refreshed when drift is detected

**Background Config Agent**
- [ ] Each skill bundles a minimal background agent that triggers at skill end when new discoveries are made
- [ ] Agent prompts user to confirm update if token cost > threshold, runs inline if minimal
- [ ] Agent is XML-prompted, focused only on config diff detection and write — no general reasoning
- [ ] Agent is shared across all skills (single implementation, embedded by reference)

**Initial Bundled Skills (user scope)**
- [ ] AWS skill: EC2, S3, IAM, ECS/EKS management via AWS CLI/SDK + project config integration
- [ ] GCP skill: Compute Engine, GKE, Cloud Run, IAM via gcloud/SDK + project config integration
- [ ] Azure skill: VMs, AKS, App Service via az CLI/SDK + project config integration
- [ ] Playwright skill + Playwright MCP (Microsoft): browser automation, E2E testing
- [ ] context-mode skill: route large outputs to sandbox, protect context window
- [ ] RTK skill: token-optimized CLI proxy for dev operations
- [ ] Serena skill: LSP-powered code intelligence and symbol navigation
- [ ] GitHub skill: PR management, issues, Actions CI/CD via gh CLI
- [ ] CI/CD skill: GitHub Actions, build pipelines, deploy workflows

### Out of Scope

- Project-scope install (per-project .claude/) — deferred to v2; user scope covers the majority of use cases
- GUI / web dashboard — CLI only for v1
- Skill authoring wizard (creating new skills via agentkit) — v2
- Private registry authentication (beyond token in config) — v2
- Windows package manager integration (winget/choco) — v2 distribution only

## Context

- Working directory: `/Users/nithin/Ejyle/coding-agent-utils`
- User is building this for Ejyle; Go is already the preferred language for CLIs in this environment
- GSD is already installed at user scope and is the primary workflow engine — agentkit must integrate cleanly with GSD's install conventions
- Existing skill structure examples in `~/.claude/skills/` show the reference/ and scripts/ patterns already in use (book-distiller, database-schema-designer, playwright-cli)
- The background agent pattern is borrowed from GSD's existing background agent approach (XML prompting, focused scope, minimal token use)
- Sites for registry research: mcpmarket.com/tools/skills, Capafy AI, The Agent Skills Directory
- GSD source: https://github.com/open-gsd/gsd-core

## Constraints

- **Language**: Go — single binary, no runtime dependency, cross-compile to Windows/Linux/macOS
- **Install scope**: User scope default (~/.claude/, ~/.config/github-copilot/, ~/.codex/, ~/.gemini/) — no root/admin required
- **Skill token budget**: SKILL.md body must stay under 500 lines; heavy content goes in references/
- **Background agent**: Must not use more than ~1,000 tokens per invocation; no general reasoning, config diff only
- **MCP compatibility**: Must write MCP config in the format expected by each target assistant

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go over Python for CLI | Single binary, no runtime, fast startup, easy cross-compile — better for distribution | — Pending |
| User scope only (v1) | Covers majority use case; project scope adds complexity for edge cases | — Pending |
| `.agent-utils/config.json` for project config | Dedicated namespace, gitignore-able, not mixed into .claude/ | — Pending |
| Background agent inline vs prompt | Token-adaptive: auto when cheap, user-confirmed when expensive | — Pending |
| Manifest-driven registries | Extensible without CLI changes; providers own their package listings | — Pending |

---
*Last updated: 2026-06-08 after initialization*

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state
