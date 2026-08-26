---
id: TASK-22
title: Add board filters and saved view state
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 15:26'
labels: []
milestone: m-2
dependencies:
  - TASK-21
priority: medium
type: feature
ordinal: 22000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Specification section 6: filters by milestone, priority, type and label. Without them the board shows 162 To Do items at once, which is the dump problem that killed the Obsidian option. Filter state must survive a restart so the working view is not rebuilt every morning.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Board filters by milestone, priority, type, label and project
- [ ] #2 Filters combine rather than replace one another
- [ ] #3 Active filters are visible and clearable in one action
- [ ] #4 Filter state persists across application restart
- [ ] #5 Text search matches task title and description
<!-- AC:END -->
