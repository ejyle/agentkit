package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/ui"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages for the target assistant",
	Long: `Display a table of all packages installed by agentkit for the target coding assistant.

Example:
  agentkit list --target claude`,
	RunE: runList,
}

func init() {
	listCmd.Flags().String("target", "claude", "Target coding assistant (claude, copilot, codex, gemini, opencode)")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, _ []string) error {
	target, _ := cmd.Flags().GetString("target")

	store := config.NewConfigStore(target)
	records, err := store.ListInstalled()
	if err != nil {
		// D-04 format error to stderr.
		fmt.Fprintf(os.Stderr, "✗ Error: could not read installed packages: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run: agentkit install <name> --target %s\n", target)
		os.Exit(1)
	}

	// Empty slice → RenderInstalledTable returns the helpful message (not an error).
	fmt.Print(ui.RenderInstalledTable(records, target))
	return nil
}
