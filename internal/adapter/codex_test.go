package adapter_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/ejyle/agentkit/internal/adapter"
	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// isErrNotSupported checks whether err wraps adapter.ErrNotSupported.
func isErrNotSupported(err error) bool {
	return errors.Is(err, adapter.ErrNotSupported)
}

// makeCodexAdapter creates a CodexAdapter configured to use tmpHome as the home directory.
func makeCodexAdapter(t *testing.T, tmpHome string) *adapter.CodexAdapter {
	t.Helper()
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "codex")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("codex", storePath)
	return adapter.NewCodexAdapterWithHome(store, tmpHome)
}

// codexEntry returns a standard test MCPServerEntry.
func codexEntry() domain.MCPServerEntry {
	return domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch"},
	}
}

// TestCodexAdapter_WriteMCPConfig_CreatesTOML verifies that WriteMCPConfig creates
// ~/.codex/config.toml with the [mcp_servers.<name>] section.
func TestCodexAdapter_WriteMCPConfig_CreatesTOML(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCodexAdapter(t, tmpHome)

	if err := ad.WriteMCPConfig(codexEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	configPath := filepath.Join(tmpHome, ".codex", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("expected %s to be created", configPath)
	}

	var raw map[string]interface{}
	if _, err := toml.DecodeFile(configPath, &raw); err != nil {
		t.Fatalf("DecodeFile(%s): %v", configPath, err)
	}

	mcpServers, ok := raw["mcp_servers"].(map[string]interface{})
	if !ok {
		t.Fatalf("mcp_servers key missing or wrong type")
	}
	serverEntry, ok := mcpServers["my-server"].(map[string]interface{})
	if !ok {
		t.Fatalf("[mcp_servers.my-server] section missing or wrong type")
	}
	if serverEntry["command"] != "uvx" {
		t.Errorf("command = %v; want uvx", serverEntry["command"])
	}
}

// TestCodexAdapter_WriteMCPConfig_PreservesNonMCPKeys verifies that existing TOML keys
// outside of mcp_servers are preserved after a WriteMCPConfig call.
func TestCodexAdapter_WriteMCPConfig_PreservesNonMCPKeys(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".codex")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	initialTOML := "[settings]\nmodel = \"gpt-4o\"\n"
	if err := os.WriteFile(configPath, []byte(initialTOML), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCodexAdapter(t, tmpHome)
	if err := ad.WriteMCPConfig(codexEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	var raw map[string]interface{}
	if _, err := toml.DecodeFile(configPath, &raw); err != nil {
		t.Fatalf("DecodeFile: %v", err)
	}

	settings, ok := raw["settings"].(map[string]interface{})
	if !ok {
		t.Fatalf("[settings] section missing after WriteMCPConfig")
	}
	if settings["model"] != "gpt-4o" {
		t.Errorf("settings.model = %v; want gpt-4o (key was clobbered)", settings["model"])
	}
}

// TestCodexAdapter_WriteMCPConfig_ErrForeignConflict verifies that WriteMCPConfig returns
// ErrForeignConflict when the key exists and is not agentkit-owned.
func TestCodexAdapter_WriteMCPConfig_ErrForeignConflict(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".codex")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	initialTOML := "[mcp_servers.my-server]\ncommand = \"other\"\nargs = []\n"
	if err := os.WriteFile(configPath, []byte(initialTOML), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCodexAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(codexEntry(), nil)
	if err == nil {
		t.Fatal("WriteMCPConfig() expected ErrForeignConflict, got nil")
	}
	var conflictErr *adapter.ErrForeignConflict
	if !adapter.AsErrForeignConflict(err, &conflictErr) {
		t.Errorf("WriteMCPConfig() error = %v; want ErrForeignConflict", err)
	}
}

// TestCodexAdapter_WriteMCPConfig_AutoOverwrite verifies that WriteMCPConfig auto-overwrites
// when the key is agentkit-owned.
func TestCodexAdapter_WriteMCPConfig_AutoOverwrite(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".codex")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	initialTOML := "[mcp_servers.my-server]\ncommand = \"uvx\"\nargs = [\"old-arg\"]\n"
	if err := os.WriteFile(configPath, []byte(initialTOML), 0644); err != nil {
		t.Fatal(err)
	}

	storeDir := filepath.Join(tmpHome, "config", "agentkit", "codex")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("codex", storePath)
	rec := domain.InstalledRecord{
		Name:        "my-server",
		Version:     "1.0.0",
		Type:        domain.PackageTypeMCP,
		InstallPath: "mcp_servers.my-server",
		SourceURL:   "https://example.com",
		Checksum:    "sha256:abc",
	}
	if err := store.RecordInstalled(rec); err != nil {
		t.Fatal(err)
	}

	ad := adapter.NewCodexAdapterWithHome(store, tmpHome)
	newEntry := domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"new-arg"},
	}
	if err := ad.WriteMCPConfig(newEntry, nil); err != nil {
		t.Fatalf("WriteMCPConfig() auto-overwrite failed: %v", err)
	}

	var raw map[string]interface{}
	if _, err := toml.DecodeFile(configPath, &raw); err != nil {
		t.Fatal(err)
	}
	mcpServers := raw["mcp_servers"].(map[string]interface{})
	srv := mcpServers["my-server"].(map[string]interface{})
	args, ok := srv["args"].([]interface{})
	if !ok || len(args) == 0 || args[0] != "new-arg" {
		t.Errorf("expected args=[new-arg], got %v", srv["args"])
	}
}

// TestCodexAdapter_ReadMCPConfig_AfterWrite verifies round-trip: write then read returns entry.
func TestCodexAdapter_ReadMCPConfig_AfterWrite(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCodexAdapter(t, tmpHome)

	if err := ad.WriteMCPConfig(codexEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	entries, err := ad.ReadMCPConfig()
	if err != nil {
		t.Fatalf("ReadMCPConfig() error: %v", err)
	}
	got, ok := entries["my-server"]
	if !ok {
		t.Fatal("ReadMCPConfig() missing my-server entry after write")
	}
	if got.Command != "uvx" {
		t.Errorf("Command = %q; want uvx", got.Command)
	}
	if len(got.Args) != 1 || got.Args[0] != "mcp-server-fetch" {
		t.Errorf("Args = %v; want [mcp-server-fetch]", got.Args)
	}
}

// TestCodexAdapter_RemoveMCPConfig verifies that RemoveMCPConfig removes only the named entry.
func TestCodexAdapter_RemoveMCPConfig(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".codex")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	initialTOML := "[mcp_servers.my-server]\ncommand = \"uvx\"\nargs = []\n[mcp_servers.other]\ncommand = \"node\"\nargs = []\n"
	if err := os.WriteFile(configPath, []byte(initialTOML), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCodexAdapter(t, tmpHome)
	if err := ad.RemoveMCPConfig("my-server"); err != nil {
		t.Fatalf("RemoveMCPConfig() error: %v", err)
	}

	var raw map[string]interface{}
	if _, err := toml.DecodeFile(configPath, &raw); err != nil {
		t.Fatal(err)
	}
	mcpServers, ok := raw["mcp_servers"].(map[string]interface{})
	if !ok {
		t.Fatalf("mcp_servers missing after remove")
	}
	if _, ok := mcpServers["my-server"]; ok {
		t.Error("my-server should have been removed")
	}
	if _, ok := mcpServers["other"]; !ok {
		t.Error("other server should remain")
	}
}

// TestCodexAdapter_WriteSkill_ErrNotSupported verifies that WriteSkill returns ErrNotSupported.
func TestCodexAdapter_WriteSkill_ErrNotSupported(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCodexAdapter(t, tmpHome)

	err := ad.WriteSkill("some-skill", map[string][]byte{"SKILL.md": []byte("content")})
	if err == nil {
		t.Fatal("WriteSkill() expected ErrNotSupported, got nil")
	}
	if !isErrNotSupported(err) {
		t.Errorf("WriteSkill() error = %v; want wrapping ErrNotSupported", err)
	}
}

// TestCodexAdapter_WriteMCPConfig_PreservesEnvSubtable verifies env entries are preserved.
func TestCodexAdapter_WriteMCPConfig_PreservesEnvSubtable(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCodexAdapter(t, tmpHome)

	entry := domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch"},
		Env:     map[string]string{"MY_KEY": "val"},
	}
	if err := ad.WriteMCPConfig(entry, nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	configPath := filepath.Join(tmpHome, ".codex", "config.toml")
	var raw map[string]interface{}
	if _, err := toml.DecodeFile(configPath, &raw); err != nil {
		t.Fatal(err)
	}

	mcpServers := raw["mcp_servers"].(map[string]interface{})
	srv := mcpServers["my-server"].(map[string]interface{})
	envMap, ok := srv["env"].(map[string]interface{})
	if !ok {
		t.Fatalf("[mcp_servers.my-server.env] section missing")
	}
	if envMap["MY_KEY"] != "val" {
		t.Errorf("MY_KEY = %v; want val", envMap["MY_KEY"])
	}
}
