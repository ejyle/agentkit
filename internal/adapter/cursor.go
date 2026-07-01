package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/fileutil"
)

// CursorAdapter implements AssistantAdapter for the Cursor editor.
// MCP config is written to ~/.cursor/mcp.json using the mcpServers JSON key.
// Entries use plain command/args/env format — no "type" field (unlike Copilot).
// Skills are written to ~/.cursor/skills/<name>/ using the standard SKILL.md
// folder structure (same mechanism as Claude Code/Gemini/pi — Cursor's native
// Agent Skills feature, not Rules/.mdc).
type CursorAdapter struct {
	jsonMCPAdapter
}

// NewCursorAdapter returns a CursorAdapter using the real home directory.
func NewCursorAdapter(store *config.ConfigStore) *CursorAdapter {
	return NewCursorAdapterWithHome(store, "")
}

// NewCursorAdapterWithHome returns a CursorAdapter with an injected home directory.
// Used in tests to avoid reads/writes to the real ~/.cursor/mcp.json.
func NewCursorAdapterWithHome(store *config.ConfigStore, homeDir string) *CursorAdapter {
	return &CursorAdapter{
		jsonMCPAdapter: jsonMCPAdapter{
			store:   store,
			homeDir: homeDir,
			mcpKey:  "mcpServers",
			configPath: func(home string) (string, error) {
				return filepath.Join(home, ".cursor", "mcp.json"), nil
			},
			extraFields: nil, // Cursor format: plain command/args/env — no "type" field
		},
	}
}

// Name returns "cursor".
func (a *CursorAdapter) Name() string { return "cursor" }

// WriteSkill writes skill files into ~/.cursor/skills/<name>/.
func (a *CursorAdapter) WriteSkill(name string, files map[string][]byte) error {
	home, err := a.home()
	if err != nil {
		return err
	}
	skillPath := filepath.Join(home, ".cursor", "skills", name)
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

// RemoveSkill removes ~/.cursor/skills/<name>/ entirely.
func (a *CursorAdapter) RemoveSkill(name string) error {
	home, err := a.home()
	if err != nil {
		return err
	}
	skillPath := filepath.Join(home, ".cursor", "skills", name)
	return os.RemoveAll(skillPath)
}
