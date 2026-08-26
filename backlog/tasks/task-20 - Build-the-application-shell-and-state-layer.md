---
id: TASK-20
title: Build the application shell and state layer
status: To Do
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-1
dependencies:
  - TASK-19
priority: high
type: feature
ordinal: 20000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The frontend skeleton every screen mounts into: layout, navigation between board and future screens, Tailwind theme and tokens, and nanostores wiring fed by the backend bridge. Visual decisions taken here set the tone of the whole product, so they are made deliberately rather than accumulated.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Application shell renders navigation and a content region
- [ ] #2 nanostores hold the task set and registry and update from backend events
- [ ] #3 Tailwind theme tokens for colour, spacing and typography are defined in one place
- [ ] #4 Loading, empty and error states are handled at the shell level
- [ ] #5 Keyboard navigation works for the primary layout without a mouse
<!-- AC:END -->
