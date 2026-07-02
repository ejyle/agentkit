---
gsd_state_version: 1.0
milestone: v0.1.0
milestone_name: milestone
status: completed
last_updated: "2026-07-02T05:35:12.003Z"
progress:
  total_phases: 4
  completed_phases: 4
  total_plans: 23
  completed_plans: 23
  percent: 100
---

# STATE: agentkit

_Last updated: 2026-07-01 - Completed quick task 260701-hsh: Add support for Cursor AI as well_

---

## Project Reference

**Core value:** Zero-friction install of the right skills, agents, and MCP servers for any AI coding assistant — one command gets you from bare machine to fully instrumented dev environment.

**Current focus:** Milestone complete

---

## Current Position

Phase: 04 (distribution-hardening) — EXECUTING
Plan: 1 of 6
**Phase:** 04
**Plan:** Not started
**Status:** Milestone complete
**Progress:** [██████████] 100%

```
Phase 1: Foundation          [██████████] 100% (complete 2026-06-08)
Phase 2: Multi-Assistant     [██████████] 100% (complete 2026-06-09)
Phase 3: Bundled Skills      [----------]   0%
Phase 4: Distribution        [----------]   0%
```

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Phases complete | 2 / 4 |
| Plans complete | 12 / 11 (Phase 1: 6, Phase 2: 5, Phase 3+: TBD) |
| Requirements mapped | 41 / 41 |
| Requirements validated | 22 / 41 (CLI-01, CLI-02, CLI-05, CLI-06, CLI-07, CLI-08, AST-01, AST-02, AST-03, AST-04, AST-05, AST-06, MCP-01, MCP-02, MCP-03, MCP-04, MCP-05, MCP-06, MCP-07, SKL-01, SKL-02, SKL-03) |

| Phase 02 P01 (infrastructure) | ~8min | 2 tasks | 6 files |
| Phase 02 P02 (copilot adapters) | ~6min | 2 tasks | 4 files |
| Phase 02 P03 (gemini+pi adapters) | ~8min | 2 tasks | 4 files |
| Phase 02 P04 (codex+opencode+factory) | ~12min | 2 tasks | 7 files |
| Phase 02 P05 (e2e verification) | ~5min | 2 tasks | 0 files |
| Phase 04-distribution-hardening P05 | 12 | 1 tasks | 1 files |

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

### Quick Tasks Completed

| # | Description | Date | Commit | Status | Directory |
|---|-------------|------|--------|--------|-----------|
| 260626-jvc | Create skill-finder skill that searches web for Claude Code skills and optionally installs them to project skill directory | 2026-06-26 | f75764b | Verified | [260626-jvc-create-skill-finder-skill-that-searches-](./quick/260626-jvc-create-skill-finder-skill-that-searches-/) |
| 260626-qna | Add Windows install script to README — currently only Linux/macOS are covered | 2026-06-26 | 7ea5585 | Complete | [260626-qna-add-windows-install-script-to-readme-cur](./quick/260626-qna-add-windows-install-script-to-readme-cur/) |
| 260701-hsh | Add support for Cursor AI as well — CursorAdapter (MCP config + skills), wired into factory/CLI/doctor/README | 2026-07-01 | 479b8dd | Verified | [260701-hsh-add-support-for-cursor-ai-as-well](./quick/260701-hsh-add-support-for-cursor-ai-as-well/) |

---

## Session Continuity

### Last Session

**2026-06-09** — Planned Phase 3 (03-bundled-skills): 6 plans in 4 waves. GitHubReleaseInstaller, --bundle command, 10 bundled skills (BND-01 through BND-09 + skill-author meta-skill), 10-12 external curated skills, goreleaser config update.

### Next Action

Execute Phase 3 (Bundled Skills): run all 6 plans — `/gsd:execute-phase 3`
