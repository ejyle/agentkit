package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages for the target assistant",
	Long: `Display a table of all packages installed by agentkit for the target coding assistant.

Example:
  agentkit list --target claude`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation comes in a later plan.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
