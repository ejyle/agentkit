---
gsd_state_version: 1.0
milestone: v0.1.0
milestone_name: milestone
status: executing
last_updated: "2026-06-09T05:09:33.615Z"
progress:
  total_phases: 4
  completed_phases: 1
  total_plans: 11
  completed_plans: 7
  percent: 25
---

# STATE: agentkit

_Last updated: 2026-06-09 (Phase 2 planned — 5 plans ready)_

---

## Project Reference

**Core value:** Zero-friction install of the right skills, agents, and MCP servers for any AI coding assistant — one command gets you from bare machine to fully instrumented dev environment.

**Current focus:** Phase 2 — multi assistant & full install

---

## Current Position

Phase: 01 (foundation) — EXECUTING
Plan: 1 of 5
**Phase:** 2
**Plan:** 1 of 5
**Status:** Ready to execute
**Progress:** [██████░░░░] 64%

```
Phase 1: Foundation          [████████░░] 83%
Phase 2: Multi-Assistant     [----------]  0%
Phase 3: Bundled Skills      [----------]  0%
Phase 4: Distribution        [----------]  0%
```

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Phases complete | 0 / 4 |
| Plans complete | 5 / 6 |
| Requirements mapped | 41 / 41 |
| Requirements validated | 15 / 41 (CLI-01, CLI-02, CLI-05, CLI-06, CLI-07, CLI-08, AST-01, MCP-01, MCP-03, MCP-05, MCP-06, MCP-07, SKL-01, SKL-02, SKL-03) |

---
| Phase 02-multi-assistant-full-install P01 | 236 | - tasks | - files |

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
| SearchService uses local searchRegistry interface | Avoids coupling service tests to real RegistryManager; enables in-memory mocks | Confirmed |
| RenderInstalledTable uses fixed column widths | PACKAGE=20, VERSION=10, TYPE=8, TARGET=12, REGISTRY=20 — go-list style, no terminal width dependency | Confirmed |
| ErrNotInstalled sentinel defined in service/uninstall.go | Shared by both UninstallService and UpdateService; avoids re-definition | Confirmed |
| installServiceAdapter bridges InstallService to updateInstaller interface | Avoids coupling cmd/update.go to service package internals | Confirmed |
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
- [x] Build list and search commands: SearchService, lipgloss table renderer (D-05/D-06), agentkit list + agentkit search (01-04)
- [x] Build uninstall and update commands: UninstallService (D-09), UpdateService (D-08), agentkit uninstall + agentkit update (01-05)

### Blockers

None

---

## Session Continuity

### Last Session

**2026-06-08** — Completed 01-05-PLAN.md: Uninstall and update commands — UninstallService (D-09 non-destructive, ErrNotInstalled), UpdateService (D-08 InstallService delegate, UpdateAll continues on error), agentkit uninstall + agentkit update wired. 65 tests pass.

### Next Action

Execute plan 01-06 (final plan in phase 01-foundation).
