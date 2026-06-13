package service_test

import (
	"errors"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/registry"
	"github.com/ejyle/agentkit/internal/service"
)

// mockSearchRegistry is a test double for the searchRegistry interface.
type mockSearchRegistry struct {
	results []registry.SearchResult
	err     error
}

func (m *mockSearchRegistry) Search(query string) ([]registry.SearchResult, error) {
	return m.results, m.err
}

// Test 1: SearchService.Search delegates to the underlying registry.
func TestSearchService_Search_DelegatesQuery(t *testing.T) {
	expected := []registry.SearchResult{
		{
			Package:      domain.Package{Name: "playwright", Version: "1.0.0", Description: "Playwright MCP server", Type: domain.PackageTypeMCP},
			Score:        100,
			RegistryName: "agentkit-registry",
		},
	}
	mock := &mockSearchRegistry{results: expected}
	svc := service.NewSearchService(mock)

	results, err := svc.Search("playwright")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(results))
	}
	if results[0].Package.Name != "playwright" {
		t.Errorf("expected result name %q, got %q", "playwright", results[0].Package.Name)
	}
}

// Test 2: SearchService.Search propagates errors from the registry.
func TestSearchService_Search_PropagatesError(t *testing.T) {
	expectedErr := errors.New("registry unavailable")
	mock := &mockSearchRegistry{err: expectedErr}
	svc := service.NewSearchService(mock)

	_, err := svc.Search("playwright")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
