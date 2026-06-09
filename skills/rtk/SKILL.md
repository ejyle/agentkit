---
name: rtk
description: >
  Use when running any dev operations at the command line — git, npm, build tools,
  test runners, or any CLI that produces terminal output. RTK (Rust Token Killer) is
  a transparent proxy that intercepts and filters command output, saving 60–90% of
  tokens on routine dev operations.
license: Apache-2.0
---

## When to Use

Activate this skill for any session where you are running CLI tools:

- Git operations (`git status`, `git log`, `git diff`, `git blame`)
- Package manager commands (`npm install`, `npm run`, `yarn`, `pnpm`)
- Build tools (`go build`, `cargo build`, `make`, `gradle`)
- Test runners (`go test`, `jest`, `pytest`, `cargo test`)
- Any other command that produces output you need to reason about

**RTK is a hook-based proxy — most commands are rewritten automatically.** You do not need to prefix commands with `rtk` for normal usage. The hook intercepts Bash calls and rewrites them transparently.

## Meta Commands

These commands must be called with `rtk` directly (not via the hook):

```bash
rtk gain              # Show token savings analytics for current session
rtk gain --history    # Show command usage history with per-command savings
rtk discover          # Analyze Claude Code history for missed optimization opportunities
rtk proxy <cmd>       # Execute a raw command without RTK filtering (for debugging)
```

## Installation Verification

Run these to confirm RTK is installed and working correctly:

```bash
rtk --version         # Should show: rtk X.Y.Z
rtk gain              # Should show savings (not "command not found")
which rtk             # Should show the binary path (not reachingforthejack/rtk)
```

**Name collision warning:** If `rtk gain` fails, you may have the `reachingforthejack/rtk`
package (Rust Type Kit) installed instead of this token-killer tool. Verify with `which rtk`
and check the binary source.

## Hook-Based Usage

Once installed, RTK hooks into Claude Code's Bash tool. All standard commands are
automatically rewritten:

```
git status       →  rtk git status
npm run build    →  rtk npm run build
go test ./...    →  rtk go test ./...
```

This is fully transparent — you write commands normally, RTK filters the output,
and only the relevant tokens reach Claude's context window. Zero overhead in your prompts.

## Reference Files

| Task | Reference file |
|------|---------------|
| Full command reference, token savings analytics, proxy debugging, hook configuration | `references/commands.md` |

## Common Gotchas

- **`rtk gain` vs hook:** `rtk gain` and other meta commands must use `rtk` directly.
  The hook does NOT wrap meta commands — only standard dev tool commands.
- **Proxy mode for debugging:** If you suspect RTK is filtering too aggressively,
  use `rtk proxy <cmd>` to run the raw command and see unfiltered output.
- **Session vs. history:** `rtk gain` shows current-session savings; `rtk gain --history`
  shows cumulative savings across all sessions.
- **Hook not firing:** If commands aren't being intercepted, verify the Claude Code hook
  is configured in your settings. Run `rtk doctor` if available.
