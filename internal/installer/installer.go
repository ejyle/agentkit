// Package installer provides MCPInstaller implementations for npx and binary install methods.
package installer

import (
	"errors"
	"fmt"

	"github.com/ejyle/agentkit/internal/domain"
)

// Sentinel errors returned by installer implementations.
var (
	// ErrNodeNotFound is returned when Node.js is not found on PATH.
	ErrNodeNotFound = errors.New("node not found on PATH; install Node.js to use npx-based MCP servers")
	// ErrChecksumMismatch is returned when the downloaded binary's SHA256 does not match the manifest.
	ErrChecksumMismatch = errors.New("SHA256 checksum mismatch: downloaded file does not match registry manifest")
	// ErrInsecureURL is returned when a binary download URL uses a non-HTTPS scheme.
	ErrInsecureURL = errors.New("insecure download URL: only https:// URLs are allowed")
	// ErrUvxNotFound is returned when uvx is not found on PATH.
	ErrUvxNotFound = errors.New("uvx not found on PATH; install uv to use Python-based MCP servers: https://docs.astral.sh/uv/")
	// ErrDockerNotFound is returned when docker is not found on PATH.
	ErrDockerNotFound = errors.New("docker not found on PATH; install Docker: https://docs.docker.com/get-docker/")
	// ErrGitHubReleaseNotFound is returned when the GitHub release tarball is not found (404).
	ErrGitHubReleaseNotFound = errors.New("github-release: tarball not found; check version tag exists on GitHub")
)

// MCPInstaller is the interface for installing MCP server packages.
type MCPInstaller interface {
	// Install installs the package described by spec.
	Install(spec domain.InstallSpec) error
	// IsAvailable reports whether the underlying runtime (node, etc.) is on PATH.
	IsAvailable() bool
	// Method returns the install method this installer handles.
	Method() domain.InstallMethod
}

// NewInstaller returns the appropriate MCPInstaller for the given install method.
func NewInstaller(method domain.InstallMethod) (MCPInstaller, error) {
	switch method {
	case domain.InstallMethodNpx:
		return NewNpxInstaller(), nil
	case domain.InstallMethodBinary:
		return NewBinaryInstaller(), nil
	case domain.InstallMethodCustom:
		return NewCustomInstaller(), nil
	case domain.InstallMethodUvx:
		return NewUvxInstaller(), nil
	case domain.InstallMethodDocker:
		return NewDockerInstaller(), nil
	default:
		return nil, fmt.Errorf("unsupported install method: %q", method)
	}
}
