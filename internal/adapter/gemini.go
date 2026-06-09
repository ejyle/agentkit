package adapter

import (
	"fmt"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// GeminiAdapter implements AssistantAdapter for Gemini CLI.
// Full implementation is delivered in plan 02-03.
// This scaffold allows factory.go to compile in the parallel wave-2 execution.
type GeminiAdapter struct {
	store *config.ConfigStore
}

// NewGeminiAdapter returns a GeminiAdapter using the real home directory.
func NewGeminiAdapter(store *config.ConfigStore) *GeminiAdapter {
	return &GeminiAdapter{store: store}
}

func (a *GeminiAdapter) Name() string { return "gemini" }

func (a *GeminiAdapter) WriteMCPConfig(_ domain.MCPServerEntry, _ *domain.InstalledRecord) error {
	return fmt.Errorf("gemini adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *GeminiAdapter) RemoveMCPConfig(_ string) error {
	return fmt.Errorf("gemini adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *GeminiAdapter) ReadMCPConfig() (map[string]domain.MCPServerEntry, error) {
	return nil, fmt.Errorf("gemini adapter: not yet implemented: %w", ErrNotSupported)
}
func (a *GeminiAdapter) WriteSkill(_ string, _ map[string][]byte) error {
	return fmt.Errorf("gemini adapter: WriteSkill not supported: %w", ErrNotSupported)
}
func (a *GeminiAdapter) RemoveSkill(_ string) error {
	return fmt.Errorf("gemini adapter: RemoveSkill not supported: %w", ErrNotSupported)
}
