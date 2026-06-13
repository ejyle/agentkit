package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InstalledStatePath returns the path to the per-assistant installed.json file.
// Path: <UserConfigDir>/agentkit/<target>/installed.json
func InstalledStatePath(target string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "agentkit", target, "installed.json"), nil
}

// ManifestCachePath returns the path to the cached manifest for a registry.
// Path: <UserCacheDir>/agentkit/<registryID>/manifest.json
func ManifestCachePath(registryID string) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "agentkit", registryID, "manifest.json"), nil
}

// AgentBinPath returns the directory for agentkit-managed binaries.
// Path: <UserConfigDir>/agentkit/bin
func AgentBinPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "agentkit", "bin"), nil
}

// TarballCachePath returns the path for a cached GitHub release tarball.
// Path: <UserCacheDir>/agentkit/releases/<repo-slug>/<version>/tarball.tar.gz
func TarballCachePath(repo, version string) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	slug := strings.ReplaceAll(repo, "/", "-")
	return filepath.Join(base, "agentkit", "releases", slug, version, "tarball.tar.gz"), nil
}

// SkillInstallPath returns the install path for a named skill for a target assistant.
// For claude: <UserHomeDir>/.claude/skills/<name>
func SkillInstallPath(target, name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch target {
	case "claude":
		return filepath.Join(home, ".claude", "skills", name), nil
	case "gemini":
		return filepath.Join(home, ".gemini", "skills", name), nil
	case "pi":
		// pi (personal intelligence) skills live under ~/.agents/skills/<name> per D-11.
		return filepath.Join(home, ".agents", "skills", name), nil
	case "copilot-cli", "copilot-vscode", "codex", "opencode":
		return "", fmt.Errorf("SkillInstallPath: %q has no user-global skill directory", target)
	default:
		return filepath.Join(home, "."+target, "skills", name), nil
	}
}
