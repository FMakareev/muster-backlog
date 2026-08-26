---
id: TASK-51
title: Build cross-project search
status: To Do
assignee: []
created_date: '2026-08-26 15:27'
labels: []
milestone: m-2
dependencies:
  - TASK-17
  - TASK-20
priority: high
type: feature
ordinal: 51000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
With several projects loaded, finding a task by memory of its title is the most common navigation. Search runs over titles and bodies of tasks, drafts, documents and decisions across every registered project, and is reachable from anywhere by keyboard.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Search matches task titles and bodies across every registered project
- [ ] #2 Documents, decisions and drafts are searchable alongside tasks
- [ ] #3 Results show the owning project and the kind of item found
- [ ] #4 Search opens by keyboard shortcut and selects a result without a mouse
- [ ] #5 Selecting a result opens the corresponding panel or viewer
- [ ] #6 Search over the full corpus returns within a documented time budget
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
