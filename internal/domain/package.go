// Package domain defines the core data types for agentkit.
// These types are the stable contract that all other packages depend on.
package domain

// InstallMethod describes how a package is installed.
type InstallMethod string

const (
	// InstallMethodNpx installs via npx (e.g. npx -y @playwright/mcp).
	InstallMethodNpx InstallMethod = "npx"
	// InstallMethodBinary downloads a pre-built binary.
	InstallMethodBinary InstallMethod = "binary"
	// InstallMethodCustom uses a custom install script.
	InstallMethodCustom InstallMethod = "custom"
	// InstallMethodUvx installs via uvx (e.g. uvx mcp-server-fetch).
	InstallMethodUvx InstallMethod = "uvx"
	// InstallMethodDocker installs via docker pull (e.g. docker pull ghcr.io/github/github-mcp-server).
	InstallMethodDocker InstallMethod = "docker"
	// InstallMethodGitHubRelease extracts a skill subdirectory from a GitHub release tarball.
	InstallMethodGitHubRelease InstallMethod = "github-release"
)

// PackageType categorises what an agentkit package provides.
type PackageType string

const (
	// PackageTypeMCP is an MCP server package.
	PackageTypeMCP PackageType = "mcp"
	// PackageTypeSkill is an AI agent skill package.
	PackageTypeSkill PackageType = "skill"
	// PackageTypeAgent is a coding agent package.
	PackageTypeAgent PackageType = "agent"
)

// InstallSpec holds the installation parameters for a package.
type InstallSpec struct {
	Method  InstallMethod `json:"method"`
	Package string        `json:"package,omitempty"`
	URL     string        `json:"url,omitempty"`
	Args    []string      `json:"args,omitempty"`
	// Repo and Path are used by the github-release install method only.
	// Example: Repo = "ejyle/agentkit", Path = "skills/aws"
	Repo string `json:"repo,omitempty"`
	Path string `json:"path,omitempty"`
	// SkillDir is the resolved target directory, set at runtime by service.Install() — not serialised.
	SkillDir string `json:"-"`
}

// MCPServerEntry is a single entry in an assistant's MCP server configuration.
type MCPServerEntry struct {
	Name    string            `json:"name,omitempty"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// Package is a single entry in the agentkit registry manifest.
type Package struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Type        PackageType    `json:"type"`
	Source      string         `json:"source"`
	Install     InstallSpec    `json:"install"`
	MCPEntry    MCPServerEntry `json:"mcp_entry,omitempty"`
	SHA256      string         `json:"sha256,omitempty"`
}

// Manifest is the top-level structure of a registry.json file.
type Manifest struct {
	Packages []Package `json:"packages"`
}
