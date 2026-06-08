package registry

import "github.com/ejyle/agentkit/internal/domain"

// Registry is the interface for fetching and searching package manifests.
type Registry interface {
	// Name returns the unique identifier for this registry.
	Name() string
	// Resolve finds a package by exact name (case-insensitive).
	Resolve(name string) (*domain.Package, error)
	// Search returns scored results for a query string.
	Search(query string) ([]domain.Package, error)
}

// SearchResult pairs a Package with its relevance score and source registry name.
type SearchResult struct {
	Package      domain.Package
	Score        int
	RegistryName string
}

// RegistryManager fans out Resolve and Search across multiple registries.
type RegistryManager struct {
	registries []Registry
}

// NewRegistryManager creates a RegistryManager pre-loaded with the two default registries
// (agentkit-registry and gsd-core). Both URLs are HTTPS only (T-02-01).
func NewRegistryManager() *RegistryManager {
	return &RegistryManager{
		registries: []Registry{
			NewGitHubManifestRegistry(
				"agentkit-registry",
				"https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json",
			),
			NewGitHubManifestRegistry(
				"gsd-core",
				"https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json",
			),
		},
	}
}

// NewRegistryManagerWithRegistries creates a RegistryManager with the given registries
// (used in tests to inject mock registries).
func NewRegistryManagerWithRegistries(registries ...Registry) *RegistryManager {
	return &RegistryManager{registries: registries}
}

// Resolve finds a package by name, iterating registries in order.
// Returns the first match found.
func (m *RegistryManager) Resolve(name string) (*domain.Package, error) {
	panic("not implemented")
}

// Search fans out across all registries, ranks, deduplicates, and returns results.
func (m *RegistryManager) Search(query string) ([]SearchResult, error) {
	panic("not implemented")
}
