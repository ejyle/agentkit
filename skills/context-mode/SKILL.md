---
name: context-mode
description: >
  Use when any task produces more than 20 lines of output, reads files for analysis,
  makes HTTP/API calls, or runs grep/search commands — context-mode routes those
  operations through sandbox tools so raw output never floods your context window.
license: Apache-2.0
---

## When to Use

Activate this skill whenever the task involves:

- Running commands that produce more than 20 lines of terminal output
- Reading files to analyze, explore, or summarize (not to edit)
- Making HTTP calls, fetching URLs, or calling external APIs
- Running grep, find, or other search commands across large codebases
- Spawning subagents or multi-step research workflows

**Do NOT activate for:** Simple edits, small config writes, git commits, mkdir/rm/mv — those are fine with native Bash and Read tools.

## Blocked Commands

These Bash patterns are intercepted at the hook level and will fail with an error:

| Pattern | Reason | Alternative |
|---------|--------|-------------|
| `curl <url>` / `wget <url>` | Floods context with raw HTTP response | `ctx_fetch_and_index(url, source)` |
| `fetch('http...` inline | Same — raw JSON/HTML enters context | `ctx_execute(language: "javascript", ...)` |
| `requests.get(` / `http.get(` | Python/Node inline HTTP | `ctx_execute(language, code)` |
| `WebFetch` tool | Intercepted entirely | `ctx_fetch_and_index(url, source)` |

## Tool Selection Hierarchy

Use tools in this order — stop when you have enough:

| Priority | Tool | Use When |
|----------|------|----------|
| 1 — GATHER | `ctx_batch_execute(commands, queries)` | Primary: runs commands + indexes + searches in ONE call |
| 2 — FOLLOW-UP | `ctx_search(queries: ["q1", "q2"])` | Query already-indexed content; batch all questions in one call |
| 3 — PROCESSING | `ctx_execute(language, code)` | API calls, log analysis, data processing in sandbox |
| 3 — FILE | `ctx_execute_file(path, language, code)` | Analyze a file without loading it into context |
| 4 — WEB | `ctx_fetch_and_index(url, source)` + `ctx_search` | Fetch URL → chunk → index → search; HTML never enters context |
| 5 — INDEX | `ctx_index(content, source)` | Store content for later search with descriptive label |

### Read vs ctx_execute_file

- **Read** a file only when you need to **Edit** it — Edit requires file content in context.
- **ctx_execute_file** when you need to **analyze, summarize, or extract** — only your printed output enters context.

## Quick Reference

```bash
# GATHER: run commands and search in one shot
ctx_batch_execute(
  commands=["git log --oneline -20", "find . -name '*.go' | head -30"],
  queries=["recent commit messages", "go file list"]
)

# FOLLOW-UP: search already-indexed content
ctx_search(queries=["routing rules", "blocked commands", "tool hierarchy"])

# PROCESSING: run code in sandbox
ctx_execute(language="shell", code="wc -l src/**/*.ts | sort -n | tail -20")

# FILE ANALYSIS: analyze file without loading it
ctx_execute_file("/path/to/file.log", language="shell", code="grep ERROR | tail -50")

# WEB FETCH: index a URL then search it
ctx_fetch_and_index("https://docs.example.com/api", source="example-api-docs")
ctx_search(queries=["authentication", "rate limits"], source="example-api-docs")
```

## Output Constraints

- Keep responses under 500 words.
- Write artifacts (code, configs, reports) to FILES — never inline.
  Return only: file path + 1-line description.
- Use descriptive `source` labels when indexing so future searches can filter by label.

## ctx Management Commands

| Command | Action |
|---------|--------|
| `ctx stats` | Call `ctx_stats` MCP tool — display verbatim |
| `ctx doctor` | Call `ctx_doctor` MCP tool — run returned shell command; show as checklist |
| `ctx upgrade` | Call `ctx_upgrade` MCP tool — run returned shell command; show as checklist |

## Subagent Routing

When spawning subagents (Agent/Task tool), the context-mode routing block is automatically
injected into their prompt. You do NOT need to manually instruct subagents — they inherit
all routing rules automatically.

## Reference Files

| Task | Reference file |
|------|---------------|
| Full blocked-command list, routing rules by command type, grep/find patterns | `references/routing-rules.md` |
