---
id: TASK-65
title: Pick a folder with the system dialog
status: To Do
assignee: []
created_date: '2026-08-26 23:40'
labels: []
milestone: m-3
dependencies: []
priority: medium
type: feature
ordinal: 65000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Adding a project means typing or pasting a path, which is the one place in the application that asks a person to know something the desktop already knows. Wails can open the native directory chooser; the text field stays, because a path can also be pasted and because there is no dialog in server mode.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A button on the add-a-folder form opens the system directory chooser
- [ ] #2 The chosen folder fills the path field and is inspected as if it had been typed
- [ ] #3 Cancelling the dialog leaves the form exactly as it was
- [ ] #4 Where no dialog is available the button is absent rather than broken, and the field still works
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
- [ ] #5 Linters and formatters pass across Go and frontend
- [ ] #6 Automated tests cover the change and the suite is green
- [ ] #7 User-facing behaviour change is reflected in README or docs
- [ ] #8 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
