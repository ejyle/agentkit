package cmd

import (
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search the agentkit registry for packages",
	Long: `Search the curated agentkit registry for skills, MCP servers, and agents matching the query.

Example:
  agentkit search playwright`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation comes in a later plan.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
