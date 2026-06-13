package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejyle/agentkit/internal/config"
)

// vsCodeEditions is the ordered list of VS Code edition directory names to check.
// Detection follows: "Code" → "Code - Insiders" → "code-server".
var vsCodeEditions = []string{"Code", "Code - Insiders", "code-server"}

// CopilotVSCodeAdapter implements AssistantAdapter for GitHub Copilot in VS Code.
// It writes MCP server config to the VS Code mcp.json file under the "servers" key
// (not "mcpServers"), with platform-aware path detection and edition fallback.
// Skill operations are not supported.
type CopilotVSCodeAdapter struct {
	jsonMCPAdapter
}

// NewCopilotVSCodeAdapter returns a CopilotVSCodeAdapter using os.UserConfigDir() as the
// VS Code config base directory.
func NewCopilotVSCodeAdapter(store *config.ConfigStore) *CopilotVSCodeAdapter {
	return newCopilotVSCodeAdapter(store, "")
}

// NewCopilotVSCodeAdapterWithHome returns a CopilotVSCodeAdapter with an injected home
// directory. The homeDir parameter is unused for VS Code (configDir is what matters),
// so callers that only need home injection should use NewCopilotVSCodeAdapterWithConfigDir.
func NewCopilotVSCodeAdapterWithHome(store *config.ConfigStore, _ string) *CopilotVSCodeAdapter {
	return newCopilotVSCodeAdapter(store, "")
}

// NewCopilotVSCodeAdapterWithConfigDir returns a CopilotVSCodeAdapter with an injected
// VS Code config base directory. Used in tests to avoid touching real ~/Library or ~/.config.
func NewCopilotVSCodeAdapterWithConfigDir(store *config.ConfigStore, configDir string) *CopilotVSCodeAdapter {
	return newCopilotVSCodeAdapter(store, configDir)
}

func newCopilotVSCodeAdapter(store *config.ConfigStore, configDir string) *CopilotVSCodeAdapter {
	a := &CopilotVSCodeAdapter{}
	a.jsonMCPAdapter = jsonMCPAdapter{
		store:       store,
		mcpKey:      "servers",
		extraFields: nil,
		configPath: func(_ string) (string, error) {
			base := configDir
			if base == "" {
				var err error
				base, err = os.UserConfigDir()
				if err != nil {
					return "", fmt.Errorf("copilot-vscode: cannot determine user config directory: %w", err)
				}
			}
			// Edition detection: check each edition dir; return first with an existing User/ dir.
			// If none found, default to "Code" (will be created on first write).
			for _, edition := range vsCodeEditions {
				userDir := filepath.Join(base, edition, "User")
				if _, err := os.Stat(userDir); err == nil {
					return filepath.Join(userDir, "mcp.json"), nil
				}
			}
			// No edition directory found — default to "Code".
			return filepath.Join(base, "Code", "User", "mcp.json"), nil
		},
	}
	return a
}

// Name returns "copilot-vscode".
func (a *CopilotVSCodeAdapter) Name() string { return "copilot-vscode" }

// WriteSkill returns ErrNotSupported — VS Code Copilot has no CLI-level skill directory.
func (a *CopilotVSCodeAdapter) WriteSkill(_ string, _ map[string][]byte) error {
	return fmt.Errorf("copilot-vscode adapter: WriteSkill not supported — VS Code Copilot has no CLI-level skill directory: %w", ErrNotSupported)
}

// RemoveSkill returns ErrNotSupported — VS Code Copilot has no CLI-level skill directory.
func (a *CopilotVSCodeAdapter) RemoveSkill(_ string) error {
	return fmt.Errorf("copilot-vscode adapter: RemoveSkill not supported — VS Code Copilot has no CLI-level skill directory: %w", ErrNotSupported)
}
