package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// CheckResult holds the outcome of a single doctor check.
type CheckResult struct {
	Label   string
	Status  string // "pass", "warn", "fail"
	Message string
	Hint    string
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check your agentkit environment",
	Long:  `Run environment checks to verify agentkit is correctly installed and configured.`,
	RunE:  runDoctor,
}

func init() {
	// Bypass the root command's --target validation: doctor does not take a target flag.
	doctorCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	var results []CheckResult

	results = append(results, checkBinaryInPath())
	results = append(results, checkConfigDirWritable())
	results = append(results, checkRegistryReachable())
	results = append(results, checkAssistantDirs()...)
	results = append(results, checkRuntimeDeps()...)

	anyFail := false
	for _, r := range results {
		printCheckResult(r)
		if r.Status == "fail" {
			anyFail = true
		}
	}

	if anyFail {
		return fmt.Errorf("one or more checks failed")
	}
	return nil
}

// checkBinaryInPath checks whether agentkit is on the PATH.
func checkBinaryInPath() CheckResult {
	path, err := exec.LookPath("agentkit")
	if err != nil {
		return CheckResult{
			Label:  "agentkit in PATH",
			Status: "fail",
			Hint:   "Add agentkit to your PATH",
		}
	}
	return CheckResult{
		Label:   "agentkit in PATH",
		Status:  "pass",
		Message: path,
	}
}

// checkConfigDirWritable checks that ~/.agentkit/ exists and is writable.
func checkConfigDirWritable() CheckResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return CheckResult{
			Label:   "~/.agentkit/ writable",
			Status:  "fail",
			Message: err.Error(),
			Hint:    "Run: mkdir -p ~/.agentkit",
		}
	}

	dir := filepath.Join(home, ".agentkit")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return CheckResult{
			Label:   "~/.agentkit/ writable",
			Status:  "fail",
			Message: err.Error(),
			Hint:    "Run: mkdir -p ~/.agentkit",
		}
	}

	testFile := filepath.Join(dir, ".write-test")
	if err := os.WriteFile(testFile, []byte{}, 0600); err != nil {
		return CheckResult{
			Label:   "~/.agentkit/ writable",
			Status:  "fail",
			Message: err.Error(),
			Hint:    "Check permissions on ~/.agentkit/",
		}
	}
	_ = os.Remove(testFile)

	return CheckResult{
		Label:  "~/.agentkit/ writable",
		Status: "pass",
	}
}

// checkRegistryReachable checks network connectivity to the agentkit registry.
func checkRegistryReachable() CheckResult {
	const registryURL = "https://raw.githubusercontent.com/ejyle/agentkit-registry/main/registry.json"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, registryURL, nil)
	if err != nil {
		return CheckResult{
			Label:   "registry reachable (agentkit-registry)",
			Status:  "fail",
			Message: err.Error(),
			Hint:    "Check network connectivity",
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{
			Label:   "registry reachable (agentkit-registry)",
			Status:  "fail",
			Message: err.Error(),
			Hint:    "Check network connectivity",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return CheckResult{
			Label:   "registry reachable (agentkit-registry)",
			Status:  "fail",
			Message: fmt.Sprintf("HTTP %d", resp.StatusCode),
			Hint:    "Check network connectivity",
		}
	}

	return CheckResult{
		Label:  "registry reachable (agentkit-registry)",
		Status: "pass",
	}
}

// assistantDir maps assistant name to home-relative directory path.
type assistantDir struct {
	name  string
	label string
	rel   []string // path segments relative to home (or config dir for opencode)
}

// checkAssistantDirs checks for the presence of each supported assistant's config directory.
func checkAssistantDirs() []CheckResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return []CheckResult{{
			Label:   "assistant dirs",
			Status:  "fail",
			Message: err.Error(),
		}}
	}

	// For ~/.config/opencode/ we need to derive the config base
	// but we use ~/.config/ explicitly as opencode always uses XDG-style config.
	entries := []struct {
		dirLabel string // e.g. "~/.claude/"
		path     string
		descName string
	}{
		{
			dirLabel: "~/.claude/",
			path:     filepath.Join(home, ".claude"),
			descName: "Claude Code",
		},
		{
			dirLabel: "~/.gemini/",
			path:     filepath.Join(home, ".gemini"),
			descName: "Gemini CLI",
		},
		{
			dirLabel: "~/.copilot/",
			path:     filepath.Join(home, ".copilot"),
			descName: "Copilot CLI",
		},
		{
			dirLabel: "~/.codex/",
			path:     filepath.Join(home, ".codex"),
			descName: "Codex",
		},
		{
			dirLabel: "~/.cursor/",
			path:     filepath.Join(home, ".cursor"),
			descName: "Cursor",
		},
		{
			dirLabel: "~/.config/opencode/",
			path:     filepath.Join(home, ".config", "opencode"),
			descName: "OpenCode",
		},
	}

	results := make([]CheckResult, 0, len(entries))
	for _, e := range entries {
		if _, err := os.Stat(e.path); err == nil {
			results = append(results, CheckResult{
				Label:  e.dirLabel + " exists",
				Status: "pass",
			})
		} else {
			results = append(results, CheckResult{
				Label:   e.dirLabel,
				Status:  "warn",
				Message: fmt.Sprintf("not found — %s not installed", e.descName),
			})
		}
	}
	return results
}

// checkRuntimeDeps checks for optional runtime dependencies (node, docker, uvx).
func checkRuntimeDeps() []CheckResult {
	type dep struct {
		binary    string
		passLabel string
		failLabel string
		failMsg   string
		hint      string
	}

	deps := []dep{
		{
			binary:    "node",
			passLabel: "node available",
			failLabel: "node",
			failMsg:   "not found — npx-based MCPs won't install",
			hint:      "Install: https://nodejs.org",
		},
		{
			binary:    "docker",
			passLabel: "docker available",
			failLabel: "docker",
			failMsg:   "not found — Docker-based MCPs won't install",
			hint:      "Install: https://docs.docker.com/get-docker/",
		},
		{
			binary:    "uvx",
			passLabel: "uvx available",
			failLabel: "uvx",
			failMsg:   "not found — Python MCPs won't install",
			hint:      "Install: pip install uv",
		},
	}

	results := make([]CheckResult, 0, len(deps))
	for _, d := range deps {
		if _, err := exec.LookPath(d.binary); err == nil {
			results = append(results, CheckResult{
				Label:  d.passLabel,
				Status: "pass",
			})
		} else {
			results = append(results, CheckResult{
				Label:   d.failLabel,
				Status:  "warn",
				Message: d.failMsg,
				Hint:    d.hint,
			})
		}
	}
	return results
}

// printCheckResult prints a single check result to stdout (or stderr for failures).
func printCheckResult(r CheckResult) {
	switch r.Status {
	case "pass":
		if r.Message != "" {
			fmt.Printf("✓ %s (%s)\n", r.Label, r.Message)
		} else {
			fmt.Printf("✓ %s\n", r.Label)
		}
	case "warn":
		if r.Message != "" {
			fmt.Printf("⚠ %s — %s\n", r.Label, r.Message)
		} else {
			fmt.Printf("⚠ %s\n", r.Label)
		}
		if r.Hint != "" {
			fmt.Printf("   → %s\n", r.Hint)
		}
	case "fail":
		if r.Message != "" {
			fmt.Fprintf(os.Stderr, "✗ %s — %s\n", r.Label, r.Message)
		} else {
			fmt.Fprintf(os.Stderr, "✗ %s\n", r.Label)
		}
		if r.Hint != "" {
			fmt.Fprintf(os.Stderr, "   → %s\n", r.Hint)
		}
	}
}
