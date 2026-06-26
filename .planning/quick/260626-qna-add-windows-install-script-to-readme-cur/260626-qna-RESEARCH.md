# Research: Windows Install Script for agentkit

## Key Findings

### install.sh Pattern (existing)
- Detects version from GitHub Releases API (`/repos/ejyle/agentkit/releases/latest`)
- Downloads `agentkit_{version}_{os}_{arch}.{ext}` from GitHub Releases
- Verifies SHA256 against `checksums.txt`
- Extracts binary, installs to `~/.local/bin`
- Prints PATH hint for user

### Windows-specific Facts
- GoReleaser builds `windows/amd64` only (no arm64)
- Windows archive format: `.zip` (GoReleaser `format_overrides`)
- Binary name inside zip: `agentkit.exe`
- Install location decision: `$env:LOCALAPPDATA\Programs\agentkit`

### PowerShell Best Practices (irm | iex pattern)
```powershell
irm https://raw.githubusercontent.com/ejyle/agentkit/main/scripts/install.ps1 | iex
```
- `Invoke-RestMethod` (irm) fetches the script; `iex` executes it
- Standard pattern: winget, Scoop, oh-my-posh all use this
- Need `Set-ExecutionPolicy RemoteSigned -Scope CurrentUser` preamble

### Checksum verification on Windows
- `Get-FileHash -Algorithm SHA256` is built-in PowerShell — no extra tools needed
- Parse `checksums.txt` with `-match` or `Select-String`

### PATH persistence on Windows
- `[Environment]::SetEnvironmentVariable('PATH', newPath, 'User')` writes to HKCU registry
- Survives shell restarts; no admin required
- Must refresh current session: `$env:PATH = $env:PATH + ";$InstallDir"`

### Pitfalls
- Zip extraction: use `Expand-Archive -DestinationPath` (PowerShell 5+ built-in)
- TLS: PowerShell 5 may need `[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12`
- GoReleaser zip contains binary at root (not nested) for windows — verify with `agentkit.exe` path
