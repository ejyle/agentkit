package installer

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/ejyle/agentkit/internal/domain"
)

// DockerInstaller installs MCP servers via "docker pull <image>".
// It eagerly pulls the image on install (D-09).
// It uses exec.Command arg-array form — never shell string interpolation (T-02-01).
type DockerInstaller struct {
	lookPath lookPathFunc
	run      runFunc
}

// NewDockerInstaller returns a DockerInstaller using the real exec.LookPath and exec.Command.
func NewDockerInstaller() *DockerInstaller {
	d := &DockerInstaller{}
	d.lookPath = exec.LookPath
	d.run = func(name string, args []string) error {
		cmd := exec.Command(name, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("docker exited non-zero: %w\noutput: %s", err, out.String())
		}
		return nil
	}
	return d
}

// NewDockerInstallerWithLookPath returns a DockerInstaller with an injected LookPath for testing.
func NewDockerInstallerWithLookPath(lp lookPathFunc) *DockerInstaller {
	d := NewDockerInstaller()
	d.lookPath = lp
	return d
}

// NewDockerInstallerWithRunner returns a DockerInstaller with an injected run function for testing.
// It uses a stub lookPath that reports docker as available so tests do not depend on docker being installed.
func NewDockerInstallerWithRunner(run runFunc) *DockerInstaller {
	return &DockerInstaller{
		lookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
		run:      run,
	}
}

// IsAvailable returns true if "docker" is found on PATH.
func (d *DockerInstaller) IsAvailable() bool {
	_, err := d.lookPath("docker")
	return err == nil
}

// Method returns InstallMethodDocker.
func (d *DockerInstaller) Method() domain.InstallMethod {
	return domain.InstallMethodDocker
}

// Install runs "docker pull <spec.Package>" using the arg-array form (never shell interpolation).
// Returns ErrDockerNotFound if docker is not on PATH.
// The image is eagerly pulled on install (D-09).
func (d *DockerInstaller) Install(spec domain.InstallSpec) error {
	if _, err := d.lookPath("docker"); err != nil {
		return ErrDockerNotFound
	}
	return d.run("docker", []string{"pull", spec.Package})
}
