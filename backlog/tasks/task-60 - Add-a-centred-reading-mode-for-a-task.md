---
id: TASK-60
title: Add a centred reading mode for a task
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 19:43'
updated_date: '2026-08-26 20:11'
labels: []
milestone: m-2
dependencies:
  - TASK-23
priority: medium
type: feature
ordinal: 60000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The task panel is a column against the right edge. On a wide screen that puts the text a long way from where the eye rests, and reading a long description means looking at the far side of the monitor. A second mode shows the same task centred over the board, for reading rather than for working alongside it. Which mode is used is a preference, not a mode switch to rediscover each time.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A task can be opened centred over the board as well as in the side panel
- [x] #2 The centred view has a comfortable measure for reading rather than filling the window
- [x] #3 Switching between the two modes is one action and the choice is remembered
- [x] #4 Both modes show the same content and close the same way
- [x] #5 The centred view is reachable and dismissable by keyboard
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A task can be read centred over the board instead of in the column against the edge.

One control in the panel header switches between them, the choice is stored in the settings file rather than rediscovered each time, and both modes show exactly the same content and close on Escape. The centred view is capped at a comfortable measure rather than filling the window, and sizes to its content rather than stretching to the full height.

Verified in a browser: the centred view's midpoint sits on the viewport's midpoint, its width is 768 pixels against a 1500-pixel window, and its height is 489 against a 950-pixel viewport rather than the full height it stretched to on the first attempt.
<!-- SECTION:FINAL_SUMMARY:END -->
