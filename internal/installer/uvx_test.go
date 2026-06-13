package installer_test

import (
	"errors"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/installer"
)

func TestUvxInstaller_IsAvailable_False(t *testing.T) {
	inst := installer.NewUvxInstallerWithLookPath(func(file string) (string, error) {
		return "", errors.New("not found")
	})
	if inst.IsAvailable() {
		t.Fatal("expected IsAvailable() == false when uvx not on PATH")
	}
}

func TestUvxInstaller_IsAvailable_True(t *testing.T) {
	inst := installer.NewUvxInstallerWithLookPath(func(file string) (string, error) {
		return "/usr/bin/uvx", nil
	})
	if !inst.IsAvailable() {
		t.Fatal("expected IsAvailable() == true when uvx is on PATH")
	}
}

func TestUvxInstaller_ErrNotFound(t *testing.T) {
	inst := installer.NewUvxInstallerWithLookPath(func(file string) (string, error) {
		return "", errors.New("not found")
	})
	spec := domain.InstallSpec{Method: domain.InstallMethodUvx, Package: "mcp-server-fetch"}
	err := inst.Install(spec)
	if !errors.Is(err, installer.ErrUvxNotFound) {
		t.Fatalf("expected ErrUvxNotFound, got: %v", err)
	}
}

func TestUvxInstaller_Install_ArgArray(t *testing.T) {
	var gotName string
	var gotArgs []string
	inst := installer.NewUvxInstallerWithRunner(func(name string, args []string) error {
		gotName = name
		gotArgs = args
		return nil
	})
	spec := domain.InstallSpec{Method: domain.InstallMethodUvx, Package: "mcp-server-fetch"}
	if err := inst.Install(spec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotName != "uvx" {
		t.Errorf("expected command name 'uvx', got %q", gotName)
	}
	if len(gotArgs) == 0 || gotArgs[0] != "mcp-server-fetch" {
		t.Errorf("expected args[0] == 'mcp-server-fetch', got %v", gotArgs)
	}
}

func TestUvxInstaller_Install_WithExtraArgs(t *testing.T) {
	var gotArgs []string
	inst := installer.NewUvxInstallerWithRunner(func(name string, args []string) error {
		gotArgs = args
		return nil
	})
	spec := domain.InstallSpec{
		Method:  domain.InstallMethodUvx,
		Package: "mcp-server-fetch",
		Args:    []string{"--db-path", "/x"},
	}
	if err := inst.Install(spec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"mcp-server-fetch", "--db-path", "/x"}
	if len(gotArgs) != len(expected) {
		t.Fatalf("expected args %v, got %v", expected, gotArgs)
	}
	for i, a := range expected {
		if gotArgs[i] != a {
			t.Errorf("args[%d]: expected %q, got %q", i, a, gotArgs[i])
		}
	}
}
