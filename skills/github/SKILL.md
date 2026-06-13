---
name: github
description: >
  Use when managing GitHub repositories, pull requests, issues, Actions workflows, or secrets
  via the gh CLI or GitHub API.
license: Apache-2.0
---

## When to Use

Activate this skill when the task involves:

- Creating, reviewing, merging, or listing pull requests with `gh pr`
- Creating, editing, triaging, or querying issues with `gh issue`
- Triggering, monitoring, or debugging GitHub Actions workflows with `gh run` / `gh workflow`
- Managing repository secrets, environments, or deployment rules
- Cloning repos, forking, or managing GitHub releases
- Writing or debugging `.github/workflows/*.yml` files
- Configuring branch protection rules or CODEOWNERS

## Quick Reference

### Authentication

```bash
# Login (opens browser)
gh auth login

# Login with token (for CI)
echo $GITHUB_TOKEN | gh auth login --with-token

# Check auth status
gh auth status
```

### Pull Requests

```bash
# Create PR from current branch
gh pr create --title "Add feature X" --body "Closes #42"

# Create draft PR
gh pr create --title "WIP: Add feature X" --draft

# List open PRs
gh pr list

# View PR details
gh pr view 123

# Checkout PR branch locally
gh pr checkout 123

# Review PR
gh pr review 123 --approve
gh pr review 123 --request-changes --body "Please address the comments"

# Merge PR
gh pr merge 123 --squash --delete-branch
gh pr merge 123 --rebase
gh pr merge 123 --merge
```

### Issues

```bash
# Create issue
gh issue create --title "Bug: login fails on mobile" \
  --body "Steps to reproduce..." --label bug --assignee alice

# List issues
gh issue list
gh issue list --label "bug" --state open

# View issue
gh issue view 42

# Close issue
gh issue close 42 --comment "Fixed in #123"
```

### GitHub Actions

```bash
# List workflow runs
gh run list

# View run details (including logs)
gh run view 12345

# Watch run in real time
gh run watch 12345

# Manually trigger a workflow
gh workflow run my-workflow.yml

# Trigger with inputs
gh workflow run deploy.yml --field environment=staging

# List workflows
gh workflow list

# Manage secrets
gh secret set MY_SECRET
gh secret list
```

### Repositories

```bash
# Clone a repo
gh repo clone owner/repo

# Fork and clone
gh repo fork owner/repo --clone

# Create new repo
gh repo create my-repo --public --source=. --push

# View repo info
gh repo view owner/repo
```

## Reference Files

Load the appropriate reference file for deep-dive tasks:

| Task | Reference file |
|------|---------------|
| PR lifecycle, review workflow, merge strategies, conflict resolution | `references/prs.md` |
| Issue management, bulk ops, templates, PR linking | `references/issues.md` |
| Actions runs, workflow triggers, secrets, artifacts, matrices | `references/actions.md` |

## Common Gotchas

- **`GH_TOKEN` vs `GITHUB_TOKEN`** — `GH_TOKEN` is the env var for `gh` CLI; `GITHUB_TOKEN` is the Actions-provided token; both work for most operations but have different scopes
- **Draft PRs don't trigger CI** — some workflows are configured to skip draft PRs; convert to ready-for-review when you want CI to run
- **`gh pr create` uses upstream by default** — in a fork, `--head` defaults to the fork's branch but `--base` defaults to the upstream default branch
- **Rate limiting** — GitHub REST API is limited to 5,000 requests/hour for authenticated users; `gh` CLI caches some responses; add `--json` output and pipe through `jq` to batch operations
