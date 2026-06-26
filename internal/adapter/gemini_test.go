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

// makeGeminiAdapter creates a GeminiAdapter configured to use tmpHome as the home directory.
func makeGeminiAdapter(t *testing.T, tmpHome string) *adapter.GeminiAdapter {
	t.Helper()
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "gemini")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("gemini", storePath)
	return adapter.NewGeminiAdapterWithHome(store, tmpHome)
}

// geminiEntry returns a standard test MCPServerEntry for a Gemini test.
func geminiEntry() domain.MCPServerEntry {
	return domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch"},
	}
}

// TestGeminiAdapter_WriteMCPConfig_CreatesFile: empty tmpHome → WriteMCPConfig → file at
// ~/.gemini/settings.json; "mcpServers" key present; NO "type" field in entry.
func TestGeminiAdapter_WriteMCPConfig_CreatesFile(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeGeminiAdapter(t, tmpHome)

	if err := ad.WriteMCPConfig(geminiEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	settingsPath := filepath.Join(tmpHome, ".gemini", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("expected file at %s, got error: %v", settingsPath, err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parsing settings.json: %v", err)
	}
	mcpServers, ok := cfg["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("mcpServers key missing or wrong type in settings.json")
	}
	entry, ok := mcpServers["my-server"].(map[string]interface{})
	if !ok {
		t.Fatal("my-server entry missing in mcpServers")
	}
	// Gemini format must NOT have a "type" field.
	if _, hasType := entry["type"]; hasType {
		t.Errorf("Gemini MCP entry must not have a 'type' field; entry = %v", entry)
	}
}

// TestGeminiAdapter_WriteMCPConfig_PreservesExistingKeys: pre-existing top-level keys survive a write.
func TestGeminiAdapter_WriteMCPConfig_PreservesExistingKeys(t *testing.T) {
	tmpHome := t.TempDir()
	geminiDir := filepath.Join(tmpHome, ".gemini")
	if err := os.MkdirAll(geminiDir, 0755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(geminiDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"theme":"dark","otherKey":42}`), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeGeminiAdapter(t, tmpHome)
	if err := ad.WriteMCPConfig(geminiEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
	var cfg map[string]interface{}
	json.Unmarshal(data, &cfg)
	if cfg["theme"] != "dark" {
		t.Errorf("theme was clobbered; got %v", cfg["theme"])
	}
	if cfg["otherKey"] != float64(42) {
		t.Errorf("otherKey was clobbered; got %v", cfg["otherKey"])
	}
}

// TestGeminiAdapter_WriteMCPConfig_ErrForeignConflict: key exists but not owned by agentkit.
func TestGeminiAdapter_WriteMCPConfig_ErrForeignConflict(t *testing.T) {
	tmpHome := t.TempDir()
	geminiDir := filepath.Join(tmpHome, ".gemini")
	if err := os.MkdirAll(geminiDir, 0755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(geminiDir, "settings.json")
	initial := `{"mcpServers":{"my-server":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(settingsPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeGeminiAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(geminiEntry(), nil)
	if err == nil {
		t.Fatal("WriteMCPConfig() expected ErrForeignConflict, got nil")
	}
	var conflictErr *adapter.ErrForeignConflict
	if !adapter.AsErrForeignConflict(err, &conflictErr) {
		t.Errorf("WriteMCPConfig() error = %v; want ErrForeignConflict", err)
	}
}

// TestGeminiAdapter_WriteMCPConfig_AutoOverwrite: key exists and IS owned by agentkit → overwrite.
func TestGeminiAdapter_WriteMCPConfig_AutoOverwrite(t *testing.T) {
	tmpHome := t.TempDir()
	geminiDir := filepath.Join(tmpHome, ".gemini")
	if err := os.MkdirAll(geminiDir, 0755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(geminiDir, "settings.json")
	initial := `{"mcpServers":{"my-server":{"command":"uvx","args":["mcp-server-fetch@0.9"]}}}`
	if err := os.WriteFile(settingsPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre-record my-server as owned by agentkit in the store.
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "gemini")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("gemini", storePath)
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

	ad := adapter.NewGeminiAdapterWithHome(store, tmpHome)
	newEntry := domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch@1.2"},
	}
	if err := ad.WriteMCPConfig(newEntry, nil); err != nil {
		t.Fatalf("WriteMCPConfig() expected auto-overwrite (agentkit-owned), got error: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
	var cfg map[string]interface{}
	json.Unmarshal(data, &cfg)
	mcpServers := cfg["mcpServers"].(map[string]interface{})
	entry := mcpServers["my-server"].(map[string]interface{})
	args := entry["args"].([]interface{})
	if len(args) == 0 || args[len(args)-1] != "mcp-server-fetch@1.2" {
		t.Errorf("expected updated args, got %v", args)
	}
}

// TestGeminiAdapter_ReadMCPConfig_AfterWrite: ReadMCPConfig returns the written entry.
func TestGeminiAdapter_ReadMCPConfig_AfterWrite(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeGeminiAdapter(t, tmpHome)

	entry := geminiEntry()
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

// TestGeminiAdapter_RemoveMCPConfig: removes only the named key; other keys untouched.
func TestGeminiAdapter_RemoveMCPConfig(t *testing.T) {
	tmpHome := t.TempDir()
	geminiDir := filepath.Join(tmpHome, ".gemini")
	if err := os.MkdirAll(geminiDir, 0755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(geminiDir, "settings.json")
	initial := `{"theme":"dark","mcpServers":{"my-server":{"command":"uvx","args":["mcp-server-fetch"]},"other-server":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(settingsPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeGeminiAdapter(t, tmpHome)
	if err := ad.RemoveMCPConfig("my-server"); err != nil {
		t.Fatalf("RemoveMCPConfig() error: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
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

// TestGeminiAdapter_WriteSkill_CreatesDirectory: WriteSkill writes files to ~/.gemini/skills/<name>/.
func TestGeminiAdapter_WriteSkill_CreatesDirectory(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeGeminiAdapter(t, tmpHome)

	skillName := "my-skill"
	content := []byte("# My Skill")
	if err := ad.WriteSkill(skillName, map[string][]byte{"SKILL.md": content}); err != nil {
		t.Fatalf("WriteSkill() error: %v", err)
	}

	destPath := filepath.Join(tmpHome, ".gemini", "skills", skillName, "SKILL.md")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("expected file at %s: %v", destPath, err)
	}
	if string(data) != string(content) {
		t.Errorf("SKILL.md content = %q; want %q", string(data), string(content))
	}
}

// TestGeminiAdapter_RemoveSkill: WriteSkill then RemoveSkill → skill directory no longer exists.
func TestGeminiAdapter_RemoveSkill(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makeGeminiAdapter(t, tmpHome)

	skillName := "my-skill"
	if err := ad.WriteSkill(skillName, map[string][]byte{"SKILL.md": []byte("content")}); err != nil {
		t.Fatalf("WriteSkill() error: %v", err)
	}
	if err := ad.RemoveSkill(skillName); err != nil {
		t.Fatalf("RemoveSkill() error: %v", err)
	}

	skillDir := filepath.Join(tmpHome, ".gemini", "skills", skillName)
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("expected skill directory %s to be removed", skillDir)
	}
}
