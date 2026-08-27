---
id: TASK-70
title: Manage milestones from the application
status: To Do
assignee: []
created_date: '2026-08-27 00:38'
labels: []
milestone: m-5
dependencies: []
priority: medium
type: feature
ordinal: 70000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Milestones are the axis this backlog is planned on - 49 across the nine projects, the board groups by them, cards and the projects screen show their progress - and there is no way to make one, rename one or retire one without a terminal.

The CLI has add, rename, remove and archive. rename updates the tasks that point at the milestone by default, and remove asks what should happen to them: clear, keep or reassign. Both of those choices have to be surfaced rather than decided quietly, because they rewrite other files.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A milestone can be added to a project from the application
- [ ] #2 A milestone can be renamed, and what happens to the tasks pointing at it is shown before it is done
- [ ] #3 A milestone can be archived or removed, with the choice of what becomes of its tasks made explicitly
- [ ] #4 Milestone progress on the board and the projects screen follows immediately
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
