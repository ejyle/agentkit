package adapter_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ejyle/agentkit/internal/adapter"
	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// makeAdapter creates a ClaudeCodeAdapter configured to use tmpHome as the home directory.
func makeAdapter(t *testing.T, tmpHome string) *adapter.ClaudeCodeAdapter {
	t.Helper()
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "claude")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("claude", storePath)
	return adapter.NewClaudeCodeAdapterWithHome(store, tmpHome)
}

// playwrightEntry returns a standard test MCPServerEntry for playwright.
func playwrightEntry() domain.MCPServerEntry {
	return domain.MCPServerEntry{
		Name:    "playwright",
		Command: "npx",
		Args:    []string{"-y", "@playwright/mcp"},
	}
}

// Test 7: WriteMCPConfig() detects ~/.claude.json when it exists; writes to that path.
func TestClaudeCodeAdapter_WriteMCPConfig_PrimaryPath(t *testing.T) {
	tmpHome := t.TempDir()
	claudeJSON := filepath.Join(tmpHome, ".claude.json")
	if err := os.WriteFile(claudeJSON, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(playwrightEntry(), nil)
	if err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Fatalf("reading %s: %v", claudeJSON, err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parsing %s: %v", claudeJSON, err)
	}
	mcpServers, ok := cfg["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatalf("mcpServers key missing or wrong type in %s", claudeJSON)
	}
	if _, ok := mcpServers["playwright"]; !ok {
		t.Errorf("playwright key not written to mcpServers")
	}
}

// Test 8: WriteMCPConfig() falls back to ~/.claude/settings.json if it exists and ~/.claude.json does not.
func TestClaudeCodeAdapter_WriteMCPConfig_FallbackPath(t *testing.T) {
	tmpHome := t.TempDir()
	settingsDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(playwrightEntry(), nil)
	if err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading %s: %v", settingsPath, err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parsing %s: %v", settingsPath, err)
	}
	mcpServers, ok := cfg["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatalf("mcpServers key missing in settings.json")
	}
	if _, ok := mcpServers["playwright"]; !ok {
		t.Errorf("playwright key not written to mcpServers in settings.json")
	}
}

// Test 9: WriteMCPConfig() creates ~/.claude.json when neither file exists.
func TestClaudeCodeAdapter_WriteMCPConfig_CreatesFile(t *testing.T) {
	tmpHome := t.TempDir()
	claudeJSON := filepath.Join(tmpHome, ".claude.json")

	ad := makeAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(playwrightEntry(), nil)
	if err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	if _, err := os.Stat(claudeJSON); os.IsNotExist(err) {
		t.Errorf("expected %s to be created, but it does not exist", claudeJSON)
	}
}

// Test 10: WriteMCPConfig() preserves ALL existing keys in ~/.claude.json when adding a new mcpServers entry.
func TestClaudeCodeAdapter_WriteMCPConfig_PreservesKeys(t *testing.T) {
	tmpHome := t.TempDir()
	claudeJSON := filepath.Join(tmpHome, ".claude.json")
	initial := `{"existingKey":"existingValue","otherKey":42}`
	if err := os.WriteFile(claudeJSON, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(playwrightEntry(), nil)
	if err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Fatalf("reading %s: %v", claudeJSON, err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parsing: %v", err)
	}
	if cfg["existingKey"] != "existingValue" {
		t.Errorf("existingKey was clobbered; got %v", cfg["existingKey"])
	}
	if cfg["otherKey"] != float64(42) {
		t.Errorf("otherKey was clobbered; got %v", cfg["otherKey"])
	}
}

// Test 11: WriteMCPConfig() detects foreign conflict (mcpServers.playwright exists but not owned by agentkit).
func TestClaudeCodeAdapter_WriteMCPConfig_ForeignConflict(t *testing.T) {
	tmpHome := t.TempDir()
	claudeJSON := filepath.Join(tmpHome, ".claude.json")
	initial := `{"mcpServers":{"playwright":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(claudeJSON, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(playwrightEntry(), nil)
	if err == nil {
		t.Fatal("WriteMCPConfig() expected ErrForeignConflict, got nil")
	}
	var conflictErr *adapter.ErrForeignConflict
	if !adapter.AsErrForeignConflict(err, &conflictErr) {
		t.Errorf("WriteMCPConfig() error = %v; want ErrForeignConflict", err)
	}
}

// Test 12: WriteMCPConfig() auto-overwrites when agentkit owns the key (record in installed.json).
func TestClaudeCodeAdapter_WriteMCPConfig_AgentKitOwned(t *testing.T) {
	tmpHome := t.TempDir()
	claudeJSON := filepath.Join(tmpHome, ".claude.json")
	initial := `{"mcpServers":{"playwright":{"command":"npx","args":["-y","@playwright/mcp@0.9"]}}}`
	if err := os.WriteFile(claudeJSON, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre-record playwright as owned by agentkit in the store.
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "claude")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("claude", storePath)
	rec := domain.InstalledRecord{
		Name:        "playwright",
		Version:     "0.9",
		Type:        domain.PackageTypeMCP,
		InstallPath: "mcpServers.playwright",
		InstalledAt: time.Now(),
		SourceURL:   "https://example.com",
		Checksum:    "sha256:abc",
	}
	if err := store.RecordInstalled(rec); err != nil {
		t.Fatal(err)
	}

	ad := adapter.NewClaudeCodeAdapterWithHome(store, tmpHome)
	newEntry := domain.MCPServerEntry{
		Name:    "playwright",
		Command: "npx",
		Args:    []string{"-y", "@playwright/mcp@1.2"},
	}
	err := ad.WriteMCPConfig(newEntry, nil)
	if err != nil {
		t.Fatalf("WriteMCPConfig() expected auto-overwrite (agentkit-owned), got error: %v", err)
	}

	// Verify the new entry is written.
	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]interface{}
	json.Unmarshal(data, &cfg)
	mcpServers := cfg["mcpServers"].(map[string]interface{})
	entry := mcpServers["playwright"].(map[string]interface{})
	args := entry["args"].([]interface{})
	if len(args) == 0 || args[len(args)-1] != "@playwright/mcp@1.2" {
		t.Errorf("expected updated args, got %v", args)
	}
}

// Test 13: ReadMCPConfig() after WriteMCPConfig() returns the written entry (post-install verify).
func TestClaudeCodeAdapter_ReadAfterWrite(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeAdapter(t, tmpHome)

	entry := playwrightEntry()
	if err := ad.WriteMCPConfig(entry, nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	entries, err := ad.ReadMCPConfig()
	if err != nil {
		t.Fatalf("ReadMCPConfig() error: %v", err)
	}
	got, ok := entries["playwright"]
	if !ok {
		t.Fatal("ReadMCPConfig() missing 'playwright' entry after write")
	}
	if got.Command != "npx" {
		t.Errorf("ReadMCPConfig() entry.Command = %q; want %q", got.Command, "npx")
	}
}

// Test 14: RemoveMCPConfig("playwright") removes only the playwright key; other keys untouched.
func TestClaudeCodeAdapter_RemoveMCPConfig(t *testing.T) {
	tmpHome := t.TempDir()
	claudeJSON := filepath.Join(tmpHome, ".claude.json")
	initial := `{"existingKey":"value","mcpServers":{"playwright":{"command":"npx","args":["-y","@playwright/mcp"]},"other-server":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(claudeJSON, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeAdapter(t, tmpHome)
	if err := ad.RemoveMCPConfig("playwright"); err != nil {
		t.Fatalf("RemoveMCPConfig() error: %v", err)
	}

	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]interface{}
	json.Unmarshal(data, &cfg)

	// Verify playwright is gone.
	mcpServers := cfg["mcpServers"].(map[string]interface{})
	if _, ok := mcpServers["playwright"]; ok {
		t.Error("playwright should have been removed from mcpServers")
	}
	// Verify other-server is still there.
	if _, ok := mcpServers["other-server"]; !ok {
		t.Error("other-server should remain in mcpServers")
	}
	// Verify top-level key preserved.
	if cfg["existingKey"] != "value" {
		t.Errorf("existingKey was clobbered; got %v", cfg["existingKey"])
	}
}
