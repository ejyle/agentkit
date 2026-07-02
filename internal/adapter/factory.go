package adapter

import (
	"fmt"

	"github.com/ejyle/agentkit/internal/config"
)

// NewAdapter returns the AssistantAdapter for the given target assistant name.
//
// Supported targets: claude, copilot-cli, copilot-vscode, gemini, pi, codex, opencode, cursor.
//
// Note: cmd/root.go validates the --target flag before this function is called;
// the default case is a defense-in-depth guard for programmatic callers.
func NewAdapter(target string, store *config.ConfigStore) (AssistantAdapter, error) {
	switch target {
	case "claude":
		return NewClaudeCodeAdapter(store), nil
	case "copilot-cli":
		return NewCopilotCLIAdapter(store), nil
	case "copilot-vscode":
		return NewCopilotVSCodeAdapter(store), nil
	case "gemini":
		return NewGeminiAdapter(store), nil
	case "pi":
		return NewPiAdapter(store), nil
	case "codex":
		return NewCodexAdapter(store), nil
	case "opencode":
		return NewOpenCodeAdapter(store), nil
	case "cursor":
		return NewCursorAdapter(store), nil
	default:
		return nil, fmt.Errorf("unsupported target assistant: %q (valid targets: claude, copilot-cli, copilot-vscode, gemini, pi, codex, opencode, cursor)", target)
	}
}
