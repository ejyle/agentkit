// Package service provides core orchestration services for agentkit.
package service

import "github.com/ejyle/agentkit/internal/registry"

// searchRegistry is the interface SearchService depends on.
// Using a local interface allows test doubles without coupling to the real RegistryManager.
type searchRegistry interface {
	Search(query string) ([]registry.SearchResult, error)
}

// SearchService wraps a searchRegistry to provide search functionality.
type SearchService struct {
	reg searchRegistry
}

// NewSearchService creates a SearchService backed by the given registry.
// reg may be a *registry.RegistryManager or any searchRegistry implementation (e.g. a mock).
func NewSearchService(reg searchRegistry) *SearchService {
	return &SearchService{reg: reg}
}

// Search executes a search query against the backing registry and returns ranked results.
func (s *SearchService) Search(query string) ([]registry.SearchResult, error) {
	return s.reg.Search(query)
}
