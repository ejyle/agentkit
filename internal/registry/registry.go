package registry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ejyle/agentkit/internal/domain"
)

// Registry is the interface for fetching and searching package manifests.
type Registry interface {
	// Name returns the unique identifier for this registry.
	Name() string
	// Resolve finds a package by exact name (case-insensitive).
	Resolve(name string) (*domain.Package, error)
	// Search returns packages matching the given query string.
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

// NewRegistryManager creates a RegistryManager pre-loaded with the two default registries:
//   - agentkit-registry (official curated registry — D-01, D-02)
//   - gsd-core (open-gsd skill library — REG-02)
//
// Both URLs are HTTPS only (T-02-01).
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

// NewRegistryManagerWithRegistries creates a RegistryManager with the given registries.
// Used in tests to inject mock/httptest registries.
func NewRegistryManagerWithRegistries(registries ...Registry) *RegistryManager {
	return &RegistryManager{registries: registries}
}

// Resolve finds a package by exact name, iterating registries in registration order.
// Returns the first match; if no registry has the package, returns an error.
func (m *RegistryManager) Resolve(name string) (*domain.Package, error) {
	for _, reg := range m.registries {
		pkg, err := reg.Resolve(name)
		if err != nil {
			// Registry unavailable — continue trying others.
			continue
		}
		if pkg != nil {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("%q not found in any registry", name)
}

// Search fans out across all registries, ranks, deduplicates, and returns results.
//
// Scoring (D-06):
//   - Exact name match (case-insensitive): score = 100
//   - Name contains query:                score = 50
//   - Description contains query:         score = 10
//
// Sort order: score descending, then name ascending (deterministic — T7).
// Deduplication: first registry to return a package wins (by name, case-insensitive).
func (m *RegistryManager) Search(query string) ([]SearchResult, error) {
	seen := make(map[string]struct{})
	var results []SearchResult
	needle := strings.ToLower(query)

	for _, reg := range m.registries {
		pkgs, err := reg.Search("")
		if err != nil {
			// Registry unavailable — skip it.
			continue
		}
		for _, p := range pkgs {
			key := strings.ToLower(p.Name)
			if _, dup := seen[key]; dup {
				continue
			}
			score := scorePackage(p, needle)
			if query == "" || score > 0 {
				seen[key] = struct{}{}
				results = append(results, SearchResult{
					Package:      p,
					Score:        score,
					RegistryName: reg.Name(),
				})
			}
		}
	}

	// Sort by score descending, then name ascending for determinism.
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Package.Name < results[j].Package.Name
	})

	return results, nil
}

// scorePackage calculates the relevance score for a package against a query needle.
// needle must already be lower-cased.
func scorePackage(p domain.Package, needle string) int {
	if needle == "" {
		return 50 // no query — include all at a neutral score
	}
	nameLower := strings.ToLower(p.Name)
	if nameLower == needle {
		return 100
	}
	if strings.Contains(nameLower, needle) {
		return 50
	}
	if strings.Contains(strings.ToLower(p.Description), needle) {
		return 10
	}
	return 0
}
