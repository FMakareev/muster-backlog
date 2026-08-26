---
id: TASK-50
title: Build the sortable task list view
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:27'
updated_date: '2026-08-26 21:59'
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Follow-up from first use: subtasks are now grouped under their parent instead of sorting away from it. The sort still decides the order at the top level and among siblings; a subtask simply follows the task it belongs to, indented and marked. A subtask whose parent is not in the list - filtered out, or archived - keeps its own place at the top level rather than disappearing with it.

Two defects found while verifying this, both fixed. The first was mine: keying the grouping on project plus id folded an archived task into its live namesake, because archiving is a soft delete that lets one id name two different tasks in a project - the table showed 189 rows where the status strip said 190. The key is now the whole ref, and the parent comes from the backend as an exact ref rather than an id to guess a class for. The second was older: the Updated column printed 0001-01-01 for a task with no updated_date, which reads as a date somebody meant; the panel already blanked it and the table now does too.

Verified in the browser across four views: 95 subtask rows all sitting under their own parent, families still together after re-sorting by another column, no row drawn twice, and the row count matching the status strip for every project and for all of them - 902/902, 190/190, 117/117, 165/165.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The task list: a sortable, filterable table across every project, opened with l.

Ten columns covering project, id, title, status, priority, type, milestone, assignee, labels and updated date; any column sorts and reverses; the visible column set is chosen from a picker and remembered across restarts. A row opens the same task panel the board does, and the same filters apply.

Verified against the nine real projects: 890 rows render, sorting by title reverses on a second click, and the column picker adds and removes columns. Storage failures are caught rather than left to break the screen - a browser refusing local storage is not a reason to show no columns.
<!-- SECTION:FINAL_SUMMARY:END -->
