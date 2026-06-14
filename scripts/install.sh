#!/usr/bin/env bash
# install.sh — curl|sh installer for agentkit
#
# Usage:
#   sh -c "$(curl -fsSL https://raw.githubusercontent.com/ejyle/agentkit/main/scripts/install.sh)"
#
# Override version:
#   AGENTKIT_VERSION=0.2.0 sh install.sh
#
# Installs to: ~/.local/bin/agentkit (no root/sudo required)

set -euo pipefail

_resolve_version() {
  local api_url="https://api.github.com/repos/ejyle/agentkit/releases/latest"
  local tag
  tag=$(curl -fsSL "$api_url" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/' | sed 's/^v//')
  if [ -z "$tag" ]; then
    printf 'Could not determine latest version from GitHub API\n' >&2
    exit 1
  fi
  printf '%s' "$tag"
}

VERSION="${AGENTKIT_VERSION:-}"
if [ -z "$VERSION" ]; then
  printf 'Detecting latest version...\n'
  VERSION=$(_resolve_version)
fi

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "${OS}" in
  linux)
    OS="linux"
    EXT="tar.gz"
    ;;
  darwin)
    OS="darwin"
    EXT="tar.gz"
    ;;
  *)
    printf 'Unsupported OS: %s\n' "${OS}" >&2
    exit 1
    ;;
esac

# Detect ARCH
ARCH=$(uname -m)
case "${ARCH}" in
  x86_64)
    ARCH="amd64"
    ;;
  aarch64 | arm64)
    ARCH="arm64"
    ;;
  *)
    printf 'Unsupported architecture: %s\n' "${ARCH}" >&2
    exit 1
    ;;
esac

# Detect checksum tool (macOS ships shasum; Linux ships sha256sum)
if command -v sha256sum >/dev/null 2>&1; then
  SHA_CMD="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  SHA_CMD="shasum -a 256"
else
  printf 'Cannot verify checksum — neither sha256sum nor shasum found\n' >&2
  exit 1
fi

FILENAME="agentkit_${VERSION}_${OS}_${ARCH}.${EXT}"
BASE_URL="https://github.com/ejyle/agentkit/releases/download/v${VERSION}"

# Create a secure temp directory and ensure it is cleaned up on exit
TMPDIR=$(mktemp -d)
trap 'rm -rf "${TMPDIR}"' EXIT

printf 'Downloading agentkit %s (%s/%s)...\n' "${VERSION}" "${OS}" "${ARCH}"
curl -fsSL "${BASE_URL}/${FILENAME}" -o "${TMPDIR}/${FILENAME}"
curl -fsSL "${BASE_URL}/checksums.txt" -o "${TMPDIR}/checksums.txt"

# Verify SHA256 checksum BEFORE any binary execution
printf 'Verifying checksum...\n'
grep "${FILENAME}" "${TMPDIR}/checksums.txt" | (cd "${TMPDIR}" && ${SHA_CMD} --check -)

# Extract binary (GoReleaser nests binary in agentkit_VERSION_OS_ARCH/agentkit)
tar -xzf "${TMPDIR}/${FILENAME}" -C "${TMPDIR}" agentkit

# Install to user-local bin (no sudo required — CLI-10)
INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "${INSTALL_DIR}"
mv "${TMPDIR}/agentkit" "${INSTALL_DIR}/agentkit"
chmod +x "${INSTALL_DIR}/agentkit"

printf 'Installed agentkit %s to %s\n' "${VERSION}" "${INSTALL_DIR}/agentkit"
printf 'Add to PATH: export PATH="$HOME/.local/bin:$PATH"\n'
