package installer

import (
	"errors"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
)

func TestCustomInstaller_Method(t *testing.T) {
	c := NewCustomInstaller()
	if c.Method() != domain.InstallMethodCustom {
		t.Fatalf("expected %q, got %q", domain.InstallMethodCustom, c.Method())
	}
}

func TestCustomInstaller_IsAvailable(t *testing.T) {
	c := NewCustomInstaller()
	if !c.IsAvailable() {
		t.Fatal("IsAvailable should always return true for CustomInstaller")
	}
}

func TestCustomInstaller_MissingCommand(t *testing.T) {
	c := NewCustomInstaller()
	err := c.Install(domain.InstallSpec{Method: domain.InstallMethodCustom, Package: ""})
	if !errors.Is(err, ErrCustomMissingCommand) {
		t.Fatalf("expected ErrCustomMissingCommand, got: %v", err)
	}
}

func TestCustomInstaller_RunsCommand(t *testing.T) {
	var capturedName string
	var capturedArgs []string

	c := NewCustomInstallerWithRunner(func(name string, args []string) error {
		capturedName = name
		capturedArgs = args
		return nil
	})

	spec := domain.InstallSpec{
		Method:  domain.InstallMethodCustom,
		Package: "my-tool",
		Args:    []string{"--flag", "value"},
	}
	if err := c.Install(spec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedName != "my-tool" {
		t.Errorf("expected command %q, got %q", "my-tool", capturedName)
	}
	if len(capturedArgs) != 2 || capturedArgs[0] != "--flag" || capturedArgs[1] != "value" {
		t.Errorf("unexpected args: %v", capturedArgs)
	}
}

func TestCustomInstaller_CommandError(t *testing.T) {
	sentinel := errors.New("command failed")
	c := NewCustomInstallerWithRunner(func(_ string, _ []string) error {
		return sentinel
	})

	err := c.Install(domain.InstallSpec{Method: domain.InstallMethodCustom, Package: "failing-tool"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
}

func TestNewInstaller_CustomMethod(t *testing.T) {
	inst, err := NewInstaller(domain.InstallMethodCustom)
	if err != nil {
		t.Fatalf("NewInstaller(custom) returned error: %v", err)
	}
	if inst.Method() != domain.InstallMethodCustom {
		t.Fatalf("expected custom method, got %q", inst.Method())
	}
}
