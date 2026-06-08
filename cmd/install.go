package cmd

import (
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a skill, MCP server, or agent",
	Long: `Install a package from the agentkit curated registry into the target coding assistant.

Example:
  agentkit install playwright --target claude`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation comes in a later plan.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
