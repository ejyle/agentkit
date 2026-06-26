# install.ps1 — PowerShell installer for agentkit (Windows)
#
# Usage (run in PowerShell as current user — no admin required):
#   irm https://raw.githubusercontent.com/ejyle/agentkit/main/scripts/install.ps1 | iex
#
# Override version:
#   $env:AGENTKIT_VERSION = "0.2.0"; irm ... | iex
#
# Installs to: %LOCALAPPDATA%\Programs\agentkit\agentkit.exe
# Adds install directory to the current user's PATH (registry, no admin needed).

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

function Resolve-LatestVersion {
    $apiUrl = 'https://api.github.com/repos/ejyle/agentkit/releases/latest'
    try {
        $release = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        return $release.tag_name -replace '^v', ''
    } catch {
        Write-Error "Could not determine latest version from GitHub API: $_"
        exit 1
    }
}

$version = if ($env:AGENTKIT_VERSION) { $env:AGENTKIT_VERSION } else {
    Write-Host 'Detecting latest version...'
    Resolve-LatestVersion
}

$arch     = 'amd64'
$filename = "agentkit_${version}_windows_${arch}.zip"
$baseUrl  = "https://github.com/ejyle/agentkit/releases/download/v${version}"

$installDir = Join-Path $env:LOCALAPPDATA 'Programs\agentkit'
$tmpDir     = Join-Path $env:TEMP "agentkit-install-$([System.Guid]::NewGuid().ToString('N'))"

New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
    $zipPath      = Join-Path $tmpDir $filename
    $checksumPath = Join-Path $tmpDir 'checksums.txt'

    Write-Host "Downloading agentkit $version (windows/$arch)..."
    Invoke-WebRequest -Uri "$baseUrl/$filename"     -OutFile $zipPath      -UseBasicParsing
    Invoke-WebRequest -Uri "$baseUrl/checksums.txt" -OutFile $checksumPath -UseBasicParsing

    # Verify SHA256 checksum before extracting
    Write-Host 'Verifying checksum...'
    $expected = (Select-String -Path $checksumPath -Pattern [regex]::Escape($filename)).Line -split '\s+' | Select-Object -First 1
    if (-not $expected) {
        Write-Error "Checksum entry for $filename not found in checksums.txt"
        exit 1
    }
    $actual = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLower()
    if ($actual -ne $expected.ToLower()) {
        Write-Error "Checksum mismatch! Expected: $expected  Got: $actual"
        exit 1
    }

    # Extract binary
    Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force

    $exeSrc = Join-Path $tmpDir 'agentkit.exe'
    if (-not (Test-Path $exeSrc)) {
        Write-Error "agentkit.exe not found in archive"
        exit 1
    }

    # Install
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    Move-Item -Path $exeSrc -Destination (Join-Path $installDir 'agentkit.exe') -Force

    # Add to User PATH (registry, no admin required)
    $userPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
    if ($userPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable('PATH', "$userPath;$installDir", 'User')
        $env:PATH = "$env:PATH;$installDir"
    }

    Write-Host ""
    Write-Host "agentkit $version installed to $installDir\agentkit.exe"
    Write-Host ""
    Write-Host "Restart your terminal (or open a new one) for PATH changes to take effect."
    Write-Host "Then run: agentkit --version"
} finally {
    Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
}
