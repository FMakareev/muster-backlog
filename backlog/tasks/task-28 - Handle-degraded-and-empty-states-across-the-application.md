---
id: TASK-28
title: Handle degraded and empty states across the application
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 18:59'
labels: []
milestone: m-1
dependencies:
  - TASK-21
  - TASK-24
priority: medium
type: enhancement
ordinal: 28000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The application points at directories it does not own: paths move, repositories get cleaned, the CLI may be missing, a project may have no backlog at all. Each of these must produce a readable state rather than a blank screen or a crash, because they will happen routinely on a real machine.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 An empty registry shows guidance on adding the first project
- [x] #2 A project path that does not exist or has no backlog directory is shown as degraded, not dropped silently
- [x] #3 A missing or unsupported backlog CLI is reported with what to install
- [x] #4 One broken project never prevents the others from loading
- [x] #5 Parser and watcher diagnostics are reachable from the UI rather than only in logs
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. The application points at folders it does not own. Paths move, repositories get cleaned, the CLI may be missing, a folder may never have been initialised. Each of those has to produce something readable rather than a blank screen, because all of them will happen on a real machine.
2. Diagnostics are currently one banner showing the first problem and a count in the status strip. Add a panel that lists every one, opened from that count, so a skipped file is reachable rather than buried.
3. A degraded project stays in the roll, struck through and without a count, with the reason on hover - a project silently missing from the board is worse than one that says why it is broken.
4. Verify each state by constructing it: an empty registry, a path that does not exist, a folder with no backlog, a project among healthy ones, and a missing CLI.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Added a problems panel opened from the count in the status strip, because a count nobody can expand is a number nobody can act on. Each entry is typed - Project, Skipped file, Backlog.md CLI, Registry - and carries the reason and the path.

A degraded project stays in the roll, struck through and showing a dash instead of a count, with the reason on hover. Dropping it would be worse: a project silently missing from a board whose whole purpose is showing every project is the kind of absence nobody notices.

Verified by constructing every state rather than reasoning about it. A registry naming one healthy project, one folder that was never initialised, one path that no longer exists, and a stray README inside the healthy project's tasks directory produced: the healthy project loading its task normally, both broken ones struck through, and the problems panel listing each exactly once. Running the same binary with the CLI absent from PATH added a fourth problem, 'Changes cannot be saved - the backlog CLI was not found: install it and make sure backlog is on PATH', and attempting to move a card repeated that same reason rather than failing silently.

That verification found a real defect: every broken project was being reported twice, once correctly as a project failure and once as a skipped file, because the store folded project-level errors into the parser's diagnostics. One broken folder looked like two problems and the second was mislabelled. Diagnostics now carry only files that were actually skipped; a covering test asserts a failed project never appears among them.

Also corrected the status strip saying '1 tasks · 1 projects'.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Every way the application can be pointed at something broken now produces a readable state instead of a blank screen.

A missing registry offers onboarding rather than an error; a folder that was never initialised or has moved away stays visible in the project roll, struck through and with its reason; a missing or too-old backlog CLI is one standing report plus the same reason whenever a write is attempted; one broken project never stops the others; and a problems panel opened from the status strip lists everything, typed and with paths, so a skipped file is reachable rather than buried in a count.

Verified by constructing each state and driving the running application, not by reading code: a registry with one healthy project, one uninitialised folder, one vanished path and a stray README produced exactly three problems and a working board, and the same binary run without the CLI on PATH added the fourth and refused the write with that reason.

Verification caught a real defect: the store folded project-level failures into its parser diagnostics, so each broken folder was reported twice and the duplicate was mislabelled as a skipped file. Fixed, with a test that asserts it stays fixed.
<!-- SECTION:FINAL_SUMMARY:END -->
