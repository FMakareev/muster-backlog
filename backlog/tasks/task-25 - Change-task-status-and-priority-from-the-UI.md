---
id: TASK-25
title: Move tasks between statuses by drag and drop
status: In Progress
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 17:30'
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
- [x] #1 Dragging a card between columns changes its status through the backlog CLI
- [x] #2 A drop into a status the card project does not declare is rejected during the drag, not after it
- [x] #3 The board settles on the result of the rescan rather than on the dropped position
- [x] #4 A failed write returns the card to its original column and shows why
- [ ] #5 Priority, assignee and labels can be set from the card and from the task panel
- [ ] #6 Status can also be changed by keyboard without dragging
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Drag-and-drop with validation and CLI writes - done alongside TASK-27, since the refusal path is what that task's criterion needed.
2. Still to do: setting priority and assignee from the card and from the task panel, which needs the panel from TASK-23.
3. Still to do: changing status by keyboard without dragging.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Landed with TASK-27, because a validated refusal cannot be demonstrated without a drag to refuse.

Dragging a card writes through the backlog CLI and the board then settles on what a rescan finds, never on where the card was dropped. Verified in a browser against a scratch bench of two projects with deliberately different status lists: a valid move wrote to disk and the file says In Progress; an invalid move was refused before anything moved, with the reason and the fix named.

The failure path was exercised rather than assumed. Making a task file read-only and dragging it produced the notice 'TASK-2 could not be moved: backlog task edit TASK-2 -s Done: EACCES: permission denied, open ...' - the task, the exact command, and the operating system's own reason - and the card returned to To Do because the rescan found it unchanged there. The file on disk was still To Do.

Creating, editing and deleting cards from the board are refused outright and SVAR's per-column add control is hidden. The board would otherwise do all three in memory, showing a task the files do not have, which is the one thing this application must never do.

Remaining for this task: priority and assignee from the card and the panel, which needs the panel from TASK-23, and changing status by keyboard without dragging.
<!-- SECTION:NOTES:END -->
