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

	if !ui.IsTerminal() {
		// Non-interactive: run synchronously, no spinner.
		results, err := searchSvc.Search(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error: search failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "Run: agentkit list\n")
			os.Exit(1)
		}
		fmt.Print(ui.RenderSearchResults(results))
		return nil
	}

	// Interactive terminal: drive the spinner via bubbletea.
	resultCh := make(chan *searchOutcome, 1)
	go func() {
		results, err := searchSvc.Search(query)
		resultCh <- &searchOutcome{results: results, err: err}
	}()

	spinnerModel := ui.NewSpinnerModel()
	p := tea.NewProgram(spinnerModel)
	doneCh := make(chan *searchOutcome, 1)

	go func() {
		out := <-resultCh
		if out.err != nil {
			p.Send(ui.ErrorMsg{Err: out.err})
		} else {
			p.Send(ui.DoneMsg{})
		}
		doneCh <- out
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "spinner error: %v\n", err)
	}

	outcome := <-doneCh
	if outcome.err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error: search failed: %v\n", outcome.err)
		fmt.Fprintf(os.Stderr, "Run: agentkit list\n")
		os.Exit(1)
	}

	fmt.Print(ui.RenderSearchResults(outcome.results))
	return nil
}
