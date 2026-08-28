---
id: TASK-85
title: Declare lefthook so a fresh clone can install the workspace
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 14:19'
updated_date: '2026-08-28 14:23'
labels: []
dependencies: []
priority: high
type: bug
ordinal: 85000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`pnpm install` fails on any machine that does not already have lefthook: the root package.json has `prepare: lefthook install` but nothing declares lefthook, so pnpm runs a program the manifest never installs. It works on the author's machine only because a global lefthook sits on PATH. CI has none, so the workspace step dies with 'lefthook: not found' — and so would a first-time contributor following the README, which claims the hooks are 'installed by pnpm install'.

Declaring it as a devDependency is not quite enough on its own: pnpm 11 blocks a dependency's build scripts and exits non-zero with ERR_PNPM_IGNORED_BUILDS unless the manifest says what to do about them.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `pnpm install` succeeds and installs the git hooks on a machine with no lefthook on PATH, using only what the manifest declares
- [x] #2 The lefthook version is pinned, and the pin is old enough to satisfy the repository's minimum-release-age policy
- [x] #3 pnpm's decision about lefthook's build script is recorded in the manifest with its reason, so the install exits zero
- [x] #4 A test fails if a lifecycle script invokes a program the workspace does not install
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Reproduce in a scratch workspace: prepare script invoking an undeclared binary, and pnpm 11's ERR_PNPM_IGNORED_BUILDS on a declared one.
2. Add lefthook 2.1.10 to the root devDependencies — the version already in use, published 2026-07-08 and so past the 7-day minimum-release-age floor.
3. Record the build-script decision in pnpm-workspace.yaml under allowBuilds, declining it: lefthook's postinstall only runs `lefthook install -f`, which the prepare script already does, and it skips itself under CI anyway.
4. Refresh the lockfile.
5. Add a test asserting every root lifecycle script names a program the workspace installs.
6. Verify by installing with node_modules removed and confirming the hooks are written by the local binary rather than a global one.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Root cause: the root package.json declared no lefthook while its prepare script ran one. It worked here only because /usr/bin/lefthook exists on this machine — a coincidence, not a dependency.

Reproduced both states in a fresh clone installed with a PATH built from symlinks that contains no lefthook at all. Before the change: '. prepare: sh: 1: lefthook: not found / [ELIFECYCLE] Command failed.', the exact CI error. After: exit 0, and 'sync hooks: pre-commit, commit-msg, pre-push'.

Pinned 2.1.10 — the version already installed here, published 2026-07-08, so it clears the 7-day minimum-release-age floor in .npmrc. 2.1.12 was published today and would be refused by that policy.

Declaring it is not sufficient on its own: pnpm 11 blocks a dependency's install scripts and exits 1 with ERR_PNPM_IGNORED_BUILDS until the manifest decides. Measured where that decision lives — the pnpm field in package.json is no longer read ('The "pnpm" field in package.json is no longer read by pnpm'), and neither onlyBuiltDependencies nor ignored-built-dependencies in .npmrc had any effect. `pnpm approve-builds` writes 'allowBuilds: {pkg}: true|false' into pnpm-workspace.yaml, which is the current form.

Declined rather than allowed. lefthook's postinstall is only 'lefthook install -f' — read it — which the prepare script already does explicitly, and it skips itself when CI is set. Nothing is lost and one fewer dependency script runs on every install.

Not taken: installing lefthook globally in the workflow, which was the suggested fix. It would have made CI green while a fresh clone stayed broken, and CI is the only thing that noticed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
pnpm install ran 'lefthook install' from a prepare script while nothing declared lefthook, so it succeeded only on a machine that already had one — CI and every fresh clone died at the install step with 'lefthook: not found'. Declared lefthook 2.1.10 as a workspace devDependency and recorded pnpm's build-script decision for it in pnpm-workspace.yaml (declined, with the reason), which takes the install back to exit 0. Verified by cloning the repository and installing it with a PATH containing no lefthook: before, the exact CI failure; after, exit 0 with all three hooks synced. TestEveryScriptRunsSomethingTheWorkspaceInstalls fails if a lifecycle script names a program the workspace does not install, fault-injected by removing the binary. wails3 task lint: 0 issues; full suite green with MUSTER_REQUIRE_BACKLOG_CLI=1.
<!-- SECTION:FINAL_SUMMARY:END -->
