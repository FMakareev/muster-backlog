---
id: TASK-62
title: Choose between tray and taskbar when the window is closed
status: To Do
assignee: []
created_date: '2026-08-26 19:43'
labels: []
milestone: m-3
dependencies: []
priority: medium
type: feature
ordinal: 62000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
This is a tool to reach for several times an hour, so where it goes when the window is closed matters. Some people want it resident in the system tray, others want it to behave like an ordinary window and live in the taskbar. Neither is right for everyone, so it is a preference rather than a decision. Wails v3 provides a system tray; on Linux the tray is a desktop-environment feature and may simply not exist, which the setting has to survive.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A preference chooses between minimising to the tray and behaving as an ordinary window
- [ ] #2 With the tray chosen, closing the window leaves the application resident and the tray icon restores it
- [ ] #3 The tray menu can show and quit the application
- [ ] #4 A desktop environment with no tray falls back to ordinary window behaviour and says so rather than vanishing
- [ ] #5 The preference is remembered across restarts
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
