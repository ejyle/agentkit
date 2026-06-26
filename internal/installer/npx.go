package installer

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/ejyle/agentkit/internal/domain"
)

// lookPathFunc is the function used to find a binary on PATH.
// Settable in tests to inject mock behaviour without shell invocation.
type lookPathFunc func(file string) (string, error)

// runFunc is called by Install to execute a command. Settable in tests.
type runFunc func(name string, args []string) error

// NpxInstaller installs MCP servers via "npx -y <package>".
// It uses exec.Command arg-array form — never shell string interpolation (T-03-01).
type NpxInstaller struct {
	lookPath lookPathFunc
	run      runFunc
}

// NewNpxInstaller returns an NpxInstaller using the real exec.LookPath and exec.Command.
func NewNpxInstaller() *NpxInstaller {
	n := &NpxInstaller{}
	n.lookPath = exec.LookPath
	n.run = func(name string, args []string) error {
		cmd := exec.Command(name, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npx exited non-zero: %w\noutput: %s", err, out.String())
		}
		return nil
	}
	return n
}

// NewNpxInstallerWithLookPath returns an NpxInstaller with an injected LookPath for testing.
func NewNpxInstallerWithLookPath(lp lookPathFunc) *NpxInstaller {
	n := NewNpxInstaller()
	n.lookPath = lp
	return n
}

// NewNpxInstallerWithRunner returns an NpxInstaller with an injected run function for testing.
func NewNpxInstallerWithRunner(run runFunc) *NpxInstaller {
	n := &NpxInstaller{
		lookPath: exec.LookPath,
		run:      run,
	}
	return n
}

// IsAvailable returns true if "node" is found on PATH.
func (n *NpxInstaller) IsAvailable() bool {
	_, err := n.lookPath("node")
	return err == nil
}

// Method returns InstallMethodNpx.
func (n *NpxInstaller) Method() domain.InstallMethod {
	return domain.InstallMethodNpx
}

// Install runs "npx -y <spec.Package> [spec.Args...]" using the arg-array form (never shell
// interpolation). spec.Args are install-time args (e.g. "--version" for packages that error
// without a subcommand). Returns ErrNodeNotFound if node is not on PATH.
func (n *NpxInstaller) Install(spec domain.InstallSpec) error {
	if _, err := n.lookPath("node"); err != nil {
		return ErrNodeNotFound
	}
	args := append([]string{"-y", spec.Package}, spec.Args...)
	return n.run("npx", args)
}

// IsErrNodeNotFound reports whether err is (or wraps) ErrNodeNotFound.
func IsErrNodeNotFound(err error) bool {
	return err != nil && err.Error() == ErrNodeNotFound.Error()
}
