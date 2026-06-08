package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/ejyle/agentkit/internal/registry"
	"github.com/ejyle/agentkit/internal/service"
	"github.com/ejyle/agentkit/internal/ui"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the agentkit registry for packages",
	Long: `Search the curated agentkit registry for skills, MCP servers, and agents matching the query.

Example:
  agentkit search playwright`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

// searchOutcome carries results or error from the background search goroutine.
type searchOutcome struct {
	results []registry.SearchResult
	err     error
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	reg := registry.NewRegistryManager()
	searchSvc := service.NewSearchService(reg)

	// Run search in background; drive the spinner UI from the main goroutine.
	var outcome *searchOutcome
	resultCh := make(chan *searchOutcome, 1)

	go func() {
		results, err := searchSvc.Search(query)
		resultCh <- &searchOutcome{results: results, err: err}
	}()

	spinnerModel := ui.NewSpinnerModel()
	p := tea.NewProgram(spinnerModel)

	// Forward search result to the bubbletea program.
	go func() {
		out := <-resultCh
		if out.err != nil {
			p.Send(ui.ErrorMsg{Err: out.err})
		} else {
			p.Send(ui.DoneMsg{})
		}
		outcome = out
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "spinner error: %v\n", err)
	}

	// Ensure we have the result (should already be set, but guard for safety).
	if outcome == nil {
		outcome = <-resultCh
	}

	if outcome.err != nil {
		// D-04 format.
		fmt.Fprintf(os.Stderr, "✗ Error: search failed: %v\n", outcome.err)
		fmt.Fprintf(os.Stderr, "Run: agentkit list\n")
		os.Exit(1)
	}

	fmt.Print(ui.RenderSearchResults(outcome.results))
	return nil
}
