---
id: TASK-61
title: Show subtasks on the board
status: To Do
assignee: []
created_date: '2026-08-26 19:43'
labels: []
milestone: m-2
dependencies: []
priority: low
type: enhancement
ordinal: 61000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Backlog.md tasks can have subtasks through parent_task_id, and 91 files in the reference corpus use them. On the board a subtask is an ordinary card with a dotted id, so the relationship is invisible and a parent looks no different from anything else. Worth showing, but not at the cost of making the board harder to read - a nested board is a worse board.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A card shows whether it has subtasks and how many are done
- [ ] #2 A subtask card shows which task it belongs to
- [ ] #3 Opening a parent lists its subtasks as links; opening a subtask links back
- [ ] #4 The board remains readable at the full corpus size with this on
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
