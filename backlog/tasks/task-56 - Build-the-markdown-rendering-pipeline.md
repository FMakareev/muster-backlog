---
id: TASK-56
title: Build the markdown rendering pipeline
status: To Do
assignee: []
created_date: '2026-08-26 15:29'
labels: []
milestone: m-1
dependencies:
  - TASK-20
priority: high
type: feature
ordinal: 56000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Task bodies, documents and decisions all render through one pipeline, so it is defined once rather than three times. It must handle what Backlog.md files actually contain: section markers, acceptance-criteria checklists, task cross-references, code blocks and Mermaid diagrams. Rendering is local and offline - no content leaves the machine.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Markdown renders with code highlighting and Mermaid diagram support
- [ ] #2 Task cross-references such as TASK-42 render as links that navigate within the application
- [ ] #3 Acceptance-criteria checklists render as checkable items reflecting their state on disk
- [ ] #4 Backlog.md section markers are handled without leaking into the output
- [ ] #5 Rendering is sandboxed against untrusted markdown and makes no network requests
- [ ] #6 Relative links to files inside a project resolve
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
