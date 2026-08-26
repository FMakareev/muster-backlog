---
id: TASK-32
title: Label active-milestone tasks with review cost
status: To Do
assignee: []
created_date: '2026-08-26 15:02'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-2
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: medium
type: task
ordinal: 32000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Specification section 8, item 3: labelling 162 tasks by hand is not an option. The first pass covers only the active milestones - roughly 20 to 25 tasks - and everything else keeps the default of three points. This is the data without which the planner has nothing to plan against.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every task in an active milestone across the registered projects carries an rv label
- [ ] #2 Labelling is done through the backlog CLI, not by editing files
- [ ] #3 The judgement used to assign each level is written down so it can be applied consistently later
- [ ] #4 Tasks outside active milestones are confirmed to fall back to the default cost
<!-- AC:END -->
