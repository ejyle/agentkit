# Serena: LSP Usage Reference

Detailed tool parameters, usage patterns, and workflows for Serena's LSP-powered
code intelligence tools.

---

## Session Initialization

### `initial_instructions()`

**Must be called first** in any Serena-heavy session. Loads the Serena Instructions Manual
with project-specific LSP configuration, language server selection, and tool-usage guidelines.

```python
initial_instructions()
# Returns: the Serena Instructions Manual for this project
# Side effect: configures Serena's internal LSP connection
```

### `onboarding()`

Run when starting work on a project for the first time, or after significant structural changes
(e.g., a major refactor, new package added, language server switched).

```python
onboarding()
# Returns: project overview — file tree, detected language, LSP status, recommended memory keys
```

---

## Navigation Tools

### `get_symbols_overview(path)`

Returns a structural overview of all symbols (functions, types, constants, interfaces) in a
file or directory. Useful for understanding a package without reading every file.

```python
get_symbols_overview("internal/installer/")
# Returns: symbol list for all .go files in internal/installer/
# Output: symbol name, kind (function/type/const), line number, file

get_symbols_overview("internal/installer/github_release.go")
# Returns: all symbols in that specific file
```

**Use before editing:** Get the overview to understand what's in a file before making changes.

### `find_symbol(name)`

Locate the definition of a named symbol across the project. Returns file path and line number.

```python
find_symbol("GitHubReleaseInstaller")
# Returns: internal/installer/github_release.go:12

find_symbol("InstallMethodGitHubRelease")
# Returns: internal/domain/package.go:45
```

**Tip:** Use for any symbol where text search might return false positives (e.g., comments,
string literals). `find_symbol` only returns actual definitions.

### `find_declaration(name)`

Find the declaration of a symbol — typically for interface definitions or type aliases where
the symbol may be defined separately from its primary implementation.

```python
find_declaration("Installer")
# Returns: internal/domain/installer.go:8  (interface Installer { ... })
```

### `find_implementations(name)`

Find all concrete implementations of an interface or all method implementations of a given
method name across the project.

```python
find_implementations("Installer")
# Returns: NpxInstaller (internal/installer/npx.go:15)
#          BinaryInstaller (internal/installer/binary.go:22)
#          GitHubReleaseInstaller (internal/installer/github_release.go:12)
```

**Use for:** Verifying interface coverage, finding all places a method must be updated.

### `find_referencing_symbols(name, kind)`

Find all symbols that reference or call the given symbol. The `kind` parameter filters by
reference type: `"method"`, `"field"`, `"type"`, `"import"`.

```python
find_referencing_symbols("Install", "method")
# Returns: all places that call the Install method
# Output: file path, line, containing function

find_referencing_symbols("InstallSpec", "type")
# Returns: all places that use the InstallSpec type
```

**Use before refactoring:** Always run this before renaming or changing a symbol signature.

---

## File and Pattern Search

### `list_dir(path)`

Directory listing with metadata (file size, last modified). More informative than `ls`.

```python
list_dir("skills/aws/")
# Returns: SKILL.md (2.4KB), references/ (dir)

list_dir("internal/")
# Returns: all subdirectories with counts
```

### `find_file(pattern)`

Find files by name or glob pattern across the project.

```python
find_file("github_release.go")
# Returns: internal/installer/github_release.go

find_file("*.json")
# Returns: all .json files in the project
```

### `search_for_pattern(pattern, path)`

Regex or text search across files. Returns file path, line number, and matching line.

```python
search_for_pattern("InstallMethod.*=", "internal/domain/")
# Returns all constant definitions for InstallMethod

search_for_pattern("github-release", ".")
# Returns all files mentioning "github-release"
```

**vs. find_symbol:** `search_for_pattern` is text-based — it finds matches in comments and
strings too. Use `find_symbol` when you need accurate type-aware symbol lookup.

---

## Structural Edits

### `replace_symbol_body(name, new_body)`

Replace the entire body of a function or method. Preserves the signature line.

```python
replace_symbol_body("Install", """
func (g *GitHubReleaseInstaller) Install(spec domain.InstallSpec) error {
    tarball, err := g.fetchTarball(spec.Repo)
    if err != nil {
        return fmt.Errorf("fetch tarball: %w", err)
    }
    return g.extract(tarball, spec.Path)
}
""")
```

**When to use:** Rewriting a function implementation entirely. Safer than text editing
because it targets the exact symbol boundaries.

### `insert_after_symbol(name, code)`

Insert code immediately after a named symbol ends.

```python
insert_after_symbol("NewGitHubReleaseInstaller", """

func (g *GitHubReleaseInstaller) Method() domain.InstallMethod {
    return domain.InstallMethodGitHubRelease
}
""")
```

### `insert_before_symbol(name, code)`

Insert code immediately before a named symbol.

```python
insert_before_symbol("GitHubReleaseInstaller", """
// cacheKey returns the cache key for a repo+version combination.
func cacheKey(repo, version string) string {
    return repo + "@" + version
}

""")
```

### `rename_symbol(old_name, new_name)`

Safe rename across all files — updates definition and all references atomically.

```python
rename_symbol("GitHubReleaseInstaller", "ReleaseInstaller")
# Updates: definition in github_release.go
#          all call sites in installer.go, install.go, tests
#          import references if applicable
```

**Never use text find/replace for renames** — it misses dynamic references, interface
implementations, and embedded struct references.

### `safe_delete_symbol(name)`

Delete a symbol after verifying no remaining references exist.

```python
safe_delete_symbol("legacyInstall")
# If callers exist: returns error listing all referencing symbols
# If no callers: deletes the symbol and its surrounding whitespace
```

**Workflow:**
1. `find_referencing_symbols("legacyInstall", "method")` — understand what calls it
2. Update or remove all callers
3. `safe_delete_symbol("legacyInstall")` — clean deletion

---

## Diagnostics

### `get_diagnostics_for_file(path)`

Get type errors, unresolved imports, and linting issues for a file without running a build.

```python
get_diagnostics_for_file("internal/installer/github_release.go")
# Returns: list of errors with file, line, severity, message

# Example output:
# ERROR   line 45: undefined: sync.Map (import "sync" missing)
# WARNING line 12: exported type GitHubReleaseInstaller without documentation
```

**Use when:** You've made structural edits and want to catch type errors before running `go build`.

---

## Memory Tools

Serena's memory tools persist notes across Claude Code sessions within the same project.

### `write_memory(key, content)`

Store project-specific notes under a named key.

```python
write_memory("installer-pattern", """
All installers implement domain.Installer interface:
- Method() domain.InstallMethod
- IsAvailable() bool
- Install(spec domain.InstallSpec) error

New installers must be registered in NewInstaller() switch statement
in internal/installer/installer.go
""")
```

### `read_memory(key)`

Retrieve a previously stored note.

```python
read_memory("installer-pattern")
# Returns: the stored text above
```

### `list_memories()`

List all stored memory keys for this project.

```python
list_memories()
# Returns: ["installer-pattern", "bundle-architecture", "registry-format"]
```

**Recommended memory keys for agentkit:**
- `installer-pattern` — interface contract and registration pattern
- `bundle-architecture` — bundle loader, JSON format, parallel install flow
- `registry-format` — registry.json schema, manifest parsing rules
- `skill-structure` — SKILL.md spec, references/ conventions, line limits

---

## Common Workflows

### Refactoring a Method Signature

1. `find_symbol("MethodName")` — locate definition
2. `find_referencing_symbols("MethodName", "method")` — find all callers
3. Update callers with `replace_symbol_body` or direct edits
4. `replace_symbol_body("MethodName", newImpl)` — update the definition
5. `get_diagnostics_for_file(file)` — verify no type errors

### Safe Interface Evolution

1. `find_declaration("InterfaceName")` — see current interface
2. `find_implementations("InterfaceName")` — find all concrete types
3. Add method to interface
4. Update each implementation found in step 2
5. `get_diagnostics_for_file` on each affected file

### Cross-Package Symbol Search

1. `get_symbols_overview("pkg/")` — understand the package surface
2. `find_symbol("SymbolName")` — locate the definition
3. `find_referencing_symbols("SymbolName", "type")` — find all usages across packages
4. Make targeted edits based on the full reference map
