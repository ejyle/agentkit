# Features Research: agentkit

**Domain:** AI agent skill / MCP / coding assistant management CLI
**Researched:** 2026-06-08
**Overall confidence:** HIGH (multiple authoritative sources: agentskills.io spec, GSD docs, mise docs, skills.sh, mcpmarket.com)

---

## Existing Tools Landscape

### Skill Distribution Hubs

| Tool | Model | Install UX | Notes |
|------|-------|------------|-------|
| **skills.sh** | Open, free | `skills install <name>` single command | Primary hub for agentskills.io spec; supports 10+ assistants |
| **agentskills.io / agentskills/agentskills** | Open spec (Anthropic-originated) | GitHub clone or hub CLI | Spec defines SKILL.md + frontmatter + references/ + scripts/; adopted by Claude, Codex, Gemini, Copilot |
| **MicrosoftDocs/Agent-Skills** | Open, curated | Manual copy or VS Code agent integration | Azure/Microsoft-specific skills, four-stage progressive disclosure (advertise → load → read resources → assets) |
| **mcpmarket.com** | Open listing, community submit | Browse by category/tag, one-click install for Cline clients; store feature for Claude/Cursor | Hosts npx, pip, binary, Docker MCP servers |
| **mcp-marketplace.io** | Open | One-click | Separate from mcpmarket.com |
| **open-gsd/gsd-core** | Open | Installer script; prompts runtime + global vs local | Converts install artefacts per runtime (TOML for Codex, JSON for OpenCode, slash-command rewrite for Gemini) |
| **Capafy AI** | Closed-source, monetised | One-click in chat; no install required | Creators set price; skill code never exposed; pay-per-use / subscribe / one-time models |
| **lobehub / gsd-2** | Open listing | Registry discovery | Lists GSD-2 agent framework skills |

### Package Managers That Define Developer Expectations

| Tool | Patterns that stick |
|------|-------------------|
| **brew** | `install`, `upgrade`, `list`, `info`, `doctor`, `uninstall`; `brew bundle` for declarative Brewfile; `brew bundle dump` to snapshot current state |
| **npm** | `install`, `update`, `list`, `search`, `info`, `--save-exact` for version pinning, `--offline` / `--prefer-offline` flags, scoped packages (`@org/pkg`) |
| **mise** | Per-project `.mise.toml` activated automatically on `cd`; cascading config (global → work → project); `mise use --global` vs `mise use` (local); `.tool-versions` asdf compat |
| **cargo** | `cargo install`, `--locked` for reproducible installs, `--offline`, Cargo.toml for project pins |

Key developer UX expectations drawn from these tools:
- **Verb-noun consistency**: install / update / list / search / uninstall / info
- **Scope flags**: user-global vs project-local (like npm's `-g`)
- **Version pinning**: exact version or semver range
- **Dry-run / --check**: see what would change before changing
- **Offline mode**: operate from local cache when no network
- **Declarative snapshot**: dump current state to a lockfile / config

---

## Table Stakes Features

Features users expect. Missing = product feels incomplete or users abandon for manual installs.

| Feature | Why Expected | Complexity | Dependencies |
|---------|--------------|------------|--------------|
| `install <name>` single command | All package managers set this bar; manual copy is the current pain point | Low | Registry resolution |
| `list` — show installed skills/MCPs with version and source | brew list / npm list; developers need an audit view | Low | Local manifest |
| `update [name]` — update one or all | brew upgrade pattern; devs expect to stay current with no friction | Medium | Registry fetch + diff |
| `search <query>` — ranked results across registries | npm search / cargo search; necessary to discover the right package | Medium | Registry index + fuzzy match |
| `uninstall <name>` — clean removal | All package managers; partial install artifacts are a common complaint in manual installs | Low | Local manifest |
| `info <name>` — metadata, description, version, source | npm info / brew info; needed before committing to install | Low | Registry fetch |
| Multi-registry support with priority ordering | Custom sources expected (like npm registries); enterprise teams need internal registries | Medium | Registry abstraction |
| Correct config file written per target assistant | Each assistant (Claude, Copilot, Codex, Gemini) has a different config path/format; wrong format = skill doesn't load | Medium | Adapter per assistant |
| User-scope default (no root/sudo) | Developers refuse tools requiring admin; mise and npm -g succeed here | Low | Path resolution |
| Single binary, no runtime dependency | Go binary expectation; Python CLIs with venv deps are a known friction source | Low (Go gives this for free) | Build/release pipeline |
| Progress output with clear success/failure states | npm install spinners, brew install progress bars; silent tools feel broken | Low | Output formatting |
| SKILL.md + references/ + scripts/ structure honoured on install | The open standard (agentskills.io) is already adopted cross-platform; deviating breaks portability | Low | Manifest schema |

---

## Differentiating Features

Features that set agentkit apart from manual installs and from skills.sh.

| Feature | Value Proposition | Complexity | Dependencies |
|---------|-------------------|------------|--------------|
| `--bundle <name>` — install curated preset collections | brew bundle pattern; "cloud bundle" installs AWS + GCP + Azure + GitHub skills atomically; saves new-machine setup time | Medium | Bundle manifest format, registry resolution |
| Cross-assistant install with `--target` flag | No other tool normalises across Claude/Copilot/Codex/Gemini/OpenCode in one CLI; GSD's installer is interactive not scriptable | High | Per-assistant adapter layer |
| Background config agent — auto-detects project infra and persists to `.agent-utils/config.json` | Token cost reduction on every future skill invocation; no other skill manager does per-project state; GSD does it at framework level but not per-skill | High | XML-prompted focused agent, config schema |
| Per-project config at `.agent-utils/config.json` with timestamp + drift detection | mise's per-project pattern applied to AI config; facts discovered once, refreshed on drift | Medium | Background agent, file I/O |
| Progressive disclosure enforced by structure (SKILL.md < 500 lines, heavy content in references/) | Reduces context rot (40-60% token savings per session per research); publisher tooling that validates this budget is novel | Low (validation), Medium (publisher guide) | Manifest schema validation |
| Three-tier loading: metadata catalog → full SKILL.md → references on demand | Microsoft Agent Framework four-stage model; reduces cold-start token burn | Medium | Skill loader integration per assistant |
| `install gsd` — installs full GSD suite in one command | GSD is the primary workflow engine in this environment; atomic install of skills + agents + config removes manual multi-step setup | Medium | GSD registry adapter |
| Custom registry add/remove (like `npm config set registry`) | Enterprise / private registry support without CLI changes; providers own their manifest | Medium | Registry config storage |
| `bundle dump` — snapshot current installed state to a bundle file | brew bundle dump pattern; onboard new teammates or restore after wipe | Medium | Local manifest, bundle format |
| `doctor` — diagnose install health (missing configs, wrong paths, version drift) | brew doctor is beloved; catches silent breakage | Medium | Per-assistant config validators |
| Token-adaptive background agent (auto-run if < threshold, confirm if expensive) | Prevents surprise token spend; no existing skill manager has this UX | Medium | Token estimator, agent runner |
| Manifest-driven registry (providers publish their own manifest, no CLI PR needed) | Extensible without gating on CLI release cycle; comparable to tap model in brew | Low (protocol), Medium (tooling for providers) | Registry schema, docs |

---

## Anti-Features (deliberately skip)

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| GUI / web dashboard | Adds surface area, slows v1; target users are CLI-native developers | CLI only; consider web listing page post-v1 |
| Skill authoring wizard (`agentkit new-skill`) | Creative flow in v1 is install/manage, not author; authoring needs user research to get right | Document the SKILL.md structure; defer wizard to v2 |
| Per-project install scope in v1 (`.claude/` per-project) | Majority use case is user scope; project scope adds config-conflict edge cases and complicates uninstall | User scope in v1; add `--scope project` flag in v2 |
| Private registry authentication beyond API token in config | OAuth flows, SSO, keychain integration are out-of-scope complexity for v1 | Token-in-config is sufficient; full auth in v2 |
| Skill execution / running skills | agentkit manages skills; AI assistants run them — conflating the two creates a competing-with-users problem | Stay in install/config layer |
| Paid/monetised skill marketplace (Capafy model) | Adds legal, billing, fraud surface area; the open-standard ecosystem is where traction is | Integrate with open registries; leave monetisation to Capafy |
| Windows package manager integration (winget/choco) in v1 | Cross-compile binary is enough for Windows; winget tap requires Microsoft review process | Provide downloadable binary; winget/choco in v2 distribution |
| Dependency resolution between skills | npm-style dep trees for skills are premature; skills are independent by design in the open standard | Flat install; document skill prerequisites in SKILL.md |
| Automatic skill invocation / injection into agent context | That's the AI assistant's job; agentkit owns the filesystem layer only | Write to correct path; let the assistant pick up the skill |

---

## GSD Integration Notes

GSD (open-gsd/gsd-core) is the primary workflow engine in this environment. Key integration constraints:

- **GSD install convention**: GSD's installer is interactive (prompts runtime + scope), not scriptable. `agentkit install gsd` must replicate this non-interactively with sensible defaults (`--target claude --scope user`).
- **Runtime transformation**: GSD's installer transforms artefacts per runtime (TOML for Codex, JSON for OpenCode, slash-command syntax rewrite for Gemini). agentkit's adapter layer must do the same.
- **GSD registry as first-class source**: `open-gsd/gsd-core` should be a built-in registry alongside mcpmarket.com and the agentskills.io manifest; not a special-case.
- **GSD slash-command naming**: Claude Code uses `/gsd-*` (hyphen form); Gemini uses `/gsd:*` (colon form). The adapter must handle this rewrite.
- **`.planning/` and `.gsd/` directories**: GSD stores project state in these; `.agent-utils/config.json` should not collide. Keep namespaces separate.
- **Background agent pattern**: GSD's background agent uses XML prompting with focused scope and minimal token footprint. agentkit's per-skill background config agent should follow the same XML-prompted, config-diff-only pattern — not import GSD as a dependency, but mirror the pattern.

---

## Bundle / Collection Patterns

Drawn from brew bundle and mise patterns:

### What a Bundle Is

A bundle is a named, versioned list of skills/MCPs to install atomically. Comparable to a Brewfile or a mise.toml `[tools]` section.

### Bundle Sources (priority order)

1. **Built-in bundles** shipped with agentkit binary (cloud, ci, playwright, gsd) — always available offline
2. **Registry-provided bundles** declared in a registry manifest — providers can publish curated collections
3. **User-defined bundles** stored in `~/.config/agentkit/bundles/` — personal presets
4. **Project bundles** in `.agent-utils/bundle.yaml` (v2) — team onboarding

### Built-in Bundle Recommendations (from PROJECT.md requirements)

| Bundle | Contents | Use Case |
|--------|----------|----------|
| `cloud` | AWS skill + GCP skill + Azure skill | Cloud infrastructure work |
| `ci` | GitHub skill + CI/CD skill | Pipeline and automation work |
| `browser` | Playwright skill + Playwright MCP | E2E testing |
| `gsd` | Full GSD suite (skills + agents + config) | GSD workflow engine users |
| `context` | context-mode skill + RTK skill | Token-optimised Claude Code setup |

### Bundle UX Pattern (from brew bundle)

```
agentkit bundle install cloud          # install named bundle
agentkit bundle dump                   # snapshot current installed to ~/.config/agentkit/bundle.yaml
agentkit bundle check                  # verify all bundle members are current
```

`bundle dump` writes a declarative file that can be committed to dotfiles repos — enabling the "new machine in one command" use case that brew bundle popularised.

### Progressive Disclosure in Bundles

Bundles install skills at their metadata-only tier initially. Full SKILL.md content is never loaded into the assistant context until the assistant actually needs the skill. This means installing a 10-skill bundle adds ~1,000 tokens to the context catalog, not 50,000.

---

## Feature Dependencies

```
Registry abstraction
  → install, update, search, info (all depend on registry resolution)
  → bundle install (depends on registry abstraction)

Per-assistant adapter
  → install (writes config to correct path/format)
  → MCP config write (format varies: JSON paths differ per assistant)

Local manifest (installed-packages record)
  → list
  → update (diff against registry)
  → uninstall
  → bundle dump

Background config agent
  → .agent-utils/config.json (depends on agent having run at least once)
  → per-project config read (depends on .agent-utils/config.json existing)

Progressive disclosure structure (SKILL.md < 500 lines, references/)
  → Token budget enforcement (validated at install time via manifest)
  → Three-tier loading (depends on structure being correct)
```

---

## MVP Recommendation

Prioritise for v1:
1. `install <name>` from GitHub manifest + mcpmarket.com + GSD registries
2. `list` — installed skills with version and source
3. `update [name]` — keep skills current
4. `search <query>` — ranked cross-registry discovery
5. `--bundle cloud/ci/gsd/context` — atomic preset installs
6. Per-assistant adapter (Claude Code first, others follow)
7. SKILL.md structure validation at install time

Defer to v2:
- `bundle dump` (snapshot to file) — valuable but not blocking adoption
- `doctor` — polish feature, not day-one
- Full background config agent — complex; ship stub that writes config, full drift detection in v2
- Project-scope install (`--scope project`)
- Private registry auth beyond token

---

## Sources

- [agentskills.io Specification](https://agentskills.io/specification) — HIGH confidence
- [Microsoft Agent Skills DeepWiki — Skill Structure](https://deepwiki.com/microsoft/agent-skills/3.2-skill-structure) — HIGH confidence
- [Progressive Disclosure Pattern — DeepWiki](https://deepwiki.com/microsoft/agent-skills/5.3-progressive-disclosure-pattern) — HIGH confidence
- [MicrosoftDocs/Agent-Skills GitHub](https://github.com/MicrosoftDocs/Agent-Skills) — HIGH confidence
- [open-gsd/gsd-core — install-on-your-runtime.md](https://github.com/open-gsd/gsd-core/blob/next/docs/how-to/install-on-your-runtime.md) — HIGH confidence
- [mcpmarket.com](https://mcpmarket.com/) — MEDIUM confidence (marketing site)
- [Capafy AI](https://capafy.ai/) — MEDIUM confidence (marketing site)
- [Homebrew Bundle docs](https://docs.brew.sh/Brew-Bundle-and-Brewfile) — HIGH confidence
- [mise Configuration](https://mise.jdx.dev/configuration.html) — HIGH confidence
- [skills.sh / agentskills ecosystem](https://www.agensi.io/learn/agent-skills-open-standard) — MEDIUM confidence
- [Progressive Disclosure token savings — Ardalis](https://ardalis.com/optimizing-ai-agents-with-progressive-disclosure/) — MEDIUM confidence
- [GSD background agent — MindStudio analysis](https://www.mindstudio.ai/blog/gsd-framework-prevent-context-rot-claude-code) — MEDIUM confidence
