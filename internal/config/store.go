package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/google/renameio/v2"
)

// ConfigStore manages the per-assistant installed.json state file.
// All writes are atomic via renameio to prevent partial-write corruption.
type ConfigStore struct {
	target   string
	basePath string
	mu       sync.Mutex
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

// RecordInstalled writes or updates an InstalledRecord in installed.json.
// Auto-creates the directory and file on first call (D-12).
// Uses an atomic rename write (T-02-04).
func (s *ConfigStore) RecordInstalled(rec domain.InstalledRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Auto-create parent directory (D-12).
	if err := os.MkdirAll(filepath.Dir(s.basePath), 0755); err != nil {
		return err
	}

	state, err := s.loadStateUnlocked()
	if err != nil {
		return err
	}

	state.Packages[rec.Name] = rec
	return s.writeStateUnlocked(state)
}

// RemoveRecord deletes a record from installed.json by name.
func (s *ConfigStore) RemoveRecord(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.loadStateUnlocked()
	if err != nil {
		return err
	}
	delete(state.Packages, name)
	return s.writeStateUnlocked(state)
}

// ListInstalled returns all installed records sorted by name.
// Returns an empty slice (not an error) if installed.json does not exist.
func (s *ConfigStore) ListInstalled() ([]domain.InstalledRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.loadStateUnlocked()
	if err != nil {
		return nil, err
	}

	records := make([]domain.InstalledRecord, 0, len(state.Packages))
	for _, r := range state.Packages {
		records = append(records, r)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].Name < records[j].Name
	})
	return records, nil
}

// GetRecord retrieves a single installed record by name.
// Returns (record, true, nil) if found, or (zero, false, nil) if absent.
func (s *ConfigStore) GetRecord(name string) (domain.InstalledRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.loadStateUnlocked()
	if err != nil {
		return domain.InstalledRecord{}, false, err
	}
	rec, ok := state.Packages[name]
	return rec, ok, nil
}

// loadStateUnlocked reads installed.json from disk.
// If the file is absent, returns an empty InstalledState (not an error).
// Caller must hold s.mu.
func (s *ConfigStore) loadStateUnlocked() (domain.InstalledState, error) {
	data, err := os.ReadFile(s.basePath)
	if os.IsNotExist(err) {
		return domain.InstalledState{Packages: make(map[string]domain.InstalledRecord)}, nil
	}
	if err != nil {
		return domain.InstalledState{}, err
	}

	var state domain.InstalledState
	if err := json.Unmarshal(data, &state); err != nil {
		return domain.InstalledState{}, err
	}
	if state.Packages == nil {
		state.Packages = make(map[string]domain.InstalledRecord)
	}
	return state, nil
}

// writeStateUnlocked serialises state and writes it atomically via renameio.
// Caller must hold s.mu.
func (s *ConfigStore) writeStateUnlocked(state domain.InstalledState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return renameio.WriteFile(s.basePath, data, 0644)
}
