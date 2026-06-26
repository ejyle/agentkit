---
status: complete
---

# Quick Task 260626-qna Summary

**Task:** Add Windows install script to README — currently only Linux/macOS were covered
**Date:** 2026-06-26
**Commit:** 7ea5585

## What was done

1. **Created `scripts/install.ps1`** — Windows PowerShell install script that:
   - Detects latest version from GitHub Releases API (or uses `$env:AGENTKIT_VERSION` override)
   - Downloads `agentkit_{version}_windows_amd64.zip` from GitHub Releases
   - Verifies SHA256 checksum via `Get-FileHash` against `checksums.txt`
   - Extracts `agentkit.exe` to `%LOCALAPPDATA%\Programs\agentkit`
   - Adds install directory to User PATH via registry (`[Environment]::SetEnvironmentVariable`) — no admin required
   - Cleans up temp files on exit
   - Enforces TLS 1.2 for compatibility with older PowerShell 5

2. **Updated `README.md` Install section**:
   - Added Windows `irm | iex` one-liner
   - Added note about install location and auto-PATH
   - Added Scoop coming-soon stub (matching Homebrew treatment)
   - Fixed install.sh URL (`dev` → `main`)

## Decisions honored
- Install dir: `%LOCALAPPDATA%\Programs\agentkit` (no admin)
- Scoop stub: added as coming-soon alongside Homebrew
