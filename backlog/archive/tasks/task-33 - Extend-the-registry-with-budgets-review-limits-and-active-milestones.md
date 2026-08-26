---
id: TASK-33
title: 'Extend the registry with budgets, review limits and active milestones'
status: To Do
assignee: []
created_date: '2026-08-26 15:02'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-2
dependencies:
  - TASK-15
priority: high
type: feature
ordinal: 33000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The planner needs the settings the MVP parsed but ignored: the daily review budget, the separate personal quota, per-project review limits and the active milestone that decides whether a project participates at all.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Daily review budget and separate personal quota are read and applied
- [ ] #2 Per-project review limit and active milestone drive planner behaviour
- [ ] #3 A project with no active milestone is treated as frozen and excluded from planning
- [ ] #4 Changing a setting takes effect without restarting the application
- [ ] #5 Every setting has a documented default and a documented meaning
<!-- AC:END -->
