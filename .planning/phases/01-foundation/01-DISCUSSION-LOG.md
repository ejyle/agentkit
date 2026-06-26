# Phase 1: Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in 01-CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-08
**Phase:** 1-Foundation
**Areas discussed:** Install output style, Multi-registry conflicts (resolved by direction change), MCP config merge behavior, Installed state tracking

---

## Install output style

### Q1: What should agentkit install look like while running?

| Option | Description | Selected |
|--------|-------------|----------|
| Spinner + progress | Bubbletea spinner per phase, clean success line | ✓ |
| Clean structured lines | fmt.Println in sequence, no TUI at install time | |
| TTY-aware (both) | Rich when TTY, plain when piped | |

**User's choice:** Spinner + progress

---

### Q2: Error output on install failure?

| Option | Description | Selected |
|--------|-------------|----------|
| Single clear error line | What failed + registries checked + suggested command | ✓ |
| Verbose with suggested fix | Error + what was tried + concrete fix command | |
| Structured JSON on error | Machine-readable JSON to stderr alongside human error | |

**User's choice:** Single clear error line

---

### Q3: What should agentkit list output look like?

| Option | Description | Selected |
|--------|-------------|----------|
| Table format | Aligned columns: PACKAGE, VERSION, TYPE, TARGET, REGISTRY | ✓ |
| One line per package | Compact: `playwright@1.2.0 (mcp, claude, agentkit-registry)` | |
| Grouped by target assistant | Separate section per assistant | |

**User's choice:** Table format

---

### Q4: Should agentkit search show a spinner?

| Option | Description | Selected |
|--------|-------------|----------|
| Spinner then ranked list | Fetch all registries in parallel, deterministic ordered output | ✓ |
| Streaming results | Print each registry's results as they arrive, non-deterministic order | |
| You decide | Researcher/planner pick | |

**User's choice:** Spinner then ranked list

---

## Multi-registry conflicts (resolved by direction change)

The gray area as framed (same package name in two registries) became moot when the user clarified the product direction.

**User clarification (free-text):** agentkit uses a single curated registry only. Skills/MCPs/agents are benchmarked via auto-research before being listed. The curated quality guarantee IS the product's selling point. Users install from the official list — they don't add arbitrary sources.

**Impact:**
- CLI-09 (`agentkit registry add/remove`) — removed from v1
- REG-03 (mcpmarket.com) — removed from v1
- REG-04 (custom registry sources) — removed from v1
- REG-01 narrowed: only the official agentkit-registry, not "any GitHub repo"

### Q: Should users ever install from a source other than the official curated registry (in v1)?

| Option | Description | Selected |
|--------|-------------|----------|
| No — curated registry only | v1 installs only from official agentkit-registry | ✓ |
| Yes — curated first, custom as power feature | Official default + power-user custom sources labeled as unverified | |

**User's choice:** No — curated registry only

---

## MCP config merge behavior

### Q1: What happens when an MCP entry with the same name already exists in settings.json?

| Option | Description | Selected |
|--------|-------------|----------|
| Overwrite silently | Replace entry, agentkit owns what it installs | |
| Warn and overwrite | Print warning showing old vs new, then overwrite | |
| Prompt to confirm | Stop and ask Overwrite? [y/N] | ✓ |

**User's choice:** Prompt to confirm

---

### Q2: Should it still prompt when upgrading an agentkit-managed package?

| Option | Description | Selected |
|--------|-------------|----------|
| Always prompt (consistent) | Same behavior whether upgrade or foreign entry | |
| Auto on upgrade, prompt on foreign | If agentkit owns it (installed.json), auto-upgrade with one-line notice | ✓ |
| Auto-overwrite always with --yes flag | Default is prompt; --yes skips for CI | |

**User's choice:** Auto on upgrade, prompt on foreign

---

### Q3: Uninstall behavior when settings.json has other data?

| Option | Description | Selected |
|--------|-------------|----------|
| Remove only agentkit's entry, preserve rest | Atomic read → delete key → write | ✓ |
| You decide | Researcher/planner pick | |

**User's choice:** Remove only agentkit's entry, preserve everything else

---

## Installed state tracking

### Q1: Where should agentkit track what's installed?

| Option | Description | Selected |
|--------|-------------|----------|
| Single global file | `~/.config/agentkit/installed.json` across all assistants | |
| Per-assistant files | `~/.config/agentkit/claude/installed.json` per target | ✓ |
| You decide | Researcher/planner pick | |

**User's choice:** Per-assistant files

---

### Q2: What fields must each installed package entry track?

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal: name, version, type, install_path | Just what's needed for 4 commands | |
| Full: add installed_at, source_url, checksum | Rich audit trail for security and reproducibility | ✓ |

**User's choice:** Full schema with installed_at, source_url, checksum

---

### Q3: First install behavior (installed.json doesn't exist yet)?

| Option | Description | Selected |
|--------|-------------|----------|
| Auto-create on first install | Create dir + file transparently, no setup step | ✓ |
| Require agentkit init | User runs init once to set up config directory | |

**User's choice:** Auto-create on first install

---

## Claude's Discretion

- Search result ranking algorithm (exact match → fuzzy → description keywords)
- Bubbletea spinner component style, frame rate, and color
- JSON schema versioning strategy for installed.json (whether to include schema_version field)

## Deferred Ideas

- CLI-09 (`agentkit registry add/remove`) — removed from v1 (curated-only model)
- REG-03 (mcpmarket.com API registry) — removed from v1
- REG-04 (custom registry sources) — removed from v1
- `--yes` / `--non-interactive` flag for CI scripting — noted but not required for Phase 1
