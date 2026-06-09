package adapter

import (
	"fmt"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// CopilotVSCodeAdapter implements AssistantAdapter for GitHub Copilot in VS Code.
// Full implementation is delivered in plan 02-02.
// This scaffold allows factory.go to compile in the parallel wave-2 execution.
type CopilotVSCodeAdapter struct {
	store *config.ConfigStore
}

// NewCopilotVSCodeAdapter returns a CopilotVSCodeAdapter using the real config directories.
func NewCopilotVSCodeAdapter(store *config.ConfigStore) *CopilotVSCodeAdapter {
	return &CopilotVSCodeAdapter{store: store}
}

func (a *CopilotVSCodeAdapter) Name() string { return "copilot-vscode" }

func (a *CopilotVSCodeAdapter) WriteMCPConfig(_ domain.MCPServerEntry, _ *domain.InstalledRecord) error {
	return fmt.Errorf("copilot-vscode adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *CopilotVSCodeAdapter) RemoveMCPConfig(_ string) error {
	return fmt.Errorf("copilot-vscode adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *CopilotVSCodeAdapter) ReadMCPConfig() (map[string]domain.MCPServerEntry, error) {
	return nil, fmt.Errorf("copilot-vscode adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *CopilotVSCodeAdapter) WriteSkill(_ string, _ map[string][]byte) error {
	return fmt.Errorf("copilot-vscode adapter: WriteSkill not supported: %w", ErrNotSupported)
}
func (a *CopilotVSCodeAdapter) RemoveSkill(_ string) error {
	return fmt.Errorf("copilot-vscode adapter: RemoveSkill not supported: %w", ErrNotSupported)
}
