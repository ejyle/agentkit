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

// makePiAdapter creates a PiAdapter configured to use tmpHome as the home directory.
func makePiAdapter(t *testing.T, tmpHome string) *adapter.PiAdapter {
	t.Helper()
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "pi")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("pi", storePath)
	return adapter.NewPiAdapterWithHome(store, tmpHome)
}

// piEntry returns a standard test MCPServerEntry for a Pi test.
func piEntry() domain.MCPServerEntry {
	return domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch"},
	}
}

// TestPiAdapter_WriteMCPConfig_CreatesFile: empty tmpHome → WriteMCPConfig → file at
// tmpHome/.pi/agent/mcp.json; "mcpServers" key present; entry has correct Command.
func TestPiAdapter_WriteMCPConfig_CreatesFile(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makePiAdapter(t, tmpHome)

	if err := ad.WriteMCPConfig(piEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	mcpPath := filepath.Join(tmpHome, ".pi", "agent", "mcp.json")
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
	if entry["command"] != "uvx" {
		t.Errorf("entry.command = %v; want %q", entry["command"], "uvx")
	}
}

// TestPiAdapter_WriteMCPConfig_PreservesExistingKeys: pre-existing top-level keys survive a write.
func TestPiAdapter_WriteMCPConfig_PreservesExistingKeys(t *testing.T) {
	tmpHome := t.TempDir()
	piDir := filepath.Join(tmpHome, ".pi", "agent")
	if err := os.MkdirAll(piDir, 0755); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(piDir, "mcp.json")
	if err := os.WriteFile(mcpPath, []byte(`{"existingKey":"value","otherKey":42}`), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makePiAdapter(t, tmpHome)
	if err := ad.WriteMCPConfig(piEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	data, _ := os.ReadFile(mcpPath)
	var cfg map[string]interface{}
	json.Unmarshal(data, &cfg)
	if cfg["existingKey"] != "value" {
		t.Errorf("existingKey was clobbered; got %v", cfg["existingKey"])
	}
	if cfg["otherKey"] != float64(42) {
		t.Errorf("otherKey was clobbered; got %v", cfg["otherKey"])
	}
}

// TestPiAdapter_WriteMCPConfig_ErrForeignConflict: key exists but not owned by agentkit.
func TestPiAdapter_WriteMCPConfig_ErrForeignConflict(t *testing.T) {
	tmpHome := t.TempDir()
	piDir := filepath.Join(tmpHome, ".pi", "agent")
	if err := os.MkdirAll(piDir, 0755); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(piDir, "mcp.json")
	initial := `{"mcpServers":{"my-server":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(mcpPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makePiAdapter(t, tmpHome)
	err := ad.WriteMCPConfig(piEntry(), nil)
	if err == nil {
		t.Fatal("WriteMCPConfig() expected ErrForeignConflict, got nil")
	}
	var conflictErr *adapter.ErrForeignConflict
	if !adapter.AsErrForeignConflict(err, &conflictErr) {
		t.Errorf("WriteMCPConfig() error = %v; want ErrForeignConflict", err)
	}
}

// TestPiAdapter_WriteMCPConfig_AutoOverwrite: key exists and IS owned by agentkit → overwrite.
func TestPiAdapter_WriteMCPConfig_AutoOverwrite(t *testing.T) {
	tmpHome := t.TempDir()
	piDir := filepath.Join(tmpHome, ".pi", "agent")
	if err := os.MkdirAll(piDir, 0755); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(piDir, "mcp.json")
	initial := `{"mcpServers":{"my-server":{"command":"uvx","args":["mcp-server-fetch@0.9"]}}}`
	if err := os.WriteFile(mcpPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre-record my-server as owned by agentkit in the store.
	storeDir := filepath.Join(tmpHome, "config", "agentkit", "pi")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("pi", storePath)
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

	ad := adapter.NewPiAdapterWithHome(store, tmpHome)
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

// TestPiAdapter_ReadMCPConfig_AfterWrite: ReadMCPConfig returns the written entry.
func TestPiAdapter_ReadMCPConfig_AfterWrite(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makePiAdapter(t, tmpHome)

	entry := piEntry()
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

// TestPiAdapter_RemoveMCPConfig: removes only the named key; other keys untouched.
func TestPiAdapter_RemoveMCPConfig(t *testing.T) {
	tmpHome := t.TempDir()
	piDir := filepath.Join(tmpHome, ".pi", "agent")
	if err := os.MkdirAll(piDir, 0755); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(piDir, "mcp.json")
	initial := `{"existingKey":"value","mcpServers":{"my-server":{"command":"uvx","args":["mcp-server-fetch"]},"other-server":{"command":"node","args":["server.js"]}}}`
	if err := os.WriteFile(mcpPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makePiAdapter(t, tmpHome)
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
	if cfg["existingKey"] != "value" {
		t.Errorf("existingKey was clobbered; got %v", cfg["existingKey"])
	}
}

// TestPiAdapter_WriteSkill_CreatesDirectory: WriteSkill writes files to ~/.agents/skills/<name>/
// (NOT ~/.pi/skills/<name>/) — validates D-05 and A4 assumption from RESEARCH.md.
func TestPiAdapter_WriteSkill_CreatesDirectory(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makePiAdapter(t, tmpHome)

	skillName := "my-skill"
	content := []byte("# My Skill")
	if err := ad.WriteSkill(skillName, map[string][]byte{"SKILL.md": content}); err != nil {
		t.Fatalf("WriteSkill() error: %v", err)
	}

	// Pi skills MUST resolve to ~/.agents/skills/<name>/ (not ~/.pi/skills/<name>/).
	destPath := filepath.Join(tmpHome, ".agents", "skills", skillName, "SKILL.md")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("expected file at %s (not .pi/skills/): %v", destPath, err)
	}
	if string(data) != string(content) {
		t.Errorf("SKILL.md content = %q; want %q", string(data), string(content))
	}

	// Negative check: confirm ~/.pi/skills/<name>/ does NOT exist.
	wrongPath := filepath.Join(tmpHome, ".pi", "skills", skillName)
	if _, err := os.Stat(wrongPath); !os.IsNotExist(err) {
		t.Errorf("skill directory should NOT exist at %s (wrong path)", wrongPath)
	}
}

// TestPiAdapter_RemoveSkill: WriteSkill then RemoveSkill → skill directory no longer exists.
func TestPiAdapter_RemoveSkill(t *testing.T) {
	tmpHome := t.TempDir()
	ad := makePiAdapter(t, tmpHome)

	skillName := "my-skill"
	if err := ad.WriteSkill(skillName, map[string][]byte{"SKILL.md": []byte("content")}); err != nil {
		t.Fatalf("WriteSkill() error: %v", err)
	}
	if err := ad.RemoveSkill(skillName); err != nil {
		t.Fatalf("RemoveSkill() error: %v", err)
	}

	skillDir := filepath.Join(tmpHome, ".agents", "skills", skillName)
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("expected skill directory %s to be removed", skillDir)
	}
}
