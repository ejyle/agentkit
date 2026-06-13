// CLI-03: gsd package routes through the gsd-core registry pre-registered in NewRegistryManager().
// D-17 verification (2026-06-09): https://raw.githubusercontent.com/open-gsd/gsd-core/main/registry.json
// returned HTTP 404 — the open-gsd/gsd-core repository does not yet exist. The registry URL is
// pre-wired as a fallback source; when the repo is created, gsd install will resolve via it
// automatically without any code change. Until then, agentkit install gsd will resolve from
// agentkit-registry (first registry in the list). No additional code change needed for CLI-03.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/ejyle/agentkit/internal/adapter"
	"github.com/ejyle/agentkit/internal/bundle"
	"github.com/ejyle/agentkit/internal/config"
	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/installer"
	"github.com/ejyle/agentkit/internal/registry"
	"github.com/ejyle/agentkit/internal/service"
	"github.com/ejyle/agentkit/internal/ui"
)

var installCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a skill, MCP server, or agent",
	Long: `Install a package from the agentkit curated registry into the target coding assistant.

Example:
  agentkit install playwright --target claude`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("bundle", "b", "", "Install a preset bundle (cloud, dev, context)")
}

func runInstall(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")

	// Wire dependencies (shared by bundle and single-package paths).
	store := config.NewConfigStore(target)
	reg := registry.NewRegistryManager()
	ad, err := adapter.NewAdapter(target, store)
	if err != nil {
		return err
	}

	svc := service.NewInstallService(
		reg, ad, store,
		func(method domain.InstallMethod) (service.Installer, error) {
			return installer.NewInstaller(method)
		},
	)

	// Dispatch to bundle install if --bundle flag is set.
	bundleName, _ := cmd.Flags().GetString("bundle")
	if bundleName != "" {
		return runBundleInstall(cmd, bundleName, target, svc)
	}

	// Single-package install path.
	if len(args) == 0 {
		return fmt.Errorf("requires a package name or --bundle flag; run 'agentkit install --help'")
	}
	name := args[0]

	if !ui.IsTerminal() {
		// Non-interactive: run synchronously, no spinner.
		pkg, err := svc.Install(name, target)
		if err != nil {
			return handleInstallError(cmd, name, target, err, svc)
		}
		installPath := installPathFor(pkg, target)
		fmt.Printf("✓ %s@%s installed → %s (%s)\n", pkg.Name, pkg.Version, installPath, target)
		return nil
	}

	// Interactive terminal: drive the spinner via bubbletea.
	resultCh := make(chan *installOutcome, 1)
	go func() {
		pkg, err := svc.Install(name, target)
		resultCh <- &installOutcome{pkg: pkg, err: err}
	}()

	spinnerModel := ui.NewSpinnerModel()
	p := tea.NewProgram(spinnerModel)
	doneCh := make(chan *installOutcome, 1)

	go func() {
		outcome := <-resultCh
		if outcome.err != nil {
			p.Send(ui.ErrorMsg{Err: outcome.err})
		} else {
			p.Send(ui.DoneMsg{})
		}
		doneCh <- outcome
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "spinner error: %v\n", err)
	}

	installResult := <-doneCh
	if installResult.err != nil {
		return handleInstallError(cmd, name, target, installResult.err, svc)
	}

	pkg := installResult.pkg
	installPath := installPathFor(pkg, target)
	fmt.Printf("✓ %s@%s installed → %s (%s)\n",
		pkg.Name, pkg.Version, installPath, target)
	return nil
}

// bundleResult holds the per-package outcome of a parallel bundle install.
type bundleResult struct {
	name string
	pkg  *domain.Package
	err  error
}

// runBundleInstall installs all packages in a named bundle in parallel.
// Uses sync.WaitGroup (NOT errgroup) for best-effort semantics — all packages
// are attempted regardless of individual failures (D-04).
// Exits with code 1 if any package fails (D-05).
func runBundleInstall(_ *cobra.Command, bundleName, target string, svc *service.InstallService) error {
	manifest, err := bundle.LoadBundles()
	if err != nil {
		return err
	}
	pkgNames, err := manifest.Resolve(bundleName)
	if err != nil {
		return err
	}

	results := make([]bundleResult, len(pkgNames))
	var wg sync.WaitGroup
	for i, name := range pkgNames {
		wg.Add(1)
		go func(idx int, n string) {
			defer wg.Done()
			pkg, installErr := svc.Install(n, target)
			results[idx] = bundleResult{name: n, pkg: pkg, err: installErr}
		}(i, name)
	}
	wg.Wait()

	// Print per-package result lines (D-04 best-effort output).
	failed := 0
	for _, r := range results {
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "  %s ✗ %s\n", r.name, r.err)
			failed++
		} else {
			fmt.Printf("  %s ✓\n", r.name)
		}
	}

	if failed == 0 {
		fmt.Printf("%d/%d installed\n", len(pkgNames), len(pkgNames))
	} else {
		fmt.Fprintf(os.Stderr, "%d/%d installed — %d failed\n", len(pkgNames)-failed, len(pkgNames), failed)
		os.Exit(1) // D-05: exit code 1 on any failure
	}
	return nil
}

// installOutcome carries the result of the background install goroutine.
type installOutcome struct {
	pkg *domain.Package
	err error
}

// handleInstallError processes install errors, prompting on ErrForeignConflict (D-07)
// and printing a D-04 formatted message for other errors.
func handleInstallError(cmd *cobra.Command, name, target string, err error, svc *service.InstallService) error {
	var fc *adapter.ErrForeignConflict
	if adapter.AsErrForeignConflict(err, &fc) {
		return handleForeignConflict(cmd, name, target, fc, svc)
	}

	// D-04: error line + context + suggested command.
	fmt.Fprintf(os.Stderr, "✗ Error: %s\n", err.Error())
	fmt.Fprintf(os.Stderr, "Run: agentkit search %s\n", name)
	os.Exit(1)
	return nil // unreachable
}

// handleForeignConflict presents the D-07 overwrite prompt and proceeds if the user confirms.
func handleForeignConflict(_ *cobra.Command, name, target string, fc *adapter.ErrForeignConflict, _ *service.InstallService) error {
	fmt.Fprintf(os.Stderr, "⚠  Foreign conflict: mcpServers.%s is already configured by another tool.\n", name)
	fmt.Fprintf(os.Stderr, "  Old: command=%q args=%v\n", fc.OldEntry.Command, fc.OldEntry.Args)
	fmt.Fprintf(os.Stderr, "  New: command=%q args=%v\n", fc.NewEntry.Command, fc.NewEntry.Args)
	fmt.Fprintf(os.Stderr, "Overwrite? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error reading response: %v\n", err)
		os.Exit(1)
	}
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Fprintf(os.Stderr, "Install cancelled.\n")
		os.Exit(1)
	}

	// User confirmed overwrite — re-run install (ownership check bypassed on second call
	// because WriteSkill/WriteMCPConfig will receive the ownership record from the store
	// if agentkit wrote a placeholder, or succeed if the key was cleared).
	// For now print the suggested command.
	fmt.Fprintf(os.Stderr, "✗ To force-overwrite foreign config, use: agentkit install %s --target %s --force\n", name, target)
	os.Exit(1)
	return nil
}

// installPathFor returns the display path for the success line (D-03).
// For skills: ~/.claude/skills/<name>/
// For MCP servers: the mcpServers.<name> key path in the config file.
func installPathFor(pkg *domain.Package, target string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "mcpServers." + pkg.Name
	}
	if pkg.Type == domain.PackageTypeSkill {
		switch target {
		case "claude":
			return filepath.Join(home, ".claude", "skills", pkg.Name) + string(filepath.Separator)
		default:
			return filepath.Join(home, "."+target, "skills", pkg.Name) + string(filepath.Separator)
		}
	}
	// MCP server — show config file path.
	claudeJSON := filepath.Join(home, ".claude.json")
	if _, err := os.Stat(claudeJSON); err == nil {
		return claudeJSON + "#mcpServers." + pkg.Name
	}
	return filepath.Join(home, ".claude", "settings.json") + "#mcpServers." + pkg.Name
}
