package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/google/renameio/v2"
)

// PiAdapter implements AssistantAdapter for Pi (Personal Intelligence).
// MCP config is written to ~/.pi/agent/mcp.json using the mcpServers JSON key.
// Entries use plain command/args/env format — no "type" field.
// Skills are written to ~/.agents/skills/<name>/ per D-05 and D-11.
//
// Note: Pi skills live under ~/.agents/skills/ (not ~/.pi/skills/) to match the
// user-global agents skill convention (A4 assumption from RESEARCH.md).
// ErrNotSupported is NOT used — both WriteMCPConfig and WriteSkill are fully implemented.
type PiAdapter struct {
	jsonMCPAdapter
}

// NewPiAdapter returns a PiAdapter using the real home directory.
func NewPiAdapter(store *config.ConfigStore) *PiAdapter {
	return NewPiAdapterWithHome(store, "")
}

// NewPiAdapterWithHome returns a PiAdapter with an injected home directory.
// Used in tests to avoid reads/writes to the real ~/.pi/agent/mcp.json.
func NewPiAdapterWithHome(store *config.ConfigStore, homeDir string) *PiAdapter {
	return &PiAdapter{
		jsonMCPAdapter: jsonMCPAdapter{
			store:   store,
			homeDir: homeDir,
			mcpKey:  "mcpServers",
			configPath: func(home string) (string, error) {
				return filepath.Join(home, ".pi", "agent", "mcp.json"), nil
			},
			extraFields: nil, // Pi format: plain command/args/env — no "type" field
		},
	}
}

// Name returns "pi".
func (a *PiAdapter) Name() string { return "pi" }

// WriteSkill writes skill files into ~/.agents/skills/<name>/.
// Pi skills use the ~/.agents/ namespace (not ~/.pi/), per D-11.
func (a *PiAdapter) WriteSkill(name string, files map[string][]byte) error {
	home, err := a.home()
	if err != nil {
		return err
	}
	skillPath := filepath.Join(home, ".agents", "skills", name)
	if err := os.MkdirAll(skillPath, 0755); err != nil {
		return err
	}
	for filename, content := range files {
		dest := filepath.Join(skillPath, filename)
		if err := renameio.WriteFile(dest, content, 0644); err != nil {
			return fmt.Errorf("writing skill file %s: %w", filename, err)
		}
	}
	return nil
}

// RemoveSkill removes ~/.agents/skills/<name>/ entirely.
func (a *PiAdapter) RemoveSkill(name string) error {
	home, err := a.home()
	if err != nil {
		return err
	}
	skillPath := filepath.Join(home, ".agents", "skills", name)
	return os.RemoveAll(skillPath)
}
