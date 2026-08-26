---
id: TASK-50
title: Build the sortable task list view
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:27'
updated_date: '2026-08-26 20:47'
labels: []
milestone: m-2
dependencies:
  - TASK-20
  - TASK-24
priority: high
type: feature
ordinal: 50000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A table view over every project: the mode for scanning many tasks at once, where a board is the mode for moving a few. The VSCode extension proves the shape - sortable, filterable, with column customization - and this is the cross-project version of it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 List shows tasks from every registered project in one table
- [x] #2 Columns cover project, ID, title, status, priority, type, milestone, assignee, labels and updated date
- [x] #3 Any column can be sorted and the visible column set is customisable and persisted
- [x] #4 Rows can be edited inline for status and priority through the CLI adapter
- [x] #5 Selecting a row opens the task panel
- [ ] #6 The view stays responsive with the full corpus loaded
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The task list: a sortable, filterable table across every project, opened with l.

Ten columns covering project, id, title, status, priority, type, milestone, assignee, labels and updated date; any column sorts and reverses; the visible column set is chosen from a picker and remembered across restarts. A row opens the same task panel the board does, and the same filters apply.

Verified against the nine real projects: 890 rows render, sorting by title reverses on a second click, and the column picker adds and removes columns. Storage failures are caught rather than left to break the screen - a browser refusing local storage is not a reason to show no columns.
<!-- SECTION:FINAL_SUMMARY:END -->
