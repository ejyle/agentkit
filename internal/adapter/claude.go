package adapter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
	"github.com/google/renameio/v2"
)

// ClaudeCodeAdapter implements AssistantAdapter for Claude Code.
// It performs runtime path detection (never hardcodes ~/.claude/settings.json),
// non-destructive merge-writes via map[string]interface{}, atomic file writes via
// renameio, and post-install verification (T-03-04, T-03-05, T-03-06, MCP-07).
type ClaudeCodeAdapter struct {
	store   *config.ConfigStore
	homeDir string // empty = use os.UserHomeDir() at runtime
}

// NewClaudeCodeAdapter returns a ClaudeCodeAdapter using the real home directory.
func NewClaudeCodeAdapter(store *config.ConfigStore) *ClaudeCodeAdapter {
	return &ClaudeCodeAdapter{store: store}
}

// NewClaudeCodeAdapterWithHome returns a ClaudeCodeAdapter with an injected home directory.
// Used in tests to avoid reads/writes to the real ~/.claude.json.
func NewClaudeCodeAdapterWithHome(store *config.ConfigStore, homeDir string) *ClaudeCodeAdapter {
	return &ClaudeCodeAdapter{store: store, homeDir: homeDir}
}

// Name returns "claude".
func (a *ClaudeCodeAdapter) Name() string { return "claude" }

// home returns the home directory, falling back to os.UserHomeDir.
func (a *ClaudeCodeAdapter) home() (string, error) {
	if a.homeDir != "" {
		return a.homeDir, nil
	}
	return os.UserHomeDir()
}

// mcpConfigPath implements runtime path detection (Pattern 2, Pitfall 1):
//  1. Stat home/.claude.json — if it exists, use it (current Claude Code user-scope MCP path)
//  2. Stat home/.claude/settings.json — if it exists, use it (legacy fallback)
//  3. Otherwise return home/.claude.json (create on first write — D-12)
func (a *ClaudeCodeAdapter) mcpConfigPath() (string, error) {
	home, err := a.home()
	if err != nil {
		return "", err
	}
	primary := filepath.Join(home, ".claude.json")
	if _, err := os.Stat(primary); err == nil {
		return primary, nil
	}
	legacy := filepath.Join(home, ".claude", "settings.json")
	if _, err := os.Stat(legacy); err == nil {
		return legacy, nil
	}
	return primary, nil
}

// readRawConfig reads the MCP config file and unmarshals it as map[string]interface{}
// to preserve all keys we do not manage. Returns an empty map if the file does not exist.
func (a *ClaudeCodeAdapter) readRawConfig() (map[string]interface{}, error) {
	path, err := a.mcpConfigPath()
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

// WriteMCPConfig writes an MCP server entry to Claude Code's config file.
//
// Conflict handling:
//   - If the key already exists and is NOT in installed.json → ErrForeignConflict (D-07, T-03-05)
//   - If the key already exists and IS in installed.json (agentkit-owned) → overwrite (D-08)
//   - Otherwise → new install
//
// Write: atomic via renameio (T-03-04).
// Post-install verify: re-reads written config and confirms key presence (T-03-06, MCP-06).
func (a *ClaudeCodeAdapter) WriteMCPConfig(entry domain.MCPServerEntry, _ *domain.InstalledRecord) error {
	raw, err := a.readRawConfig()
	if err != nil {
		return err
	}

	// Extract or initialise mcpServers map.
	var mcpServers map[string]interface{}
	if v, ok := raw["mcpServers"]; ok {
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
			// Foreign conflict — read back old entry details.
			old := extractEntry(mcpServers[entry.Name])
			old.Name = entry.Name
			return &ErrForeignConflict{OldEntry: old, NewEntry: entry}
		}
		// agentkit-owned: auto-overwrite (D-08).
	}

	// Build the mcpServers entry map — only write command and args (+ env if non-empty).
	entryMap := map[string]interface{}{
		"command": entry.Command,
		"args":    entry.Args,
	}
	if len(entry.Env) > 0 {
		entryMap["env"] = entry.Env
	}
	mcpServers[entry.Name] = entryMap
	raw["mcpServers"] = mcpServers

	// Atomic write via renameio (T-03-04).
	path, err := a.mcpConfigPath()
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
	if err := renameio.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing MCP config: %w", err)
	}

	// Post-install verify (T-03-06, MCP-06): re-read and confirm key is present.
	result, err := a.ReadMCPConfig()
	if err != nil {
		return fmt.Errorf("post-install verify failed to parse config: %w", err)
	}
	if _, ok := result[entry.Name]; !ok {
		return fmt.Errorf("post-install verify failed: mcpServers.%s not found after write", entry.Name)
	}
	return nil
}

// RemoveMCPConfig removes the named MCP server entry from Claude Code's config file (D-09).
// Only the named key under mcpServers is removed; all other keys are untouched.
func (a *ClaudeCodeAdapter) RemoveMCPConfig(name string) error {
	raw, err := a.readRawConfig()
	if err != nil {
		return err
	}

	if v, ok := raw["mcpServers"]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			delete(m, name)
			raw["mcpServers"] = m
		}
	}

	path, err := a.mcpConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return renameio.WriteFile(path, data, 0644)
}

// ReadMCPConfig reads and parses all mcpServers entries from Claude Code's config file.
func (a *ClaudeCodeAdapter) ReadMCPConfig() (map[string]domain.MCPServerEntry, error) {
	raw, err := a.readRawConfig()
	if err != nil {
		return nil, err
	}

	result := map[string]domain.MCPServerEntry{}
	v, ok := raw["mcpServers"]
	if !ok {
		return result, nil
	}
	servers, ok := v.(map[string]interface{})
	if !ok {
		return result, nil
	}
	for name, val := range servers {
		entry := extractEntry(val)
		entry.Name = name
		result[name] = entry
	}
	return result, nil
}

// WriteSkill writes skill files into ~/.claude/skills/<name>/.
func (a *ClaudeCodeAdapter) WriteSkill(name string, files map[string][]byte) error {
	skillPath, err := config.SkillInstallPath("claude", name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(skillPath, 0755); err != nil {
		return err
	}
	for filename, content := range files {
		dest := filepath.Join(skillPath, filename)
		if err := renameio.WriteFile(dest, content, 0644); err != nil {
			return fmt.Errorf("writing skill file %s: %w", filename, err)
		}
	}
	return nil
}

// RemoveSkill removes ~/.claude/skills/<name>/ entirely.
func (a *ClaudeCodeAdapter) RemoveSkill(name string) error {
	skillPath, err := config.SkillInstallPath("claude", name)
	if err != nil {
		return err
	}
	return os.RemoveAll(skillPath)
}

// extractEntry converts a raw JSON decoded interface{} map to a domain.MCPServerEntry.
func extractEntry(val interface{}) domain.MCPServerEntry {
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
