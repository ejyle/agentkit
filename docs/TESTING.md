<!-- generated-by: gsd-doc-writer -->
# Testing

## Test Framework and Setup

agentkit uses Go's standard `testing` package — no third-party test framework is required. The module is defined at `github.com/ejyle/agentkit` in `go.mod` with Go 1.26.3.

All tests are co-located with the packages they test under `internal/`, following Go conventions. No global test setup or teardown is needed beyond a working Go installation.

**Packages with test coverage:**

| Package | Test file(s) |
|---------|-------------|
| `internal/adapter` | `claude_test.go`, `codex_test.go`, `copilot_cli_test.go`, `copilot_vscode_test.go`, `gemini_test.go`, `opencode_test.go`, `pi_test.go` |
| `internal/bundle` | `bundles_test.go` |
| `internal/config` | `store_test.go` |
| `internal/domain` | `installed_test.go`, `package_test.go` |
| `internal/installer` | `binary_test.go`, `custom_test.go`, `docker_test.go`, `github_default_branch_test.go`, `github_release_integration_test.go`, `github_release_test.go`, `npx_test.go`, `uvx_test.go` |
| `internal/registry` | `github_test.go` |
| `internal/service` | `install_test.go`, `search_test.go`, `uninstall_test.go`, `update_test.go` |
| `internal/skill` | `validate_test.go` |
| `internal/ui` | `table_test.go` |

## Running Tests

**Run the full test suite:**

```bash
go test ./...
```

**Run with verbose output (shows individual test names and results):**

```bash
go test -v ./...
```

**Run tests for a single package:**

```bash
go test ./internal/service/...
go test ./internal/adapter/...
go test ./internal/registry/...
```

**Run a specific test by name:**

```bash
go test -run TestBinaryInstaller_Install_Success ./internal/installer/...
go test -run TestInstall ./internal/service/...
```

**Run with race detector (recommended before merging):**

```bash
go test -race ./...
```

**Run with coverage report:**

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Run with coverage summary in terminal:**

```bash
go test -cover ./...
```

## Writing New Tests

**File naming convention:** Test files use the `*_test.go` suffix and live in the same directory as the package under test (e.g., `internal/service/install_test.go` tests `internal/service/`).

**Package naming:** Tests use external test packages (e.g., `package service_test`, `package adapter_test`) to enforce testing through the public API surface. This prevents tests from relying on unexported internals.

**Test helper patterns:**

- Use `t.TempDir()` for any test that needs a temporary directory — Go cleans it up automatically after the test.
- Use `net/http/httptest.NewServer` or `httptest.NewTLSServer` to spin up local HTTP servers for registry and installer tests. See `internal/registry/github_test.go` and `internal/installer/binary_test.go` for examples.
- Use mock structs implementing domain interfaces for service-layer tests. See `internal/service/install_test.go` for the `mockRegistry`, `mockInstaller`, and `mockAdapter` patterns.
- Use `t.Helper()` in helper functions so failure lines point to the calling test.

**Testdata:** Static fixture files live in `testdata/` at the project root (e.g., `testdata/registry.json`). Reference them with relative paths from the test file's directory.

**Example test structure:**

```go
package mypackage_test

import (
    "testing"

    "github.com/ejyle/agentkit/internal/mypackage"
)

func TestMyFunc_Success(t *testing.T) {
    got, err := mypackage.MyFunc("input")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got != "expected" {
        t.Errorf("got %q, want %q", got, "expected")
    }
}
```

## Coverage Requirements

No coverage threshold is configured. The project uses Go's built-in coverage tooling without enforced minimums.

To inspect per-package coverage locally:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## CI Integration

Tests are not currently run as a dedicated CI step. The `.github/workflows/release.yml` workflow runs GoReleaser on tag pushes (`v*`) and snapshot builds on pushes to `main` — neither job includes an explicit `go test` step.

To add a CI test job, create `.github/workflows/test.yml` triggered on `push` and `pull_request`:

```yaml
name: Test

on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - run: go test -race ./...
```
