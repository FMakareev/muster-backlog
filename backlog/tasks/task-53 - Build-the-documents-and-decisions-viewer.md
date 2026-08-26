---
id: TASK-53
title: Build the documents and decisions viewer
status: To Do
assignee: []
created_date: '2026-08-26 15:27'
updated_date: '2026-08-26 15:29'
labels: []
milestone: m-2
dependencies:
  - TASK-16
  - TASK-20
  - TASK-56
priority: medium
type: feature
ordinal: 53000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Backlog.md projects carry documents and decision records next to tasks, and they are read far more often than they are written. Rendering them with the same markdown pipeline as task bodies, including Mermaid diagrams, makes the application the single place a project is read from.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Documents and decisions of every registered project are browsable in one tree
- [ ] #2 Markdown renders with the same pipeline as task bodies, including Mermaid diagrams
- [ ] #3 Links between documents and to tasks resolve and navigate within the application
- [ ] #4 The viewer reflects file changes live
- [ ] #5 Documents are reachable from search results
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
