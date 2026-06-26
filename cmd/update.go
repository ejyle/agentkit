package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ejyle/agentkit/internal/adapter"
	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/installer"
	"github.com/ejyle/agentkit/internal/registry"
	"github.com/ejyle/agentkit/internal/service"
)

var updateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update installed packages to the latest registry versions",
	Long: `Update one or all packages installed by agentkit for the target coding assistant.
Omit name to update all packages.

Example:
  agentkit update playwright --target claude
  agentkit update --target claude`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

// installServiceAdapter wraps *service.InstallService to satisfy the service.updateInstaller interface.
type installServiceAdapter struct {
	svc *service.InstallService
}

func (a *installServiceAdapter) Install(name, target string) (*domain.Package, error) {
	return a.svc.Install(name, target)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")

	// Wire dependencies.
	store := config.NewConfigStore(target)
	reg := registry.NewRegistryManager()
	ad := adapter.NewClaudeCodeAdapter(store)

	installSvc := service.NewInstallService(
		reg, ad, store,
		func(method domain.InstallMethod) (service.Installer, error) {
			return installer.NewInstaller(method)
		},
	)

	updateSvc := service.NewUpdateService(reg, store, &installServiceAdapter{svc: installSvc})

	if len(args) == 1 {
		name := args[0]
		msg, err := updateSvc.Update(name, target)
		if err != nil {
			if errors.Is(err, service.ErrNotInstalled) {
				fmt.Fprintf(os.Stderr, "✗ Error: %s is not installed for target %s\n", name, target)
				fmt.Fprintf(os.Stderr, "Run: agentkit list --target %s\n", target)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "✗ Error: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "Run: agentkit list --target %s\n", target)
			os.Exit(1)
		}
		// D-08 upgrade notice format.
		if msg == "already up to date" {
			fmt.Printf("✓ %s: already up to date\n", name)
		} else {
			// msg is "updated <name>: <old> → <new>"
			fmt.Printf("⚠ %s\n", msg)
		}
		return nil
	}

	// UpdateAll: update all installed packages.
	msgs, err := updateSvc.UpdateAll(target)
	for _, m := range msgs {
		if m == "already up to date" {
			fmt.Printf("✓ already up to date\n")
		} else {
			fmt.Printf("⚠ %s\n", m)
		}
	}
	if err != nil {
		// D-04 format.
		fmt.Fprintf(os.Stderr, "✗ Error: some packages failed to update: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run: agentkit list --target %s\n", target)
		os.Exit(1)
	}
	if len(msgs) == 0 {
		fmt.Printf("No packages installed for target %s.\n", target)
	}
	return nil
}
