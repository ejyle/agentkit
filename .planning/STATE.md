---
gsd_state_version: 1.0
milestone: v0.1.0
milestone_name: milestone
status: Not started
last_updated: "2026-06-08T14:00:49.859Z"
progress:
  total_phases: 4
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# STATE: agentkit

_Last updated: 2026-06-08_

---

## Project Reference

**Core value:** Zero-friction install of the right skills, agents, and MCP servers for any AI coding assistant — one command gets you from bare machine to fully instrumented dev environment.

**Current focus:** Phase 1 — Foundation (CLI core + Claude Code adapter + default registries)

---

## Current Position

**Phase:** 1 — Foundation
**Plan:** None started
**Status:** Not started
**Progress:** [----------] 0%

```
Phase 1: Foundation          [----------]  0%
Phase 2: Multi-Assistant     [----------]  0%
Phase 3: Bundled Skills      [----------]  0%
Phase 4: Distribution        [----------]  0%
```

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Phases complete | 0 / 4 |
| Plans complete | 0 / 8 |
| Requirements mapped | 41 / 41 |
| Requirements validated | 0 / 41 |

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
- [ ] Scaffold Go module with `cobra` + `bubbletea` + `go-retryablehttp` dependencies
- [ ] Define domain types: `Package`, `Manifest`, `MCPServerEntry` — stabilize before any CLI commands

### Blockers

None

---

## Session Continuity

### Last Session

_No sessions yet._

### Next Action

Run `/gsd:plan-phase 1` to create the Phase 1 execution plan.
