---
id: TASK-88
title: Stop a test from depending on Claude Code being installed
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 15:43'
updated_date: '2026-08-28 15:47'
labels: []
dependencies: []
priority: medium
type: bug
ordinal: 88000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TestThePlanSaysWhereTheCommandWillRun asks for a plan for the claude-code client, which is a 'cli' provider: planning it looks for the `claude` binary and fails when there is none. The test passes on a machine that has Claude Code installed and fails everywhere else, CI included, with 'claude was not found'.

What the test is actually about is that the plan says which environment the registered command will run in, alongside the client's own note. Whether the machine has that client installed is beside the point.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The test passes on a machine with no agent client installed, and still fails if the plan stops naming the environment
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Confirm the dependency: whichbin.Look for the provider's bin is what fails.
2. Put a stand-in for the client on PATH inside the test so planning gets past the lookup, without changing what is asserted.
3. Verify by running the test with a PATH that has no real client on it.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reproduced the CI failure locally by running the test with a PATH holding no claude and a HOME that whichbin's install-directory search finds nothing under — this machine keeps claude in ~/.local/bin, which is one of the directories that search covers, so stripping PATH alone was not enough to reproduce it. In that environment the test fails with exactly the CI message, and passes with the stand-in.

PATH is replaced rather than prepended so a machine that does have the client cannot be the reason it passes. Incidentally the test also got faster: without a client to find, planning was spending 0.66s falling through to the login-shell lookup.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
TestThePlanSaysWhereTheCommandWillRun planned for a cli-kind client, so it looked for the claude binary and failed on any machine without Claude Code installed — CI included, with 'claude was not found', while the behaviour under test was fine. The test now writes a stand-in for the client into a temporary directory and sets PATH to that directory alone. Verified by reproducing the CI failure locally with a PATH and HOME that hide every real client, then passing with the stand-in in the same environment.
<!-- SECTION:FINAL_SUMMARY:END -->
