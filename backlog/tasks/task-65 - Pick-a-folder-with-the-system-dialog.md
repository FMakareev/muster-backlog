---
id: TASK-65
title: Pick a folder with the system dialog
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 23:40'
updated_date: '2026-08-27 17:16'
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
- [x] #1 A button on the add-a-folder form opens the system directory chooser
- [x] #2 The chosen folder fills the path field and is inspected as if it had been typed
- [x] #3 Cancelling the dialog leaves the form exactly as it was
- [x] #4 Where no dialog is available the button is absent rather than broken, and the field still works
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
- [ ] #5 Linters and formatters pass across Go and frontend
- [ ] #6 Automated tests cover the change and the suite is green
- [ ] #7 User-facing behaviour change is reflected in README or docs
- [ ] #8 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The chosen path goes through the same field and the same inspection as a typed one, so there is one way into the form rather than two, and the field keeps working where there is no dialog. Availability is a build-time fact rather than a runtime probe: the server build has no dialogs at all and answers every request with an error, so the button is absent there rather than present and broken. The chooser opens beside the projects already registered, since a second project usually lives next to the first, and at the home directory otherwise.

This needed a way to exercise the desktop build, which the browser tests cannot reach - they drive the server build, which is exactly the one with no dialogs. There is no xdotool on this machine, but there is Xvfb, ffmpeg and libXtst, which is enough: the real binary runs on a virtual display, XTest sends the keys and clicks through a small ctypes helper, and ffmpeg takes the screenshot. All four criteria were checked that way against the actual desktop build rather than by construction.

What that showed, in order: the browse button present on the add form; the GTK directory chooser opening already inside the folder holding the registered projects and listing only directories; choosing 'plain' putting its full path in the field, with the inspection underneath reading 'There is no backlog here yet. This is not a git repository, so it will be initialised without git integration' and the name pre-filled from the folder; and cancelling a second chooser leaving the form byte-for-byte as it was.

The same method reaches the two tray criteria on TASK-62 that were left unchecked for want of a click.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A browse button on the add-a-folder form opens the desktop's own directory chooser, starting beside the projects already registered. What it returns goes through the same field and the same inspection as a typed path, so nothing downstream has two cases to handle, and cancelling changes nothing.

Where there is no dialog - the server build - the button is absent rather than broken, which is decided at build time rather than guessed at runtime.

Verified against the real desktop binary rather than by construction: with no xdotool available, the app runs on an Xvfb display driven through XTest by a small ctypes helper, with ffmpeg for screenshots. That showed the button, the chooser opening in the right place, the chosen folder inspected in the form, and a cancel leaving it untouched. The same rig can now reach the tray criteria that TASK-62 left unchecked.
<!-- SECTION:FINAL_SUMMARY:END -->
