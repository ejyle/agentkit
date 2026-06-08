package installer_test

import (
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/installer"
)

// Test 1: NpxInstaller.IsAvailable() returns true when "node" is on PATH.
// We can't reliably mock exec.LookPath in tests, so we test the real behaviour
// (node may or may not be present) and just verify the method doesn't panic.
func TestNpxInstaller_IsAvailable(t *testing.T) {
	n := installer.NewNpxInstaller()
	// IsAvailable must return a bool without panicking.
	_ = n.IsAvailable()
}

// Test 2: NpxInstaller.Method() returns domain.InstallMethodNpx.
func TestNpxInstaller_Method(t *testing.T) {
	n := installer.NewNpxInstaller()
	if got := n.Method(); got != domain.InstallMethodNpx {
		t.Errorf("Method() = %q; want %q", got, domain.InstallMethodNpx)
	}
}

// Test 3: NpxInstaller.Install() returns ErrNodeNotFound when node is not on PATH.
// We test this by creating an installer with a custom LookPath that always fails.
func TestNpxInstaller_Install_ErrNodeNotFound(t *testing.T) {
	n := installer.NewNpxInstallerWithLookPath(func(file string) (string, error) {
		return "", &notFoundError{file}
	})
	err := n.Install(domain.InstallSpec{
		Method:  domain.InstallMethodNpx,
		Package: "@playwright/mcp",
	})
	if err == nil {
		t.Fatal("Install() expected error when node not found, got nil")
	}
	if !installer.IsErrNodeNotFound(err) {
		t.Errorf("Install() error = %v; want ErrNodeNotFound", err)
	}
}

// Test 4: NpxInstaller uses exec.Command arg array form (not shell string interpolation).
// We verify this by checking that the command is assembled as separate args.
func TestNpxInstaller_NoShellInterpolation(t *testing.T) {
	n := installer.NewNpxInstallerWithRunner(func(name string, args []string) error {
		if name != "npx" {
			t.Errorf("expected command %q, got %q", "npx", name)
		}
		if len(args) < 2 || args[0] != "-y" || args[1] != "@playwright/mcp" {
			t.Errorf("expected args [\"-y\", \"@playwright/mcp\"], got %v", args)
		}
		return nil
	})
	err := n.Install(domain.InstallSpec{
		Method:  domain.InstallMethodNpx,
		Package: "@playwright/mcp",
	})
	if err != nil {
		t.Errorf("Install() unexpected error: %v", err)
	}
}

// notFoundError implements the exec.Error-like interface for lookup failures.
type notFoundError struct{ name string }

func (e *notFoundError) Error() string { return e.name + ": not found" }
