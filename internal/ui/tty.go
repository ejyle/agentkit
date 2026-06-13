package ui

import "os"

// IsTerminal returns true when stdout is connected to a real interactive terminal.
// Uses only stdlib — avoids adding a golang.org/x/term dependency.
func IsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
