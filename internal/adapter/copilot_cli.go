package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// CopilotCLIAdapter implements AssistantAdapter for GitHub Copilot CLI.
// It writes MCP server config to ~/.copilot/mcp-config.json under the "mcpServers" key,
// injecting the Copilot-CLI-required fields "type": "local" and "tools": ["*"] into
// each entry. Skill operations are not supported (Copilot CLI has no user-global skill dir).
type CopilotCLIAdapter struct {
	jsonMCPAdapter
}

// NewCopilotCLIAdapter returns a CopilotCLIAdapter using the real home directory.
func NewCopilotCLIAdapter(store *config.ConfigStore) *CopilotCLIAdapter {
	return newCopilotCLIAdapter(store, "")
}

// NewCopilotCLIAdapterWithHome returns a CopilotCLIAdapter with an injected home directory.
// Used in tests to avoid reads/writes to the real ~/.copilot/mcp-config.json.
func NewCopilotCLIAdapterWithHome(store *config.ConfigStore, homeDir string) *CopilotCLIAdapter {
	return newCopilotCLIAdapter(store, homeDir)
}

func newCopilotCLIAdapter(store *config.ConfigStore, homeDir string) *CopilotCLIAdapter {
	a := &CopilotCLIAdapter{}
	a.jsonMCPAdapter = jsonMCPAdapter{
		store:   store,
		homeDir: homeDir,
		mcpKey:  "mcpServers",
		configPath: func(home string) (string, error) {
			// Check $COPILOT_HOME env var first (T-02-04: treated as directory, not shell-exec).
			if copilotHome := os.Getenv("COPILOT_HOME"); copilotHome != "" {
				return filepath.Join(copilotHome, "mcp-config.json"), nil
			}
			return filepath.Join(home, ".copilot", "mcp-config.json"), nil
		},
		extraFields: func(_ domain.MCPServerEntry) map[string]interface{} {
			return map[string]interface{}{
				"type":  "local",
				"tools": []string{"*"},
			}
		},
	}
	return a
}

// Name returns "copilot-cli".
func (a *CopilotCLIAdapter) Name() string { return "copilot-cli" }

// WriteSkill returns ErrNotSupported — Copilot CLI has no user-global skill directory.
func (a *CopilotCLIAdapter) WriteSkill(_ string, _ map[string][]byte) error {
	return fmt.Errorf("copilot-cli adapter: WriteSkill not supported — Copilot CLI has no CLI-level skill directory: %w", ErrNotSupported)
}

// RemoveSkill returns ErrNotSupported — Copilot CLI has no user-global skill directory.
func (a *CopilotCLIAdapter) RemoveSkill(_ string) error {
	return fmt.Errorf("copilot-cli adapter: RemoveSkill not supported — Copilot CLI has no CLI-level skill directory: %w", ErrNotSupported)
}
