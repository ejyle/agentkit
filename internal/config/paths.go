package config

import (
	"os"
	"path/filepath"
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
	default:
		return filepath.Join(home, "."+target, "skills", name), nil
	}
}
