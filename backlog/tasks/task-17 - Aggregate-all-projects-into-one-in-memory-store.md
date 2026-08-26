---
id: TASK-17
title: Aggregate all projects into one in-memory store
status: To Do
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-1
dependencies:
  - TASK-16
  - TASK-15
priority: high
type: feature
ordinal: 17000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The multi-repo view needs one queryable collection across every registered project, held in memory - 640 files are milliseconds to load and no database is warranted. Task identity must be qualified by project, because task IDs collide across repositories.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Store holds tasks from every registered project keyed by a project-qualified identity
- [ ] #2 Store answers queries by status, project, milestone, priority, type and label
- [ ] #3 Cross-project ID collisions are handled without shadowing
- [ ] #4 Reload of a single project updates the store without rescanning the others
- [ ] #5 Loading the full corpus at startup stays within a documented time budget
<!-- AC:END -->
