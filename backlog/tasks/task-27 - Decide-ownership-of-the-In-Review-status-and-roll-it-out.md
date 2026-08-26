---
id: TASK-27
title: Reconcile differing status sets into unified board columns
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 15:26'
labels: []
milestone: m-1
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 27000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Statuses are per-project configuration in Backlog.md, and registered projects will not agree on them. The board therefore derives its columns from the union of every project status list rather than imposing one. Muster never edits another project config.yml to make the view simpler.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Board columns are the union of the status lists declared by the registered projects
- [ ] #2 Column ordering is derived deterministically and documented, including how conflicting orders are resolved
- [ ] #3 A project that does not declare a status shows as empty in that column rather than as an error
- [ ] #4 A card cannot be dragged into a status its own project does not declare, and the reason is shown
- [ ] #5 Adding or removing a status in a project config is picked up without restarting the application
- [ ] #6 No code path writes to another project configuration file
<!-- AC:END -->
