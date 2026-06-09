package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agentkit",
	Short: "AI agent skill and MCP server manager",
	Long: `agentkit installs, updates, and manages AI agent skills, MCP servers, and
coding agents across all major AI coding assistants.`,
}

var validTargets = []string{"claude", "copilot-cli", "copilot-vscode", "codex", "gemini", "opencode", "pi"}

// Execute runs the root cobra command and returns any error.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP(
		"target", "t", "claude",
		"Target coding assistant (claude|copilot-cli|copilot-vscode|codex|gemini|opencode|pi)",
	)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		target, err := cmd.Flags().GetString("target")
		if err != nil {
			return err
		}
		for _, v := range validTargets {
			if target == v {
				return nil
			}
		}
		return fmt.Errorf(
			"invalid target %q: must be one of claude, copilot-cli, copilot-vscode, codex, gemini, opencode, pi",
			target,
		)
	}
}
