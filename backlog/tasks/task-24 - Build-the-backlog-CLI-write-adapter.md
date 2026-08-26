---
id: TASK-24
title: Build the backlog CLI write adapter
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 17:15'
labels: []
milestone: m-1
dependencies:
  - TASK-17
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 24000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Specification section 5: every write goes through the backlog CLI, never through a writer of our own, because the file format is alive and would have to be chased at each CLI release. This adapter is the single choke point - it locates the binary, runs commands in the right working directory, and turns failures into something the UI can show.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Adapter resolves the backlog binary and verifies a supported version at startup
- [x] #2 Commands run with the target project as the working directory
- [x] #3 Non-zero exits and stderr are captured and surfaced as structured errors
- [x] #4 Arguments are passed without shell interpolation
- [x] #5 A missing or unsupported CLI is reported once, clearly, instead of failing per action
- [x] #6 Every write is followed by a rescan so the store reflects the CLI result rather than an assumption
- [x] #7 Concurrent writes to the same project are serialised
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/backlogcli is the single choke point for every write. Nothing else in the application may touch a task file, so the whole no-second-writer rule lives or dies here.
2. Resolve the binary once at startup and check its version, so a missing or too-old CLI is one clear report rather than a failure per action.
3. Run through exec with an argument slice - never a shell string. Task titles in this corpus contain quotes, backticks, semicolons and Cyrillic; a shell would be a command injection waiting for the right title.
4. Serialise writes per project with a mutex keyed on the project path. Two concurrent edits in one project can collide over ids and ordinals; different projects have no reason to wait for each other.
5. Capture stdout, stderr and the exit code into a structured error carrying the command, so the UI can show what was run and what it said rather than a bare string.
6. Every write ends with a rescan of that project, so the store reflects what the CLI actually did rather than what the UI assumed. This is what lets the board settle on the result instead of on the dropped position.
7. Test against the real CLI on scratch projects, not a mock: the point of this adapter is that the CLI owns the format, so a mock would be testing my own assumptions.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
internal/backlogcli is the single choke point. Nothing else in the application writes anything, which is what makes the no-second-writer rule enforceable rather than aspirational.

Arguments go to the process as a slice, never assembled into a shell string. This is not theoretical hygiene: titles in the author's own projects contain quotes, backticks, semicolons and dollar signs. A test creates tasks titled 'Fix $(touch /path/canary) in the parser', 'Handle `rm -rf /` safely' and 'Semicolons; ampersands & pipes | all of it', then asserts the canary file was never created and that every title survived the round trip byte for byte, Cyrillic included.

Writes are serialised per project with a mutex keyed on the path. Two edits in one project can collide over ids and ordinals; two different projects have no reason to wait for each other. Verified by creating eight tasks concurrently and asserting all eight exist with distinct ids - a lost update would show up as two tasks claiming one id.

Every write is followed by a rescan of that project before the result is returned, so the store reflects what the files say rather than what the caller asked for. The test asserts the new status is visible immediately after the call with no sleeping and no waiting on the watcher, and separately that a failed write leaves the store showing the untouched task - which is what will let the board return a card to where it was.

Version checking compares numerically, so 1.48.0 is newer than 1.9.0 rather than older as a string comparison would have it. A binary that runs but reports 0.1.0 is rejected with a message naming the version needed.

Found while testing: 'backlog init' refuses --agent-instructions when --integration-mode is none. That is a real constraint of 1.48.0, so Init now catches the combination before running and returns a typed error, which lets the form in TASK-55 prevent it rather than show a failure afterwards.

Tests run the real CLI against scratch projects rather than a mock. The premise of this package is that the CLI owns the format, so a mock would only be testing my own assumptions about it. Coverage 87% of the adapter.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added internal/backlogcli, the only thing in Muster that writes, and wired it into the service so every write is followed by a re-read.

The runner resolves the binary once at startup and checks its version numerically, runs commands with the project as the working directory, passes arguments straight to the process, serialises writes per project, and turns failures into a CommandError carrying the command, the exit code and what the CLI said. On top of it sit the operations the board needs: status, priority, assignee, milestone, labels, acceptance criteria, draft capture and promotion, and a non-interactive init. The service exposes them as writes that reload the project and emit the change event before returning.

Verified against the real CLI on scratch projects rather than a mock - 12 adapter tests and 5 service tests, 87% and 75% coverage. The ones that matter: five titles containing command substitution, backticks, semicolons, pipes and Cyrillic all survive the round trip unchanged and the canary file the injected command would have created never appears; eight concurrent creates in one project produce eight distinct ids; a status change is visible in the store immediately after the call with no sleeping, because the write re-read the files; and a write against a task that does not exist returns a renderable problem naming TASK-404 while leaving the real task untouched.

Testing found a real constraint of the CLI: init refuses agent instructions when AI integration is set to none. Init now catches that before running, so the Projects screen can prevent the combination rather than surface a failure afterwards.
<!-- SECTION:FINAL_SUMMARY:END -->
