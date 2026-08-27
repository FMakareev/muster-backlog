---
id: TASK-21
title: Render the multi-project kanban board
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-1
dependencies:
  - TASK-20
priority: high
type: feature
ordinal: 21000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The central screen: one board over every registered project at once. This is what the VSCode extension cannot do - it switches between backlog folders rather than combining them. Built on SVAR Svelte Kanban (MIT, native Svelte 5, virtualised) so drag-and-drop, grouping and large-board performance come from the library instead of being hand-built.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 SVAR Svelte Kanban is integrated and themed consistently with the application
- [x] #2 Board renders tasks from every registered project in the unified columns
- [x] #3 Project of origin is visible on every card and tasks can be grouped by project
- [x] #4 Cards show title from frontmatter, priority, type, labels, assignee and milestone
- [x] #5 Board reflects file system changes live without user action
- [x] #6 Board stays responsive with the full corpus of roughly 640 tasks loaded
- [x] #7 Clicking a card opens the task panel
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Cards carry the project's colour as a small square plus the id, priority, type, milestone, labels and assignee, all in the same dense register as the rest of the interface. With nine projects interleaved in one column, a card that does not say where it came from is a card you have to click to understand.

Grouping is expressed as an ordering rather than as swimlanes, because the open edition of SVAR has no rows - confirmed by reading the 2.6.0 component API, whatever its readme implies. With grouping on, a column's cards clump by project instead of interleaving; within a project the board keeps the order the store gave it, which is Backlog.md's own comparator. The toggle is in the top bar with its key shown next to it, like the screen keys.

Measured in a browser against the nine real projects, 884 tasks:
- 52 to 57 cards in the DOM at any moment against 884 loaded, so virtualisation is doing its job rather than being switched on hopefully.
- Toggling grouping, which re-sorts every card: 95 ms.
- Opening the task panel: 90 ms.
- JS heap: 18 MB.
- Scrolling a long column keeps the DOM at the same few dozen cards and the titles change as expected.

The wheel-step figure came out at a flat 50 ms, which is Playwright's own step pacing rather than render time, so it is not quoted as a rendering measurement.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The multi-project board is complete: one set of columns over every registered project at once, which is the thing no existing tool does.

Cards show the project's colour, id, title, priority, type, milestone, labels and assignee; clicking or pressing Enter opens the task panel; the board follows file changes without polling; and a grouping toggle keeps each project's cards together inside a column, since SVAR's open edition has no swimlanes.

Verified in a browser against the nine real projects and 884 tasks. Columns are the union of the projects' declared statuses; cards from all nine appear together, each identifiable by colour; grouping clumps a column by project; the panel opens on click. Responsiveness measured rather than asserted: 52 to 57 cards in the DOM against 884 loaded, a full re-sort in 95 ms, the panel opening in 90 ms, and an 18 MB JS heap.
<!-- SECTION:FINAL_SUMMARY:END -->
