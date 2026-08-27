---
id: TASK-23
title: Build the task detail panel
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-1
dependencies:
  - TASK-21
  - TASK-56
priority: high
type: feature
ordinal: 23000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The key screen. Its absence is precisely what disqualified the Obsidian plus Bases alternative: a table without fast access to the task body reads as a dump. The panel opens from anywhere and shows the whole task - description, acceptance criteria, implementation plan, notes, dependencies and file references.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Panel opens from a board card and closes without losing board state
- [x] #2 Description, acceptance criteria, implementation plan and notes render as formatted markdown
- [x] #3 Acceptance criteria display their checked state
- [x] #4 Dependencies render as links that navigate to the referenced task
- [x] #5 Project, milestone, priority, type, labels, assignee and updated date are visible
- [x] #6 Panel content updates live when the underlying file changes
- [x] #7 Panel opens and closes by keyboard
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. A panel beside the board rather than a modal over it. This is the screen whose absence disqualified the Obsidian alternative, and its job is fast access to a body while the board stays visible - a dialog that hides the board would repeat the mistake in a different shape.
2. Selection is a task reference held in one store, so the panel, the board and anything later all agree on what is open.
3. Content comes from the same task list the board reads, derived by reference rather than fetched separately. That is what makes it live for free: when the watcher reloads a project, the panel follows.
4. Every section renders through the markdown pipeline from TASK-56, so a task body, a document and a decision all look the same and are sanitised the same way.
5. Dependencies resolve inside the owning project only, because ids collide across projects. A dependency that does not resolve is shown as the raw id rather than a dead link.
6. Escape closes, and the panel takes focus when it opens so a keyboard user is not left behind on the board.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
A panel beside the board rather than a dialog over it. The Obsidian alternative was rejected because a table gives no fast access to the body; a modal that hides the board would repeat that mistake in a different shape.

Content is derived from the same task list the board reads, keyed by reference rather than fetched separately. That is what makes it live for nothing: when the watcher reloads a project, the panel follows.

Found and fixed while verifying, none of which reading the code would have shown:

- SVAR's themes load an icon font and Open Sans from cdn.svar.dev. A local-first application was reaching out to a CDN on every launch, and would have shown broken icons with the network unplugged. Caught only because the browser check was recording requests. The theme takes a fonts prop, so it is mounted with fonts disabled and the seven icons the board actually uses are declared from a bundled 5 KB woff2. Recorded in the specification so a future dependency gets the same check.
- The panel did not open at all: SVAR prefixes the identity it writes into data-id with a colon, so every lookup missed. Cards carry role=button and tabindex, so once that was fixed they open by Enter or Space as well as by click.
- The panel was cut off by the right edge of the window. SVAR's columns have fixed widths, so the board container would not shrink; it needed min-width zero for flex to give the panel its space.

Verified by driving the running application against the nine real projects: opening a card shows title, status, priority, type, milestone, assignee and both dates, seven acceptance criteria with their checked state, and description, notes and final summary rendered through the markdown pipeline. Dependencies TASK-5 and TASK-6 render as links and clicking one navigates to that task. Escape closes the panel and the board still holds its 58 cards. Zero offsite requests.

For the live criterion: with the panel open on TASK-7, an external 'backlog task edit TASK-7 --add-label live-check' made the label appear in the panel with no interaction, and removing it reverted the panel, which stayed open throughout.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the task panel - the screen whose absence disqualified the Obsidian alternative.

It opens beside the board from a click or from Enter on a focused card, shows project, id, title, status, priority, type, milestone, assignee, labels and both dates, lists acceptance criteria with their state on disk, renders description, plan, notes, final summary and comments through the markdown pipeline, and links dependencies to the tasks they name. Escape closes it and the board is untouched. Content is derived from the same task list the board reads, so it follows a file change on its own.

Verified against the nine real projects in a browser rather than in tests: TASK-7 opens with all metadata, six criteria and their checked states, rendered body sections, and TASK-5 and TASK-6 as working dependency links - clicking one lands on 'Adopt Conventional Commits with automated message linting'. With the panel open, an external CLI edit added a label and it appeared without interaction; removing it reverted.

Verification found three defects. The panel would not open at all, because SVAR prefixes the identity it writes into data-id with a colon. It was clipped by the window edge, because a fixed-width board will not shrink without min-width zero. And most importantly, SVAR's themes were loading an icon font and a typeface from cdn.svar.dev - a local-first application making a network request on every launch, found only because the browser check happened to be recording requests. The theme now mounts with fonts disabled and the icons are bundled; the specification records the check so the next dependency gets it too.
<!-- SECTION:FINAL_SUMMARY:END -->
