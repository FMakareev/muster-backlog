---
id: TASK-44
title: Build the Projects screen
status: To Do
assignee: []
created_date: '2026-08-26 15:03'
updated_date: '2026-08-26 15:28'
labels: []
milestone: m-3
dependencies:
  - TASK-24
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 44000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The registry stops being a hand-edited file. Projects are added, removed, reordered and recoloured from the application, and the screen is where a folder is turned into a Backlog.md project in the first place.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A project can be added by folder path and removed from the registry in the UI
- [ ] #2 Display name, colour and ordering are editable per project
- [ ] #3 Each project shows its task counts, milestone progress and the Backlog.md configuration it declares
- [ ] #4 A folder without a backlog offers initialisation from this screen
- [ ] #5 Registry edits are written back preserving comments where possible
- [ ] #6 An invalid or unreadable path is rejected with an explanation before it is saved
- [ ] #7 A project can be temporarily hidden from the board without being removed from the registry
<!-- AC:END -->
