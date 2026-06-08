---
gsd_state_version: 1.0
milestone: v0.1.0
milestone_name: milestone
status: executing
last_updated: "2026-06-08T16:00:00.000Z"
progress:
  total_phases: 4
  completed_phases: 0
  total_plans: 6
  completed_plans: 3
  percent: 50
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
**Plan:** 3 complete (01-03), starting 01-04
**Status:** Executing Phase 01
**Progress:** [█████░░░░░] 50%

```
Phase 1: Foundation          [█████░░░░░] 50%
Phase 2: Multi-Assistant     [----------]  0%
Phase 3: Bundled Skills      [----------]  0%
Phase 4: Distribution        [----------]  0%
```

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Phases complete | 0 / 4 |
| Plans complete | 3 / 6 |
| Requirements mapped | 41 / 41 |
| Requirements validated | 11 / 41 (CLI-01, CLI-02, AST-01, MCP-01, MCP-03, MCP-05, MCP-06, MCP-07, SKL-01, SKL-02, SKL-03) |

---

## Accumulated Context

### Key Decisions (pending confirmation)

| Decision | Rationale | Status |
|----------|-----------|--------|
| Go over Python for CLI | Single binary, no runtime, fast startup, easy cross-compile | Confirmed |
| Registry ID sanitized to [a-zA-Z0-9_-] before filesystem use | Prevents path traversal via malicious registry IDs (T-02-05) | Confirmed |
| Test-mode timeout detection disables retries | responseTimeout < 1s signals test mode; avoids retry multiplication in unit tests | Confirmed |
| User scope only (v1) | Covers majority use case; project scope adds complexity | Pending |
| InstallService uses local interface types for injectable mocks | Avoids coupling service tests to real filesystem/network; enables pure unit tests | Confirmed |
| InstallService skips WriteMCPConfig for skill packages | Skill packages use WriteSkill path; MCP entry not applicable to skill-type installs | Confirmed |
| charmbracelet/bubbles added for spinner component | SpinnerModel requires bubbles.spinner.Model; same Charm ecosystem as existing bubbletea/lipgloss | Confirmed |
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
- [x] Build ConfigStore (atomic installed.json CRUD) and GitHubManifestRegistry (ETag caching) (01-02, commits ead5b9b..76fdfb0)
- [x] Build install vertical slice: npx/binary installers, ClaudeCodeAdapter, InstallService, bubbletea spinner, agentkit install command (01-03, commits f439bb2..a6f33db)

### Blockers

None

---

## Session Continuity

### Last Session

**2026-06-08** — Completed 01-03-PLAN.md: Install vertical slice — npx/binary installers, ClaudeCodeAdapter (runtime path detection + atomic merge + post-install verify + foreign conflict), InstallService (9-step flow), bubbletea spinner, agentkit install command (commits f439bb2, cc807f6, 2220788, a6f33db)

### Next Action

Execute plan 01-04 (next plan in phase 01-foundation).
