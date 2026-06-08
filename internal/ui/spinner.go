// Package ui provides bubbletea-based terminal UI models for agentkit.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerPhase describes the current phase of an install operation.
type SpinnerPhase string

const (
	// PhaseFetchRegistry is shown while fetching the registry manifest.
	PhaseFetchRegistry SpinnerPhase = "Fetching registry..."
	// PhaseResolving is shown while resolving the package in the manifest.
	PhaseResolving SpinnerPhase = "Resolving package..."
	// PhaseInstalling is shown while running the install adapter (npx, binary download).
	PhaseInstalling SpinnerPhase = "Running install adapter..."
)

// PhaseUpdateMsg is sent to SpinnerModel to advance the displayed phase.
type PhaseUpdateMsg struct{ Phase SpinnerPhase }

// DoneMsg is sent to SpinnerModel when the install operation completes successfully.
type DoneMsg struct{}

// ErrorMsg is sent to SpinnerModel when the install operation fails.
type ErrorMsg struct{ Err error }

// SpinnerModel is a bubbletea.Model that displays a spinner alongside the current phase.
// The caller advances phases by sending PhaseUpdateMsg via tea.Cmd, signals completion
// with DoneMsg, and signals failure with ErrorMsg.
type SpinnerModel struct {
	spinner spinner.Model
	phase   SpinnerPhase
	done    bool
	err     error
}

// NewSpinnerModel returns a SpinnerModel starting at PhaseFetchRegistry.
func NewSpinnerModel() SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return SpinnerModel{
		spinner: s,
		phase:   PhaseFetchRegistry,
	}
}

// Init starts the spinner tick.
func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages: spinner tick, phase advance, done, error.
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case PhaseUpdateMsg:
		m.phase = msg.Phase
		return m, nil
	case DoneMsg:
		m.done = true
		return m, tea.Quit
	case ErrorMsg:
		m.err = msg.Err
		return m, tea.Quit
	}
	return m, nil
}

// View renders the spinner + phase text. Returns an empty string when done
// (the caller prints the success/error line after the program exits).
func (m SpinnerModel) View() string {
	if m.done || m.err != nil {
		return ""
	}
	return m.spinner.View() + " " + string(m.phase)
}

// Done reports whether the model received a DoneMsg.
func (m SpinnerModel) Done() bool { return m.done }

// Err returns any error received via ErrorMsg.
func (m SpinnerModel) Err() error { return m.err }
