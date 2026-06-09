package adapter

import (
	"fmt"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// PiAdapter implements AssistantAdapter for Pi (pi.ai).
// Full implementation is delivered in plan 02-03.
// This scaffold allows factory.go to compile in the parallel wave-2 execution.
type PiAdapter struct {
	store *config.ConfigStore
}

// NewPiAdapter returns a PiAdapter using the real home directory.
func NewPiAdapter(store *config.ConfigStore) *PiAdapter {
	return &PiAdapter{store: store}
}

func (a *PiAdapter) Name() string { return "pi" }

func (a *PiAdapter) WriteMCPConfig(_ domain.MCPServerEntry, _ *domain.InstalledRecord) error {
	return fmt.Errorf("pi adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *PiAdapter) RemoveMCPConfig(_ string) error {
	return fmt.Errorf("pi adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *PiAdapter) ReadMCPConfig() (map[string]domain.MCPServerEntry, error) {
	return nil, fmt.Errorf("pi adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *PiAdapter) WriteSkill(_ string, _ map[string][]byte) error {
	return fmt.Errorf("pi adapter: WriteSkill not supported: %w", ErrNotSupported)
}
func (a *PiAdapter) RemoveSkill(_ string) error {
	return fmt.Errorf("pi adapter: RemoveSkill not supported: %w", ErrNotSupported)
}
