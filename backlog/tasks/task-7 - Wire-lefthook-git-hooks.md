---
id: TASK-7
title: Wire lefthook git hooks
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 14:59'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-0
dependencies:
  - TASK-5
  - TASK-6
priority: high
type: chore
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Hooks catch mistakes before they reach CI or history. lefthook runs the linters from the previous task on staged files and validates the commit message. Hooks must stay fast enough that nobody reaches for --no-verify.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 lefthook.yml is checked in and installs via a documented command
- [x] #2 pre-commit runs format and lint on staged files only
- [x] #3 commit-msg validates the message against the commit convention
- [x] #4 pre-push runs the test suite
- [x] #5 Hook installation is documented in CONTRIBUTING and happens on dependency install where possible
- [x] #6 Full pre-commit run on a typical change stays under a documented time budget
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Write lefthook.yml with three hooks: pre-commit running formatters and linters on staged files only, commit-msg running commitlint, pre-push running the test suite.
2. Keep pre-commit genuinely staged-file scoped - golangci-lint on the changed packages and eslint/prettier on the changed files - so the hook cost scales with the change, not with the repository. A hook slow enough to provoke --no-verify is a hook that does not run.
3. Install lefthook on dependency install via a prepare script so a fresh clone gets hooks without a separate documented step, and document the manual command as well.
4. Measure a real pre-commit run and record the figure, since the acceptance criterion is a documented time budget rather than a vague claim.
5. Verify each hook fires: a badly formatted file blocked at commit, a malformed message blocked at commit-msg, and the push hook running the Go tests.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
lefthook 2.1.10, already present on this machine. Three hooks: pre-commit (parallel gofmt, golangci-lint, prettier, eslint), commit-msg (commitlint), pre-push (go test, svelte-check).

Two problems surfaced during verification and were fixed:

1. The root node_modules was not ignored. .gitignore inherited from the Wails template only listed frontend/node_modules, so making the repository a pnpm workspace in TASK-5 quietly created an unignored root node_modules - the first staging attempt picked up 11459 files. .gitignore now ignores node_modules at any depth. This would have been a very unpleasant first public commit.

2. eslint and prettier are dependencies of the frontend workspace package, so 'pnpm exec' from the repository root could not find them. The jobs now use lefthook's per-job root: "frontend/", which both runs in the right place and rewrites {staged_files} relative to it. Added --no-warn-ignored to eslint because the committed Wails bindings match its ignore pattern and produced warning noise on every commit.

golangci-lint runs with --new-from-rev=HEAD rather than on {staged_files}: Go linting is only meaningful per package, and a file-scoped run reports false positives for symbols defined in sibling files. Formatters use stage_fixed: true so a commit is never left half-formatted.

Measured on the current tree with 176 staged files: pre-commit 1.9-2.2s (eslint dominates at ~1.9s), commit-msg 0.7s, pre-push 2.1s. Budget documented in CONTRIBUTING as under 3 seconds for pre-commit.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Wired lefthook so the linters from TASK-6 and the commit convention from TASK-5 are enforced before anything reaches CI or history.

Added lefthook.yml with pre-commit (gofmt, golangci-lint, prettier, eslint - staged files only, formatters restage what they fix), commit-msg (commitlint) and pre-push (go test, svelte-check); added a prepare script so pnpm install installs the hooks; documented installation and the time budget in CONTRIBUTING and README.

Verified each hook by running it rather than by inspecting configuration. pre-commit passes clean (exit 0) and blocks on a real defect - an injected unused Go function produced 'func leftover is unused' and exit 1. commit-msg exits 1 on 'made changes' and 0 on 'chore(release): wire lefthook hooks'. pre-push runs go test and svelte-check to completion, exit 0. Timings measured over 176 staged files: pre-commit 1.9-2.2s, commit-msg 0.7s, pre-push 2.1s, all inside the documented 3-second budget.

Verification also caught a real problem: the workspace change in TASK-5 left the root node_modules unignored, and the first staging attempt swept up 11459 files. .gitignore now ignores node_modules at any depth.
<!-- SECTION:FINAL_SUMMARY:END -->
