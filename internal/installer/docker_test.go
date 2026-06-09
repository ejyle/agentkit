package installer_test

import (
	"errors"
	"testing"

	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/installer"
)

func TestDockerInstaller_IsAvailable_False(t *testing.T) {
	inst := installer.NewDockerInstallerWithLookPath(func(file string) (string, error) {
		return "", errors.New("not found")
	})
	if inst.IsAvailable() {
		t.Fatal("expected IsAvailable() == false when docker not on PATH")
	}
}

func TestDockerInstaller_IsAvailable_True(t *testing.T) {
	inst := installer.NewDockerInstallerWithLookPath(func(file string) (string, error) {
		return "/usr/bin/docker", nil
	})
	if !inst.IsAvailable() {
		t.Fatal("expected IsAvailable() == true when docker is on PATH")
	}
}

func TestDockerInstaller_ErrNotFound(t *testing.T) {
	inst := installer.NewDockerInstallerWithLookPath(func(file string) (string, error) {
		return "", errors.New("not found")
	})
	spec := domain.InstallSpec{Method: domain.InstallMethodDocker, Package: "ghcr.io/github/github-mcp-server"}
	err := inst.Install(spec)
	if !errors.Is(err, installer.ErrDockerNotFound) {
		t.Fatalf("expected ErrDockerNotFound, got: %v", err)
	}
}

func TestDockerInstaller_Install_PullsEagerly(t *testing.T) {
	var gotName string
	var gotArgs []string
	inst := installer.NewDockerInstallerWithRunner(func(name string, args []string) error {
		gotName = name
		gotArgs = args
		return nil
	})
	spec := domain.InstallSpec{
		Method:  domain.InstallMethodDocker,
		Package: "ghcr.io/github/github-mcp-server",
	}
	if err := inst.Install(spec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotName != "docker" {
		t.Errorf("expected command name 'docker', got %q", gotName)
	}
	expected := []string{"pull", "ghcr.io/github/github-mcp-server"}
	if len(gotArgs) != len(expected) {
		t.Fatalf("expected args %v, got %v", expected, gotArgs)
	}
	for i, a := range expected {
		if gotArgs[i] != a {
			t.Errorf("args[%d]: expected %q, got %q", i, a, gotArgs[i])
		}
	}
}
