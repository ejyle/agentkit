package installer

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"github.com/ejyle/agentkit/internal/domain"
)

// ErrCustomMissingCommand is returned when a custom install spec has no command in Package.
var ErrCustomMissingCommand = errors.New("custom install spec must set 'package' to the command to run")

// CustomInstaller executes arbitrary commands specified in the package manifest.
// It uses exec.Command arg-array form — never shell string interpolation (T-03-01).
// The spec.Package field is the executable; spec.Args are its arguments.
type CustomInstaller struct {
	run runFunc
}

// NewCustomInstaller returns a CustomInstaller using the real exec.Command.
func NewCustomInstaller() *CustomInstaller {
	c := &CustomInstaller{}
	c.run = func(name string, args []string) error {
		cmd := exec.Command(name, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("custom install command exited non-zero: %w\noutput: %s", err, out.String())
		}
		return nil
	}
	return c
}

// NewCustomInstallerWithRunner returns a CustomInstaller with an injected run function for testing.
func NewCustomInstallerWithRunner(run runFunc) *CustomInstaller {
	return &CustomInstaller{run: run}
}

// IsAvailable always returns true — custom installers use whatever command is in the spec.
func (c *CustomInstaller) IsAvailable() bool { return true }

// Method returns InstallMethodCustom.
func (c *CustomInstaller) Method() domain.InstallMethod { return domain.InstallMethodCustom }

// Install runs spec.Package with spec.Args using exec.Command arg-array form.
// spec.Package must be non-empty (it is the executable path or name).
func (c *CustomInstaller) Install(spec domain.InstallSpec) error {
	if spec.Package == "" {
		return ErrCustomMissingCommand
	}
	return c.run(spec.Package, spec.Args)
}
