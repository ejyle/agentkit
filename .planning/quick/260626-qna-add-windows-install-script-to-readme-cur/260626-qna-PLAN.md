---
quick_id: 260626-qna
slug: add-windows-install-script-to-readme-cur
description: Add Windows install script to README — currently only Linux/macOS are covered
date: 2026-06-26
must_haves:
  truths:
    - scripts/install.ps1 exists and follows the same structure as install.sh
    - README.md Install section includes Windows PowerShell one-liner
    - README.md includes Scoop stub (coming soon) matching Homebrew's treatment
    - install.ps1 installs to %LOCALAPPDATA%\Programs\agentkit
    - install.ps1 adds install dir to User PATH permanently via registry
    - install.ps1 verifies SHA256 checksum before executing the binary
  artifacts:
    - scripts/install.ps1
    - README.md (updated Install section)
  key_links:
    - scripts/install.sh (pattern reference)
    - .goreleaser.yaml (archive format: zip for windows/amd64)
---

# Plan: Add Windows Install Script

## Task 1 — Create scripts/install.ps1

**files:** `scripts/install.ps1`
**action:** Create PowerShell install script mirroring install.sh behavior for Windows
**verify:** File exists, contains GitHub API version detection, SHA256 verification, Expand-Archive, SetEnvironmentVariable PATH update
**done:** `scripts/install.ps1` created with correct content

Script requirements:
- Detect latest version from GitHub Releases API
- Allow override via `$env:AGENTKIT_VERSION`
- Download `agentkit_{version}_windows_amd64.zip` from GitHub Releases
- Download `checksums.txt` and verify SHA256 with `Get-FileHash`
- Extract `agentkit.exe` to `$env:LOCALAPPDATA\Programs\agentkit`
- Add install dir to User PATH via `[Environment]::SetEnvironmentVariable`
- Print instructions to restart terminal

## Task 2 — Update README.md Install section

**files:** `README.md`
**action:** Add Windows PowerShell install block and Scoop stub to the Install section
**verify:** README Install section has Windows irm|iex one-liner and Scoop coming-soon entry
**done:** README.md updated with Windows instructions

Changes:
- Add Windows block after macOS/Linux block
- Add Scoop stub (coming soon) matching Homebrew's format
