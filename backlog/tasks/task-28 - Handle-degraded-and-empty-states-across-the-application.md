---
id: TASK-28
title: Handle degraded and empty states across the application
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-1
dependencies:
  - TASK-21
  - TASK-24
priority: medium
type: enhancement
ordinal: 28000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The application points at directories it does not own: paths move, repositories get cleaned, the CLI may be missing, a project may have no backlog at all. Each of these must produce a readable state rather than a blank screen or a crash, because they will happen routinely on a real machine.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 An empty registry shows guidance on adding the first project
- [ ] #2 A project path that does not exist or has no backlog directory is shown as degraded, not dropped silently
- [ ] #3 A missing or unsupported backlog CLI is reported with what to install
- [ ] #4 One broken project never prevents the others from loading
- [ ] #5 Parser and watcher diagnostics are reachable from the UI rather than only in logs
<!-- AC:END -->
