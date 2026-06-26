---
phase: 01-foundation
plan: "04"
subsystem: list-search-ui
tags: [cli, ui, lipgloss, bubbletea, search]
dependency_graph:
  requires: [01-02, 01-03]
  provides: [list-command, search-command, table-renderer, search-service]
  affects: [cmd/list.go, cmd/search.go, internal/ui/table.go, internal/service/search.go]
tech_stack:
  added: []
  patterns: [lipgloss-table-rendering, searchregistry-interface, tdd-red-green]
key_files:
  created:
    - internal/service/search.go
    - internal/service/search_test.go
    - internal/ui/table.go
    - internal/ui/table_test.go
  modified:
    - cmd/list.go
    - cmd/search.go
decisions:
  - SearchService uses local searchRegistry interface for injectable mocks (not *RegistryManager directly)
  - RenderInstalledTable uses fixed column widths (PACKAGE=20, VERSION=10, TYPE=8, TARGET=12, REGISTRY=20) with lipgloss alternating row styles
  - registryNameFromURL extracts repo name from raw.githubusercontent.com URL pattern; falls back to full URL
metrics:
  duration: "~15 minutes"
  completed: "2026-06-08"
---

# Phase 01 Plan 04: List and Search Commands Summary

Implemented `agentkit list` (D-05 aligned table) and `agentkit search` (D-06 spinner + ranked results) reusing ConfigStore and RegistryManager from 01-02.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 (RED) | Failing tests for SearchService and table renderer | test commit | internal/service/search_test.go, internal/ui/table_test.go |
| 1 (GREEN) | SearchService + lipgloss table renderer | feat commit | internal/service/search.go, internal/ui/table.go |
| 2 | Wire list and search commands | feat commit | cmd/list.go, cmd/search.go |

## What Was Built

**SearchService** (`internal/service/search.go`) — Thin service wrapping a `searchRegistry` interface. Delegates `Search(query)` to the backing registry. Interface-based so tests use in-memory mocks without network.

**Table renderer** (`internal/ui/table.go`) — Two functions:
- `RenderInstalledTable(records, target)`: D-05 columns PACKAGE/VERSION/TYPE/TARGET/REGISTRY with lipgloss bold headers, alternating row styles, and "No packages installed" empty state.
- `RenderSearchResults(results)`: D-06 name+type+[registry]+description list, top 20, "No results found." empty state.

**`agentkit list`** (`cmd/list.go`) — Reads `ConfigStore.ListInstalled()` for the given `--target`, renders table. No `os.Exit(1)` on empty results — prints helpful message and returns 0.

**`agentkit search`** (`cmd/search.go`) — Shows bubbletea `SpinnerModel` (PhaseFetchRegistry) while background goroutine runs `SearchService.Search()`, then prints `RenderSearchResults` output. D-04 error format to stderr on failure.

## Verification Results

- `go build ./...` — 0
- `go test ./...` — 55 passed, 0 failed
- `go vet ./...` — 0
- `./agentkit list --help` — shows `--target` flag
- `./agentkit search --help` — shows `[query]` in usage

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None. Both commands are fully wired to real data sources (ConfigStore and RegistryManager).

## Threat Flags

None. Search query is passed to `strings.Contains`/`strings.EqualFold` only; no exec or SQL exposure. List output shows only the user's own installed packages.

## Self-Check: PASSED

- internal/service/search.go: FOUND
- internal/service/search_test.go: FOUND
- internal/ui/table.go: FOUND
- internal/ui/table_test.go: FOUND
- cmd/list.go: FOUND (modified)
- cmd/search.go: FOUND (modified)
- All commits verified via git log
