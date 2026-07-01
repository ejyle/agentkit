---
phase: 260701-hsh-add-support-for-cursor-ai-as-well
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/adapter/cursor.go
  - internal/adapter/cursor_test.go
  - internal/adapter/factory.go
  - cmd/root.go
  - cmd/doctor.go
  - README.md
autonomous: false
requirements: []

must_haves:
  truths:
    - "Running `agentkit install <pkg> --target cursor` writes an MCP server entry to ~/.cursor/mcp.json under the mcpServers key with no extra type field"
    - "Running `agentkit install <skill> --target cursor` writes a standard SKILL.md-based skill folder to ~/.cursor/skills/<name>/, identical in shape to the Claude/Gemini/pi skill folders"
    - "Passing --target cursor no longer fails root command validation"
    - "agentkit doctor reports whether ~/.cursor/ exists, matching the pattern used for other assistants"
  artifacts:
    - internal/adapter/cursor.go
    - internal/adapter/cursor_test.go
  key_links:
    - "cmd/root.go validTargets includes \"cursor\" -> internal/adapter/factory.go NewAdapter(\"cursor\", ...) routes to NewCursorAdapter -> internal/adapter/cursor.go CursorAdapter embeds jsonMCPAdapter with configPath ~/.cursor/mcp.json and WriteSkill/RemoveSkill targeting ~/.cursor/skills/<name>/"
---

<objective>
Add Cursor AI (the Cursor editor) as a new supported assistant target in agentkit, following the exact adapter pattern already established for Gemini (MCP config via embedded `jsonMCPAdapter`, skills via standard `SKILL.md` folders under a user-home path).

Purpose: Extend agentkit's multi-assistant support to Cursor so users can `agentkit install <pkg> --target cursor` and get both MCP server registration (`~/.cursor/mcp.json`) and skill installs (`~/.cursor/skills/<name>/`) working identically to Claude Code/Gemini/pi.

Output: `internal/adapter/cursor.go` (new `CursorAdapter`), `internal/adapter/cursor_test.go` (mirrors `gemini_test.go`), `"cursor"` wired into `cmd/root.go` validTargets and `internal/adapter/factory.go`, `cmd/doctor.go` assistant-dir check, and README target-list updates.
</objective>

<execution_context>
@$HOME/.claude/gsd-core/workflows/execute-plan.md
@$HOME/.claude/gsd-core/templates/summary.md
</execution_context>

<context>
@.planning/quick/260701-hsh-add-support-for-cursor-ai-as-well/260701-hsh-CONTEXT.md
@.planning/quick/260701-hsh-add-support-for-cursor-ai-as-well/260701-hsh-RESEARCH.md
@internal/adapter/gemini.go
@internal/adapter/gemini_test.go
@internal/adapter/jsonbase.go
@internal/adapter/factory.go
@cmd/root.go
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Create CursorAdapter with MCP config and skill support</name>
  <files>internal/adapter/cursor.go, internal/adapter/cursor_test.go</files>
  <behavior>
    Mirror `internal/adapter/gemini_test.go` exactly (same test names with s/Gemini/Cursor/, same table shapes), adjusted only for Cursor's paths:
    - Test 1: `TestCursorAdapter_WriteMCPConfig_CreatesFile` — empty tmpHome, WriteMCPConfig, expect file at `~/.cursor/mcp.json`, `mcpServers` key present, entry has NO `type` field.
    - Test 2: `TestCursorAdapter_WriteMCPConfig_PreservesExistingKeys` — pre-existing top-level keys (e.g. `theme`, `otherKey`) survive a write.
    - Test 3: `TestCursorAdapter_WriteMCPConfig_ErrForeignConflict` — key exists in file but not recorded in the ConfigStore -> `*adapter.ErrForeignConflict` via `adapter.AsErrForeignConflict`.
    - Test 4: `TestCursorAdapter_WriteMCPConfig_AutoOverwrite` — key exists AND is agentkit-owned (pre-recorded via `store.RecordInstalled`) -> overwrite succeeds, new args land.
    - Test 5: `TestCursorAdapter_ReadMCPConfig_AfterWrite` — round-trip read returns the written entry's Command.
    - Test 6: `TestCursorAdapter_RemoveMCPConfig` — removes only the named key; sibling `other-server` key and unrelated top-level `theme` key are untouched.
    - Test 7: `TestCursorAdapter_WriteSkill_CreatesDirectory` — `WriteSkill("my-skill", map[string][]byte{"SKILL.md": content})` writes to `~/.cursor/skills/my-skill/SKILL.md` with matching content.
    - Test 8: `TestCursorAdapter_RemoveSkill` — WriteSkill then RemoveSkill removes the `~/.cursor/skills/<name>/` directory entirely (`os.IsNotExist` on stat).
  </behavior>
  <action>
    RED: Create `internal/adapter/cursor_test.go` in package `adapter_test`, copying the structure/helpers of `gemini_test.go` (`makeCursorAdapter(t, tmpHome)` helper building a `config.ConfigStore` at `<tmpHome>/config/agentkit/cursor/installed.json`, plus a `cursorEntry()` helper returning a standard `domain.MCPServerEntry{Name: "my-server", Command: "uvx", Args: []string{"mcp-server-fetch"}}`). Reference `adapter.NewCursorAdapterWithHome` and `adapter.CursorAdapter`, which do not exist yet. Run tests; confirm compile failure (RED, since the type doesn't exist).

    GREEN: Create `internal/adapter/cursor.go` implementing `CursorAdapter` by embedding `jsonMCPAdapter`, following the `gemini.go` pattern exactly (per D of RESEARCH.md Pattern 1 and CONTEXT.md's corrected skills decision):
    - `type CursorAdapter struct { jsonMCPAdapter }`
    - `NewCursorAdapter(store *config.ConfigStore) *CursorAdapter` delegates to `NewCursorAdapterWithHome(store, "")`
    - `NewCursorAdapterWithHome(store *config.ConfigStore, homeDir string) *CursorAdapter` sets `mcpKey: "mcpServers"`, `configPath` returning `filepath.Join(home, ".cursor", "mcp.json")`, `extraFields: nil` (Cursor's MCP entry shape has no extra `type` field, matching Gemini not Copilot/OpenCode).
    - `func (a *CursorAdapter) Name() string { return "cursor" }`
    - `WriteSkill(name string, files map[string][]byte) error` — writes each file under `~/.cursor/skills/<name>/`, creating the directory via `os.MkdirAll(skillPath, 0755)` and writing each file via `fileutil.WriteFile`, identical logic to `GeminiAdapter.WriteSkill` (per CONTEXT.md's corrected decision: standard SKILL.md folder, no `.mdc`/Rules conversion).
    - `RemoveSkill(name string) error` — `os.RemoveAll(filepath.Join(home, ".cursor", "skills", name))`, identical to `GeminiAdapter.RemoveSkill`.
    Add a doc comment above the struct stating: MCP config written to `~/.cursor/mcp.json` using the `mcpServers` key; entries use plain command/args/env format with no extra field; skills written to `~/.cursor/skills/<name>/` using the standard SKILL.md folder structure (same mechanism as Claude Code/Gemini/pi — Cursor's native Agent Skills feature, not Rules/`.mdc`).
    Run tests again; confirm all 8 pass (GREEN).

    REFACTOR: none expected — this is a direct structural mirror of an existing, tested adapter with no novel logic.
  </action>
  <verify>
    <automated>go test ./internal/adapter/... -run TestCursorAdapter -v</automated>
  </verify>
  <done>internal/adapter/cursor.go exists with CursorAdapter (embeds jsonMCPAdapter, WriteSkill/RemoveSkill for ~/.cursor/skills/), internal/adapter/cursor_test.go exists with 8 passing tests mirroring gemini_test.go, `go test ./internal/adapter/... -run TestCursorAdapter -v` passes with 0 failures</done>
</task>

<task type="auto">
  <name>Task 2: Wire "cursor" into factory, root command validation, and doctor</name>
  <files>internal/adapter/factory.go, cmd/root.go, cmd/doctor.go</files>
  <action>
    In `internal/adapter/factory.go`: add `case "cursor": return NewCursorAdapter(store), nil` to the switch in `NewAdapter`. Update the function's doc comment ("Supported targets: ...") and the `default` case's error message to include `cursor` in the target list, keeping the existing alphabetical-ish ordering convention used in the file (append after `pi`, before `codex`, matching current list order: claude, copilot-cli, copilot-vscode, gemini, pi, codex, opencode, cursor).

    In `cmd/root.go`: add `"cursor"` to the `validTargets` slice. Update the `--target` flag's help string and the `PersistentPreRunE` invalid-target error message to list `cursor` alongside the existing targets, keeping consistent ordering with `validTargets`.

    In `cmd/doctor.go`'s `checkAssistantDirs`: add a new entry to the `entries` slice with `dirLabel: "~/.cursor/"`, `path: filepath.Join(home, ".cursor")`, `descName: "Cursor"`, placed after the `~/.codex/` entry and before the OpenCode entry (matching the existing per-assistant ordering pattern in that slice).

    Do NOT touch `internal/service/install.go`'s `targetFlag` function — it has no Cursor-specific external-installer flag concept in RESEARCH.md/CONTEXT.md, and its existing `default: return "--claude"` fallback already handles the "cursor" case safely with no regression for other targets.
  </action>
  <verify>
    <automated>go build ./... &amp;&amp; go vet ./... &amp;&amp; go test ./cmd/... ./internal/adapter/... -v</automated>
  </verify>
  <done>`agentkit install &lt;pkg&gt; --target cursor` no longer fails PersistentPreRunE validation; `adapter.NewAdapter("cursor", store)` returns a *CursorAdapter with no error; `agentkit doctor` includes a "~/.cursor/" check result; go build/vet/test all pass with zero errors</done>
</task>

<task type="checkpoint:human-verify" gate="blocking">
  <name>Checkpoint: Verify Cursor adapter wiring end-to-end</name>
  <what-built>
    Cursor AI added as a fully-wired agentkit target: CursorAdapter (MCP config + skill install), factory routing, CLI target validation, and doctor health check. Also confirm README target-list wording was updated (Task 3 covers README; verify here alongside the adapter behavior).
  </what-built>
  <how-to-verify>
    1. Run `go build -o /tmp/agentkit-cursor-check ./...` from the repo root — confirm it builds cleanly.
    2. Run `go test ./internal/adapter/... -run TestCursorAdapter -v` — confirm all 8 tests pass.
    3. Run `go test ./cmd/... ./internal/adapter/...` — confirm the full package suite is still green (no regressions to existing claude/gemini/copilot/codex/opencode/pi adapters).
    4. Run `/tmp/agentkit-cursor-check --target cursor --help` — confirm it does NOT print the "invalid target" error.
    5. Skim `README.md`'s "Supported targets" line and feature bullet — confirm `cursor` is listed alongside claude/copilot-cli/copilot-vscode/codex/gemini/opencode/pi.
  </how-to-verify>
  <resume-signal>Type "approved" or describe issues</resume-signal>
</task>

<task type="auto">
  <name>Task 3: Update README target documentation</name>
  <files>README.md</files>
  <action>
    Add "Cursor" to the top feature bullet listing supported assistants ("Supports Claude Code, GitHub Copilot (CLI + VS Code), Codex, Gemini CLI, OpenCode" -> append ", Cursor"). Update the intro paragraph's assistant list similarly. Update the "Supported targets:" line under "### Install a package" to include `cursor` in the backtick-separated list, keeping the existing comma-separated backtick format and ordering convention (append after `pi`).
  </action>
  <verify>
    <automated>grep -c 'cursor' README.md</automated>
  </verify>
  <done>README.md lists "cursor" as a supported target in both the features bullet and the "Supported targets:" line under Install a package</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| CLI arg -> adapter | `--target cursor` string flows from user-controlled CLI flag into `NewAdapter` switch and filesystem path construction |
| Skill/MCP name -> filesystem path | Package/skill names flow into `filepath.Join(home, ".cursor", "skills", name)` and `mcpServers[entry.Name]` |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-260701-01 | Tampering | internal/adapter/cursor.go WriteSkill | low | accept | Skill/package names are already sanitized to `[a-zA-Z0-9_-]` upstream before reaching any adapter (per STATE.md's existing "Registry ID sanitized" decision, confirmed in RESEARCH.md Security Domain) — CursorAdapter introduces no new, divergent sanitization path; it reuses the same upstream-validated `name` parameter that GeminiAdapter/ClaudeCodeAdapter already trust |
| T-260701-02 | Tampering | internal/adapter/cursor.go WriteMCPConfig | low | accept | Reuses `jsonMCPAdapter`'s existing atomic-write-via-renameio and `ErrForeignConflict` ownership check (D-07/D-08) unchanged — no new write logic introduced, identical trust posture to GeminiAdapter |
</threat_model>

<verification>
Run `go build ./...`, `go vet ./...`, and `go test ./...` from the repo root. Confirm `TestCursorAdapter_*` (8 tests) all pass, and confirm no existing adapter test (claude/gemini/copilot/codex/opencode/pi) regresses. Manually confirm `cursor` appears in `cmd/root.go` validTargets, `internal/adapter/factory.go`'s switch, `cmd/doctor.go`'s assistant dir list, and `README.md`.
</verification>

<success_criteria>
- `internal/adapter/cursor.go` exists, implementing `AssistantAdapter` for Cursor via embedded `jsonMCPAdapter` (MCP config at `~/.cursor/mcp.json`, `mcpServers` key, no extra fields) plus standalone `WriteSkill`/`RemoveSkill` targeting `~/.cursor/skills/<name>/`.
- `internal/adapter/cursor_test.go` exists with 8 tests mirroring `gemini_test.go`'s coverage, all passing.
- `cmd/root.go`'s `validTargets` includes `"cursor"`; `--target cursor` passes validation.
- `internal/adapter/factory.go`'s `NewAdapter("cursor", store)` returns a working `*CursorAdapter`.
- `cmd/doctor.go` reports on `~/.cursor/` existence.
- `README.md` documents `cursor` as a supported target.
- Full test suite (`go test ./...`) is green.
</success_criteria>

<output>
Create `.planning/quick/260701-hsh-add-support-for-cursor-ai-as-well/260701-hsh-SUMMARY.md` when done
</output>
