## Overview

GitHub Issues track bugs, feature requests, and tasks. The `gh issue` CLI covers creation, listing, editing, and closing. Issues can be linked to milestones, projects, labels, and assignees. Bulk operations are best done with `gh issue list --json` piped through `jq`.

## Common Commands

### Creating Issues

```bash
# Create issue interactively
gh issue create

# Create with all fields
gh issue create \
  --title "Bug: login fails on mobile Safari" \
  --body "## Steps to reproduce
1. Open app on iOS Safari
2. Enter credentials
3. Tap Sign In

## Expected
User is redirected to dashboard.

## Actual
Page reloads, user remains on login page." \
  --label bug \
  --label "mobile" \
  --assignee alice \
  --milestone "v1.2"

# Create from a template (if .github/ISSUE_TEMPLATE/ exists)
gh issue create --template bug_report.md
```

### Listing and Filtering Issues

```bash
# List all open issues
gh issue list

# Filter by label
gh issue list --label bug
gh issue list --label "high-priority,regression"

# Filter by state
gh issue list --state open
gh issue list --state closed
gh issue list --state all

# Filter by assignee
gh issue list --assignee alice
gh issue list --assignee @me

# Full-text search
gh issue list --search "login mobile"
gh issue list --search "label:bug is:open sort:created-asc"

# Output as JSON for scripting
gh issue list --json number,title,state,labels,assignees \
  --jq '.[] | "\(.number): [\(.labels | map(.name) | join(", "))] \(.title)"'
```

### Viewing and Editing Issues

```bash
# View issue details
gh issue view 42

# View in browser
gh issue view 42 --web

# Edit issue (change title, body, labels, assignee, milestone)
gh issue edit 42 --title "Updated: login fails on mobile"
gh issue edit 42 --add-label "confirmed"
gh issue edit 42 --remove-label "needs-triage"
gh issue edit 42 --add-assignee bob
gh issue edit 42 --milestone "v1.3"

# Add a comment
gh issue comment 42 --body "Reproduced on iOS 17.4. Priority bump to P1."
```

### Closing and Reopening

```bash
# Close issue
gh issue close 42

# Close with a comment
gh issue close 42 --comment "Fixed in PR #123. Will ship in v1.2."

# Reopen issue
gh issue reopen 42
gh issue reopen 42 --comment "Regression found in v1.2.1"
```

### Bulk Operations with JSON + jq

```bash
# Close all issues matching a label
gh issue list --label "wont-fix" --state open --json number \
  --jq '.[].number' | xargs -I {} gh issue close {} --comment "Closing as wont-fix."

# List issues without assignees
gh issue list --state open --json number,title,assignees \
  --jq '.[] | select(.assignees == []) | "\(.number): \(.title)"'

# Export issues to CSV
gh issue list --state all --limit 200 \
  --json number,title,state,createdAt,closedAt,labels \
  --jq '.[] | [.number, .title, .state, .createdAt, .closedAt] | @csv'

# Count issues per label
gh issue list --state open --limit 200 --json labels \
  --jq '[.[].labels[].name] | group_by(.) | map({label: .[0], count: length}) | sort_by(.count) | reverse[]'
```

## Patterns

### Issue Templates

Create `.github/ISSUE_TEMPLATE/` directory with multiple template files:

```
.github/
  ISSUE_TEMPLATE/
    bug_report.md
    feature_request.md
    question.md
    config.yml
```

`config.yml` can disable blank issues and add external links:

```yaml
blank_issues_enabled: false
contact_links:
  - name: Community Forum
    url: https://community.example.com
    about: Ask questions here first
```

### Linking Issues to PRs

In PR body or commit messages:

```
Closes #42          → closes on merge to default branch
Fixes #42           → same
Resolves #42        → same
Related to #42      → cross-reference only (no auto-close)
Closes owner/repo#42  → cross-repo close
```

### Sprint / Milestone Workflow

```bash
# Create milestone
gh api repos/OWNER/REPO/milestones \
  --method POST \
  --field title="v1.3" \
  --field due_on="2025-03-31T00:00:00Z"

# List milestones
gh api repos/OWNER/REPO/milestones --jq '.[].title'

# Move issues to milestone
gh issue edit 42 --milestone "v1.3"

# View milestone progress
gh api repos/OWNER/REPO/milestones \
  --jq '.[] | select(.title=="v1.3") | {open: .open_issues, closed: .closed_issues}'
```

## Gotchas

- **Labels must exist before use** — `gh issue create --label new-label` fails if the label doesn't exist on the repo; create it first via `gh label create new-label --color ff0000`
- **`--search` uses GitHub Search syntax** — the search string supports qualifiers like `is:open`, `label:bug`, `author:alice`, `sort:updated-asc`; see GitHub docs for full syntax
- **Bulk edits via `xargs` may hit rate limits** — GitHub allows 5,000 API requests/hour; for large repos, add `sleep 0.1` between requests in the xargs pipeline
- **Closing with a comment requires separate call** — `gh issue close 42 --comment "..."` sends two API requests; the comment may appear before or after the close event in the timeline
- **Issue numbers are per-repo** — if you fork a repo, issue numbers restart at 1 in the fork; cross-reference with `owner/repo#N` syntax
