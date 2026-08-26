---
id: TASK-41
title: Refine the capture window into a usable inbox
status: To Do
assignee: []
created_date: '2026-08-26 15:02'
updated_date: '2026-08-26 15:28'
labels: []
milestone: m-3
dependencies:
  - TASK-26
priority: medium
type: enhancement
ordinal: 41000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Capture only works if it costs nothing to use: sensible defaults, the last project remembered, multi-line text, and no lost input when something goes wrong. Everything here is polish on top of the working hotkey capture, and all of it decides whether the inbox is actually used.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Capture window opens with a sensible default project and remembers the previous choice
- [ ] #2 Project can be switched by keyboard without leaving the text field
- [ ] #3 Multi-line capture is supported and the whole text reaches the draft
- [ ] #4 Successful capture is confirmed unobtrusively without stealing focus back
- [ ] #5 A capture made while the target project is unavailable is retained rather than lost
<!-- AC:END -->
