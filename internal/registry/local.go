package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ejyle/agentkit/internal/domain"
)

// LocalFileRegistry reads a registry.json manifest from the local filesystem.
// Used for development, acceptance testing, and offline scenarios.
type LocalFileRegistry struct {
	name string
	path string
}

// NewLocalFileRegistry creates a registry backed by a local JSON file.
// Returns an error if the file does not exist or cannot be parsed.
func NewLocalFileRegistry(name, path string) (*LocalFileRegistry, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("registry file %q not found: %w", path, err)
	}
	return &LocalFileRegistry{name: name, path: path}, nil
}

func (r *LocalFileRegistry) Name() string { return r.name }

func (r *LocalFileRegistry) load() (*domain.Manifest, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return nil, fmt.Errorf("reading registry file %q: %w", r.path, err)
	}
	var m domain.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing registry file %q: %w", r.path, err)
	}
	return &m, nil
}

func (r *LocalFileRegistry) Resolve(name string) (*domain.Package, error) {
	m, err := r.load()
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(name)
	for i := range m.Packages {
		if strings.ToLower(m.Packages[i].Name) == lower {
			return &m.Packages[i], nil
		}
	}
	return nil, nil
}

func (r *LocalFileRegistry) Search(_ string) ([]domain.Package, error) {
	m, err := r.load()
	if err != nil {
		return nil, err
	}
	return m.Packages, nil
}
