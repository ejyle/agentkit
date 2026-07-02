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

// makeCursorAdapter creates a CursorAdapter configured to use tmpHome as the home directory.
func makeCursorAdapter(t *testing.T, tmpHome string) *adapter.CursorAdapter {
	t.Helper()
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "cursor")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("cursor", storePath)
	return adapter.NewCursorAdapterWithHome(store, tmpHome)
}

// cursorEntry returns a standard test MCPServerEntry for a Cursor test.
func cursorEntry() domain.MCPServerEntry {
	return domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch"},
	}
}

// TestCursorAdapter_WriteMCPConfig_CreatesFile: empty tmpHome → WriteMCPConfig → file at
// ~/.cursor/mcp.json; "mcpServers" key present; NO "type" field in entry.
func TestCursorAdapter_WriteMCPConfig_CreatesFile(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCursorAdapter(t, tmpHome)

	if err := ad.WriteMCPConfig(cursorEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	mcpPath := filepath.Join(tmpHome, ".cursor", "mcp.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("expected file at %s, got error: %v", mcpPath, err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parsing mcp.json: %v", err)
	}
	mcpServers, ok := cfg["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("mcpServers key missing or wrong type in mcp.json")
	}
	entry, ok := mcpServers["my-server"].(map[string]interface{})
	if !ok {
		t.Fatal("my-server entry missing in mcpServers")
	}
	// Cursor format must NOT have a "type" field.
	if _, hasType := entry["type"]; hasType {
		t.Errorf("Cursor MCP entry must not have a 'type' field; entry = %v", entry)
	}
}

// TestCursorAdapter_WriteMCPConfig_PreservesExistingKeys: pre-existing top-level keys survive a write.
func TestCursorAdapter_WriteMCPConfig_PreservesExistingKeys(t *testing.T) {
	tmpHome := t.TempDir()
	cursorDir := filepath.Join(tmpHome, ".cursor")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(cursorDir, "mcp.json")
	if err := os.WriteFile(mcpPath, []byte(`{"theme":"dark","otherKey":42}`), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCursorAdapter(t, tmpHome)
	if err := ad.WriteMCPConfig(cursorEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	data, _ := os.ReadFile(mcpPath)
	var cfg map[string]interface{}
	json.Unmarshal(data, &cfg)
	if cfg["theme"] != "dark" {
		t.Errorf("theme was clobbered; got %v", cfg["theme"])
	}
	if cfg["otherKey"] != float64(42) {
		t.Errorf("otherKey was clobbered; got %v", cfg["otherKey"])
	}
}

// TestCursorAdapter_WriteMCPConfig_ErrForeignConflict: key exists but not owned by agentkit.
func TestCursorAdapter_WriteMCPConfig_ErrForeignConflict(t *testing.T) {
	tmpHome := t.TempDir()
	cursorDir := filepath.Join(tmpHome, ".cursor")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(cursorDir, "mcp.json")
	initial := `{"mcpServers":{"my-server":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(mcpPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCursorAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(cursorEntry(), nil)
	if err == nil {
		t.Fatal("WriteMCPConfig() expected ErrForeignConflict, got nil")
	}
	var conflictErr *adapter.ErrForeignConflict
	if !adapter.AsErrForeignConflict(err, &conflictErr) {
		t.Errorf("WriteMCPConfig() error = %v; want ErrForeignConflict", err)
	}
}

// TestCursorAdapter_WriteMCPConfig_AutoOverwrite: key exists and IS owned by agentkit → overwrite.
func TestCursorAdapter_WriteMCPConfig_AutoOverwrite(t *testing.T) {
	tmpHome := t.TempDir()
	cursorDir := filepath.Join(tmpHome, ".cursor")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(cursorDir, "mcp.json")
	initial := `{"mcpServers":{"my-server":{"command":"uvx","args":["mcp-server-fetch@0.9"]}}}`
	if err := os.WriteFile(mcpPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre-record my-server as owned by agentkit in the store.
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "cursor")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("cursor", storePath)
	rec := domain.InstalledRecord{
		Name:        "my-server",
		Version:     "0.9",
		Type:        domain.PackageTypeMCP,
		InstallPath: "mcpServers.my-server",
		InstalledAt: time.Now(),
		SourceURL:   "https://example.com",
		Checksum:    "sha256:abc",
	}
	if err := store.RecordInstalled(rec); err != nil {
		t.Fatal(err)
	}

	ad := adapter.NewCursorAdapterWithHome(store, tmpHome)
	newEntry := domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch@1.2"},
	}
	if err := ad.WriteMCPConfig(newEntry, nil); err != nil {
		t.Fatalf("WriteMCPConfig() expected auto-overwrite (agentkit-owned), got error: %v", err)
	}

	data, _ := os.ReadFile(mcpPath)
	var cfg map[string]interface{}
	json.Unmarshal(data, &cfg)
	mcpServers := cfg["mcpServers"].(map[string]interface{})
	entry := mcpServers["my-server"].(map[string]interface{})
	args := entry["args"].([]interface{})
	if len(args) == 0 || args[len(args)-1] != "mcp-server-fetch@1.2" {
		t.Errorf("expected updated args, got %v", args)
	}
}

// TestCursorAdapter_ReadMCPConfig_AfterWrite: ReadMCPConfig returns the written entry.
func TestCursorAdapter_ReadMCPConfig_AfterWrite(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCursorAdapter(t, tmpHome)

	entry := cursorEntry()
	if err := ad.WriteMCPConfig(entry, nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	entries, err := ad.ReadMCPConfig()
	if err != nil {
		t.Fatalf("ReadMCPConfig() error: %v", err)
	}
	got, ok := entries["my-server"]
	if !ok {
		t.Fatal("ReadMCPConfig() missing 'my-server' entry after write")
	}
	if got.Command != "uvx" {
		t.Errorf("ReadMCPConfig() entry.Command = %q; want %q", got.Command, "uvx")
	}
}

// TestCursorAdapter_RemoveMCPConfig: removes only the named key; other keys untouched.
func TestCursorAdapter_RemoveMCPConfig(t *testing.T) {
	tmpHome := t.TempDir()
	cursorDir := filepath.Join(tmpHome, ".cursor")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(cursorDir, "mcp.json")
	initial := `{"theme":"dark","mcpServers":{"my-server":{"command":"uvx","args":["mcp-server-fetch"]},"other-server":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(mcpPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeCursorAdapter(t, tmpHome)
	if err := ad.RemoveMCPConfig("my-server"); err != nil {
		t.Fatalf("RemoveMCPConfig() error: %v", err)
	}

	data, _ := os.ReadFile(mcpPath)
	var cfg map[string]interface{}
	json.Unmarshal(data, &cfg)
	mcpServers := cfg["mcpServers"].(map[string]interface{})
	if _, ok := mcpServers["my-server"]; ok {
		t.Error("my-server should have been removed from mcpServers")
	}
	if _, ok := mcpServers["other-server"]; !ok {
		t.Error("other-server should remain in mcpServers")
	}
	if cfg["theme"] != "dark" {
		t.Errorf("theme was clobbered; got %v", cfg["theme"])
	}
}

// TestCursorAdapter_WriteSkill_CreatesDirectory: WriteSkill writes files to ~/.cursor/skills/<name>/.
func TestCursorAdapter_WriteSkill_CreatesDirectory(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCursorAdapter(t, tmpHome)

	skillName := "my-skill"
	content := []byte("# My Skill")
	if err := ad.WriteSkill(skillName, map[string][]byte{"SKILL.md": content}); err != nil {
		t.Fatalf("WriteSkill() error: %v", err)
	}

	destPath := filepath.Join(tmpHome, ".cursor", "skills", skillName, "SKILL.md")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("expected file at %s: %v", destPath, err)
	}
	if string(data) != string(content) {
		t.Errorf("SKILL.md content = %q; want %q", string(data), string(content))
	}
}

// TestCursorAdapter_RemoveSkill: WriteSkill then RemoveSkill → skill directory no longer exists.
func TestCursorAdapter_RemoveSkill(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeCursorAdapter(t, tmpHome)

	skillName := "my-skill"
	if err := ad.WriteSkill(skillName, map[string][]byte{"SKILL.md": []byte("content")}); err != nil {
		t.Fatalf("WriteSkill() error: %v", err)
	}
	if err := ad.RemoveSkill(skillName); err != nil {
		t.Fatalf("RemoveSkill() error: %v", err)
	}

	skillDir := filepath.Join(tmpHome, ".cursor", "skills", skillName)
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("expected skill directory %s to be removed", skillDir)
	}
}
