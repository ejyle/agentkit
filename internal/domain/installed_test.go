package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ejyle/agentkit/internal/domain"
)

func TestInstalledRecord_JSONRoundtrip(t *testing.T) {
	ts := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	original := domain.InstalledRecord{
		Name:        "playwright",
		Version:     "1.2.0",
		Type:        domain.PackageTypeMCP,
		InstallPath: "mcpServers.playwright",
		InstalledAt: ts,
		SourceURL:   "https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json",
		Checksum:    "sha256:abc123",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal InstalledRecord: %v", err)
	}

	var decoded domain.InstalledRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal InstalledRecord: %v", err)
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
	if decoded.InstallPath != original.InstallPath {
		t.Errorf("InstallPath: got %q, want %q", decoded.InstallPath, original.InstallPath)
	}
	if !decoded.InstalledAt.Equal(original.InstalledAt) {
		t.Errorf("InstalledAt: got %v, want %v", decoded.InstalledAt, original.InstalledAt)
	}
	if decoded.SourceURL != original.SourceURL {
		t.Errorf("SourceURL: got %q, want %q", decoded.SourceURL, original.SourceURL)
	}
	if decoded.Checksum != original.Checksum {
		t.Errorf("Checksum: got %q, want %q", decoded.Checksum, original.Checksum)
	}
}

// TestInstalledRecord_JSONFieldNames verifies the D-11 snake_case field names are used in JSON.
func TestInstalledRecord_JSONFieldNames(t *testing.T) {
	ts := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	record := domain.InstalledRecord{
		Name:        "playwright",
		Version:     "1.2.0",
		Type:        domain.PackageTypeMCP,
		InstallPath: "mcpServers.playwright",
		InstalledAt: ts,
		SourceURL:   "https://example.com/registry.json",
		Checksum:    "sha256:abc",
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	requiredKeys := []string{
		"name", "version", "type", "install_path", "installed_at", "source_url", "checksum",
	}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing JSON field %q (D-11 schema requirement)", key)
		}
	}

	// Ensure camelCase variants are NOT present.
	forbiddenKeys := []string{"installPath", "installedAt", "sourceURL", "sourceUrl"}
	for _, key := range forbiddenKeys {
		if _, ok := raw[key]; ok {
			t.Errorf("found camelCase JSON field %q — must be snake_case per D-11", key)
		}
	}
}

// TestInstalledState_ZeroValue verifies nil-safety of the zero-value InstalledState.
func TestInstalledState_ZeroValue(t *testing.T) {
	var state domain.InstalledState
	// Packages is nil on zero-value — must not panic.
	if state.Packages != nil {
		t.Errorf("expected nil Packages on zero-value InstalledState, got %v", state.Packages)
	}
	// Nil-map read must not panic.
	_ = state.Packages["playwright"]
}

func TestInstalledState_JSONRoundtrip(t *testing.T) {
	ts := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	original := domain.InstalledState{
		Packages: map[string]domain.InstalledRecord{
			"playwright": {
				Name:        "playwright",
				Version:     "1.2.0",
				Type:        domain.PackageTypeMCP,
				InstallPath: "mcpServers.playwright",
				InstalledAt: ts,
				SourceURL:   "https://example.com/registry.json",
				Checksum:    "sha256:abc",
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal InstalledState: %v", err)
	}

	var decoded domain.InstalledState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal InstalledState: %v", err)
	}

	rec, ok := decoded.Packages["playwright"]
	if !ok {
		t.Fatal("playwright key missing after roundtrip")
	}
	if rec.Version != "1.2.0" {
		t.Errorf("Version: got %q, want %q", rec.Version, "1.2.0")
	}
}
