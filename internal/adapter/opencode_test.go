package adapter_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ejyle/agentkit/internal/adapter"
	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
)

// makeOpenCodeAdapter creates an OpenCodeAdapter using tmpConfigDir as the config directory.
func makeOpenCodeAdapter(t *testing.T, tmpConfigDir string) *adapter.OpenCodeAdapter {
	t.Helper()
	storeDir := filepath.Join(tmpConfigDir, "agentkit", "opencode")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(storeDir, "installed.json")
	store := config.NewConfigStoreWithPath("opencode", storePath)
	return adapter.NewOpenCodeAdapterWithConfigDir(store, tmpConfigDir)
}

// openCodeEntry returns a standard test MCPServerEntry for opencode.
func openCodeEntry() domain.MCPServerEntry {
	return domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch"},
	}
}

// TestOpenCodeAdapter_WriteMCPConfig_CreatesFile verifies that WriteMCPConfig creates
// <configDir>/opencode/opencode.json with the "mcp" key.
func TestOpenCodeAdapter_WriteMCPConfig_CreatesFile(t *testing.T) {
	tmpConfigDir := t.TempDir()
	ad := makeOpenCodeAdapter(t, tmpConfigDir)

	if err := ad.WriteMCPConfig(openCodeEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	configPath := filepath.Join(tmpConfigDir, "opencode", "opencode.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("expected %s to be created", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	mcpMap, ok := raw["mcp"].(map[string]interface{})
	if !ok {
		t.Fatalf("\"mcp\" key missing or wrong type (got %T)", raw["mcp"])
	}
	if _, ok := mcpMap["my-server"]; !ok {
		t.Fatalf("my-server entry missing from mcp map")
	}
	if _, ok := raw["mcpServers"]; ok {
		t.Error("should not have \"mcpServers\" key; OpenCode uses \"mcp\"")
	}
}

// TestOpenCodeAdapter_WriteMCPConfig_CommandIsArray verifies the command field is a JSON array.
func TestOpenCodeAdapter_WriteMCPConfig_CommandIsArray(t *testing.T) {
	tmpConfigDir := t.TempDir()
	ad := makeOpenCodeAdapter(t, tmpConfigDir)

	if err := ad.WriteMCPConfig(openCodeEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	configPath := filepath.Join(tmpConfigDir, "opencode", "opencode.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	mcpMap := raw["mcp"].(map[string]interface{})
	srv := mcpMap["my-server"].(map[string]interface{})
	cmdVal := srv["command"]
	cmdArr, ok := cmdVal.([]interface{})
	if !ok {
		t.Fatalf("command field is %T, want []interface{}", cmdVal)
	}
	if len(cmdArr) == 0 {
		t.Fatal("command array is empty")
	}
	if cmdArr[0] != "uvx" {
		t.Errorf("command[0] = %v; want uvx", cmdArr[0])
	}
	if len(cmdArr) < 2 || cmdArr[1] != "mcp-server-fetch" {
		t.Errorf("command[1] = %v; want mcp-server-fetch", cmdArr[1])
	}
}

// TestOpenCodeAdapter_WriteMCPConfig_EnvironmentKey verifies the env is written under "environment".
func TestOpenCodeAdapter_WriteMCPConfig_EnvironmentKey(t *testing.T) {
	tmpConfigDir := t.TempDir()
	ad := makeOpenCodeAdapter(t, tmpConfigDir)

	entry := domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch"},
		Env:     map[string]string{"K": "V"},
	}
	if err := ad.WriteMCPConfig(entry, nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	configPath := filepath.Join(tmpConfigDir, "opencode", "opencode.json")
	data, _ := os.ReadFile(configPath)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	mcpMap := raw["mcp"].(map[string]interface{})
	srv := mcpMap["my-server"].(map[string]interface{})

	if _, hasEnv := srv["env"]; hasEnv {
		t.Error("should NOT have \"env\" key; OpenCode uses \"environment\"")
	}
	envMap, ok := srv["environment"].(map[string]interface{})
	if !ok {
		t.Fatalf("\"environment\" key missing or wrong type")
	}
	if envMap["K"] != "V" {
		t.Errorf("environment.K = %v; want V", envMap["K"])
	}
}

// TestOpenCodeAdapter_WriteMCPConfig_PreservesExistingKeys verifies existing JSON keys survive.
func TestOpenCodeAdapter_WriteMCPConfig_PreservesExistingKeys(t *testing.T) {
	tmpConfigDir := t.TempDir()
	configDir := filepath.Join(tmpConfigDir, "opencode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "opencode.json")
	initial := `{"theme":"dark","other":42}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeOpenCodeAdapter(t, tmpConfigDir)
	if err := ad.WriteMCPConfig(openCodeEntry(), nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if raw["theme"] != "dark" {
		t.Errorf("theme clobbered; got %v", raw["theme"])
	}
	if raw["other"] != float64(42) {
		t.Errorf("other clobbered; got %v", raw["other"])
	}
}

// TestOpenCodeAdapter_WriteMCPConfig_ErrForeignConflict verifies foreign conflict detection.
func TestOpenCodeAdapter_WriteMCPConfig_ErrForeignConflict(t *testing.T) {
	tmpConfigDir := t.TempDir()
	configDir := filepath.Join(tmpConfigDir, "opencode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "opencode.json")
	initial := `{"mcp":{"my-server":{"type":"local","command":["other"],"enabled":true}}}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeOpenCodeAdapter(t, tmpConfigDir)
	err := ad.WriteMCPConfig(openCodeEntry(), nil)
	if err == nil {
		t.Fatal("WriteMCPConfig() expected ErrForeignConflict, got nil")
	}
	var conflictErr *adapter.ErrForeignConflict
	if !adapter.AsErrForeignConflict(err, &conflictErr) {
		t.Errorf("WriteMCPConfig() error = %v; want ErrForeignConflict", err)
	}
}

// TestOpenCodeAdapter_ReadMCPConfig_AfterWrite verifies round-trip: Write then Read.
func TestOpenCodeAdapter_ReadMCPConfig_AfterWrite(t *testing.T) {
	tmpConfigDir := t.TempDir()
	ad := makeOpenCodeAdapter(t, tmpConfigDir)

	entry := domain.MCPServerEntry{
		Name:    "my-server",
		Command: "uvx",
		Args:    []string{"mcp-server-fetch", "--port", "8080"},
	}
	if err := ad.WriteMCPConfig(entry, nil); err != nil {
		t.Fatalf("WriteMCPConfig() error: %v", err)
	}

	entries, err := ad.ReadMCPConfig()
	if err != nil {
		t.Fatalf("ReadMCPConfig() error: %v", err)
	}
	got, ok := entries["my-server"]
	if !ok {
		t.Fatal("ReadMCPConfig() missing my-server after write")
	}
	if got.Command != "uvx" {
		t.Errorf("Command = %q; want uvx", got.Command)
	}
	if len(got.Args) != 3 || got.Args[0] != "mcp-server-fetch" || got.Args[1] != "--port" || got.Args[2] != "8080" {
		t.Errorf("Args = %v; want [mcp-server-fetch --port 8080]", got.Args)
	}
}

// TestOpenCodeAdapter_RemoveMCPConfig verifies removal of named entry only.
func TestOpenCodeAdapter_RemoveMCPConfig(t *testing.T) {
	tmpConfigDir := t.TempDir()
	configDir := filepath.Join(tmpConfigDir, "opencode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "opencode.json")
	initial := `{"mcp":{"my-server":{"type":"local","command":["uvx"],"enabled":true},"other":{"type":"local","command":["node"],"enabled":true}}}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	ad := makeOpenCodeAdapter(t, tmpConfigDir)
	if err := ad.RemoveMCPConfig("my-server"); err != nil {
		t.Fatalf("RemoveMCPConfig() error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	mcpMap := raw["mcp"].(map[string]interface{})
	if _, ok := mcpMap["my-server"]; ok {
		t.Error("my-server should have been removed")
	}
	if _, ok := mcpMap["other"]; !ok {
		t.Error("other server should remain")
	}
}

// TestOpenCodeAdapter_WriteSkill_ErrNotSupported verifies ErrNotSupported for WriteSkill.
func TestOpenCodeAdapter_WriteSkill_ErrNotSupported(t *testing.T) {
	tmpConfigDir := t.TempDir()
	ad := makeOpenCodeAdapter(t, tmpConfigDir)

	err := ad.WriteSkill("some-skill", map[string][]byte{"SKILL.md": []byte("content")})
	if err == nil {
		t.Fatal("WriteSkill() expected ErrNotSupported, got nil")
	}
	if !isErrNotSupported(err) {
		t.Errorf("WriteSkill() error = %v; want wrapping ErrNotSupported", err)
	}
}
