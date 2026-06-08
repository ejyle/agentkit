package cmd

import (
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <name>",
	Short: "Uninstall a package from the target assistant",
	Long: `Remove a previously installed package from the target coding assistant and update its config.

Example:
  agentkit uninstall playwright --target claude`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation comes in a later plan.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
