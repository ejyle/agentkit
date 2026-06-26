package ui_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/registry"
	"github.com/ejyle/agentkit/internal/ui"
)

// Test 3: RenderInstalledTable with records contains all expected column headers.
func TestRenderInstalledTable_Headers(t *testing.T) {
	records := []domain.InstalledRecord{
		{Name: "playwright", Version: "1.0.0", Type: domain.PackageTypeMCP, InstalledAt: time.Now(), SourceURL: "https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json"},
		{Name: "gsd", Version: "2.0.0", Type: domain.PackageTypeSkill, InstalledAt: time.Now(), SourceURL: "https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json"},
	}

	output := ui.RenderInstalledTable(records, "claude")

	for _, header := range []string{"PACKAGE", "VERSION", "TYPE", "TARGET", "REGISTRY"} {
		if !strings.Contains(output, header) {
			t.Errorf("expected output to contain header %q, got:\n%s", header, output)
		}
	}
}

// Test 4: RenderInstalledTable with records contains both package names.
func TestRenderInstalledTable_ContainsPackageNames(t *testing.T) {
	records := []domain.InstalledRecord{
		{Name: "playwright", Version: "1.0.0", Type: domain.PackageTypeMCP, InstalledAt: time.Now(), SourceURL: "https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json"},
		{Name: "gsd", Version: "2.0.0", Type: domain.PackageTypeSkill, InstalledAt: time.Now(), SourceURL: "https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json"},
	}

	output := ui.RenderInstalledTable(records, "claude")

	if !strings.Contains(output, "playwright") {
		t.Errorf("expected output to contain %q, got:\n%s", "playwright", output)
	}
	if !strings.Contains(output, "gsd") {
		t.Errorf("expected output to contain %q, got:\n%s", "gsd", output)
	}
}

// Test 5: RenderInstalledTable with empty slice returns "No packages installed" message.
func TestRenderInstalledTable_EmptyState(t *testing.T) {
	output := ui.RenderInstalledTable(nil, "claude")

	lowerOutput := strings.ToLower(output)
	if !strings.Contains(lowerOutput, "no packages installed") {
		t.Errorf("expected output to contain 'no packages installed' (case-insensitive), got:\n%s", output)
	}
}

// Test 6: RenderSearchResults with results contains package names, registry names, and descriptions.
func TestRenderSearchResults_ContainsFields(t *testing.T) {
	results := []registry.SearchResult{
		{
			Package:      domain.Package{Name: "playwright", Version: "1.0.0", Description: "Playwright MCP server for browser automation", Type: domain.PackageTypeMCP},
			Score:        100,
			RegistryName: "agentkit-registry",
		},
		{
			Package:      domain.Package{Name: "gsd", Version: "2.0.0", Description: "Get Shit Done skill pack", Type: domain.PackageTypeSkill},
			Score:        50,
			RegistryName: "gsd-core",
		},
		{
			Package:      domain.Package{Name: "codex-agent", Version: "0.5.0", Description: "Codex coding agent", Type: domain.PackageTypeAgent},
			Score:        10,
			RegistryName: "agentkit-registry",
		},
	}

	output := ui.RenderSearchResults(results)

	for _, name := range []string{"playwright", "gsd", "codex-agent"} {
		if !strings.Contains(output, name) {
			t.Errorf("expected output to contain package name %q, got:\n%s", name, output)
		}
	}
	for _, regName := range []string{"agentkit-registry", "gsd-core"} {
		if !strings.Contains(output, regName) {
			t.Errorf("expected output to contain registry name %q, got:\n%s", regName, output)
		}
	}
	for _, desc := range []string{"browser automation", "Get Shit Done", "Codex coding agent"} {
		if !strings.Contains(output, desc) {
			t.Errorf("expected output to contain description fragment %q, got:\n%s", desc, output)
		}
	}
}

// Test 7: RenderSearchResults with empty slice returns "No results found".
func TestRenderSearchResults_EmptyState(t *testing.T) {
	output := ui.RenderSearchResults(nil)

	if !strings.Contains(output, "No results found") {
		t.Errorf("expected output to contain 'No results found', got:\n%s", output)
	}
}
