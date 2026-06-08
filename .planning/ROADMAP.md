# Roadmap: agentkit

_Version: 0.1.0 | Generated: 2026-06-08_

---

## Phases

- [ ] **Phase 1: Foundation** - CLI core, Claude Code adapter, and default registries deliver `agentkit install <name> --target claude` end-to-end
- [ ] **Phase 2: Multi-Assistant & Full Install** - All 5+ target assistants and all 4 MCP install methods; complete CLI surface (search/update/uninstall); all registries live
- [ ] **Phase 3: Bundled Skills** - All 9 initial skills authored and validated; `--bundle` command delivers one-command environment setup
- [ ] **Phase 4: Distribution & Hardening** - GoReleaser v2 pipeline, Homebrew tap, `agentkit doctor`, cross-platform release; public v0.1.0 shipped

---

## Phase Details

### Phase 1: Foundation

**Goal:** Users can install a skill or MCP server by name targeting Claude Code, see what is installed, and uninstall cleanly — using a locally-built binary with no runtime dependency.
**Mode:** mvp
**Depends on:** Nothing
**Requirements:** CLI-01, CLI-02, CLI-05, CLI-06, CLI-07, CLI-08, REG-01, REG-02, REG-05, REG-06, AST-01, MCP-01, MCP-03, MCP-05, MCP-06, MCP-07, SKL-01, SKL-02, SKL-03
**Note:** CLI-09 (registry add/remove) deferred per D-01/D-02 (curated-only model in v1).
**Success Criteria**:

  1. `agentkit install playwright --target claude` resolves the package from the curated agentkit-registry, runs the npx or binary install adapter, writes the correct MCP entry to `~/.claude.json` (runtime-detected path), and prints a success line with version and target.
  2. `agentkit list` shows every installed package with its version, source registry, and target assistant — no installed package is missing from the output.
  3. `agentkit search playwright` returns ranked results from the agentkit-registry with source labels.
  4. `agentkit uninstall playwright` removes the skill or MCP entry with no leftover artifacts.
  5. `agentkit update playwright` checks for newer version and upgrades or confirms up-to-date.

**Plans:** 5/6 plans executed

Plans:
**Wave 1**

- [x] 01-01-PLAN.md — Go scaffold, cobra CLI wiring, domain types (Package, InstalledRecord)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 01-02-PLAN.md — Config store (installed.json CRUD) and registry client (GitHub manifest + ETag cache)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 01-03-PLAN.md — Install slice: MCP installers, ClaudeCodeAdapter, InstallService, agentkit install command

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 01-04-PLAN.md — Read commands: agentkit list and agentkit search
- [x] 01-05-PLAN.md — Write commands: agentkit uninstall and agentkit update

**Wave 5** *(blocked on Wave 4 completion)*

- [ ] 01-06-PLAN.md — Human verification checkpoint: end-to-end Phase 1 success criteria

### Phase 2: Multi-Assistant & Full Install

**Goal:** Users can install skills and MCP servers targeting any of the 5 supported coding assistants using any of the 4 install methods, with full registry coverage including mcpmarket.com and custom sources.
**Mode:** mvp
**Depends on:** Phase 1
**Requirements:** AST-02, AST-03, AST-04, AST-05, AST-06, MCP-02, MCP-04, REG-03, REG-04
**Success Criteria**:

  1. `agentkit install <name> --target copilot` writes MCP config to the correct Copilot CLI path (runtime-detected, not hardcoded); same for `--target codex`, `--target gemini`, `--target opencode`, and `--target pi`.
  2. `agentkit install <name>` using a pip-based or Docker-based MCP server completes without error and produces a valid, re-parseable config entry in the target assistant's config file.
  3. `agentkit search <query>` returns results from mcpmarket.com registry when online; the command succeeds with a cache-fallback warning when mcpmarket.com is unreachable.
  4. A user-added custom registry (`agentkit registry add <url>`) is searched alongside built-in registries and its packages are installable.

**Plans**: TBD

### Phase 3: Bundled Skills

**Goal:** Users can install curated skill bundles with one command, and each of the 9 initial skills is authored to the agentskills.io spec and validated at install time.
**Mode:** mvp
**Depends on:** Phase 1
**Requirements:** CLI-03, CLI-04, BND-01, BND-02, BND-03, BND-04, BND-05, BND-06, BND-07, BND-08, BND-09
**Success Criteria**:

  1. `agentkit install gsd` installs the full GSD suite (skills + agents + config) in one command with a single success confirmation.
  2. `agentkit install --bundle cloud` installs the AWS, GCP, and Azure skills atomically; `--bundle dev` installs Playwright, GitHub, and CI/CD; `--bundle context` installs context-mode, RTK, and Serena.
  3. Each of the 9 bundled skills passes the install-time validator: `SKILL.md` present, line count checked (warning if >500), `references/` directory present for multi-domain skills.
  4. After installing the AWS skill, `~/.claude/skills/aws/SKILL.md` and `~/.claude/skills/aws/references/` exist with the expected domain reference files (ec2.md, s3.md, iam.md).

**Plans**: TBD
**UI hint**: no

### Phase 4: Distribution & Hardening

**Goal:** agentkit v0.1.0 is publicly releasable: cross-platform binaries built via GoReleaser v2, installable via Homebrew tap, with a `doctor` command that validates the local install environment.
**Mode:** mvp
**Depends on:** Phase 1, Phase 2, Phase 3
**Requirements:** CLI-10
**Success Criteria**:

  1. `brew install agentkit/tap/agentkit` succeeds on macOS and produces a working binary with `agentkit --version` returning the correct release tag.
  2. A GitHub Actions release workflow runs GoReleaser v2 on tag push and produces signed binaries for macOS (arm64/amd64), Linux (amd64/arm64), and Windows (amd64).
  3. `agentkit doctor` checks: binary is in PATH, `~/.agentkit/` config directory is writable, network can reach the default registry, and each target assistant's config directory exists — printing a pass/warn/fail line for each check.
  4. Direct binary download from the GitHub release page installs without a package manager or root/sudo on all three platforms.

**Plans**: TBD

---

## Progress Table

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 4/6 | In Progress|  |
| 2. Multi-Assistant & Full Install | 0/TBD | Not started | - |
| 3. Bundled Skills | 0/TBD | Not started | - |
| 4. Distribution & Hardening | 0/TBD | Not started | - |
