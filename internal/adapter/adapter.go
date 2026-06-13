// Package adapter provides AssistantAdapter implementations for writing MCP server config
// and skill files to each supported AI coding assistant.
package adapter

import (
	"errors"

	"github.com/ejyle/agentkit/internal/domain"
)

// AssistantAdapter is the interface for reading and writing MCP config and skill files
// for a particular AI coding assistant.
type AssistantAdapter interface {
	// WriteMCPConfig writes an MCP server entry to the assistant's config file.
	// ownership may be nil on first install; non-nil ownership means agentkit already
	// owns this key and the write is an upgrade (D-08).
	WriteMCPConfig(entry domain.MCPServerEntry, ownership *domain.InstalledRecord) error
	// RemoveMCPConfig removes the named MCP server entry from the assistant's config file.
	RemoveMCPConfig(name string) error
	// ReadMCPConfig returns all MCP server entries currently in the assistant's config file.
	ReadMCPConfig() (map[string]domain.MCPServerEntry, error)
	// WriteSkill writes skill files into the assistant's skill install directory.
	WriteSkill(name string, files map[string][]byte) error
	// RemoveSkill removes a skill directory from the assistant's skill install directory.
	RemoveSkill(name string) error
	// Name returns the canonical name for this adapter's target assistant.
	Name() string
}

// ErrForeignConflict is returned by WriteMCPConfig when the target mcpServers key
// already exists in the config file but was not installed by agentkit (D-07).
type ErrForeignConflict struct {
	OldEntry domain.MCPServerEntry
	NewEntry domain.MCPServerEntry
}

func (e *ErrForeignConflict) Error() string {
	return "foreign conflict: mcpServers." + e.NewEntry.Name + " is already set by a non-agentkit source"
}

// AsErrForeignConflict is a helper that unwraps err into *ErrForeignConflict.
// Returns true and populates target if err is or wraps *ErrForeignConflict.
func AsErrForeignConflict(err error, target **ErrForeignConflict) bool {
	return errors.As(err, target)
}

// ErrNotSupported is returned by adapter methods that are not applicable for a given target assistant.
var ErrNotSupported = errors.New("operation not supported")
