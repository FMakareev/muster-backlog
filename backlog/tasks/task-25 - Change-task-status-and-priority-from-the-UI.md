---
id: TASK-25
title: Move tasks between statuses by drag and drop
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 15:26'
labels: []
milestone: m-1
dependencies:
  - TASK-24
  - TASK-23
priority: high
type: feature
ordinal: 25000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Drag-and-drop status change is the single interaction that makes a board feel like a board. The drop is validated against the target project status list, then written through the backlog CLI; the card settles where the rescan confirms it landed, not where it was dropped.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Dragging a card between columns changes its status through the backlog CLI
- [ ] #2 A drop into a status the card project does not declare is rejected during the drag, not after it
- [ ] #3 The board settles on the result of the rescan rather than on the dropped position
- [ ] #4 A failed write returns the card to its original column and shows why
- [ ] #5 Priority, assignee and labels can be set from the card and from the task panel
- [ ] #6 Status can also be changed by keyboard without dragging
<!-- AC:END -->
