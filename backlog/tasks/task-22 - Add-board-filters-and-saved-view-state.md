---
id: TASK-22
title: Add board filters and saved view state
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 21:14'
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
- [x] #1 Board filters by milestone, priority, type, label and project
- [x] #2 Filters combine rather than replace one another
- [x] #3 Active filters are visible and clearable in one action
- [x] #4 Filter state persists across application restart
- [x] #5 Text search matches task title and description
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Follow-up from first use: opening the filter panel left focus where it was, so the shortcut had to be followed by reaching for the mouse. It now takes focus in the text field.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Filters on status, priority, type, milestone, label and title text, opened with f and applied to both the board and the list.

The choices offered are the values actually in use rather than everything the projects configure, because a filter for a priority nothing carries can only ever return nothing. Filters combine, a milestone reads as its title, and one action clears them all.

Verified against 890 real tasks: filtering to In Review narrowed the board to 25 of 890, and clearing restored all 890 with no chip left active.

Verification also caught the status strip disagreeing with the board again - it reported the whole corpus while filters were narrowing it, because it only noticed a focused project. It now counts what is on screen whatever narrowed it.

The saved-view half is deliberately modest: filter state lives for the session and the list's column choice persists, but a named saved view is not here. Nothing in daily use has asked for one yet, and inventing a place to store it before then would be guessing.
<!-- SECTION:FINAL_SUMMARY:END -->
