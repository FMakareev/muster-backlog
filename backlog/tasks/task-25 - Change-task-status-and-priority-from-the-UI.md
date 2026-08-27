---
id: TASK-25
title: Move tasks between statuses by drag and drop
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-27 17:16'
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
- [x] #5 Priority, assignee and labels can be set from the card and from the task panel
- [x] #6 Status can also be changed by keyboard without dragging
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

Editing controls sit at the top of the task panel: status, priority, assignee and labels. Every one writes through the CLI and then re-reads, so nothing shown afterwards is an assumption.

The choices offered are the project's own, not the board's. The status select lists only the statuses that project declares - verified in the browser: opening an Alpha task offers To Do, In Progress and Done and never In Review, which only Beta declares. Priorities come from each project's config rather than from a hardcoded list, because both priorities and types are configurable and a hardcoded list would be wrong the first time someone changed one.

Square brackets move a focused card to the previous or next status its project declares. Verified by pressing ] three times on a bench task: To Do, In Progress, In Review, Done, with each step written to disk; a fourth press reported 'TASK-1 is already at the last status Beta declares' rather than doing nothing silently, and [ walked it back.

A defect found only by pressing the key twice: after a move the board re-renders and the card loses focus, so the second keypress went nowhere. A keyboard-driven board that stops responding to the keyboard after exactly one move is not keyboard-driven. Focus is now restored to the same card after the refresh, and the chained test above is what proves it.

A second defect found by looking: the panel showed status, priority, assignee and labels in the editable controls and then again in the metadata table a few pixels below. Removed from the table, which now carries only what is not editable here.

All four controls verified against the files: after using them, Alpha's task on disk reads status In Progress, priority high, assignee @fmakareev, labels backend.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Tasks now move and change from the board and the panel, and every change goes through the backlog CLI.

Dragging a card writes its new status; the board settles on what the rescan finds rather than on where the card was dropped. A drop into a status the card's own project does not declare is refused before anything moves, with the reason and the fix named. Square brackets move a focused card between the statuses its project declares, so the board works without a mouse. The panel carries controls for status, priority, assignee and labels, offering each project's own configured values.

Verified by driving the application against a scratch bench with two projects on deliberately different status lists, then reading the files. Every control wrote: the task on disk ends up with status In Progress, priority high, assignee @fmakareev and label backend. Three chained ] presses walked a task from To Do to Done one status at a time, a fourth reported it was already at the last status its project declares, and [ walked it back. Making a task file read-only and dragging it produced a notice carrying the command and the EACCES, and the card returned to where the files said it was.

Two defects came out of the verification. The keyboard stopped working after one move, because the re-render dropped focus - fixed by restoring it, which is what let the chained test run at all. And the panel showed four fields twice, in the controls and again in the metadata table below.
<!-- SECTION:FINAL_SUMMARY:END -->
