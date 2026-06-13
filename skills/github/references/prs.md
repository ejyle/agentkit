## Overview

Pull requests are the primary code review and merge mechanism on GitHub. The `gh pr` CLI covers the full lifecycle from creation to merge. A PR compares a head branch (with changes) against a base branch (merge target). Reviews, approvals, and branch protection rules gate the merge.

## Common Commands

### Creating PRs

```bash
# Create PR interactively (prompts for title and body)
gh pr create

# Create with all details
gh pr create \
  --title "feat: add user authentication" \
  --body "## Summary
- Implements JWT login
- Adds refresh token rotation

Closes #42" \
  --base main \
  --head feature/auth \
  --label enhancement \
  --assignee @me \
  --reviewer alice,bob

# Create draft PR (signals WIP; some CI skips drafts)
gh pr create --title "WIP: authentication" --draft

# Convert draft to ready
gh pr ready 123
```

### Listing and Viewing PRs

```bash
# List open PRs
gh pr list
gh pr list --author @me
gh pr list --reviewer @me
gh pr list --label "needs-review" --state open

# View PR summary
gh pr view 123

# View PR in browser
gh pr view 123 --web

# Check CI status on PR
gh pr checks 123

# List PR as JSON for scripting
gh pr list --json number,title,author,state --jq '.[] | "\(.number): \(.title)"'
```

### Reviewing PRs

```bash
# Approve a PR
gh pr review 123 --approve

# Request changes with comment
gh pr review 123 --request-changes \
  --body "Please address the security issue in auth.ts line 42."

# Leave a comment without approval decision
gh pr review 123 --comment \
  --body "Looks good overall, one minor nit inline."

# Add inline comment via API
gh api repos/OWNER/REPO/pulls/123/comments \
  --method POST \
  --field body="Prefer `const` here" \
  --field path="src/auth.ts" \
  --field position=10 \
  --field commit_id=$(gh pr view 123 --json headRefOid -q .headRefOid)
```

### Checking Out PRs

```bash
# Checkout PR branch (creates local tracking branch)
gh pr checkout 123

# Checkout into a new local branch name
gh pr checkout 123 --branch local-review-branch

# After reviewing locally, go back to your branch
git checkout main
```

### Merging PRs

```bash
# Squash and merge (squashes all commits into one)
gh pr merge 123 --squash --delete-branch

# Rebase and merge (linear history, commits preserved)
gh pr merge 123 --rebase --delete-branch

# Merge commit (preserves branch history)
gh pr merge 123 --merge

# Auto-merge when checks pass
gh pr merge 123 --squash --auto
```

### Conflict Resolution

```bash
# Update local branch with latest base
git fetch origin main
git rebase origin/main

# Or merge-based update
git merge origin/main

# After resolving conflicts
git add .
git rebase --continue   # (if using rebase)
# or
git commit              # (if using merge)
git push origin HEAD --force-with-lease   # safe force-push after rebase

# Check if PR is up to date
gh pr view 123 --json mergeStateStatus -q .mergeStateStatus
# "CLEAN" = mergeable; "BEHIND" = needs update
```

### Branch Protection (Viewing)

```bash
# View branch protection rules via API
gh api repos/OWNER/REPO/branches/main/protection

# List required status checks
gh api repos/OWNER/REPO/branches/main/protection/required_status_checks
```

## Patterns

### PR Template Setup

Create `.github/pull_request_template.md` in the repo root:

```markdown
## Summary
<!-- What does this PR do? -->

## Test Plan
- [ ] Unit tests pass
- [ ] Manual testing performed on: 

## Checklist
- [ ] Code follows project conventions
- [ ] No secrets committed
- [ ] Documentation updated if needed

Closes #
```

### Linking PRs to Issues

- Body text: `Closes #42`, `Fixes #42`, `Resolves #42` — auto-closes issue on merge
- Multiple: `Closes #42, Closes #43`
- Cross-repo: `Closes owner/repo#42`

### Rebase-and-Merge Workflow

1. Feature branch off main: `git checkout -b feature/my-feature main`
2. Commit small atomic commits with clear messages
3. Before PR: `git fetch origin main && git rebase origin/main`
4. Push: `git push origin feature/my-feature --force-with-lease`
5. Create PR targeting main
6. After approval: `gh pr merge --rebase --delete-branch`

## Gotchas

- **`--force-with-lease` vs `--force`** — `--force-with-lease` fails if someone else pushed to the branch since your last fetch; safer than `--force` which overwrites without checking
- **Auto-merge requires status checks** — `gh pr merge --auto` only works if branch protection requires at least one status check; otherwise the PR merges immediately
- **Squash loses individual commit messages** — the squash commit message defaults to the PR title; add a description in the PR body so it appears in the squash commit body
- **Draft PRs may not trigger CI** — workflow triggers can have `if: github.event.pull_request.draft == false`; check the workflow files before assuming CI ran on a draft
