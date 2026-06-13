---
phase: 04-distribution-hardening
reviewed: 2026-06-13T00:00:00Z
depth: standard
files_reviewed: 6
files_reviewed_list:
  - internal/version/version.go
  - cmd/root.go
  - cmd/doctor.go
  - .goreleaser.yaml
  - .github/workflows/release.yml
  - scripts/install.sh
findings:
  critical: 3
  warning: 4
  info: 2
  total: 9
status: issues_found
---

# Phase 04: Code Review Report

**Reviewed:** 2026-06-13T00:00:00Z
**Depth:** standard
**Files Reviewed:** 6
**Status:** issues_found

## Summary

Reviewed six files covering the distribution and hardening phase: version metadata, CLI root/doctor commands, GoReleaser config, GitHub Actions release workflow, and the curl-pipe install script.

The install script contains two severe defects that can be exploited or cause silent security failures. The GoReleaser config uses the wrong Homebrew key for a CLI tool, which will produce broken Homebrew installs. The doctor command misclassifies HTTP 4xx responses as passing registry checks. These issues must be resolved before shipping a public release.

---

## Critical Issues

### CR-01: `--ignore-missing` defeats checksum verification entirely

**File:** `scripts/install.sh:72`
**Issue:** `${SHA_CMD} --check --ignore-missing checksums.txt` will exit 0 (success) even when the downloaded binary does not appear in checksums.txt at all. An attacker who replaces the release binary on GitHub (via a compromised release or MITM) without also replacing checksums.txt would pass this check. The `--ignore-missing` flag was designed for verifying a subset of files from a multi-file checksum list — it must not be used here where verifying the specific binary is the entire security contract. If the binary name is not in the list, the command silently passes with no output.

**Fix:**
```sh
# After cd "${TMPDIR}", grep the expected checksum line explicitly:
grep "${FILENAME}" checksums.txt | ${SHA_CMD} --check -
```
This fails loudly (non-zero exit) if the filename is absent from checksums.txt.

---

### CR-02: `set -euo pipefail` in a `#!/usr/bin/env sh` script breaks on dash/ash

**File:** `scripts/install.sh:1,12`
**Issue:** `pipefail` is a bash-only option. The shebang is `#!/usr/bin/env sh`, which resolves to `dash` on Debian/Ubuntu (and many Docker base images). When `dash` sources `set -euo pipefail`, it exits immediately with `Illegal option - o pipefail`, killing the install on those systems. The script is advertised as `sh -c "$(curl ...)"` in its own header comment, meaning most Linux CI environments will hit this.

**Fix:** Either change the shebang to `#!/usr/bin/env bash` and require bash explicitly in the usage comment, or remove `pipefail` and replace any pipes that need failure propagation with explicit checks:
```sh
#!/usr/bin/env bash
set -euo pipefail
```

---

### CR-03: Hardcoded `VERSION=0.1.0` makes install.sh perpetually stale

**File:** `scripts/install.sh:14`
**Issue:** `VERSION="${AGENTKIT_VERSION:-0.1.0}"` defaults to a hardcoded version literal that will never be updated automatically. After v0.2.0 ships, every user who runs the script without setting `AGENTKIT_VERSION` installs v0.1.0. This is a correctness defect (wrong software installed) and a potential security issue (users believe they have the latest but have an older, potentially vulnerable version).

**Fix:** Resolve the latest version from GitHub at install time:
```sh
LATEST_URL="https://api.github.com/repos/ejyle/agentkit/releases/latest"
VERSION="${AGENTKIT_VERSION:-$(curl -fsSL "${LATEST_URL}" | grep '"tag_name"' | sed 's/.*"v\([^"]*\)".*/\1/')}"
```
Or pin the script in the release process by having GoReleaser template the default version into a versioned install script per release.

---

## Warnings

### WR-01: HTTP 4xx responses treated as "pass" in registry reachability check

**File:** `cmd/doctor.go:144-151`
**Issue:** `checkRegistryReachable` only fails on `resp.StatusCode >= 500`. A 404 (registry URL moved/wrong), 403 (access denied), 401 (auth required), or 429 (rate limited) all return `Status: "pass"`. A user whose environment cannot reach the registry would see a green check. The check label says "reachable" but it actually only means "server responded with anything other than a 5xx."

**Fix:**
```go
if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    return CheckResult{
        Label:   "registry reachable (agentkit-registry)",
        Status:  "fail",
        Message: fmt.Sprintf("HTTP %d", resp.StatusCode),
        Hint:    "Check network connectivity or registry URL",
    }
}
```

---

### WR-02: Wrong GoReleaser key for a CLI Homebrew formula

**File:** `.goreleaser.yaml:59`
**Issue:** `homebrew_casks:` is the GoReleaser key for publishing a Homebrew Cask (macOS GUI application bundle). A CLI binary must use `brews:` (Homebrew Formula). Using `homebrew_casks` will create a Cask formula that Homebrew treats as a GUI app — `brew install ejyle/agentkit/agentkit` will not work as expected; Homebrew may refuse or install to the wrong location.

**Fix:**
```yaml
brews:
  - repository:
      owner: ejyle
      name: homebrew-agentkit
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    homepage: https://github.com/ejyle/agentkit
    description: "AI agent skill and MCP server manager"
    license: MIT
    commit_author:
      name: goreleaserbot
      email: bot@goreleaser.com
```

---

### WR-03: `tar` extraction assumes binary is at archive root

**File:** `scripts/install.sh:75`
**Issue:** `tar -xzf "${TMPDIR}/${FILENAME}" -C "${TMPDIR}" agentkit` extracts only a file named `agentkit` at the top level of the archive. GoReleaser's default archive layout places binaries inside a directory like `agentkit_0.1.0_linux_amd64/agentkit`. The extraction will silently fail to find `agentkit` at the root level, `tar` exits non-zero, and the script aborts — but the error message is confusing. On some tar implementations the command exits 0 even when the file is not found, silently producing no binary, then `mv` fails with "No such file or directory."

**Fix:** Use `--strip-components=1` to unwrap the top-level directory:
```sh
tar -xzf "${TMPDIR}/${FILENAME}" -C "${TMPDIR}" --strip-components=1
```
Or explicitly test for the extracted binary before moving:
```sh
tar -xzf "${TMPDIR}/${FILENAME}" -C "${TMPDIR}"
BINARY=$(find "${TMPDIR}" -name "agentkit" -type f | head -1)
[ -z "${BINARY}" ] && { printf 'Binary not found in archive\n' >&2; exit 1; }
mv "${BINARY}" "${INSTALL_DIR}/agentkit"
```

---

### WR-04: Undocumented `"pi"` target in validTargets

**File:** `cmd/root.go:18`
**Issue:** `validTargets` includes `"pi"` but the flag help text (line 30) only lists `claude|copilot-cli|copilot-vscode|codex|gemini|opencode|pi` — it does appear in help, but `"pi"` has no corresponding config path, install logic, or documentation anywhere in the codebase or CLAUDE.md. Accepting an undocumented target silently passes validation and produces undefined behavior downstream when install/config commands attempt to resolve its paths.

**Fix:** Either remove `"pi"` from `validTargets` and the flag description until it is fully implemented, or add a guard in any command that consumes the target flag to return an explicit "not yet supported" error for `"pi"`.

---

## Info

### IN-01: Write-test file removal error silently discarded

**File:** `cmd/doctor.go:109`
**Issue:** `_ = os.Remove(testFile)` discards the error from removing the sentinel file. If removal fails (permissions, locked file), the `.write-test` artifact persists in `~/.agentkit/`. Subsequent doctor runs would succeed (the file exists and was writable in a past run) but the stale file is misleading. This is a minor reliability issue.

**Fix:**
```go
if err := os.Remove(testFile); err != nil {
    // non-fatal: log as warning but do not fail the writable check
    fmt.Fprintf(os.Stderr, "warning: could not remove test file %s: %v\n", testFile, err)
}
```

---

### IN-02: `PersistentPreRunE` bypass on `doctorCmd` will silently propagate to future subcommands

**File:** `cmd/doctor.go:32-35`
**Issue:** The bypass is set as `doctorCmd.PersistentPreRunE`, meaning any subcommand added under `doctor` in the future will also skip `--target` validation without any explicit opt-in. This is a latent footgun — a future developer adding `agentkit doctor network` would inherit the bypass without realizing it.

**Fix:** Use `PreRunE` (non-persistent) instead of `PersistentPreRunE` so the bypass applies only to the `doctor` command itself:
```go
doctorCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
    return nil
}
```

---

_Reviewed: 2026-06-13T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
