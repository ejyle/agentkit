# Phase 2: Multi-Assistant & Full Install - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-08
**Phase:** 2-Multi-Assistant & Full Install
**Areas discussed:** Pi adapter behavior, Custom registry model, Docker adapter config entry, Gemini WriteSkill convention

---

## Pi Adapter Behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Stub adapter | Implement stub returning ErrNotSupported for all operations | |
| Omit --target pi entirely | Don't register Pi as a valid target | |
| Best-effort partial adapter | Researcher discovers what Pi supports; implement those ops | ✓ |

**User's choice:** Best-effort partial adapter

| Option | Description | Selected |
|--------|-------------|----------|
| Return ErrNotSupported | Fail loudly with clear error message | ✓ |
| Silently no-op | Proceed without error, skip write | |
| Warn and continue | Print warning to stderr, exit code 0 | |

**User's choice:** Return ErrNotSupported

| Option | Description | Selected |
|--------|-------------|----------|
| WriteMCPConfig only | MCP config write highest priority | |
| WriteMCPConfig + WriteSkill | Full install support — both operations | ✓ |
| Researcher decides | Defer priority decision to researcher | |

**User's choice:** WriteMCPConfig + WriteSkill (if Pi has both systems)

| Option | Description | Selected |
|--------|-------------|----------|
| Runtime detect (recommended) | Check file existence order at runtime | ✓ |
| Single fixed path | Code one path if Pi is stable | |
| Researcher decides | Let researcher find path first | |

**User's choice:** Yes — runtime detect

---

## Custom Registry Model

| Option | Description | Selected |
|--------|-------------|----------|
| Curated + user extras | Official registry primary; user can add extras | |
| Full reversal — all sources equal | No ranking preference; all registries peers | |
| Scoped toggle | Custom sources opt-in via config flag | ✓ |

**User's choice:** Scoped toggle (initially)

| Option | Description | Selected |
|--------|-------------|----------|
| Official registry wins | Official entry takes precedence on name conflict | |
| Most recently added source wins | Last added registry wins | |
| Prompt user on conflict | Show both entries and ask | ✓ |

**User's choice:** Prompt user on conflict

**Notes:** User then changed direction entirely: "change of plans — we only have the curated skills/MCPs; if needed the user needs to raise a feature request to add it to the registry, we should do all the quality checks we do for our curated skills/MCPs etc." This reversed the scoped toggle answer and confirmed full curated-only model.

| Confirmation | Description | Selected |
|------|-------------|----------|
| Curated only, always | REG-03 and REG-04 out of v1 entirely | ✓ |
| Correct, keep REG-03 as deferred v2 | mcpmarket.com for future | |

**Final decision:** Curated-only model confirmed and extended. REG-03 and REG-04 removed from v1 scope entirely.

---

## Docker Adapter Config Entry

| Option | Description | Selected |
|--------|-------------|----------|
| command: docker, args: [run, -i, --rm, image:tag] | Direct invocation, no extra args | |
| command: docker, with extra env/volume args from manifest | Manifest can specify additional Docker args | ✓ |
| Wrapper shell script | Generate wrapper script in agentkit bin dir | |

**User's choice:** command: docker, with extra env/volume args from manifest

| Option | Description | Selected |
|--------|-------------|----------|
| docker pull at install time | Pull image during agentkit install | ✓ |
| Lazy pull — write config only | Docker pulls on first use | |
| docker pull --quiet + user notice | Pull at install time with spinner | |

**User's choice:** docker pull at install time

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — fail loudly with install instructions | ErrDockerNotFound with URL | ✓ |
| Yes — warn but continue | Write config even if Docker missing | |
| No check — let docker fail naturally | Skip availability check | |

**User's choice:** Yes — fail loudly with install instructions

| Option | Description | Selected |
|--------|-------------|----------|
| pip mirrors Docker pattern | pip install, command='python -m' | |
| Use uvx instead of pip | uvx for isolated Python MCP installs | ✓ |
| Discuss pip separately | Treat pip as separate area | |

**User's choice:** Use uvx instead of pip. MCP-02 is now a uvx installer, not pip.

| Option | Description | Selected |
|--------|-------------|----------|
| command: uvx, args: [package-name, ...] | uvx as both install and runtime command | ✓ (researcher to verify) |
| command: python, uvx as install-only | uvx installs but Python is the runtime | |
| Researcher decides | Let researcher check real Python MCP invocation | |

**User's choice:** command: uvx with args, but researcher to verify exact pattern

---

## Gemini WriteSkill Convention

| Option | Description | Selected |
|--------|-------------|----------|
| Rename SKILL.md to GEMINI.md at install | WriteSkill renames entrypoint file | |
| Copy as-is (no rename) | Write SKILL.md as SKILL.md | |
| Researcher decides | Verify what Gemini CLI looks for | ✓ |

**User's choice:** Researcher decides

| Option | Description | Selected |
|--------|-------------|----------|
| Researcher verifies exact Gemini skill path | Don't assume ~/.gemini/skills/ | ✓ |
| Assume ~/.gemini/skills/<name> | Use default pattern | |
| Add gemini case to SkillInstallPath now | Explicit case even if path matches default | |

**User's choice:** Researcher verifies the exact Gemini skill path

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — shared base struct (recommended) | Extract shared JSON merge logic | |
| Copy-paste per adapter | Each adapter self-contained | |
| Researcher decides | Verify Gemini format first | ✓ |

**User's choice:** Researcher decides (after verifying format divergence)

| Option | Description | Selected |
|--------|-------------|----------|
| Full OpenCode adapter required | Correct 'mcp' key, 'type' field, runtime detection | ✓ |
| Best-effort stub for OpenCode | Basic adapter with guessed schema | |
| Defer OpenCode to v2 | Drop AST-05 from Phase 2 | |

**User's choice:** Full OpenCode adapter required

---

## Claude's Discretion

- Whether to extract a shared JSON merge base struct — depends on researcher findings about adapter format divergence
- Exact uvx invocation args for common Python MCP packages — researcher to verify
- Copilot CLI vs VS Code divergence handling — if they diverge, implement as separate sub-adapters (copilot-cli, copilot-vscode)

## Deferred Ideas

- REG-03 (mcpmarket.com) — removed from v1 scope
- REG-04 (agentkit registry add/remove) — removed from v1 scope
- pip-based Python install — replaced by uvx
