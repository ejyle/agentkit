# Quick Task 260626-qna: Add Windows install script to README - Context

**Gathered:** 2026-06-26
**Status:** Ready for planning

<domain>
## Task Boundary

Add a Windows PowerShell install script (`scripts/install.ps1`) and update README.md to include Windows install instructions alongside the existing macOS/Linux section.

</domain>

<decisions>
## Implementation Decisions

### Install directory
- `$env:LOCALAPPDATA\Programs\agentkit` — standard Windows user-level location, no admin required

### PATH persistence
- Permanent via registry using `[Environment]::SetEnvironmentVariable` (User scope)
- Script must notify user to restart terminal after install

### Scoop stub
- Add a "coming soon" Scoop stub to README, matching Homebrew's existing treatment

### Claude's Discretion
- Script download URL format (follow GoReleaser convention: `agentkit_{version}_windows_amd64.zip`)
- Whether to use `irm | iex` pattern or a download-then-run approach
- Error handling and user feedback in the PowerShell script

</decisions>

<specifics>
## Specific Requirements

- GoReleaser builds `windows/amd64` (no arm64) and produces `.zip` archives
- Existing install.sh is in `scripts/install.sh` — PowerShell script goes in `scripts/install.ps1`
- README currently shows: macOS/Linux curl|sh, Homebrew (coming soon), Go install
- Windows section should slot in between macOS/Linux and Homebrew

</specifics>

<canonical_refs>
## Canonical References

- `scripts/install.sh` — existing Unix install script (pattern reference)
- `.goreleaser.yaml` — build/archive config (windows/amd64 zip)
- `README.md` — Install section to update

</canonical_refs>
