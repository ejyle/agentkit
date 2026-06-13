// Package ui provides terminal rendering utilities for agentkit output.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ejyle/agentkit/internal/domain"
	"github.com/ejyle/agentkit/internal/registry"
)

// Column widths for the installed table (D-05: go-list style aligned columns).
const (
	colPackage  = 20
	colVersion  = 10
	colType     = 8
	colTarget   = 12
	colRegistry = 20
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	rowStyle    = lipgloss.NewStyle()
	altRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	emptyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	nameStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	regStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	typeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
)

// padRight pads s to exactly width runes using spaces (truncates if too long).
func padRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// registryNameFromURL extracts a human-readable registry name from a source URL.
// It looks for the last path component before /registry.json.
// Falls back to the full URL if the pattern does not match.
func registryNameFromURL(sourceURL string) string {
	// Pattern: https://raw.githubusercontent.com/{owner}/{repo}/{ref}/registry.json
	// We want {repo}.
	parts := strings.Split(sourceURL, "/")
	for i, p := range parts {
		if p == "registry.json" && i >= 2 {
			// The repo name is two levels back: owner/repo/ref/registry.json
			// parts[i-2] is the repo name
			return parts[i-2]
		}
	}
	// Fallback: return the full URL
	return sourceURL
}

// RenderInstalledTable renders a D-05 aligned table of installed records for the given target.
// Returns a styled "No packages installed" message when records is empty.
func RenderInstalledTable(records []domain.InstalledRecord, target string) string {
	if len(records) == 0 {
		return emptyStyle.Render(fmt.Sprintf("No packages installed for target: %s", target))
	}

	var sb strings.Builder

	// Header row.
	header := headerStyle.Render(
		padRight("PACKAGE", colPackage) +
			padRight("VERSION", colVersion) +
			padRight("TYPE", colType) +
			padRight("TARGET", colTarget) +
			padRight("REGISTRY", colRegistry),
	)
	sb.WriteString(header)
	sb.WriteString("\n")

	// Data rows with alternating style.
	for i, rec := range records {
		regName := registryNameFromURL(rec.SourceURL)
		row := padRight(rec.Name, colPackage) +
			padRight(rec.Version, colVersion) +
			padRight(string(rec.Type), colType) +
			padRight(target, colTarget) +
			padRight(regName, colRegistry)

		if i%2 == 0 {
			sb.WriteString(rowStyle.Render(row))
		} else {
			sb.WriteString(altRowStyle.Render(row))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderSearchResults renders D-06 search results as an aligned list.
// Shows at most 20 results. Returns "No results found." for an empty slice.
func RenderSearchResults(results []registry.SearchResult) string {
	if len(results) == 0 {
		return emptyStyle.Render("No results found.")
	}

	// Cap at 20.
	shown := results
	if len(shown) > 20 {
		shown = shown[:20]
	}

	var sb strings.Builder
	for _, r := range shown {
		line := fmt.Sprintf("  %s  %s  [%s]  %s",
			nameStyle.Render(padRight(r.Package.Name, 20)),
			typeStyle.Render(padRight(string(r.Package.Type), 7)),
			regStyle.Render(r.RegistryName),
			r.Package.Description,
		)
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}
