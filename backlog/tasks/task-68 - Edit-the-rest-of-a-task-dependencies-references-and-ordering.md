---
id: TASK-68
title: 'Edit the rest of a task: dependencies, references and ordering'
status: To Do
assignee: []
created_date: '2026-08-27 00:38'
labels: []
milestone: m-5
dependencies: []
priority: high
type: feature
ordinal: 68000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The panel edits status, priority, assignee, milestone, labels, title, body sections and acceptance criteria. What it cannot touch is everything that relates a task to something else: dependencies, references, documentation links, the parent, and the ordinal that decides manual order.

Dependencies are the one that bites. 317 of 875 live tasks carry them, better than a third, and the analytics screen already reports which tasks are blocked and by what - while the relationship it reports can only be created or removed in a terminal. 1.50.1 added --clear-deps, --clear-refs and --clear-docs, so removal is now as expressible as addition.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Dependencies can be added and removed from the panel, against tasks in the same project
- [ ] #2 References, documentation links and modified files can be edited
- [ ] #3 A task's parent can be set and cleared
- [ ] #4 Manual order can be changed, and the board and list agree on the result
- [ ] #5 A dependency that would not resolve is refused with the reason, before it is written
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
- [ ] #5 Linters and formatters pass across Go and frontend
- [ ] #6 Automated tests cover the change and the suite is green
- [ ] #7 User-facing behaviour change is reflected in README or docs
- [ ] #8 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
