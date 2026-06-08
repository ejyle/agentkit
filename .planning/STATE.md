---
gsd_state_version: 1.0
milestone: v0.1.0
milestone_name: milestone
status: executing
last_updated: "2026-06-08T14:51:46.681Z"
progress:
  total_phases: 4
  completed_phases: 0
  total_plans: 6
  completed_plans: 1
  percent: 0
---

# STATE: agentkit

_Last updated: 2026-06-08_

---

## Project Reference

**Core value:** Zero-friction install of the right skills, agents, and MCP servers for any AI coding assistant — one command gets you from bare machine to fully instrumented dev environment.

**Current focus:** Phase 01 — foundation

---

## Current Position

Phase: 01 (foundation) — EXECUTING
Plan: 1 of 6
**Phase:** 1 — Foundation
**Plan:** 1 complete (01-01), starting 01-02
**Status:** Executing Phase 01
**Progress:** [██░░░░░░░░] 17%

```
Phase 1: Foundation          [██░░░░░░░░] 17%
Phase 2: Multi-Assistant     [----------]  0%
Phase 3: Bundled Skills      [----------]  0%
Phase 4: Distribution        [----------]  0%
```

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Phases complete | 0 / 4 |
| Plans complete | 1 / 6 |
| Requirements mapped | 41 / 41 |
| Requirements validated | 2 / 41 (CLI-01, CLI-02) |

---

## Accumulated Context

### Key Decisions (pending confirmation)

| Decision | Rationale | Status |
|----------|-----------|--------|
| Go over Python for CLI | Single binary, no runtime, fast startup, easy cross-compile | Pending |
| User scope only (v1) | Covers majority use case; project scope adds complexity | Pending |
| `.agent-utils/config.json` for project config | Dedicated namespace, gitignore-able | Pending |
| Manifest-driven registries | Extensible without CLI changes | Pending |

### Architecture Constraints

- Use `os.UserHomeDir()` for all path resolution — never `os.Getenv("HOME")`
- Use `filepath.Join()` for all path construction
- Atomic file writes via temp + rename (use `google/renameio` on Windows)
- Each assistant adapter detects config path at runtime — never hardcoded
- Post-install verify: re-read written MCP config after every write
- Registry HTTP: explicit 3s connect / 10s read timeouts via `go-retryablehttp`
- ETag-based manifest cache; `--offline` fallback with visible warning

### Research Flags

- **Phase 2 start:** Verify Codex CLI MCP config key names at latest version before coding adapter
- **Phase 2 start:** Verify Copilot CLI vs VS Code Copilot adapter divergence state
- **Phase 2 start:** Verify mcpmarket.com API pagination and rate limits (no official docs found)
- **Phase 3 (deferred to v2):** Per-assistant subagent invocation flags for background agent — only `claude --agent` confirmed

### Todos

- [ ] Create `agentkit-registry` GitHub repo with initial `registry.json` listing 9 bundled skills
- [x] Scaffold Go module with `cobra` + `bubbletea` + `go-retryablehttp` dependencies (01-01, commit 4335d31)
- [x] Define domain types: `Package`, `Manifest`, `MCPServerEntry` — stabilize before any CLI commands (01-01, commit 4335d31)

### Blockers

None

---

## Session Continuity

### Last Session

**2026-06-08** — Completed 01-01-PLAN.md: Go module scaffold and domain types (commit 4335d31)

### Next Action

Execute plan 01-02 (next plan in phase 01-foundation).
