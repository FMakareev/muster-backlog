---
id: TASK-53
title: Build the documents and decisions viewer
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:27'
updated_date: '2026-08-26 20:48'
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
- [x] #1 Documents and decisions of every registered project are browsable in one tree
- [x] #2 Markdown renders with the same pipeline as task bodies, including Mermaid diagrams
- [x] #3 Links between documents and to tasks resolve and navigate within the application
- [x] #4 The viewer reflects file changes live
- [x] #5 Documents are reachable from search results
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
The documents and decisions viewer, opened with d.

Every document and decision across every project in one tree, grouped and coloured by project, rendered through the same markdown pipeline as task bodies - Mermaid, code highlighting, sanitisation and all. Task references inside them link to the task. The viewer re-reads whenever a project changes, so it is as live as the board.

Verified against the nine real projects: 88 documents and decisions listed, and opening one renders its content.
<!-- SECTION:FINAL_SUMMARY:END -->
