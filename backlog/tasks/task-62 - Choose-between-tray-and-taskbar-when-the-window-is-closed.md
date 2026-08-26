---
id: TASK-62
title: Choose between tray and taskbar when the window is closed
status: In Progress
assignee:
  - '@claude'
created_date: '2026-08-26 19:43'
updated_date: '2026-08-26 21:47'
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
- [x] #1 A preference chooses between minimising to the tray and behaving as an ordinary window
- [ ] #2 With the tray chosen, closing the window leaves the application resident and the tray icon restores it
- [ ] #3 The tray menu can show and quit the application
- [x] #4 A desktop environment with no tray falls back to ordinary window behaviour and says so rather than vanishing
- [x] #5 The preference is remembered across restarts
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The preference lives in a new settings file, ~/.config/muster/settings.yml, separate from the project registry. The registry says where projects are and a person edits it by hand; this one is written by the application, and mixing them would mean rewriting a hand-edited file every time a toggle is flipped.

Availability is checked before the icon is created. On Linux the tray is not part of the windowing system: it is a status-notifier host on the session bus, and many desktops do not run one. Asking first is the difference between a preference that works and an application that disappears with no way to get it back.

Verified what can be verified without a person looking at their tray:
- The preference is offered, saved and reloaded: setting it wrote 'on_window_close: tray' and the value survived a restart.
- Availability detection is correct on this machine, which does have a host (org.kde.StatusNotifierWatcher), and the desktop binary starts and holds its window with tray behaviour on.
- The fallback path was exercised by running with DBUS_SESSION_BUS_ADDRESS pointing at nothing: the application reports 'This desktop has no system tray - Muster is set to stay in the tray, but nothing on the session bus offers one, so the window behaves ordinarily', and keeps ordinary behaviour.

That verification found a real defect with nothing to do with the tray: reload rebuilt the whole problem list from scratch, silently discarding conditions of the machine rather than of the data - a missing CLI, unreadable preferences, an absent tray. Problems are now kept in two lists, standing and scan-derived, with the standing ones first because they are the ones a person has to act on. Covered by a test.

Two criteria are left unchecked on purpose. Whether closing the window really leaves the application resident, and whether the tray menu shows and quits it, cannot be established from here: it needs someone looking at their own tray and clicking it. The code is there and the icon is created; the interaction is unproven.

Follow-up from first use: the preferences window could not be closed from the keyboard. It listened for Escape on its own element and never moved focus into itself, so the key had nowhere to land - the same flaw the search and filter panels had in weaker forms, where a stray click was enough to make them unclosable.

All five overlays now share one action. Escape is listened for on the window rather than on the overlay, so dismissal does not depend on where the caret went; overlays are kept in a stack so Escape over preferences opened above a task closes the preferences and leaves the task; and focus moves into an overlay that has not placed the caret somewhere better itself. Verified in the browser: preferences takes focus on opening and closes on Escape, all four of preferences, filters, search and new task close after focus is deliberately blurred, and Escape over a task panel closes the overlay above it without closing the task.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @claude
created: 2026-08-26 20:15
---
Two criteria need you rather than me: close the window with the tray preference on and check the icon appears, restores the window, and that its menu shows and quits. Everything around them is verified.
---
<!-- COMMENTS:END -->
