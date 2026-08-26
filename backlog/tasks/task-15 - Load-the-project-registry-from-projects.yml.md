---
id: TASK-15
title: Load the project registry
status: To Do
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 15:26'
labels: []
milestone: m-1
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The registry is the only configuration Muster owns: which folders hold Backlog.md projects, and how they are displayed. It holds no task data and no settings that belong to Backlog.md - each project keeps its own config.yml as the authority on statuses, priorities, types and labels.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Registry is read from the XDG config location with a documented fallback
- [ ] #2 Each entry carries a folder path and an optional display name and colour
- [ ] #3 Per-project Backlog.md configuration is read from that project, never duplicated into the registry
- [ ] #4 A malformed or missing registry produces an actionable message instead of a crash
- [ ] #5 A path that is not a Backlog.md project is reported as such and does not block the other projects
- [ ] #6 Registry ordering determines project ordering in the UI
<!-- AC:END -->
