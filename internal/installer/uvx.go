package installer

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/ejyle/agentkit/internal/domain"
)

// UvxInstaller installs MCP servers via "uvx <package>".
// It uses exec.Command arg-array form — never shell string interpolation (T-02-01).
type UvxInstaller struct {
	lookPath lookPathFunc
	run      runFunc
}

// NewUvxInstaller returns a UvxInstaller using the real exec.LookPath and exec.Command.
func NewUvxInstaller() *UvxInstaller {
	u := &UvxInstaller{}
	u.lookPath = exec.LookPath
	u.run = func(name string, args []string) error {
		cmd := exec.Command(name, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("uvx exited non-zero: %w\noutput: %s", err, out.String())
		}
		return nil
	}
	return u
}

// NewUvxInstallerWithLookPath returns a UvxInstaller with an injected LookPath for testing.
func NewUvxInstallerWithLookPath(lp lookPathFunc) *UvxInstaller {
	u := NewUvxInstaller()
	u.lookPath = lp
	return u
}

// NewUvxInstallerWithRunner returns a UvxInstaller with an injected run function for testing.
// It uses a stub lookPath that reports uvx as available so tests do not depend on uvx being installed.
func NewUvxInstallerWithRunner(run runFunc) *UvxInstaller {
	return &UvxInstaller{
		lookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
		run:      run,
	}
}

// IsAvailable returns true if "uvx" is found on PATH.
func (u *UvxInstaller) IsAvailable() bool {
	_, err := u.lookPath("uvx")
	return err == nil
}

// Method returns InstallMethodUvx.
func (u *UvxInstaller) Method() domain.InstallMethod {
	return domain.InstallMethodUvx
}

// Install runs "uvx <spec.Package> [spec.Args...]" using the arg-array form (never shell interpolation).
// Returns ErrUvxNotFound if uvx is not on PATH.
func (u *UvxInstaller) Install(spec domain.InstallSpec) error {
	if _, err := u.lookPath("uvx"); err != nil {
		return ErrUvxNotFound
	}
	args := append([]string{spec.Package}, spec.Args...)
	return u.run("uvx", args)
}
