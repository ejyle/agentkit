package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

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
  agentkit uninstall playwright --target claude
  agentkit uninstall --all --target claude`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().Bool("all", false, "Uninstall all installed packages")
	uninstallCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt (for use with --all)")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")

	if allFlag, _ := cmd.Flags().GetBool("all"); allFlag {
		return runUninstallAll(cmd, target)
	}

	if len(args) == 0 {
		return fmt.Errorf("requires a package name or --all; run 'agentkit uninstall --help'")
	}
	name := args[0]

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

// runUninstallAll removes all packages installed for the given target.
// Requires confirmation unless --yes is passed.
func runUninstallAll(cmd *cobra.Command, target string) error {
	store := config.NewConfigStore(target)
	ad, err := adapter.NewAdapter(target, store)
	if err != nil {
		return err
	}
	svc := service.NewUninstallService(ad, store)

	records, err := store.ListInstalled()
	if err != nil {
		return fmt.Errorf("reading installed packages: %w", err)
	}
	if len(records) == 0 {
		fmt.Printf("No packages installed for target %s.\n", target)
		return nil
	}

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Printf("This will uninstall %d package(s) from %s:\n", len(records), target)
		for _, r := range records {
			fmt.Printf("  - %s@%s\n", r.Name, r.Version)
		}
		fmt.Print("Continue? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error reading response: %v\n", err)
			os.Exit(1)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Uninstall cancelled.")
			return nil
		}
	}

	failed := 0
	for _, r := range records {
		if err := svc.Uninstall(r.Name); err != nil {
			fmt.Fprintf(os.Stderr, "  %s ✗ %s\n", r.Name, err)
			failed++
		} else {
			fmt.Printf("  %s ✓\n", r.Name)
		}
	}

	if failed == 0 {
		fmt.Printf("%d/%d uninstalled\n", len(records), len(records))
	} else {
		fmt.Fprintf(os.Stderr, "%d/%d uninstalled — %d failed\n", len(records)-failed, len(records), failed)
		os.Exit(1)
	}
	return nil
}
