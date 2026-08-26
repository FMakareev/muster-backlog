---
id: TASK-37
title: Persist the day plan across restarts
status: To Do
assignee: []
created_date: '2026-08-26 15:02'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-2
dependencies: []
priority: medium
type: feature
ordinal: 37000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A plan that evaporates when the window closes is not a plan. It has to survive restarts, roll over cleanly to a new day, and remain correct when the underlying tasks change status or disappear underneath it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The current plan is restored after an application restart
- [ ] #2 A new day starts with an empty plan while the previous plan remains reviewable
- [ ] #3 A planned task that moves to Done is reflected in the plan rather than shown stale
- [ ] #4 A planned task deleted from disk is removed from the plan with a notice
- [ ] #5 Plan state is stored outside the tracked repositories
<!-- AC:END -->
