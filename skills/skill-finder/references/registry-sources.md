# Registry Sources for skill-finder

This file defines the priority-ordered search strategy for skill-finder. Query registries
in the order listed below. Stop querying additional registries once you have 20 or more
candidate skills, then deduplicate and rank. Use the web search fallback (priority 10)
only when a domain-specific query returns fewer than 5 results from curated registries.

## Priority-Ordered Registry Table

| Priority | Registry | URL | Trust Level | Fetch Approach | Notes |
|----------|----------|-----|-------------|----------------|-------|
| 1 | anthropics/skills | github.com/anthropics/skills | Official Anthropic | GitHub raw/API (public, no auth needed) | 141K+ repo stars; launched Oct 2025; skills include skill-creator, claude-api, docx, pdf, pptx, xlsx. Use repo stars as source trust proxy — not per-skill popularity |
| 2 | skillsdirectory.com | www.skillsdirectory.com | Security-vetted | Web search + page scrape (no public JSON API) | Skills scanned for malware, prompt injection, and credential theft. Grade-A verified badge is a strong quality signal |
| 3 | agentskills.io | agentskills.io/home | Spec-conformant | Web search | Covers Claude Code, Codex, and Gemini targets. Open standard site maintained by community |
| 4 | travisvn/awesome-claude-skills | github.com/travisvn/awesome-claude-skills | Curated community | Parse README.md for curated links | 13K stars; actively maintained; high curation quality |
| 5 | VoltAgent/awesome-agent-skills | github.com/VoltAgent/awesome-agent-skills | Broad community | Parse README.md | 1000+ skills from official dev teams and community; multi-assistant (Claude, Codex, Gemini); active 2026 |
| 6 | mcpmarket.com/tools/skills | mcpmarket.com/tools/skills | Indexed directory | Web search (no confirmed public JSON API — see assumption A1) | Daily-updated; each skill has its own detail page with description; search by category |
| 7 | claudemarketplaces.com | claudemarketplaces.com | Community-curated | Web search | Daily updates; category filter available; not security-vetted |
| 8 | majiayu000/claude-skill-registry | github.com/majiayu000/claude-skill-registry | Aggregated index | Web search or README parse | Self-described comprehensive index; aggregated from GitHub and community; no official affiliation |
| 9 | daymade/claude-code-skills | github.com/daymade/claude-code-skills | Community marketplace | README parse | Production-ready focus; dev workflow emphasis; recent activity |
| 10 | Web search fallback | google.com or ctx_fetch_and_index | Unknown | Web search query "Claude Code skill [topic]" | Use only when curated sources return fewer than 5 results for a domain-specific query; apply extra slopsquatting checks on all results |

## Notable Skill Collections

**obra/superpowers** — approximately 40,900 repo stars; the community's most-starred standalone
skill collection. Covers the full dev lifecycle: brainstorm, git workflows, planning,
implementation, TDD, and code review. When searching for general-purpose productivity skills,
check this collection first. Parse the repo's README.md or directory listing to enumerate
individual skills.

**open-gsd/gsd-core** — 60+ skills across six namespaces: ns-workflow, ns-project, ns-review,
ns-context, ns-ideate, and ns-manage. Skills use an internal capability-registry.cjs to register
themselves. This is not a flat JSON manifest — do not attempt to parse the .cjs at runtime.
Instead, list skills from this collection manually using the known namespace taxonomy when
presenting skill-finder results. The skill names follow the pattern `gsd-<verb>` (e.g.,
gsd-plan-phase, gsd-execute-phase, gsd-review).

## Registry Search Strategy

1. Query registries in the priority order shown above.
2. For each registry, collect all skill entries that match the user's search term or category,
   or all available skills when no filter is specified.
3. Stop querying additional registries once the candidate set reaches 20 or more skills.
4. If fewer than 5 results come from the curated registries (priorities 1-9) for a
   domain-specific query, supplement with web search (priority 10).
5. Deduplicate across registries by skill name (case-insensitive, hyphen-normalized). When
   the same skill appears in multiple registries, keep the highest-trust source entry and
   note the alternate source in the table's Notes column.
6. Apply the quality scoring formula from `references/quality-signals.md` to all remaining
   candidates, then sort descending by score.

## Fetch Approach Details

**GitHub repos (priorities 1, 4, 5, 8, 9):**
Use the GitHub API at `api.github.com/repos/{owner}/{repo}/git/trees/HEAD?recursive=1` to
list files. Filter for entries matching `*/SKILL.md`. Fetch each SKILL.md file using the
GitHub raw URL pattern: `raw.githubusercontent.com/{owner}/{repo}/HEAD/{path}`. No
authentication required for public repos; rate limit is 60 requests/hour unauthenticated
or 5000/hour with a token.

**anthropics/skills (priority 1):**
Each skill in this repo is a top-level directory containing a SKILL.md file. Enumerate
directories from the repo tree API. Extract skill names from SKILL.md frontmatter. Apply
a flat 0.9 stars_factor to all anthropics/skills entries (source trust, not per-skill
popularity).

**Web search and page scrape (priorities 2, 3, 6, 7):**
Use `ctx_fetch_and_index(url, source)` to retrieve and index registry pages. Then use
`ctx_search(queries)` to extract skill names, URLs, and descriptions. Parse result pages
for structured listings. If no structured listing is found, fall back to searching
`"Claude Code skill <topic> site:<domain>"` and extracting the skill names from
search result snippets.

## Assumptions

**A1 — mcpmarket.com JSON API:** No confirmed public JSON API exists for mcpmarket.com/tools/skills
as of June 2026. The current fetch approach is web search and page scraping. If a machine-readable
API is discovered, update this file and replace the scraping approach with direct JSON fetching.

**A4 — open-gsd/gsd-core capability registry:** The gsd-core repo uses a compiled capability
registry at `.claude/gsd-core/bin/capability-registry.cjs`. This is not a flat registry.json
that skill-finder can index directly. If the project adds a flat `registry.json` or `manifest.json`
at the repo root in a future release, agentkit could index it automatically. Until then, enumerate
gsd-core skills manually from the known namespace list.
