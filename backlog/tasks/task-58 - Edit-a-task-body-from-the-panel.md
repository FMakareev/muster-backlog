---
id: TASK-58
title: Edit a task body from the panel
status: To Do
assignee: []
created_date: '2026-08-26 19:43'
labels: []
milestone: m-2
dependencies:
  - TASK-24
  - TASK-23
priority: high
type: feature
ordinal: 58000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The panel shows description, acceptance criteria, implementation plan and notes but cannot change any of them, so any real edit still means opening the file elsewhere. Editing them here is what makes the panel a working surface rather than a reader. Every write goes through the backlog CLI, which owns the section markup.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Description, implementation plan and notes can be edited and saved from the panel
- [ ] #2 Acceptance criteria can be added, removed, reordered and ticked
- [ ] #3 The title can be changed
- [ ] #4 Editing shows the markdown source and saves it verbatim, without reformatting what was there
- [ ] #5 A save that fails leaves the text on screen rather than discarding it
- [ ] #6 An edit made elsewhere while the panel is open does not silently overwrite what is being typed
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
