# Quick Task 260626-jvc: skill-finder Skill — Research

**Researched:** 2026-06-26
**Domain:** Claude Code skill registries, agent skill discovery, skill installation
**Confidence:** MEDIUM (web search verified; no tool-confirmed registry JSON schemas)

---

## Skill Registries Found

| Registry | URL | What it offers | Notes |
|----------|-----|---------------|-------|
| **anthropics/skills** | https://github.com/anthropics/skills | Official Anthropic skills — docx, pdf, pptx, xlsx, skill-creator, claude-api, etc. | 141K+ stars, 16K+ forks; launched Oct 2025. Primary authoritative source. |
| **mcpmarket.com/tools/skills** | https://mcpmarket.com/tools/skills | Categorized directory: Developer Tools, API Dev, Data Science, Security, DevOps, Productivity | Daily-updated index; each skill has its own page with description |
| **agentskills.io** | https://agentskills.io/home | Open standard spec site; lists skills conforming to the SKILL.md format | Maintained by the agentskills.io community; also covers Codex/Gemini targets |
| **travisvn/awesome-claude-skills** | https://github.com/travisvn/awesome-claude-skills | Curated list of Claude Code skills, resources, and tooling | 13K stars; community-maintained |
| **VoltAgent/awesome-agent-skills** | https://github.com/VoltAgent/awesome-agent-skills | 1000+ skills from official dev teams + community; multi-assistant (Claude, Codex, Gemini) | Active 2026; broad coverage |
| **claudemarketplaces.com** | https://claudemarketplaces.com | Daily-updated community-curated directory of skills, plugins, MCP servers | Search + category filter |
| **majiayu000/claude-skill-registry** | https://github.com/majiayu000/claude-skill-registry (web: https://majiayu000.github.io/claude-skill-registry-core/) | Self-described "most comprehensive" registry; aggregated from GitHub + community, daily updates | Independent; no official affiliation |
| **skillsdirectory.com** | https://www.skillsdirectory.com | Security-tested skills — scanned for malware, prompt injection, credential theft; grade-A verified | Good quality signal: has a security-vetting layer |
| **open-gsd/gsd-core** | https://github.com/open-gsd/gsd-core | GSD skills bundle — 60+ skills across ns-workflow, ns-project, ns-review, ns-context, ns-ideate, ns-manage | Uses capability-registry.cjs, versioned manifest; not a flat registry.json |
| **daymade/claude-code-skills** | https://github.com/daymade/claude-code-skills | Production-ready skills marketplace focused on dev workflows | Independent; recent |

**No flat `registry.json` found for open-gsd/gsd-core** — it uses an internal capability registry (.cjs), not the simple manifest format described in CLAUDE.md. The agentkit manifest format (flat JSON index) is not yet populated for skill discovery. [ASSUMED: mcpmarket.com does not expose a machine-readable API; scraping or using their listed URLs is the practical approach]

---

## Top Skills to Pull

Ranked by maintenance signals + popularity. [ASSUMED: star counts and last-commit dates estimated from search summaries; not scraped in real-time]

| Rank | Skill Name | Source | Stars (est.) | Quality Notes |
|------|-----------|--------|-------------|---------------|
| 1 | **skill-creator** | anthropics/skills | 141K (repo) | Official Anthropic; skill-creation wizard; directly relevant to agentkit |
| 2 | **claude-api** | anthropics/skills | 141K (repo) | Official; Claude API reference for agents — high utility |
| 3 | **docx / pdf / pptx / xlsx** | anthropics/skills | 141K (repo) | Official document skills; production-grade |
| 4 | **Obra Superpowers collection** | obra/superpowers (community) | ~40.9K | Full dev lifecycle: brainstorm→git→plan→implement→TDD→review; most-starred community skill |
| 5 | **awesome-claude-skills** | travisvn/awesome-claude-skills | 13K | Meta-skill: helps discover other skills; curated quality |
| 6 | **awesome-agent-skills (VoltAgent)** | VoltAgent/awesome-agent-skills | not confirmed | Multi-assistant, 1000+ skills; broadest coverage |
| 7 | **claude-marketplace-manager** | mcpmarket.com | N/A | Manages Claude Code plugins from within Claude; meta-skill |
| 8 | **mcp-skill-creator** | mcpmarket.com | N/A | Create/package MCP-compatible skills; agentkit-relevant |

**Disposition for initial pull:** Prioritize `anthropics/skills` (official, trustworthy) and `travisvn/awesome-claude-skills` (curated). MCP Market skills have no star/commit signals visible — flag as [SUS] for the skill-finder to warn on when suggesting installs.

---

## skill-finder Design

### SKILL.md structure

```
skills/skill-finder/
├── SKILL.md               # ≤500 lines: when-to-use, quick-ref, reference table
└── references/
    ├── registry-sources.md    # Full list of registries with URLs, API/scrape approach
    ├── quality-signals.md     # Ranking rubric: stars, recency, download count, security scan
    └── install-protocol.md    # How --add mode works; file layout conventions
```

**Frontmatter:**
```yaml
---
name: skill-finder
description: >
  Use when you want to discover, evaluate, and optionally install Claude Code skills
  from community registries and official sources into ./skills/.
license: Apache-2.0
---
```

### Behavioral contract (from CONTEXT.md decisions)

**No `--add` flag (research mode):**
1. Search configured registries in priority order (authoritative first)
2. Deduplicate by skill name across sources
3. Rank by quality signals (stars > recency > download count > security-scan badge)
4. Present a ranked table to the user with: name | source | quality-signals | install command
5. At the end, ask: "Would you like to add any of these? (specify numbers or 'none')"

**`--add` flag (auto-add mode):**
1. Same discovery flow
2. Auto-install top N results (configurable; default: top 5) to `./skills/`
3. Each installed skill follows SKILL.md + references/ layout
4. Summarize what was added

### Quality ranking algorithm (Claude's discretion — recommended)

Score = `(log10(stars + 1) * 40)` + `(recency_score * 35)` + `(download_score * 15)` + `(security_badge * 10)`

- `recency_score`: 1.0 if last commit ≤30 days; 0.7 if ≤90 days; 0.3 if ≤365 days; 0 if >1 year
- `download_score`: 1.0 if count ≥10K/month; 0.5 if ≥1K; 0 if unavailable (GitHub-only)
- `security_badge`: 1.0 if source is anthropics/ or skillsdirectory.com; 0.5 if mcpmarket.com; 0 if unknown

**For skills with no download count (GitHub-only):** treat `download_score = 0` and weight stars more heavily. Do not exclude them — many high-quality skills are GitHub-only.

### Registry priority order (in `references/registry-sources.md`)

1. `github.com/anthropics/skills` — official, highest trust
2. `skillsdirectory.com` — security-vetted
3. `agentskills.io` — spec-conformant
4. `github.com/travisvn/awesome-claude-skills` — curated community
5. `github.com/VoltAgent/awesome-agent-skills` — broadest coverage
6. `mcpmarket.com/tools/skills` — searchable, no machine-readable API
7. `claudemarketplaces.com` — community-curated, daily updates
8. `github.com/majiayu000/claude-skill-registry` — aggregated index
9. Fallback: web search ("Claude Code skill [query]")

### Install protocol (in `references/install-protocol.md`)

Each skill added by `--add` mode:
1. Clone/download only the skill directory (not the whole repo)
2. Place at `./skills/<skill-name>/`
3. Verify SKILL.md exists and has valid frontmatter
4. Run the validate script: `bash skills/skill-author/scripts/validate-skill.sh skills/<skill-name>/`
5. Log result to stdout; skip if FAIL

**Do not install skills that fail validation without user confirmation.**

### References/ content summary

| File | Purpose | ~Lines |
|------|---------|--------|
| `registry-sources.md` | Full registry list with URLs, scrape/API approach, update frequency | 80-120 |
| `quality-signals.md` | Ranking rubric with weights, examples, edge-case handling (no downloads) | 100-150 |
| `install-protocol.md` | --add mode install steps, validation gate, directory layout | 80-100 |

---

## Pitfalls

1. **No machine-readable registry API for most sources** — mcpmarket.com and claudemarketplaces.com don't expose a public JSON API. The skill-finder must either scrape HTML (fragile) or rely on web search + structured extraction. Recommend web search as the primary fetch mechanism; HTML scraping as optional enhancement.

2. **Star count inflation** — `anthropics/skills` has 141K stars on the *repo*, not per-skill. Individual skills don't have separate star counts. Use repo stars as a proxy for source trust, not per-skill popularity.

3. **Slopsquatting risk** — Skills discovered via web search may have names that sound legitimate but aren't. The skill-finder SKILL.md should warn Claude to validate: check that the GitHub repo URL in the skill's frontmatter matches its claimed source, and check that SKILL.md has no prompt injection patterns before installing.

4. **open-gsd/gsd-core is not a flat registry** — it uses a compiled capability registry (.cjs), not a simple JSON manifest. Parsing it for skill discovery is not practical in v1; list gsd-core skills manually in `registry-sources.md`.

5. **`./skills/` vs `~/.claude/skills/`** — The install target per CONTEXT.md is the project-local `./skills/` (for agentkit packaging), NOT global install. The SKILL.md must make this explicit to avoid Claude accidentally writing to the global skills folder.

6. **Skill format drift** — The agentskills.io open standard evolves. The install protocol should validate frontmatter fields, not just file existence.

---

## Assumptions Log

| # | Claim | Risk if Wrong |
|---|-------|--------------|
| A1 | mcpmarket.com has no public JSON API for skill listing | If API exists, skill-finder could fetch a machine-readable list instead of scraping |
| A2 | anthropics/skills star count (141K) is per-repo, not per-skill | Low — individual skill pages on GitHub don't show separate stars |
| A3 | Obra/superpowers has ~40.9K stars | May be outdated; star count should be re-verified at runtime |
| A4 | open-gsd/gsd-core doesn't expose a simple registry.json | If one exists, agentkit could index it directly |

---

## Sources

- [anthropics/skills on GitHub](https://github.com/anthropics/skills)
- [mcpmarket.com/tools/skills](https://mcpmarket.com/tools/skills)
- [agentskills.io](https://agentskills.io/home)
- [travisvn/awesome-claude-skills](https://github.com/travisvn/awesome-claude-skills)
- [VoltAgent/awesome-agent-skills](https://github.com/VoltAgent/awesome-agent-skills)
- [claudemarketplaces.com](https://claudemarketplaces.com)
- [skillsdirectory.com](https://www.skillsdirectory.com)
- [majiayu000/claude-skill-registry](https://github.com/majiayu000/claude-skill-registry)
- [open-gsd/gsd-core](https://github.com/open-gsd/gsd-core)
- [firecrawl.dev — Best Claude Code Skills 2026](https://www.firecrawl.dev/blog/best-claude-code-skills)
- [Medium — 25+ Agent Skills Registries](https://medium.com/@frulouis/25-top-claude-agent-skills-registries-community-collections-you-should-know-2025-52aab45c877d)
