---
id: TASK-26
title: Capture into the drafts inbox by global hotkey
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-27 18:00'
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
TASK-41 was done first and took the parts of it that were not about the window: the default project and remembering the last one, switching project by keyboard, multi-line text, the unobtrusive confirmation, and not losing what was typed when the write fails. All of that now lives in the capture form and applies to whatever opens it.

What is left here is what only a window can do: a configurable global hotkey, opening over another application, and returning focus to it afterwards. Worth measuring before building - global hotkeys on Wayland go through the desktop's own portal rather than the application, and this machine runs Wayland.
<!-- SECTION:NOTES:END -->
