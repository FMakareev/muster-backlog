---
id: TASK-50
title: Build the sortable task list view
status: To Do
assignee: []
created_date: '2026-08-26 15:27'
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
- [ ] #1 List shows tasks from every registered project in one table
- [ ] #2 Columns cover project, ID, title, status, priority, type, milestone, assignee, labels and updated date
- [ ] #3 Any column can be sorted and the visible column set is customisable and persisted
- [ ] #4 Rows can be edited inline for status and priority through the CLI adapter
- [ ] #5 Selecting a row opens the task panel
- [ ] #6 The view stays responsive with the full corpus loaded
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
