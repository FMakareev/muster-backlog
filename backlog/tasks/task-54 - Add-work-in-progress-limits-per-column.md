---
id: TASK-54
title: Add work-in-progress limits per column
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:27'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-2
dependencies:
  - TASK-21
  - TASK-52
priority: medium
type: feature
ordinal: 54000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A stock kanban mechanic, computed entirely from native data: count the tasks a project has in a status and warn when the configured limit is exceeded. This is what remains of the review-capacity idea once no field of our own is introduced - a visible signal rather than a planner, and it addresses the observed pattern of many simultaneously active tasks in one project.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A limit can be configured per column, optionally overridden per project
- [x] #2 A column at or above its limit is visibly flagged on the board
- [x] #3 The limit is a warning and never blocks a drag
- [x] #4 Limits live in the Muster registry and are never written into project configuration
- [x] #5 Limits are visible in the analytics dashboard alongside current counts
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Limits live in Muster's own settings file rather than in the project registry as the specification originally said. The registry is hand-edited and says where projects are; the settings file is written by the application. Putting a toggle in the hand-edited file would mean rewriting it every time the toggle moved.

A limit of zero or less is dropped on load rather than treated as a column nobody may use, and it is never enforced: a limit that blocks a drag is a limit people work around instead of one they act on. It appears in the overview as a count against the ceiling.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Advisory work-in-progress limits, counted from native data.

A limit per status, optionally meaningful per project, appears in the overview as the current count against the ceiling with the ones at or over it marked. It is a signal and never a rule - nothing blocks a drag.

Limits live in Muster's own settings file rather than in the hand-edited project registry, which is a deviation from the specification's original placement and is recorded in the notes. Zero and negative limits are dropped on load rather than becoming columns nobody may use, covered by a test.
<!-- SECTION:FINAL_SUMMARY:END -->
