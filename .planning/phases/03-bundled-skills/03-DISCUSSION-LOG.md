# Phase 3: Bundled Skills - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-09
**Phase:** 3-Bundled Skills
**Areas discussed:** Skill file delivery, Bundle failure semantics, The 9 skill contents, GSD package scope, More skills (external curation)

---

## Skill File Delivery

| Option | Description | Selected |
|--------|-------------|----------|
| Embedded in binary | `//go:embed` skill files in binary; zero network dependency | |
| Hosted in agentkit-registry repo | Skills in ejyle/agentkit-registry; registry.json has download_url | |
| Skills in this repo, GitHub releases | Skills in ejyle/agentkit source repo; fetched from GitHub release tarball | ✓ |
| Separate skills GitHub repo | Skills in a dedicated ejyle/agentkit-skills repo | |

**User's choice:** Skills in the current repo (agentkit source). Fetched from GitHub releases tarball at install time.

**Follow-up — registry entry format:**

| Option | Description | Selected |
|--------|-------------|----------|
| `github-release` install method | `install.method='github-release'`, `install.repo`, `install.path` fields | ✓ |
| Explicit download_url | registry.json has hardcoded URL; must update per release | |
| Bundled bypass registry | Hardcoded list in CLI; registry only for third-party | |

**Notes:** User phrased it as "is it better to have it in the current repo not include it in the binary or separate repo?" — interpreted as preference for same-repo approach without binary embedding.

---

## Bundle Failure Semantics

| Option | Description | Selected |
|--------|-------------|----------|
| Best-effort | Install all concurrently; collect failures; report at end | ✓ |
| Stop on first failure | Abort on first fail; no rollback of already-installed | |
| Atomic rollback | All-or-nothing; rollback on any failure | |

**Follow-up — parallelism:**

| Option | Description | Selected |
|--------|-------------|----------|
| Parallel | All bundle packages install concurrently | ✓ |
| Sequential | One at a time | |

**Follow-up — exit code:**

| Option | Description | Selected |
|--------|-------------|----------|
| Exit 1 if any failed | Non-zero on partial failure | ✓ |
| Exit 0 always | Always success even on partial | |

**Notes:** Straightforward — user took all recommended options.

---

## The 9 Skill Contents

| Option | Description | Selected |
|--------|-------------|----------|
| Full reference-quality | Real SKILL.md + reference files; immediately useful | ✓ |
| Minimal stubs | Placeholder content; passes validation only | |
| Mixed | Full for cloud/dev, thin for context-mode/RTK/Serena | |

**User's choice:** Full reference-quality. Also: create a framework to evaluate and improve skills so it's easy to add new skills, MCPs, agents to the curated registry.

**Follow-up — evaluation framework:**

| Option | Description | Selected |
|--------|-------------|----------|
| `agentkit validate` CLI command | Local authoring validation | |
| CI quality gate in registry repo | GitHub Actions on PR | |
| Both | CLI + CI | |
| Skill-creator pipeline | Use Claude + skill-creator for evaluation/improvement | ✓ |

**User's response:** "use claude skill-creator to do this pipeline, it can have reference, scripts as needed"

**Follow-up — skill-author meta-skill: Phase 3 or deferred:**

| Option | Description | Selected |
|--------|-------------|----------|
| 10th bundled skill in Phase 3 | Add skill-author skill now | ✓ |
| Deferred to v0.2.0 | Keep to 9 BND skills | |

**User changed answer:** Initially said defer, then changed to "we'll create the meta-skill now"

**Follow-up — context-mode/RTK/Serena sourcing:**

**User's choice:** Adapt from existing installs.

**Follow-up — skill-author meta-skill contents:**

| Option | Description | Selected |
|--------|-------------|----------|
| SKILL.md + skill-creator.md + agentkit-spec.md | Lightweight references | |
| Full: SKILL.md + domain references + scripts + auto-researcher agent | ✓ |

**Notes:** User specified: "SKILL.md, domain reference for progressive discovery, helper scripts, auto-research agent for self evaluate and improve"

**Follow-up — auto-researcher agent:**

| Option | Description | Selected |
|--------|-------------|----------|
| References existing auto-researcher | External reference only | |
| New agentkit-specific agent | New standalone agent | |
| Bundle existing in this repo | Include in this repo | ✓ |

**User's response:** "i have it installed in my system but include it in this repo"

---

## GSD Package Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Skills + agents only | Install GSD skill files + agent defs | |
| Skills + agents + slash commands | Also writes slash command entries | |
| Meta-package with sub-packages | Registry models dependencies | |
| Delegate to gsd-core | Use GSD's own install from gsd-core registry | ✓ |

**User's response:** "we dont need keep anything for GSD in our repo, we just use the gsd install method that is in their github, we just call that they will take care of the installations ref: https://github.com/open-gsd/gsd-core"

**Follow-up — how to invoke:**

| Option | Description | Selected |
|--------|-------------|----------|
| Custom install method in registry.json | GSD entry has install.method='custom' | |
| Research before implementing | Check gsd-core registry.json first | ✓ |

---

## More Skills (External Curation)

**Triggered by:** User note "might need to add more skills" at area selection.

**User-named sources:** Anthropic skills (frontend-design), Vercel Labs (vercel-react-best-practices, agent-browser), skills.sh top list, Azure top skills.

| Option | Description | Selected |
|--------|-------------|----------|
| Curate in registry.json pointing to their repos | Registry entries only; no content copy | |
| Copy content into this repo | We own it; manual updates | ✓ |
| Defer to post-launch | Not Phase 3 | |

**User's response:** "we own it, but we should have mechanism to just update these external skills with some cmd"

**Follow-up — update mechanism scope:**

| Option | Description | Selected |
|--------|-------------|----------|
| Defer: initial copy only in Phase 3 | Update mechanism is later | |
| Part of Phase 3 | Include update command | ✓ (with caveat) |

**User clarification:** "we might need to run the meta-skill + auto-research and make one more approval to make a skill, agent, or MCP live"

**Follow-up — scope boundary:**

| Option | Description | Selected |
|--------|-------------|----------|
| 10 authored skills + --bundle + initial external curation | ✓ |
| Full scope including update + publishing pipeline | |

**Final scope:** Phase 3 = 10 authored skills + --bundle + initial external curation (10–12 external skills). Update mechanism and publishing pipeline deferred to v0.2.0.

**Follow-up — number of external skills:**

| Option | Description | Selected |
|--------|-------------|----------|
| 3–5 high-quality external skills | Manageable scope | |
| 10–12 from all named sources | Broader launch library | ✓ |

---

## Claude's Discretion

- Exact `skills/` directory structure layout (flat vs nested by category)
- Whether `github-release` installer caches the tarball across a bundle session
- Spinner output format for parallel bundle installs
- Exact agentkit-registry entries for external skills (names, descriptions, version tags)
- Final list of 10–12 external skills (researcher evaluates candidates from named sources)

## Deferred Ideas

- External skill update mechanism (`agentkit skill sync <name>`) — v0.2.0
- Skill publishing pipeline with approval gating — v0.2.0 (meta-skill is in Phase 3; pipeline automation is not)
- `agentkit skill validate` / `agentkit skill improve` as CLI subcommands — v0.2.0
- Additional bundle definitions beyond cloud/dev/context — v0.2.0
- Background config agent / per-project facts system — v0.2.0 (per REQUIREMENTS.md)
