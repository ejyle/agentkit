# RTK: Command Reference

Full command reference for RTK (Rust Token Killer) — token-optimized CLI proxy.

---

## What RTK Does

RTK intercepts CLI commands via a Claude Code hook and filters output before it enters
the context window. A raw `git log --oneline -100` might produce 100 lines; RTK truncates,
deduplicates, and summarizes it to the essential signal. Token savings: 60–90% on typical
dev operations.

---

## Meta Commands (Always Call with `rtk` Directly)

These bypass the hook and must be invoked with the explicit `rtk` prefix:

### `rtk gain`

Show token savings analytics for the current session.

```
$ rtk gain
Session token savings:
  Commands intercepted: 47
  Raw tokens (estimated): 84,320
  Filtered tokens:          9,150
  Savings:                 89.1%  (75,170 tokens)
  Top commands: git log (28.4K saved), npm install (19.7K saved)
```

### `rtk gain --history`

Show cumulative savings across all sessions, with per-command breakdown:

```
$ rtk gain --history
All-time token savings:
  Total sessions: 12
  Total saved:    1.2M tokens
  
  Command       | Calls | Avg Savings
  --------------|-------|------------
  git log       |   342 | 96%
  git diff      |   218 | 78%
  npm install   |    89 | 91%
  go test       |   156 | 67%
```

### `rtk discover`

Analyze Claude Code session history for commands that WERE NOT intercepted but could have
been. Use this to find missed optimization opportunities and update hook configuration:

```
$ rtk discover
Missed opportunities in last 5 sessions:
  - cargo build (12 occurrences, estimated 45K tokens wasted)
  - docker logs  (8 occurrences, estimated 23K tokens wasted)
  - kubectl get pods (6 occurrences)
  
Suggested hook additions:
  cargo build → rtk cargo build
  docker logs  → rtk docker logs
```

### `rtk proxy <cmd>`

Execute a raw command without RTK filtering — useful for debugging when you need to see
unfiltered output:

```bash
rtk proxy git log --oneline -50     # raw git log, no filtering
rtk proxy npm install               # raw npm output
rtk proxy go build ./...            # raw build output
```

Use `rtk proxy` when:
- You suspect RTK is filtering important error messages
- You need to see the full stack trace that RTK may be truncating
- Debugging hook configuration

---

## Hook-Based Commands (Automatic Interception)

These are rewritten transparently — write them normally, RTK intercepts automatically:

### Git Commands

| Raw Command | Intercepted As | Token Savings |
|-------------|---------------|---------------|
| `git status` | `rtk git status` | ~40% (short output filtered) |
| `git log --oneline -50` | `rtk git log ...` | ~70% |
| `git log` (full) | `rtk git log` | ~95% |
| `git diff` | `rtk git diff` | ~60% |
| `git diff HEAD~5` | `rtk git diff ...` | ~80% |
| `git blame <file>` | `rtk git blame ...` | ~85% |
| `git show <hash>` | `rtk git show ...` | ~75% |

### Package Manager Commands

| Raw Command | Token Savings |
|-------------|---------------|
| `npm install` | ~90% (install noise removed) |
| `npm run build` | ~70% (build output filtered) |
| `npm test` | ~65% (test output summarized) |
| `yarn install` | ~90% |
| `pnpm install` | ~90% |
| `go get ./...` | ~80% |
| `pip install -r requirements.txt` | ~85% |

### Build Tools

| Raw Command | Token Savings |
|-------------|---------------|
| `go build ./...` | ~60% |
| `cargo build` | ~70% |
| `make` | ~65% |
| `gradle build` | ~80% |
| `mvn package` | ~75% |

### Test Runners

| Raw Command | Token Savings |
|-------------|---------------|
| `go test ./...` | ~65% (pass/fail summary vs. verbose output) |
| `jest` | ~70% |
| `pytest` | ~75% |
| `cargo test` | ~70% |

---

## Installation Verification

```bash
rtk --version
# Expected: rtk X.Y.Z

rtk gain
# Expected: session savings report (not "command not found")

which rtk
# Expected: /usr/local/bin/rtk or ~/.local/bin/rtk
# NOT expected: reachingforthejack/rtk (Rust Type Kit — name collision)
```

### Name Collision: reachingforthejack/rtk

Two packages share the `rtk` binary name:
- **This tool:** Rust Token Killer — a Claude Code optimization proxy
- **reachingforthejack/rtk:** Rust Type Kit — a different tool entirely

If `rtk gain` returns an error about unknown subcommand, you have the wrong binary.
Uninstall `reachingforthejack/rtk` and install the correct Rust Token Killer.

---

## Hook Configuration

RTK hooks into Claude Code via the `hooks` section in `~/.claude/settings.json`.
The hook intercepts Bash tool calls matching configured patterns and prepends `rtk`.

If the hook is not firing:
1. Check `~/.claude/settings.json` for the `hooks` section
2. Run `rtk discover` to see which commands are being missed
3. Verify `which rtk` returns the correct binary

---

## Filtering Behavior

RTK applies different filters per command type:

- **Git log:** Keeps commit hashes + messages, strips decoration and color codes
- **Git diff:** Keeps changed lines with context, removes binary file noise
- **npm install:** Keeps package count and warnings; strips progress bars, URL lines, audit noise
- **Build tools:** Keeps errors and warnings; strips successful compilation noise
- **Test runners:** Keeps test names + pass/fail; strips internal framework output

When RTK's filtering is too aggressive for your use case, use `rtk proxy <cmd>` to get
raw output.
