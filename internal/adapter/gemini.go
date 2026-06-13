package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/fileutil"
)

// GeminiAdapter implements AssistantAdapter for Gemini CLI.
// MCP config is written to ~/.gemini/settings.json using the mcpServers JSON key.
// Entries use plain command/args/env format — no "type" field (unlike Copilot).
// Skills are written to ~/.gemini/skills/<name>/.
type GeminiAdapter struct {
	jsonMCPAdapter
}

// NewGeminiAdapter returns a GeminiAdapter using the real home directory.
func NewGeminiAdapter(store *config.ConfigStore) *GeminiAdapter {
	return NewGeminiAdapterWithHome(store, "")
}

// NewGeminiAdapterWithHome returns a GeminiAdapter with an injected home directory.
// Used in tests to avoid reads/writes to the real ~/.gemini/settings.json.
func NewGeminiAdapterWithHome(store *config.ConfigStore, homeDir string) *GeminiAdapter {
	return &GeminiAdapter{
		jsonMCPAdapter: jsonMCPAdapter{
			store:   store,
			homeDir: homeDir,
			mcpKey:  "mcpServers",
			configPath: func(home string) (string, error) {
				return filepath.Join(home, ".gemini", "settings.json"), nil
			},
			extraFields: nil, // Gemini format: plain command/args/env — no "type" field
		},
	}
}

// Name returns "gemini".
func (a *GeminiAdapter) Name() string { return "gemini" }

// WriteSkill writes skill files into ~/.gemini/skills/<name>/.
func (a *GeminiAdapter) WriteSkill(name string, files map[string][]byte) error {
	home, err := a.home()
	if err != nil {
		return err
	}
	skillPath := filepath.Join(home, ".gemini", "skills", name)
	if err := os.MkdirAll(skillPath, 0755); err != nil {
		return err
	}
	for filename, content := range files {
		dest := filepath.Join(skillPath, filename)
		if err := fileutil.WriteFile(dest, content, 0644); err != nil {
			return fmt.Errorf("writing skill file %s: %w", filename, err)
		}
	}
	return nil
}

// RemoveSkill removes ~/.gemini/skills/<name>/ entirely.
func (a *GeminiAdapter) RemoveSkill(name string) error {
	home, err := a.home()
	if err != nil {
		return err
	}
	skillPath := filepath.Join(home, ".gemini", "skills", name)
	return os.RemoveAll(skillPath)
}
