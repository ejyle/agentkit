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

// makeCopilotCLIAdapter creates a CopilotCLIAdapter configured to use tmpHome as the home directory.
func makeCopilotCLIAdapter(t *testing.T, tmpHome string) *adapter.CopilotCLIAdapter {
	t.Helper()
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "copilot-cli")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("copilot-cli", storePath)
	return adapter.NewCopilotCLIAdapterWithHome(store, tmpHome)
}

// copilotEntry returns a standard test MCPServerEntry for copilot tests.
func copilotEntry() domain.MCPServerEntry {
	return domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch"},
	}
}

// TestCopilotCLIAdapter_WriteMCPConfig_CreatesFile verifies that WriteMCPConfig creates
// ~/.copilot/mcp-config.json with the correct mcpServers structure including type and tools.
func TestCopilotCLIAdapter_WriteMCPConfig_CreatesFile(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCopilotCLIAdapter(t, tmpHome)

	err := ad.WriteMCPConfig(copilotEntry(), nil)
	if err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	configPath := filepath.Join(tmpHome, ".copilot", "mcp-config.json")
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

	mcpServers, ok := cfg["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("mcpServers key missing or wrong type")
	}
	entry, ok := mcpServers["my-server"].(map[string]interface{})
	if !ok {
		t.Fatal("my-server entry missing")
	}
	if entry["type"] != "local" {
		t.Errorf("expected type=local, got %v", entry["type"])
	}
	tools, ok := entry["tools"].([]interface{})
	if !ok || len(tools) == 0 || tools[0] != "*" {
		t.Errorf("expected tools=[*], got %v", entry["tools"])
	}
}

// TestCopilotCLIAdapter_WriteMCPConfig_PreservesExistingKeys verifies that pre-existing
// foreign keys in mcp-config.json survive a WriteMCPConfig call.
func TestCopilotCLIAdapter_WriteMCPConfig_PreservesExistingKeys(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".copilot")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "mcp-config.json")
	initial := `{"mcpServers":{"other-tool":{"type":"local","command":"node","args":["other.js"],"tools":["*"]}}}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCopilotCLIAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(copilotEntry(), nil)
	if err != nil {
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
	mcpServers, ok := cfg["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("mcpServers missing")
	}
	if _, ok := mcpServers["other-tool"]; !ok {
		t.Error("other-tool should be preserved after WriteMCPConfig")
	}
}

// TestCopilotCLIAdapter_WriteMCPConfig_ErrForeignConflict verifies that WriteMCPConfig
// returns *ErrForeignConflict when the key exists but is not owned by agentkit.
func TestCopilotCLIAdapter_WriteMCPConfig_ErrForeignConflict(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".copilot")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "mcp-config.json")
	initial := `{"mcpServers":{"my-server":{"type":"local","command":"node","args":["server.js"],"tools":["*"]}}}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCopilotCLIAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(copilotEntry(), nil)
	if err == nil {
		t.Fatal("expected ErrForeignConflict, got nil")
	}
	var conflictErr *adapter.ErrForeignConflict
	if !adapter.AsErrForeignConflict(err, &conflictErr) {
		t.Errorf("expected *ErrForeignConflict, got: %v", err)
	}
}

// TestCopilotCLIAdapter_WriteMCPConfig_AutoOverwrite verifies that WriteMCPConfig
// succeeds (no conflict) when agentkit already owns the key.
func TestCopilotCLIAdapter_WriteMCPConfig_AutoOverwrite(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".copilot")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "mcp-config.json")
	initial := `{"mcpServers":{"my-server":{"type":"local","command":"uvx","args":["mcp-server-fetch@0.1"],"tools":["*"]}}}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Register "my-server" as owned by agentkit in the store.
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "copilot-cli")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("copilot-cli", storePath)
	rec := domain.InstalledRecord{
		Name:        "my-server",
		Version:     "0.1",
		Type:        domain.PackageTypeMCP,
		InstallPath: "mcpServers.my-server",
		InstalledAt: time.Now(),
		SourceURL:   "https://example.com",
		Checksum:    "sha256:abc",
	}
	if err := store.RecordInstalled(rec); err != nil {
		t.Fatal(err)
	}

	ad := adapter.NewCopilotCLIAdapterWithHome(store, tmpHome)
	newEntry := domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch@0.2"},
	}
	err := ad.WriteMCPConfig(newEntry, nil)
	if err != nil {
		t.Fatalf("WriteMCPConfig() auto-overwrite failed: %v", err)
	}
}

// TestCopilotCLIAdapter_ReadMCPConfig_AfterWrite verifies that ReadMCPConfig returns
// the entry written by WriteMCPConfig with correct Command and Args.
func TestCopilotCLIAdapter_ReadMCPConfig_AfterWrite(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCopilotCLIAdapter(t, tmpHome)

	entry := copilotEntry()
	if err := ad.WriteMCPConfig(entry, nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	entries, err := ad.ReadMCPConfig()
	if err != nil {
		t.Fatalf("ReadMCPConfig() error: %v", err)
	}
	got, ok := entries["my-server"]
	if !ok {
		t.Fatal("ReadMCPConfig() missing 'my-server' after write")
	}
	if got.Command != "uvx" {
		t.Errorf("Command = %q; want %q", got.Command, "uvx")
	}
	if len(got.Args) == 0 || got.Args[0] != "mcp-server-fetch" {
		t.Errorf("Args = %v; want [mcp-server-fetch]", got.Args)
	}
}

// TestCopilotCLIAdapter_RemoveMCPConfig verifies that RemoveMCPConfig removes the entry
// and leaves the file as valid JSON.
func TestCopilotCLIAdapter_RemoveMCPConfig(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCopilotCLIAdapter(t, tmpHome)

	if err := ad.WriteMCPConfig(copilotEntry(), nil); err != nil {
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

// TestCopilotCLIAdapter_WriteSkill_ErrNotSupported verifies that WriteSkill returns
// an error that wraps ErrNotSupported.
func TestCopilotCLIAdapter_WriteSkill_ErrNotSupported(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCopilotCLIAdapter(t, tmpHome)

	err := ad.WriteSkill("some-skill", map[string][]byte{"SKILL.md": []byte("hello")})
	if err == nil {
		t.Fatal("WriteSkill() expected error, got nil")
	}
	if !errors.Is(err, adapter.ErrNotSupported) {
		t.Errorf("WriteSkill() error should wrap ErrNotSupported; got: %v", err)
	}
}
