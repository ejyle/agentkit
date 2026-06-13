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

// jsonMCPAdapter is a shared base for AI coding assistants that store MCP server config
// as a JSON file with a top-level key (e.g. "mcpServers") containing a map of entries.
//
// It implements the non-destructive read-merge-write pattern from ClaudeCodeAdapter,
// extracted into a reusable struct with injected configPath and extraFields functions
// so each embedding adapter can customise its file path and entry shape.
type jsonMCPAdapter struct {
	store      *config.ConfigStore
	homeDir    string // empty = use os.UserHomeDir() at runtime
	configPath func(home string) (string, error)
	mcpKey     string // top-level JSON key for MCP servers (typically "mcpServers")
	// extraFields returns additional key-value pairs to inject into each MCP entry map.
	// May be nil — in which case no extra fields are written.
	extraFields func(entry domain.MCPServerEntry) map[string]interface{}
}

// home returns the effective home directory.
func (a *jsonMCPAdapter) home() (string, error) {
	if a.homeDir != "" {
		return a.homeDir, nil
	}
	return os.UserHomeDir()
}

// readRawConfig reads the config file and returns its contents as map[string]interface{}.
// Returns an empty map if the file does not exist (first install).
func (a *jsonMCPAdapter) readRawConfig() (map[string]interface{}, error) {
	home, err := a.home()
	if err != nil {
		return nil, err
	}
	path, err := a.configPath(home)
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
		return nil, fmt.Errorf("parsing MCP config at %s: %w", path, err)
	}
	return raw, nil
}

// WriteMCPConfig writes an MCP server entry to the assistant's config file.
//
// Conflict handling:
//   - If the key already exists and is NOT in installed.json → ErrForeignConflict (D-07)
//   - If the key already exists and IS in installed.json (agentkit-owned) → overwrite (D-08)
//   - Otherwise → new install
//
// Write: atomic via renameio.
// Post-install verify: re-reads written config and confirms key presence (MCP-06).
func (a *jsonMCPAdapter) WriteMCPConfig(entry domain.MCPServerEntry, _ *domain.InstalledRecord) error {
	raw, err := a.readRawConfig()
	if err != nil {
		return err
	}

	// Extract or initialise the MCP servers map.
	var mcpServers map[string]interface{}
	if v, ok := raw[a.mcpKey]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			mcpServers = m
		}
	}
	if mcpServers == nil {
		mcpServers = map[string]interface{}{}
	}

	// Conflict check: does this key already exist?
	if _, exists := mcpServers[entry.Name]; exists {
		_, owned, err := a.store.GetRecord(entry.Name)
		if err != nil {
			return fmt.Errorf("checking ownership for %q: %w", entry.Name, err)
		}
		if !owned {
			old := extractEntryFromRaw(mcpServers[entry.Name])
			old.Name = entry.Name
			return &ErrForeignConflict{OldEntry: old, NewEntry: entry}
		}
		// agentkit-owned: auto-overwrite (D-08).
	}

	// Build the entry map — command, args, optional env, and any adapter-specific extra fields.
	entryMap := map[string]interface{}{
		"command": entry.Command,
		"args":    entry.Args,
	}
	if len(entry.Env) > 0 {
		entryMap["env"] = entry.Env
	}
	if a.extraFields != nil {
		for k, v := range a.extraFields(entry) {
			entryMap[k] = v
		}
	}
	mcpServers[entry.Name] = entryMap
	raw[a.mcpKey] = mcpServers

	// Atomic write via renameio.
	home, err := a.home()
	if err != nil {
		return err
	}
	path, err := a.configPath(home)
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
		return fmt.Errorf("writing MCP config: %w", err)
	}

	// Post-install verify: re-read and confirm key is present (MCP-06).
	result, err := a.ReadMCPConfig()
	if err != nil {
		return fmt.Errorf("post-install verify failed to parse config: %w", err)
	}
	if _, ok := result[entry.Name]; !ok {
		return fmt.Errorf("post-install verify failed: %s.%s not found after write", a.mcpKey, entry.Name)
	}
	return nil
}

// RemoveMCPConfig removes the named MCP server entry from the config file (D-09).
// Only the named key is removed; all other keys are untouched.
func (a *jsonMCPAdapter) RemoveMCPConfig(name string) error {
	raw, err := a.readRawConfig()
	if err != nil {
		return err
	}

	if v, ok := raw[a.mcpKey]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			delete(m, name)
			raw[a.mcpKey] = m
		}
	}

	home, err := a.home()
	if err != nil {
		return err
	}
	path, err := a.configPath(home)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteFile(path, data, 0644)
}

// ReadMCPConfig reads and parses all MCP server entries from the config file.
// It handles both string command (standard) and []interface{} command array (OpenCode).
func (a *jsonMCPAdapter) ReadMCPConfig() (map[string]domain.MCPServerEntry, error) {
	raw, err := a.readRawConfig()
	if err != nil {
		return nil, err
	}

	result := map[string]domain.MCPServerEntry{}
	v, ok := raw[a.mcpKey]
	if !ok {
		return result, nil
	}
	servers, ok := v.(map[string]interface{})
	if !ok {
		return result, nil
	}
	for name, val := range servers {
		entry := extractEntryFromRaw(val)
		entry.Name = name
		result[name] = entry
	}
	return result, nil
}

// extractEntryFromRaw converts a raw JSON decoded interface{} value to a domain.MCPServerEntry.
// It handles both the standard {"command": "...", "args": [...]} form and the OpenCode
// array-command form where command is a []interface{} with [0]=command and rest=args.
func extractEntryFromRaw(val interface{}) domain.MCPServerEntry {
	m, ok := val.(map[string]interface{})
	if !ok {
		return domain.MCPServerEntry{}
	}
	entry := domain.MCPServerEntry{}

	// Handle both string command (standard) and array command (OpenCode).
	switch cmd := m["command"].(type) {
	case string:
		entry.Command = cmd
	case []interface{}:
		// OpenCode stores command as an array: [command, arg1, arg2, ...]
		if len(cmd) > 0 {
			if s, ok := cmd[0].(string); ok {
				entry.Command = s
			}
			for _, a := range cmd[1:] {
				if s, ok := a.(string); ok {
					entry.Args = append(entry.Args, s)
				}
			}
		}
	}

	if argsRaw, ok := m["args"].([]interface{}); ok {
		for _, a := range argsRaw {
			if s, ok := a.(string); ok {
				entry.Args = append(entry.Args, s)
			}
		}
	}
	if envRaw, ok := m["env"].(map[string]interface{}); ok {
		entry.Env = make(map[string]string, len(envRaw))
		for k, v := range envRaw {
			if s, ok := v.(string); ok {
				entry.Env[k] = s
			}
		}
	}
	return entry
}
