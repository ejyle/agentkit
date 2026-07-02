# Quick Task 260701-hsh: Add Cursor AI support — Research

**Researched:** 2026-07-01 (Rules/skills section corrected 2026-07-01 after user-provided source)
**Domain:** Go CLI adapter for Cursor AI code editor (MCP config + skill install)
**Confidence:** HIGH (MCP config format), HIGH (skills format/location — corrected)

## CORRECTION (supersedes original Rules-mapping findings below)

The original research pass below concluded Cursor has no user-global directory for skill-like content and proposed mapping skills to `.cursor/rules/*.mdc` (project-scoped only). **This was wrong about the mechanism, not just the scope.** Cursor ships native **Agent Skills** support (https://cursor.com/docs/skills), which is a *separate feature from Rules*. Skills are auto-discovered from these directories, both project- and user-scoped:
- `.cursor/skills/` and `.agents/skills/` (project-level, including nested subdirectories in monorepos)
- `~/.cursor/skills/` and `~/.agents/skills/` (**user-level/global** — this is what agentkit needs)
- For compatibility, Cursor also reads `.claude/skills/`, `.codex/skills/`, `~/.claude/skills/`, `~/.codex/skills/`

Each skill is a **folder containing a `SKILL.md` file**, with optional `scripts/`, `references/`, `assets/` subdirectories — the *exact same structure* agentkit already writes for Claude Code, Gemini, and pi. **No format conversion, no `.mdc` frontmatter, no Rules mapping needed.**

**Corrected recommendation:** `WriteSkill`/`RemoveSkill` for `CursorAdapter` should write/remove a standard skill folder at `~/.cursor/skills/<name>/`, using the identical code path already used by `GeminiAdapter`/`ClaudeCodeAdapter`/`PiAdapter` for skill writes — just a different base path constant. This eliminates the scope conflict entirely: both MCP config (`~/.cursor/mcp.json`) and skills (`~/.cursor/skills/<name>/`) are user-home-scoped, consistent with every other adapter in this codebase.

Everything below this point describing `.mdc`/Rules mapping (Pattern 2, Pitfall 1/2/3's Rules-specific framing, Open Question #1, and the Rules-related Sources) is **superseded** — kept only for the MCP-config findings, which remain valid and HIGH confidence.

## Summary (MCP config portion — still valid)

Cursor's MCP config format matches the existing `jsonMCPAdapter` shape used by Gemini/Pi/Copilot exactly: a JSON file with a top-level `"mcpServers"` key, each entry `{"command": "...", "args": [...], "env": {...}}` — no extra `"type"` field. The global config lives at `~/.cursor/mcp.json` (Windows: `%USERPROFILE%\.cursor\mcp.json`), confirmed by Cursor's official docs. This means `CursorAdapter` can be implemented as a thin wrapper embedding `jsonMCPAdapter`, identical in structure to `GeminiAdapter`.

**Skills (corrected): write standard `SKILL.md` folders to `~/.cursor/skills/<name>/`** — see CORRECTION above. No Rules/.mdc conversion, no scope conflict, no product decision needed.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| MCP server config write/read/remove | Adapter (Go, filesystem) | — | Same tier as every other `AssistantAdapter`; pure JSON file I/O against `~/.cursor/mcp.json` |
| Skill → Rule mapping | Adapter (Go, filesystem) | CLI (target validation, warnings) | File write tier owns the `.mdc` content; CLI tier owns surfacing the project-vs-user-scope caveat to the operator |
| Target registration | `cmd/root.go` + `internal/adapter/factory.go` | — | Existing pattern: string literal added to `validTargets` + switch case |

## Standard Stack

### Core
No new external dependencies required. Cursor's MCP format is plain JSON (stdlib `encoding/json`, already used by every other JSON adapter). `.mdc` files are plain markdown with YAML frontmatter — no YAML library needed since agentkit only needs to *write* frontmatter as a fixed-format string block (not round-trip parse arbitrary YAML), consistent with how other adapters write fixed-shape JSON via `map[string]interface{}` rather than a schema struct.

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stdlib `encoding/json` | — | MCP config R/W | Already used by `jsonMCPAdapter` |
| stdlib `path/filepath`, `os` | — | Path construction, dir creation | Already used by every adapter |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Reuse `jsonMCPAdapter` via embedding | Write a standalone `CursorAdapter` (no embedding) | No benefit — Cursor's MCP shape is a 1:1 match for `jsonMCPAdapter` (same as Gemini); standalone code would just duplicate `jsonbase.go` logic for zero gain |

**Installation:** No new packages to install — this is pure Go stdlib, matching existing adapters.

## Package Legitimacy Audit

Not applicable — no external packages are introduced by this change.

## Architecture Patterns

### System Architecture Diagram

```
cmd/root.go (validTargets: add "cursor")
        │
        ▼
internal/adapter/factory.go (NewAdapter("cursor", store))
        │
        ▼
internal/adapter/cursor.go — CursorAdapter{ jsonMCPAdapter }
        │
        ├─ WriteMCPConfig / ReadMCPConfig / RemoveMCPConfig
        │       │
        │       ▼
        │   ~/.cursor/mcp.json   { "mcpServers": { "<name>": {command, args, env} } }
        │
        └─ WriteSkill / RemoveSkill
                │
                ▼
        [DECISION NEEDED] — no documented user-global Rules location exists.
        Option A: ErrNotSupported (mirror OpenCodeAdapter's skill-unsupported pattern)
        Option B: write to CWD-relative .cursor/rules/<name>.mdc with explicit scope warning
```

### Recommended Project Structure
```
internal/adapter/
├── cursor.go        # CursorAdapter — embeds jsonMCPAdapter, mirrors gemini.go
└── cursor_test.go   # mirrors gemini_test.go test shape exactly
```

### Pattern 1: Embed `jsonMCPAdapter` for standard `mcpServers` shape
**What:** `CursorAdapter` struct embeds `jsonMCPAdapter` and supplies `configPath`, `mcpKey: "mcpServers"`, `extraFields: nil` (Cursor's entries have no extra `"type"` field, matching Gemini, not Copilot/OpenCode).
**When to use:** Any assistant whose MCP config is `{"mcpServers": {"<name>": {"command":..., "args":[...], "env":{...}}}}` with no extra per-entry fields.
**Example:**
```go
// Source: internal/adapter/gemini.go (existing pattern in this repo) — Cursor mirrors this exactly
type CursorAdapter struct {
    jsonMCPAdapter
}

func NewCursorAdapter(store *config.ConfigStore) *CursorAdapter {
    return NewCursorAdapterWithHome(store, "")
}

func NewCursorAdapterWithHome(store *config.ConfigStore, homeDir string) *CursorAdapter {
    return &CursorAdapter{
        jsonMCPAdapter: jsonMCPAdapter{
            store:   store,
            homeDir: homeDir,
            mcpKey:  "mcpServers",
            configPath: func(home string) (string, error) {
                return filepath.Join(home, ".cursor", "mcp.json"), nil
            },
            extraFields: nil,
        },
    }
}

func (a *CursorAdapter) Name() string { return "cursor" }
```

### Pattern 2: `.mdc` frontmatter for a "skill" mapped to a Cursor Rule
**What:** Cursor Rules use YAML frontmatter with three fields: `description` (human-readable "when to use this"), `globs` (file-pattern activation), `alwaysApply` (boolean). Four rule types map to combinations of these fields:

| Rule Type | `alwaysApply` | `globs` | `description` | Best fit for "installed skill" semantics |
|-----------|---------------|---------|----------------|---|
| Always | `true` | — | optional | No — loads on every request regardless of relevance; wastes context ("token tax") |
| Auto Attached | `false` | set (e.g. `["**/*.ts"]`) | optional | Only if the skill is file-type-specific; most agentkit skills are not |
| **Agent Requested** | `false` | empty | **required, must clearly describe when to use it** | **Best fit** — matches "agent pulls in when relevant" exactly; Cursor's agent reads the `description` to decide whether to load the rule body, same semantic as Claude Code's SKILL.md description-based discovery |
| Manual | `false` | empty, no description | — | No — requires explicit `@ruleName` invocation by the user, defeats "installed skill available automatically" |

**Recommendation:** Map every installed skill to an **Agent Requested** rule — `alwaysApply: false`, no `globs`, `description` populated from the skill's own description/frontmatter (agentkit skills already carry a description in `SKILL.md`).

```markdown
---
description: <skill's own description field, verbatim>
alwaysApply: false
---
<skill body content>
```
Source: cross-referenced from Cursor's official rules docs (`cursor.com/docs/context/rules`) + community documentation. `[CITED: cursor.com/docs/context/rules]`

### Anti-Patterns to Avoid
- **Treating Cursor Settings "User Rules" as a file target:** They are stored in Cursor's internal settings/DB, not a JSON or text file agentkit can locate and write deterministically. Do not attempt to write to a guessed settings path.
- **Assuming `~/.cursor/rules/` exists today:** It is a *requested*, not shipped, feature as of this research date. Do not implement `WriteSkill` against that path as if it were documented/stable.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| MCP JSON read/merge/write with conflict detection | A new parser/writer for Cursor | Existing `jsonMCPAdapter` (embed it) | Already handles atomic write, `ErrForeignConflict`, post-write verify (MCP-06) — Cursor's shape is identical to Gemini's, zero new logic needed |

**Key insight:** This phase's MCP half is a copy-paste-rename of `gemini.go`; the only genuinely new problem is the Rules/skill mapping, which cannot be solved by reusing existing code because no adapter in this codebase currently writes to a project-relative path — every existing `WriteSkill` implementation targets a user-home path (`~/.gemini/skills/`, `~/.claude/skills/`) or returns `ErrNotSupported` (OpenCode). Cursor is the first case where the *only* viable Rules location is project-relative, which is a genuinely new scope decision for this codebase, not an implementation detail.

## Runtime State Inventory

Not applicable — this is a greenfield adapter addition, not a rename/refactor/migration.

## Common Pitfalls

### Pitfall 1: Conflating "Cursor Settings > Rules" (global, plain text) with `.cursor/rules/*.mdc` (project, structured)
**What goes wrong:** Documentation and blog posts use "global rules" loosely to mean the Settings UI feature, which has nothing to do with the file-based `.mdc` Rules system. A planner skimming secondary sources could wrongly conclude a `~/.cursor/rules/` directory is real and shippable today.
**Why it happens:** Cursor's own marketing/docs terminology overlap ("Rules" is used for both features).
**How to avoid:** Anchor on the fact that `.mdc`-format, frontmatter-bearing rules are documented by Cursor as project-scoped only; there is no officially documented global equivalent.
**Warning signs:** Any source claiming "just put your .mdc file in `~/.cursor/rules/`" without citing an official Cursor doc — this is community aspiration, not shipped behavior (confirmed via open, unresolved forum feature requests).

### Pitfall 2: Writing to CWD without an explicit CLI signal
**What goes wrong:** If the planner chooses Option B (write to `./.cursor/rules/`), a silent write to the current working directory breaks the "no project scope" mental model users have for every other target and could pollute an unrelated directory if agentkit is invoked from the wrong place.
**Why it happens:** `WriteSkill`'s interface signature (`name string, files map[string][]byte`) doesn't carry a "target directory" parameter — the only path-determining state available to the adapter is home dir or CWD, no explicit project root.
**How to avoid:** If Option B is chosen, require the CLI layer to print a warning before writing (`"cursor" target skills are written to ./.cursor/rules/ relative to the current directory — not user-global`) and consider gating behind a confirmation prompt, matching this project's `bubbletea` confirm-prompt pattern used elsewhere.

### Pitfall 3: Flattening `references/`/`scripts/` subdirectories loses functionality, not just format
**What goes wrong:** agentkit skills use progressive disclosure (SKILL.md + references/ + scripts/) specifically so the assistant loads only what's needed. A Cursor Rule is a single `.mdc` file — there is no sibling-file loading mechanism inside a rule; scripts referenced by path (`scripts/foo.py`) would be dead references once flattened into one file, or would need separate physical files placed next to the `.mdc` (untested territory, not part of Cursor's documented Rules contract).
**How to avoid:** For skills with `references/`/`scripts/`, either (a) concatenate `references/*.md` content directly into the rule body with clear section headers (lossy but functional for text-only skills), or (b) explicitly skip/simplify skills that depend on `scripts/` execution when installing to Cursor, and document this limitation rather than pretend feature parity. This is a planning-time policy decision, not solvable purely by code.

## Code Examples

### Existing pattern to mirror exactly (Gemini adapter + test)
```go
// Source: internal/adapter/gemini.go (this repo)
type GeminiAdapter struct {
    jsonMCPAdapter
}
// WriteSkill / RemoveSkill write/remove a directory under ~/.gemini/skills/<name>/
```
```go
// Source: internal/adapter/gemini_test.go (this repo) — test naming/structure convention
// TestCursorAdapter_WriteMCPConfig_CreatesFile
// TestCursorAdapter_WriteMCPConfig_PreservesExistingKeys
// TestCursorAdapter_WriteMCPConfig_ErrForeignConflict
// TestCursorAdapter_WriteMCPConfig_AutoOverwrite
// TestCursorAdapter_ReadMCPConfig_AfterWrite
// TestCursorAdapter_RemoveMCPConfig
// TestCursorAdapter_WriteSkill_* / RemoveSkill_* — shape depends on planner's Rules-scope decision
```

### `.mdc` frontmatter example
```markdown
<!-- Source: cursor.com/docs/context/rules + community-verified examples -->
---
description: Use when writing or reviewing React component code
alwaysApply: false
---

# React Component Rules

Always use functional components with TypeScript interfaces.
```

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Cursor Rules "Agent Requested" type is the best semantic fit for "installed skill" | Architecture Patterns / Pattern 2 | Low — if wrong, `alwaysApply: false` + empty globs + description is still a safe, working rule; only the *optimality* of the mapping is assumed, not its validity |
| A2 | No file-based path exists for Cursor Settings "User Rules" (i.e., it cannot be targeted by writing a file) | Summary, Pitfall 1 | Medium — if Cursor exposes an undocumented settings file format in a future version, this could change; current multiple independent sources (official docs + active unresolved forum requests) corroborate this as of research date |
| A3 | Concatenating `references/*.md` into the rule body is an acceptable degradation strategy for progressive-disclosure skills | Pitfall 3 | Low-Medium — functional for text-only skills; skills relying on `scripts/` execution have no working equivalent regardless of strategy chosen, so this only affects the read-only-reference subset |

## Open Questions

1. **Should agentkit write Cursor skills to the CWD's `.cursor/rules/` at all, given the project's "user scope only" v1 constraint?**
   - What we know: No documented user-global Rules mechanism exists in Cursor today.
   - What's unclear: Whether the planner/product owner considers "MCP-only for Cursor v1" an acceptable scope reduction, or whether project-scoped writes should be a one-off exception for this target.
   - Recommendation: Surface this explicitly to the user/planner as a binary decision before implementation — do not silently pick one. If forced to default, prefer MCP-only (Option A / `ErrNotSupported`, mirroring `OpenCodeAdapter`'s existing precedent) since it requires zero deviation from the project's stated architecture constraints.

## Environment Availability

Not applicable — no new external tool/service dependency; this task only adds Go source files within the existing module.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` |
| Config file | none — uses `go test ./...` |
| Quick run command | `go test ./internal/adapter/... -run TestCursorAdapter -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| (none assigned — quick task, no formal REQ IDs) | `cursor` target added to `validTargets` and routes to `CursorAdapter` | unit | `go test ./cmd/... -run TestValidTargets` (or equivalent) | ❌ create in `cursor_test.go` / existing root_test.go if present |
| — | WriteMCPConfig writes `~/.cursor/mcp.json` with `mcpServers` key, no `type` field | unit | `go test ./internal/adapter/... -run TestCursorAdapter_WriteMCPConfig` | ❌ new `cursor_test.go` |
| — | ReadMCPConfig / RemoveMCPConfig round-trip | unit | `go test ./internal/adapter/... -run TestCursorAdapter` | ❌ new `cursor_test.go` |
| — | WriteSkill/RemoveSkill behavior per planner's scope decision | unit | `go test ./internal/adapter/... -run TestCursorAdapter_WriteSkill` | ❌ new `cursor_test.go` — depends on Open Question #1 resolution |

### Sampling Rate
- **Per task commit:** `go test ./internal/adapter/... -run TestCursorAdapter -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before considering task complete

### Wave 0 Gaps
- [ ] `internal/adapter/cursor_test.go` — mirror `gemini_test.go` structure for all MCP methods
- [ ] `cmd/root_test.go` (if it exists) — add `cursor` to any table-driven validTargets test

## Security Domain

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | yes | Same as existing adapters — skill/MCP names are validated/sanitized upstream before reaching the adapter (per STATE.md's "Registry ID sanitized to [a-zA-Z0-9_-]" decision); no new validation surface introduced |
| V6 Cryptography | no | Not applicable — no secrets/crypto in this change |

### Known Threat Patterns for this stack
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal via skill/server name into `.mdc` filename | Tampering | Reuse existing name-sanitization pattern already enforced upstream (per STATE.md); do not introduce a second, divergent sanitization path in `cursor.go` |
| Writing to unintended directory (CWD ambiguity) if Option B chosen | Tampering / repudiation of scope | Explicit CLI warning + confirmation before any project-relative write (see Pitfall 2) |

## Sources

### Primary (HIGH confidence)
- [Cursor MCP Docs](https://cursor.com/docs/mcp) — global config path `~/.cursor/mcp.json`, `mcpServers` JSON shape
- [Cursor Rules Docs](https://cursor.com/docs/context/rules) / [cursor.com/docs/rules.md](https://cursor.com/docs/rules.md) — frontmatter fields (`description`, `globs`, `alwaysApply`), four rule types, User Rules vs Project Rules distinction
- Codebase: `internal/adapter/jsonbase.go`, `internal/adapter/gemini.go`, `internal/adapter/gemini_test.go`, `internal/adapter/opencode.go`, `internal/adapter/adapter.go`, `internal/adapter/factory.go`, `cmd/root.go` — direct read of existing implementation to verify interface contract and reuse opportunity `[VERIFIED: codebase]`

### Secondary (MEDIUM confidence)
- [Cursor Community Forum — Global .cursor/rules directory feature request](https://forum.cursor.com/t/global-cursor-rules-directory/50049) — confirms no shipped global rules directory; active unresolved request
- [Cursor Community Forum — Support for ~/.cursor/rules for global .mdc rules](https://forum.cursor.com/t/support-for-cursor-rules-for-global-mdc-rules/144819) — corroborates absence of the feature
- [Cursor Rules: Complete .mdc Guide (vibecodingacademy.ai)](https://www.vibecodingacademy.ai/blog/cursor-rules-complete-guide) — frontmatter field details, cross-checked against official docs

### Tertiary (LOW confidence)
- Various SEO/marketing blog posts on Cursor MCP setup (natoma.ai, truefoundry.com, fast.io) — used only to corroborate the `~/.cursor/mcp.json` path already confirmed by official docs; not relied upon for any claim not independently verified

## Metadata

**Confidence breakdown:**
- MCP config format/path: HIGH — confirmed directly against official Cursor docs, and structurally identical to an already-implemented, tested pattern in this codebase (`jsonMCPAdapter`)
- Rules/`.mdc` format: HIGH — official docs cited for frontmatter fields and rule types
- Rules global-scope availability: HIGH confidence in the negative finding (no global directory exists) — corroborated by official docs (User Rules = plain text only) AND multiple independent, unresolved community feature requests spanning different discussion threads
- Skill-to-Rule mapping convention (Pattern 2, Pitfall 3): MEDIUM — technically sound reasoning from documented rule-type semantics, but "best fit" judgment and the flatten-vs-skip strategy for `references/`/`scripts/` are not covered by any Cursor documentation (Cursor has no concept of "skills" natively) — genuinely novel design decisions for this codebase

**Research date:** 2026-07-01
**Valid until:** 2026-08-01 (Cursor's Rules/MCP docs are actively evolving; re-verify before implementing if significant time has passed, especially re-check whether a global rules directory has since shipped)
