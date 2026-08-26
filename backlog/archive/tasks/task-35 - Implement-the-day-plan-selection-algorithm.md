---
id: TASK-35
title: Implement the day plan selection algorithm
status: To Do
assignee: []
created_date: '2026-08-26 15:02'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-2
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 35000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The muster button proposes a pool from active milestones under three constraints at once: unmet dependencies exclude a task, a project at or above its review limit is excluded entirely, and the total review cost must fit the daily budget. Active milestone plus budget together are what shrink the visible surface from 162 tasks to 20 to 25.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Candidates are drawn only from the active milestone of each non-frozen project
- [ ] #2 A task with unmet dependencies is never proposed
- [ ] #3 A project at or above its review limit contributes nothing to the proposal
- [ ] #4 The total cost of a proposal never exceeds the daily budget
- [ ] #5 Personal tasks are charged to the personal quota and not to the review budget
- [ ] #6 Selection favours priority and ordinal in a documented, deterministic order
- [ ] #7 The algorithm is unit tested against the real corpus shape, including the case where nothing is eligible
<!-- AC:END -->
