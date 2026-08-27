---
id: TASK-26
title: Capture into the drafts inbox by global hotkey
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-3
dependencies:
  - TASK-24
priority: medium
type: feature
ordinal: 26000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Drafts are the inbox: Backlog.md keeps them out of the board, so anything can be captured without deciding whether it deserves to be tracked. One window, one field, project plus text, written through backlog draft create. This is what replaces scattered notes that never reach planning.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A configurable global hotkey opens the capture window from any application
- [ ] #2 Capture window takes a target project and free text and needs no mouse
- [ ] #3 Submitting creates a draft through the backlog CLI in the selected project
- [ ] #4 Window closes on submit or on escape and returns focus to the previous application
- [ ] #5 Captured drafts stay out of the board, matching Backlog.md behaviour
- [ ] #6 A hotkey already taken by the system is reported instead of failing silently
<!-- AC:END -->
