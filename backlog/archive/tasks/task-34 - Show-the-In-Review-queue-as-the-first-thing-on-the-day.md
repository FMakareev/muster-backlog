---
id: TASK-34
title: Show the In Review queue as the first thing on the day
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
ordinal: 34000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Specification section 6: the In Review queue sits above the day plan and is cleared before anything new is picked. This is the mechanic that answers the observed symptom of thirteen simultaneous In Progress tasks, six of them in one project.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Today screen shows the In Review queue across every project above the day plan
- [ ] #2 Each entry shows project, task, review cost and how long it has been waiting
- [ ] #3 A project at or above its review limit is visibly flagged as blocked
- [ ] #4 A queue item can be moved to Done or back to In Progress from the queue
- [ ] #5 The queue updates live as agents move tasks into review
<!-- AC:END -->
