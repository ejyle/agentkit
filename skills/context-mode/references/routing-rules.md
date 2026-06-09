# context-mode: Routing Rules Reference

Full routing rules and tool patterns for context window protection.

---

## Blocked Commands — Complete List

### curl / wget

**Trigger pattern:** Any Bash call containing `curl` or `wget`
**Hook behavior:** Intercepted and replaced with an error — do NOT retry with Bash.

**Alternatives:**
```python
# Index a URL for later searching
ctx_fetch_and_index(url="https://example.com/docs", source="example-docs")
ctx_search(queries=["topic 1", "topic 2"], source="example-docs")

# Make an API call and capture only stdout
ctx_execute(language="javascript", code="""
const r = await fetch('https://api.example.com/v1/endpoint', {
  headers: { Authorization: 'Bearer TOKEN' }
});
const data = await r.json();
console.log(JSON.stringify(data.items.slice(0, 5), null, 2));
""")
```

### Inline HTTP in Bash

**Trigger patterns:**
- `fetch('http...`
- `requests.get(` / `requests.post(`
- `http.get(` / `http.request(`

**Hook behavior:** Intercepted — do NOT retry with Bash.

**Alternative:** Same as curl — use `ctx_execute(language, code)`.

### WebFetch Tool

**Hook behavior:** Denied entirely. URL is extracted and you receive an error.

**Alternative:** `ctx_fetch_and_index(url, source)` then `ctx_search(queries)`.

---

## Redirected Tools — Routing Rules

### Bash — Allowed vs. Redirected

**Always allowed:**
```bash
git status / log / diff / add / commit / push / pull
mkdir / rm / mv / cp / ln
ls / pwd / cd
npm install / pip install / go get / cargo add
chmod / chown
echo (short output only)
wc -l (on a single file)
```

**Redirect to ctx_execute when output > 20 lines:**
```bash
# WRONG — floods context
find . -name "*.go" -exec grep -l "InstallSpec" {} \;

# CORRECT — sandbox only
ctx_execute(language="shell", code="""
find . -name '*.go' -exec grep -l 'InstallSpec' {} \\;
""")
```

**Redirect to ctx_execute_file for file analysis:**
```bash
# WRONG — 300-line file enters context
Read("/path/to/large/file.go")  # use for editing only

# CORRECT — only your print() output enters context
ctx_execute_file("/path/to/large/file.go", language="shell", code="""
grep -n 'func\|type\|interface' | head -40
""")
```

### Read Tool — When to Use

| Situation | Use |
|-----------|-----|
| You need to Edit the file | Read (Edit requires content in context) |
| You need to Write a new version | Read (need current content to avoid accidental overwrites) |
| You need to analyze, summarize, explore | ctx_execute_file |
| You need to check a specific line range | Read with `offset` + `limit` |

### Grep — Routing Rules

All grep commands that could return more than 20 lines must run in sandbox:

```python
# WRONG
Bash("grep -r 'InstallMethod' internal/")

# CORRECT — only matches enter context
ctx_execute(language="shell", code="""
grep -rn 'InstallMethod' internal/ --include='*.go' | head -30
""")
```

---

## ctx_batch_execute — Primary Tool

Use this as the default first step for any research or exploration task.

```python
# Template: gather + search in one call
ctx_batch_execute(
  commands=[
    "git log --oneline -20",
    "find . -name '*.go' | grep -v test | head -40",
    "wc -l internal/**/*.go | sort -n | tail -20",
  ],
  queries=[
    "what changed recently",
    "which files are largest",
    "installer pattern",
  ]
)
```

**Key benefits:**
- Runs all commands in parallel
- Auto-indexes all output
- Returns search results immediately
- One call replaces 20–30 individual Read/Bash/Grep calls

---

## ctx_search — Follow-Up Queries

After any batch_execute or index operation, search the indexed content:

```python
# Always pass multiple questions in ONE call
ctx_search(queries=[
  "routing rules for curl",
  "blocked command patterns",
  "tool hierarchy order",
  "subagent inheritance",
])
```

**Never call ctx_search in a loop** — pass all questions at once.

---

## ctx_execute — Sandbox Processing

For any processing that produces large output or requires computation:

```python
# Shell commands
ctx_execute(language="shell", code="""
wc -l skills/*/SKILL.md | sort -n
grep -r 'TODO\|stub\|placeholder' skills/ --include='*.md'
""")

# JavaScript for API calls
ctx_execute(language="javascript", code="""
const r = await fetch('https://api.github.com/repos/owner/repo/releases/latest');
const rel = await r.json();
console.log(rel.tag_name, rel.published_at);
""")

# Python for data processing
ctx_execute(language="python", code="""
import json, sys
data = json.load(open('registry.json'))
for pkg in data['packages'][:5]:
    print(pkg['name'], pkg['version'])
""")
```

---

## ctx_fetch_and_index — Web Content

Full web fetch pipeline:

```python
# Step 1: Fetch and index (raw HTML stays in sandbox)
ctx_fetch_and_index(
  url="https://docs.agentskills.io/specification",
  source="agentskills-spec"
)

# Step 2: Search the indexed content
ctx_search(
  queries=["required frontmatter fields", "line limit", "references structure"],
  source="agentskills-spec"
)
```

**Use descriptive source labels** — they enable targeted search across multiple indexed URLs.

---

## Output Constraints

When context-mode is active, the agent must follow these output rules:

1. **500-word response limit** — concise summaries only
2. **Artifacts go to files** — never inline code/configs/reports
3. **Return format:** file path + 1-line description only
4. **Source labels** — always use descriptive `source=` values when indexing

---

## ctx Management

```bash
# Check token savings since session start
ctx stats        # calls ctx_stats MCP tool; display verbatim

# Diagnose context-mode installation
ctx doctor       # calls ctx_doctor MCP tool; run returned shell command; show as checklist

# Update context-mode to latest version
ctx upgrade      # calls ctx_upgrade MCP tool; run returned shell command; show as checklist
```
