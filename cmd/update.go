package cmd

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update installed packages to the latest registry versions",
	Long: `Update one or all packages installed by agentkit for the target coding assistant.
Omit name to update all packages.

Example:
  agentkit update playwright --target claude
  agentkit update --target claude`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation comes in a later plan.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
