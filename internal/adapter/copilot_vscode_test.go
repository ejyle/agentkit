package adapter_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ejyle/agentkit/internal/adapter"
	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// makeCopilotVSCodeAdapter creates a CopilotVSCodeAdapter configured to use tmpConfigDir
// as the VS Code config base directory (instead of os.UserConfigDir()).
func makeCopilotVSCodeAdapter(t *testing.T, tmpConfigDir string) *adapter.CopilotVSCodeAdapter {
	t.Helper()
	storeDir := filepath.Join(tmpConfigDir, "agentkit", "copilot-vscode")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("copilot-vscode", storePath)
	return adapter.NewCopilotVSCodeAdapterWithConfigDir(store, tmpConfigDir)
}

// vscodeEntry returns a standard test MCPServerEntry for VS Code Copilot tests.
func vscodeEntry() domain.MCPServerEntry {
	return domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch"},
	}
}

// TestCopilotVSCodeAdapter_WriteMCPConfig_CreatesFile verifies that WriteMCPConfig creates
// Code/User/mcp.json with the correct "servers" structure.
func TestCopilotVSCodeAdapter_WriteMCPConfig_CreatesFile(t *testing.T) {
	tmpConfigDir := t.TempDir()
	ad := makeCopilotVSCodeAdapter(t, tmpConfigDir)

	err := ad.WriteMCPConfig(vscodeEntry(), nil)
	if err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	configPath := filepath.Join(tmpConfigDir, "Code", "User", "mcp.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("expected %s to be created", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parsing config: %v", err)
	}

	servers, ok := cfg["servers"].(map[string]interface{})
	if !ok {
		t.Fatal("servers key missing or wrong type")
	}
	entry, ok := servers["my-server"].(map[string]interface{})
	if !ok {
		t.Fatal("my-server entry missing in servers")
	}
	if entry["command"] != "uvx" {
		t.Errorf("command = %v; want uvx", entry["command"])
	}
}

// TestCopilotVSCodeAdapter_WriteMCPConfig_UsesServersKey verifies that the top-level key
// in the written JSON is "servers" and not "mcpServers".
func TestCopilotVSCodeAdapter_WriteMCPConfig_UsesServersKey(t *testing.T) {
	tmpConfigDir := t.TempDir()
	ad := makeCopilotVSCodeAdapter(t, tmpConfigDir)

	if err := ad.WriteMCPConfig(vscodeEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	configPath := filepath.Join(tmpConfigDir, "Code", "User", "mcp.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing config: %v", err)
	}

	if _, ok := raw["servers"]; !ok {
		t.Error("top-level key should be 'servers'")
	}
	if _, ok := raw["mcpServers"]; ok {
		t.Error("top-level key must NOT be 'mcpServers' for VS Code Copilot")
	}
}

// TestCopilotVSCodeAdapter_WriteMCPConfig_PreservesExistingKeys verifies that foreign keys
// in the "servers" map are preserved after writing a new entry.
func TestCopilotVSCodeAdapter_WriteMCPConfig_PreservesExistingKeys(t *testing.T) {
	tmpConfigDir := t.TempDir()
	codeDir := filepath.Join(tmpConfigDir, "Code", "User")
	if err := os.MkdirAll(codeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(codeDir, "mcp.json")
	initial := `{"servers":{"existing-server":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCopilotVSCodeAdapter(t, tmpConfigDir)
	if err := ad.WriteMCPConfig(vscodeEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parsing config: %v", err)
	}
	servers, ok := cfg["servers"].(map[string]interface{})
	if !ok {
		t.Fatal("servers missing")
	}
	if _, ok := servers["existing-server"]; !ok {
		t.Error("existing-server should be preserved")
	}
}

// TestCopilotVSCodeAdapter_WriteMCPConfig_ErrForeignConflict verifies that WriteMCPConfig
// returns *ErrForeignConflict when the key exists but is not owned by agentkit.
func TestCopilotVSCodeAdapter_WriteMCPConfig_ErrForeignConflict(t *testing.T) {
	tmpConfigDir := t.TempDir()
	codeDir := filepath.Join(tmpConfigDir, "Code", "User")
	if err := os.MkdirAll(codeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(codeDir, "mcp.json")
	initial := `{"servers":{"my-server":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCopilotVSCodeAdapter(t, tmpConfigDir)
	err := ad.WriteMCPConfig(vscodeEntry(), nil)
	if err == nil {
		t.Fatal("expected ErrForeignConflict, got nil")
	}
	var conflictErr *adapter.ErrForeignConflict
	if !adapter.AsErrForeignConflict(err, &conflictErr) {
		t.Errorf("expected *ErrForeignConflict, got: %v", err)
	}
}

// TestCopilotVSCodeAdapter_RemoveMCPConfig verifies that RemoveMCPConfig removes the
// entry and leaves the config file as valid JSON.
func TestCopilotVSCodeAdapter_RemoveMCPConfig(t *testing.T) {
	tmpConfigDir := t.TempDir()
	ad := makeCopilotVSCodeAdapter(t, tmpConfigDir)

	if err := ad.WriteMCPConfig(vscodeEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}
	if err := ad.RemoveMCPConfig("my-server"); err != nil {
		t.Fatalf("RemoveMCPConfig() error: %v", err)
	}

	entries, err := ad.ReadMCPConfig()
	if err != nil {
		t.Fatalf("ReadMCPConfig() error: %v", err)
	}
	if _, ok := entries["my-server"]; ok {
		t.Error("my-server should have been removed")
	}
}

// TestCopilotVSCodeAdapter_WriteSkill_ErrNotSupported verifies that WriteSkill returns
// an error that wraps ErrNotSupported.
func TestCopilotVSCodeAdapter_WriteSkill_ErrNotSupported(t *testing.T) {
	tmpConfigDir := t.TempDir()
	ad := makeCopilotVSCodeAdapter(t, tmpConfigDir)

	err := ad.WriteSkill("some-skill", map[string][]byte{"SKILL.md": []byte("hello")})
	if err == nil {
		t.Fatal("WriteSkill() expected error, got nil")
	}
	if !errors.Is(err, adapter.ErrNotSupported) {
		t.Errorf("WriteSkill() error should wrap ErrNotSupported; got: %v", err)
	}
}

// TestCopilotVSCodeAdapter_EditionDetection verifies that when "Code - Insiders/User/" exists
// in tmpConfigDir, the adapter uses that path instead of the "Code" default.
func TestCopilotVSCodeAdapter_EditionDetection(t *testing.T) {
	tmpConfigDir := t.TempDir()

	// Pre-create the "Code - Insiders" user dir to trigger edition detection.
	insidersDir := filepath.Join(tmpConfigDir, "Code - Insiders", "User")
	if err := os.MkdirAll(insidersDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Register ownership so we can do a clean write.
	storeDir := filepath.Join(tmpConfigDir, "agentkit", "copilot-vscode")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("copilot-vscode", storePath)
	rec := domain.InstalledRecord{
		Name:        "my-server",
		Version:     "1.0",
		Type:        domain.PackageTypeMCP,
		InstallPath: "servers.my-server",
		InstalledAt: time.Now(),
		SourceURL:   "https://example.com",
		Checksum:    "sha256:def",
	}
	_ = store.RecordInstalled(rec) // pre-register (not required for first write, just defensive)

	ad := adapter.NewCopilotVSCodeAdapterWithConfigDir(store, tmpConfigDir)
	if err := ad.WriteMCPConfig(vscodeEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	// File should exist under "Code - Insiders", NOT under "Code".
	insidersPath := filepath.Join(tmpConfigDir, "Code - Insiders", "User", "mcp.json")
	if _, err := os.Stat(insidersPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s (Code - Insiders), not found", insidersPath)
	}
	codePath := filepath.Join(tmpConfigDir, "Code", "User", "mcp.json")
	if _, err := os.Stat(codePath); err == nil {
		t.Errorf("file should NOT be at %s (Code) when Code - Insiders dir exists", codePath)
	}
}
