package adapter

import (
	"fmt"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// CopilotCLIAdapter implements AssistantAdapter for GitHub Copilot CLI.
// Full implementation is delivered in plan 02-02.
// This scaffold allows factory.go to compile in the parallel wave-2 execution.
type CopilotCLIAdapter struct {
	store *config.ConfigStore
}

// NewCopilotCLIAdapter returns a CopilotCLIAdapter using the real home directory.
func NewCopilotCLIAdapter(store *config.ConfigStore) *CopilotCLIAdapter {
	return &CopilotCLIAdapter{store: store}
}

func (a *CopilotCLIAdapter) Name() string { return "copilot-cli" }

func (a *CopilotCLIAdapter) WriteMCPConfig(_ domain.MCPServerEntry, _ *domain.InstalledRecord) error {
	return fmt.Errorf("copilot-cli adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *CopilotCLIAdapter) RemoveMCPConfig(_ string) error {
	return fmt.Errorf("copilot-cli adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *CopilotCLIAdapter) ReadMCPConfig() (map[string]domain.MCPServerEntry, error) {
	return nil, fmt.Errorf("copilot-cli adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *CopilotCLIAdapter) WriteSkill(_ string, _ map[string][]byte) error {
	return fmt.Errorf("copilot-cli adapter: WriteSkill not supported: %w", ErrNotSupported)
}
func (a *CopilotCLIAdapter) RemoveSkill(_ string) error {
	return fmt.Errorf("copilot-cli adapter: RemoveSkill not supported: %w", ErrNotSupported)
}
