---
id: TASK-72
title: A note in the panel offers edits that cannot work
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-27 04:16'
updated_date: '2026-08-27 04:18'
labels: []
milestone: m-5
dependencies: []
priority: high
type: bug
ordinal: 72000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reported from first use: opening a note and changing its description failed with 'DRAFT-4 could not be saved: backlog task edit DRAFT-4 -d ...: Task DRAFT-4 not found.'

Letting notes open in the panel was right - it is the only place the whole body can be read - but it brought every task-editing control with it, and all of them write through task edit, which Backlog.md refuses for a DRAFT id. So the panel offered a dozen things that could only fail, and failed with a message naming a note that plainly exists, which reads like a bug in Muster rather than a limit of the format.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A note in the panel offers nothing that cannot work on a note
- [x] #2 The whole body stays readable, which is why a note opens there
- [x] #3 The panel says why, and offers what can be done instead
- [x] #4 Promoting from the panel opens the resulting task, editable
- [x] #5 The service refuses a task edit against a note whatever calls it, with a message that names the reason
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
My regression, from letting notes open in the panel. That was worth doing - the inbox row has space for two lines and a note can be much longer - but it brought the whole task editor with it, and every part of that writes through task edit. The CLI answers 'Task DRAFT-4 not found', which is the worst possible message here: it names a note the person is looking at, so it reads as a bug in Muster rather than as a limit of the format.

Fixed in two places, deliberately. The panel offers a note only what a note can have done to it - promoted, or rewritten in the inbox - and says why in a sentence rather than leaving someone to infer it from a failure. And underneath, every mutation that runs task edit now refuses a note before running anything, with a message that names the reason and points at the two things that work. The interface not offering it is the design; the service refusing it is the floor, and it holds whatever else ends up calling these.

Promoting from the panel is the natural resolution rather than a detour: the note becomes a task and the panel follows it, so the thing being read is the thing that can now be edited.

Verified in the browser by reproducing the report - opening the note from the inbox - then checking that the panel offers no select, no text field and no edit button on a note, that the body is still readable in full, that it says why, and that promoting from there writes the task, carries the body over and leaves the panel on an editable task. Six Go tests cover the refusals against the real CLI, including that the CLI's own confusing message never reaches the interface and that editing a real task still works.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Opening a note in the panel used to offer the whole task editor, and every part of it failed: those controls write through task edit, which Backlog.md refuses for a DRAFT id, with a message naming a note that is plainly there.

The panel now offers a note only what can be done to one - promote it, or rewrite it in the inbox - and says why. The body stays readable in full, which is the reason a note opens there at all. Promoting from the panel opens the resulting task, editable, so the thing being read becomes the thing being worked on.

Underneath, every mutation that runs task edit refuses a note before running anything. The interface not offering it is the design; this is the floor under it.

Verified by reproducing the report in the browser and checking the files, plus six Go tests against the real CLI.
<!-- SECTION:FINAL_SUMMARY:END -->
