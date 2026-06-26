---
slug: github-default-branch
date: 2026-06-26
status: complete
---

# Quick Task: Add `github-default-branch` install method

## Tasks

- [x] 1. Add `InstallMethodGitHubDefaultBranch` constant to `internal/domain/package.go`
- [x] 2. Add `ErrGitHubDefaultBranchNotFound` sentinel + wire factory in `internal/installer/installer.go`
- [x] 3. Create `internal/installer/github_default_branch.go` installer implementation
- [x] 4. Create `internal/installer/github_default_branch_test.go` unit tests (7 tests, all pass)
- [x] 5. Update `azure-skills` entry in `testdata/registry.json` to use new method
- [x] 6. Build and verify — 141/141 tests pass
