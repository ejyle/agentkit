---
name: serena
description: >
  Use when navigating code structure, finding symbol definitions or references, performing
  safe renames, or doing any task that requires accurate cross-file code intelligence.
  Serena provides LSP-powered symbol navigation, reference search, and structural edits
  that are more reliable than text search alone.
license: Apache-2.0
---

## When to Use

Activate this skill when the task involves:

- Finding where a function, type, interface, or variable is defined
- Listing all callers/references of a symbol across the codebase
- Navigating to implementations of an interface or abstract method
- Renaming a symbol safely across all files
- Deleting a function/type and verifying no remaining references
- Getting a high-level structural overview of a file or package
- Running diagnostics to find type errors or linting issues without a full build

**Important:** Always call `initial_instructions` at the start of any Serena-heavy session.
This loads the Serena Instructions Manual with tool-specific usage patterns.

## Quick Start

```python
# Step 1 (mandatory): Load Serena instructions for this session
initial_instructions()

# Step 2 (optional): Run onboarding if this is a new project
onboarding()

# Step 3: Use navigation tools
get_symbols_overview("internal/installer/")      # structural overview of a directory
find_symbol("GitHubReleaseInstaller")            # find a symbol definition
find_referencing_symbols("Install", "method")    # find all callers
```

## Key Tool Groups

### Session Setup
| Tool | When |
|------|------|
| `initial_instructions()` | First call in any Serena session — loads the Instructions Manual |
| `onboarding()` | New project or after significant structural changes |

### Navigation
| Tool | When |
|------|------|
| `get_symbols_overview(path)` | High-level view of functions/types in a file or directory |
| `find_symbol(name)` | Locate definition of a function, type, or variable |
| `find_declaration(name)` | Find the declaration (e.g., interface definition) |
| `find_implementations(name)` | Find all structs/classes implementing an interface |
| `find_referencing_symbols(name, kind)` | Find all callers/users of a symbol |

### File & Pattern Search
| Tool | When |
|------|------|
| `list_dir(path)` | Directory listing with file metadata |
| `find_file(pattern)` | Locate a file by name or glob pattern |
| `search_for_pattern(pattern, path)` | Regex or text search across files |

### Structural Edits
| Tool | When |
|------|------|
| `replace_symbol_body(name, new_body)` | Replace the entire body of a function or method |
| `insert_after_symbol(name, code)` | Insert code immediately after a symbol |
| `insert_before_symbol(name, code)` | Insert code immediately before a symbol |
| `rename_symbol(old_name, new_name)` | Safe rename across all files |
| `safe_delete_symbol(name)` | Delete a symbol after verifying no remaining references |

### Diagnostics
| Tool | When |
|------|------|
| `get_diagnostics_for_file(path)` | Get type errors, linting issues for a file |

### Memory
| Tool | When |
|------|------|
| `write_memory(key, content)` | Store project-specific notes for future sessions |
| `read_memory(key)` | Retrieve previously stored notes |
| `list_memories()` | List all stored memory keys |

## Reference Files

| Task | Reference file |
|------|---------------|
| Detailed tool parameters, usage patterns, rename workflows, diagnostic integration | `references/lsp-usage.md` |

## Common Gotchas

- **Call `initial_instructions()` first** — Serena's tools behave differently depending on the
  project language and LSP server. The Instructions Manual loads this context.
- **`find_symbol` vs `search_for_pattern`** — `find_symbol` uses the LSP index (accurate, follows
  types); `search_for_pattern` is text search (fast, but may find comments and string literals).
- **`rename_symbol` is safe-rename** — it updates all references atomically. Do NOT use text
  find/replace for renames; it will miss dynamic references and break imports.
- **`safe_delete_symbol` checks references first** — it will refuse to delete if callers exist.
  Use `find_referencing_symbols` to understand what needs updating before deleting.
- **Memory is session-persistent** — use `write_memory` to store architectural notes, onboarding
  decisions, or project quirks so they survive across Claude Code sessions.
