package adapter

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/fileutil"
)

// CodexAdapter implements AssistantAdapter for OpenAI Codex CLI.
// It stores MCP server config as a TOML file at ~/.codex/config.toml using the
// [mcp_servers.<name>] section format. Non-mcp_servers TOML keys are preserved
// on every write (T-02-10).
type CodexAdapter struct {
	store   *config.ConfigStore
	homeDir string // empty = use os.UserHomeDir() at runtime
}

// NewCodexAdapter returns a CodexAdapter using the real home directory.
func NewCodexAdapter(store *config.ConfigStore) *CodexAdapter {
	return &CodexAdapter{store: store}
}

// NewCodexAdapterWithHome returns a CodexAdapter with an injected home directory.
// Used in tests to avoid reads/writes to the real ~/.codex/config.toml.
func NewCodexAdapterWithHome(store *config.ConfigStore, homeDir string) *CodexAdapter {
	return &CodexAdapter{store: store, homeDir: homeDir}
}

// Name returns "codex".
func (a *CodexAdapter) Name() string { return "codex" }

// home returns the effective home directory.
func (a *CodexAdapter) home() (string, error) {
	if a.homeDir != "" {
		return a.homeDir, nil
	}
	return os.UserHomeDir()
}

// configPath returns the path to ~/.codex/config.toml.
func (a *CodexAdapter) configPath() (string, error) {
	home, err := a.home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex", "config.toml"), nil
}

// readRawConfig reads the TOML config file and returns its contents as map[string]interface{}.
// Returns an empty map if the file does not exist (first install).
// Decoding into map[string]interface{} preserves all non-mcp_servers TOML keys.
func (a *CodexAdapter) readRawConfig() (map[string]interface{}, error) {
	path, err := a.configPath()
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("parsing TOML config at %s: %w", path, err)
	}
	return raw, nil
}

// writeRawConfig encodes raw as TOML and atomically writes it to the config path.
func (a *CodexAdapter) writeRawConfig(raw map[string]interface{}) error {
	path, err := a.configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(raw); err != nil {
		return fmt.Errorf("encoding TOML config: %w", err)
	}
	if err := fileutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing TOML config: %w", err)
	}
	return nil
}

// WriteMCPConfig writes an MCP server entry to Codex's config.toml.
//
// Conflict handling:
//   - If the key already exists and is NOT in installed.json → ErrForeignConflict (D-07)
//   - If the key already exists and IS in installed.json (agentkit-owned) → overwrite (D-08)
//   - Otherwise → new install
//
// Write: atomic via renameio. Post-install verify: re-reads and confirms key presence (MCP-06).
// Non-mcp_servers TOML keys are preserved (T-02-10).
func (a *CodexAdapter) WriteMCPConfig(entry domain.MCPServerEntry, _ *domain.InstalledRecord) error {
	raw, err := a.readRawConfig()
	if err != nil {
		return err
	}

	// Extract or initialise mcp_servers map.
	var mcpServers map[string]interface{}
	if v, ok := raw["mcp_servers"]; ok {
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
			old := extractCodexEntry(mcpServers[entry.Name])
			old.Name = entry.Name
			return &ErrForeignConflict{OldEntry: old, NewEntry: entry}
		}
		// agentkit-owned: auto-overwrite (D-08).
	}

	// Build the entry map — command, args, and optional env subtable.
	entryMap := map[string]interface{}{
		"command": entry.Command,
		"args":    entry.Args,
	}
	if len(entry.Env) > 0 {
		// Write env as a sub-table so TOML renders it as [mcp_servers.<name>.env].
		envMap := make(map[string]interface{}, len(entry.Env))
		for k, v := range entry.Env {
			envMap[k] = v
		}
		entryMap["env"] = envMap
	}
	mcpServers[entry.Name] = entryMap
	raw["mcp_servers"] = mcpServers

	if err := a.writeRawConfig(raw); err != nil {
		return err
	}

	// Post-install verify: re-read and confirm key is present (MCP-06).
	result, err := a.ReadMCPConfig()
	if err != nil {
		return fmt.Errorf("post-install verify failed to parse config: %w", err)
	}
	if _, ok := result[entry.Name]; !ok {
		return fmt.Errorf("post-install verify failed: mcp_servers.%s not found after write", entry.Name)
	}
	return nil
}

// RemoveMCPConfig removes the named MCP server entry from config.toml (D-09).
// Only the named key is removed; all other TOML keys are untouched.
func (a *CodexAdapter) RemoveMCPConfig(name string) error {
	raw, err := a.readRawConfig()
	if err != nil {
		return err
	}

	if v, ok := raw["mcp_servers"]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			delete(m, name)
			raw["mcp_servers"] = m
		}
	}

	return a.writeRawConfig(raw)
}

// ReadMCPConfig reads and parses all mcp_servers entries from config.toml.
func (a *CodexAdapter) ReadMCPConfig() (map[string]domain.MCPServerEntry, error) {
	raw, err := a.readRawConfig()
	if err != nil {
		return nil, err
	}

	result := map[string]domain.MCPServerEntry{}
	v, ok := raw["mcp_servers"]
	if !ok {
		return result, nil
	}
	servers, ok := v.(map[string]interface{})
	if !ok {
		return result, nil
	}
	for name, val := range servers {
		entry := extractCodexEntry(val)
		entry.Name = name
		result[name] = entry
	}
	return result, nil
}

// WriteSkill is not supported for Codex CLI — it has no user-global skill directory.
func (a *CodexAdapter) WriteSkill(name string, _ map[string][]byte) error {
	return fmt.Errorf("codex adapter: WriteSkill not supported — Codex CLI has no user-global skill directory: %w", ErrNotSupported)
}

// RemoveSkill is not supported for Codex CLI.
func (a *CodexAdapter) RemoveSkill(_ string) error {
	return fmt.Errorf("codex adapter: RemoveSkill not supported — Codex CLI has no user-global skill directory: %w", ErrNotSupported)
}

// extractCodexEntry converts a raw TOML-decoded interface{} value to a domain.MCPServerEntry.
func extractCodexEntry(val interface{}) domain.MCPServerEntry {
	m, ok := val.(map[string]interface{})
	if !ok {
		return domain.MCPServerEntry{}
	}
	entry := domain.MCPServerEntry{}
	if cmd, ok := m["command"].(string); ok {
		entry.Command = cmd
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
