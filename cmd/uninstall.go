package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ejyle/agentkit/internal/adapter"
	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/service"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [name]",
	Short: "Uninstall a package from the target assistant",
	Long: `Remove a previously installed package from the target coding assistant and update its config.

Example:
  agentkit uninstall playwright --target claude`,
	Args: cobra.ExactArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	name := args[0]
	target, _ := cmd.Flags().GetString("target")

	// Wire dependencies.
	store := config.NewConfigStore(target)
	ad := adapter.NewClaudeCodeAdapter(store)
	svc := service.NewUninstallService(ad, store)

	if err := svc.Uninstall(name); err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			// D-04 error format with suggested command.
			fmt.Fprintf(os.Stderr, "✗ Error: %s is not installed for target %s\n", name, target)
			fmt.Fprintf(os.Stderr, "Run: agentkit list --target %s\n", target)
			os.Exit(1)
		}
		// D-04 generic error format.
		fmt.Fprintf(os.Stderr, "✗ Error: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run: agentkit list --target %s\n", target)
		os.Exit(1)
	}

	fmt.Printf("✓ %s uninstalled (%s)\n", name, target)
	return nil
}
