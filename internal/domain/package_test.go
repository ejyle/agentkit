package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
)

func TestPackage_JSONRoundtrip(t *testing.T) {
	original := domain.Package{
		Name:        "playwright",
		Version:     "1.2.0",
		Description: "Browser automation MCP server",
		Type:        domain.PackageTypeMCP,
		Source:      "github.com/microsoft/playwright-mcp",
		Install: domain.InstallSpec{
			Method:  domain.InstallMethodNpx,
			Package: "@playwright/mcp",
		},
		MCPEntry: domain.MCPServerEntry{
			Command: "npx",
			Args:    []string{"-y", "@playwright/mcp"},
		},
		SHA256: "abc123",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal Package: %v", err)
	}

	var decoded domain.Package
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal Package: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Version != original.Version {
		t.Errorf("Version: got %q, want %q", decoded.Version, original.Version)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type: got %q, want %q", decoded.Type, original.Type)
	}
	if decoded.Install.Method != original.Install.Method {
		t.Errorf("Install.Method: got %q, want %q", decoded.Install.Method, original.Install.Method)
	}
	if decoded.MCPEntry.Command != original.MCPEntry.Command {
		t.Errorf("MCPEntry.Command: got %q, want %q", decoded.MCPEntry.Command, original.MCPEntry.Command)
	}
}

func TestManifest_JSONRoundtrip(t *testing.T) {
	original := domain.Manifest{
		Packages: []domain.Package{
			{Name: "playwright", Version: "1.0.0", Type: domain.PackageTypeMCP},
			{Name: "gsd", Version: "2.0.0", Type: domain.PackageTypeSkill},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal Manifest: %v", err)
	}

	var decoded domain.Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal Manifest: %v", err)
	}

	if len(decoded.Packages) != len(original.Packages) {
		t.Fatalf("Packages len: got %d, want %d", len(decoded.Packages), len(original.Packages))
	}
}

func TestInstallMethod_Constants(t *testing.T) {
	if domain.InstallMethodNpx != "npx" {
		t.Errorf("InstallMethodNpx: got %q, want %q", domain.InstallMethodNpx, "npx")
	}
	if domain.InstallMethodBinary != "binary" {
		t.Errorf("InstallMethodBinary: got %q, want %q", domain.InstallMethodBinary, "binary")
	}
	if domain.InstallMethodCustom != "custom" {
		t.Errorf("InstallMethodCustom: got %q, want %q", domain.InstallMethodCustom, "custom")
	}
}

func TestPackageType_Constants(t *testing.T) {
	if domain.PackageTypeMCP != "mcp" {
		t.Errorf("PackageTypeMCP: got %q, want %q", domain.PackageTypeMCP, "mcp")
	}
	if domain.PackageTypeSkill != "skill" {
		t.Errorf("PackageTypeSkill: got %q, want %q", domain.PackageTypeSkill, "skill")
	}
	if domain.PackageTypeAgent != "agent" {
		t.Errorf("PackageTypeAgent: got %q, want %q", domain.PackageTypeAgent, "agent")
	}
}
