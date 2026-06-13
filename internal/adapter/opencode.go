package adapter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/fileutil"
)

// OpenCodeAdapter implements AssistantAdapter for OpenCode.
// It stores MCP server config as a JSON file at <UserConfigDir>/opencode/opencode.json
// using the "mcp" key (not "mcpServers"). The command field is a JSON array combining
// command and args (T-02-11). The env key is "environment" (not "env") (T-02-12).
//
// OpenCode's schema is fundamentally different from other JSON adapters and cannot
// embed jsonMCPAdapter — it would require overriding every method.
type OpenCodeAdapter struct {
	store     *config.ConfigStore
	configDir string // empty = use os.UserConfigDir() at runtime
}

// NewOpenCodeAdapter returns an OpenCodeAdapter using the real user config directory.
func NewOpenCodeAdapter(store *config.ConfigStore) *OpenCodeAdapter {
	return &OpenCodeAdapter{store: store}
}

// NewOpenCodeAdapterWithConfigDir returns an OpenCodeAdapter with an injected config directory.
// Used in tests to avoid reads/writes to the real opencode config.
func NewOpenCodeAdapterWithConfigDir(store *config.ConfigStore, configDir string) *OpenCodeAdapter {
	return &OpenCodeAdapter{store: store, configDir: configDir}
}

// Name returns "opencode".
func (a *OpenCodeAdapter) Name() string { return "opencode" }

// effectiveConfigDir returns the config directory, falling back to os.UserConfigDir.
func (a *OpenCodeAdapter) effectiveConfigDir() (string, error) {
	if a.configDir != "" {
		return a.configDir, nil
	}
	return os.UserConfigDir()
}

// configPath returns the path to <configDir>/opencode/opencode.json.
func (a *OpenCodeAdapter) configPath() (string, error) {
	dir, err := a.effectiveConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "opencode", "opencode.json"), nil
}

// readRawConfig reads the JSON config file and returns its contents as map[string]interface{}.
// Returns an empty map if the file does not exist (first install).
func (a *OpenCodeAdapter) readRawConfig() (map[string]interface{}, error) {
	path, err := a.configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]interface{}{}, nil
	}
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing OpenCode config at %s: %w", path, err)
	}
	return raw, nil
}

// WriteMCPConfig writes an MCP server entry to OpenCode's opencode.json.
//
// Key schema differences from other JSON adapters:
//   - Top-level key is "mcp" (not "mcpServers")
//   - command field is a JSON array: append([]string{entry.Command}, entry.Args...) (T-02-11)
//   - env key is "environment" (not "env") (T-02-12)
//   - type field is "local" and enabled field is true
//
// Conflict handling:
//   - If the key already exists and is NOT in installed.json → ErrForeignConflict (D-07)
//   - If the key already exists and IS in installed.json (agentkit-owned) → overwrite (D-08)
//
// Write: atomic via renameio. Post-install verify: re-reads and confirms key presence (MCP-06).
func (a *OpenCodeAdapter) WriteMCPConfig(entry domain.MCPServerEntry, _ *domain.InstalledRecord) error {
	raw, err := a.readRawConfig()
	if err != nil {
		return err
	}

	// Extract or initialise "mcp" map (NOT "mcpServers").
	var mcpMap map[string]interface{}
	if v, ok := raw["mcp"]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			mcpMap = m
		}
	}
	if mcpMap == nil {
		mcpMap = map[string]interface{}{}
	}

	// Conflict check: does this key already exist?
	if _, exists := mcpMap[entry.Name]; exists {
		_, owned, err := a.store.GetRecord(entry.Name)
		if err != nil {
			return fmt.Errorf("checking ownership for %q: %w", entry.Name, err)
		}
		if !owned {
			old := extractOpenCodeEntry(mcpMap[entry.Name])
			old.Name = entry.Name
			return &ErrForeignConflict{OldEntry: old, NewEntry: entry}
		}
		// agentkit-owned: auto-overwrite (D-08).
	}

	// Build command as a JSON array: [command, arg1, arg2, ...] (T-02-11).
	cmdArray := append([]string{entry.Command}, entry.Args...)
	cmdIface := make([]interface{}, len(cmdArray))
	for i, s := range cmdArray {
		cmdIface[i] = s
	}

	entryMap := map[string]interface{}{
		"type":    "local",
		"command": cmdIface,
		"enabled": true,
	}
	// Use "environment" key (NOT "env") (T-02-12).
	if len(entry.Env) > 0 {
		envMap := make(map[string]interface{}, len(entry.Env))
		for k, v := range entry.Env {
			envMap[k] = v
		}
		entryMap["environment"] = envMap
	}
	mcpMap[entry.Name] = entryMap
	raw["mcp"] = mcpMap

	path, err := a.configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := fileutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing OpenCode config: %w", err)
	}

	// Post-install verify: re-read and confirm key is present (MCP-06).
	result, err := a.ReadMCPConfig()
	if err != nil {
		return fmt.Errorf("post-install verify failed to parse config: %w", err)
	}
	if _, ok := result[entry.Name]; !ok {
		return fmt.Errorf("post-install verify failed: mcp.%s not found after write", entry.Name)
	}
	return nil
}

// RemoveMCPConfig removes the named MCP server entry from opencode.json (D-09).
// Only the named key is removed; all other JSON keys are untouched.
func (a *OpenCodeAdapter) RemoveMCPConfig(name string) error {
	raw, err := a.readRawConfig()
	if err != nil {
		return err
	}

	if v, ok := raw["mcp"]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			delete(m, name)
			raw["mcp"] = m
		}
	}

	path, err := a.configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteFile(path, data, 0644)
}

// ReadMCPConfig reads and parses all mcp entries from opencode.json.
// The command array is split: arr[0]=Command, arr[1:]=Args.
// The env source key is "environment" (not "env").
func (a *OpenCodeAdapter) ReadMCPConfig() (map[string]domain.MCPServerEntry, error) {
	raw, err := a.readRawConfig()
	if err != nil {
		return nil, err
	}

	result := map[string]domain.MCPServerEntry{}
	v, ok := raw["mcp"]
	if !ok {
		return result, nil
	}
	servers, ok := v.(map[string]interface{})
	if !ok {
		return result, nil
	}
	for name, val := range servers {
		entry := extractOpenCodeEntry(val)
		entry.Name = name
		result[name] = entry
	}
	return result, nil
}

// WriteSkill is not supported for OpenCode — it has no user-global skill directory.
func (a *OpenCodeAdapter) WriteSkill(name string, _ map[string][]byte) error {
	return fmt.Errorf("opencode adapter: WriteSkill not supported — OpenCode has no user-global skill directory: %w", ErrNotSupported)
}

// RemoveSkill is not supported for OpenCode.
func (a *OpenCodeAdapter) RemoveSkill(_ string) error {
	return fmt.Errorf("opencode adapter: RemoveSkill not supported — OpenCode has no user-global skill directory: %w", ErrNotSupported)
}

// extractOpenCodeEntry converts a raw JSON-decoded interface{} value to a domain.MCPServerEntry.
// The command field is a []interface{} array where arr[0]=Command and arr[1:]=Args (T-02-11).
// The env source key is "environment" (not "env") (T-02-12).
func extractOpenCodeEntry(val interface{}) domain.MCPServerEntry {
	m, ok := val.(map[string]interface{})
	if !ok {
		return domain.MCPServerEntry{}
	}
	entry := domain.MCPServerEntry{}

	// command is a JSON array: [command, arg1, arg2, ...] (T-02-11).
	if cmdRaw, ok := m["command"].([]interface{}); ok {
		if len(cmdRaw) > 0 {
			if s, ok := cmdRaw[0].(string); ok {
				entry.Command = s
			}
			for _, a := range cmdRaw[1:] {
				if s, ok := a.(string); ok {
					entry.Args = append(entry.Args, s)
				}
			}
		}
	}

	// env source key is "environment" (not "env") (T-02-12).
	if envRaw, ok := m["environment"].(map[string]interface{}); ok {
		entry.Env = make(map[string]string, len(envRaw))
		for k, v := range envRaw {
			if s, ok := v.(string); ok {
				entry.Env[k] = s
			}
		}
	}
	return entry
}
