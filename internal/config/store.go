package config

// Stub — will be implemented in GREEN phase.
// This file provides the exported identifiers so the test file compiles.

import "github.com/ejyle/agentkit/internal/domain"

// ConfigStore manages the per-assistant installed.json state file.
type ConfigStore struct {
	target   string
	basePath string
}

// NewConfigStore creates a ConfigStore using the standard path for the given target.
func NewConfigStore(target string) *ConfigStore {
	path, _ := InstalledStatePath(target)
	return &ConfigStore{target: target, basePath: path}
}

// NewConfigStoreWithPath creates a ConfigStore with an injected path (for testing).
func NewConfigStoreWithPath(target, path string) *ConfigStore {
	return &ConfigStore{target: target, basePath: path}
}

// RecordInstalled stub — always errors so RED tests fail as expected.
func (s *ConfigStore) RecordInstalled(_ domain.InstalledRecord) error {
	panic("not implemented")
}

// RemoveRecord stub.
func (s *ConfigStore) RemoveRecord(_ string) error {
	panic("not implemented")
}

// ListInstalled stub.
func (s *ConfigStore) ListInstalled() ([]domain.InstalledRecord, error) {
	panic("not implemented")
}

// GetRecord stub.
func (s *ConfigStore) GetRecord(_ string) (domain.InstalledRecord, bool, error) {
	panic("not implemented")
}
