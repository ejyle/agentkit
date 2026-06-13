---
status: partial
phase: 03-bundled-skills
source: [03-VERIFICATION.md]
started: 2026-06-09T11:51:15Z
updated: 2026-06-09T11:51:15Z
---

## Current Test

[awaiting live registry — D-17 gate: publish v0.1.0 release to unblock]

## Tests

### 1. agentkit install gsd (ROADMAP criterion 1)
expected: Installs full GSD suite from gsd-core registry in one command
result: [pending — blocked by D-17: gsd-core registry not yet published]

### 2. agentkit install --bundle cloud (ROADMAP criterion 2)
expected: Installs aws, gcp, azure skills atomically from registry
result: [pending — blocked by D-17: ejyle/agentkit-registry not yet published]

### 3. ~/.claude/skills/aws/ populated after install (ROADMAP criterion 4)
expected: SKILL.md + references/ec2.md, s3.md, iam.md exist after install
result: [pending — blocked by D-17: github-release tarball requires published release]

### 4. agentkit install --bundle dev and --bundle context (ROADMAP criterion 2b)
expected: dev bundle installs playwright/github/cicd; context bundle installs context-mode/rtk/serena
result: [pending — blocked by D-17: same as above]

## Summary

total: 4
passed: 0
issues: 0
pending: 4
skipped: 0
blocked: 4

## Gaps

All 4 items are blocked by the D-17 gate (no published GitHub release tarball). This is a documented, expected constraint — not a code defect. The GitHubReleaseInstaller, bundle resolution, and ValidateSkill wiring are all implemented and unit-tested. These items will auto-resolve when Phase 4 publishes v0.1.0 via GoReleaser.
