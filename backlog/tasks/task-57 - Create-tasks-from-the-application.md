---
id: TASK-57
title: Create tasks from the application
status: To Do
assignee: []
created_date: '2026-08-26 19:43'
labels: []
milestone: m-2
dependencies:
  - TASK-24
priority: high
type: feature
ordinal: 57000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A task manager that cannot create a task is a viewer. Today every task has to be created in a terminal, which breaks the loop the application exists to close: see the board, decide what is missing, add it. Creation goes through the backlog CLI like every other write, and the form offers what that project configures rather than a fixed list.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A task can be created from the board and from the project roll, into a chosen project
- [ ] #2 The form covers title, description, status, priority, type, milestone, assignee, labels and acceptance criteria
- [ ] #3 Status, priority and type choices come from the target project's own configuration
- [ ] #4 The task is written by the backlog CLI and appears on the board once the rescan confirms it
- [ ] #5 A failed creation explains why and loses nothing the person typed
- [ ] #6 The form opens and submits by keyboard
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
