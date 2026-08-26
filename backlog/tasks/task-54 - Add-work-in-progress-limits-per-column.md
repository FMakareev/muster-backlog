---
id: TASK-54
title: Add work-in-progress limits per column
status: To Do
assignee: []
created_date: '2026-08-26 15:27'
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
- [ ] #1 A limit can be configured per column, optionally overridden per project
- [ ] #2 A column at or above its limit is visibly flagged on the board
- [ ] #3 The limit is a warning and never blocks a drag
- [ ] #4 Limits live in the Muster registry and are never written into project configuration
- [ ] #5 Limits are visible in the analytics dashboard alongside current counts
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
