---
id: TASK-66
title: Make a draft as workable as a task
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 23:40'
updated_date: '2026-08-27 00:06'
labels: []
milestone: m-3
dependencies: []
priority: high
type: feature
ordinal: 66000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Drafts went unused because there was nothing to do with one. Capture is cheap and everything after it was not: a draft could not be created from the application at all, could not carry a priority or a milestone, and could only be read as a title and a body.

What the CLI actually allows is more than the drafts inbox uses. task create --draft takes the whole task field surface - description, labels, assignee, priority, type, milestone, acceptance criteria, dependencies, references, ordinal - and writes a draft with it. What it still does not allow is editing one: task edit refuses a DRAFT id on 1.48.0 and on 1.50.1 alike, so rewriting stays capture-and-discard, and the fields carried across should be all of them rather than four.

A promoted draft is an ordinary task and the panel already edits those completely, so promotion is the other half of the answer: promote and open, rather than promote and go looking for it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A note can be captured into the inbox from the application, with the fields the CLI accepts for a draft
- [x] #2 Rewriting a draft carries every field the CLI can set, not only the title and body
- [x] #3 A draft opens in the task panel so its whole content can be read
- [x] #4 Promoting a draft opens the resulting task, so the note can be finished in one go
- [x] #5 The inbox says what a draft cannot hold and why, rather than silently dropping it
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
The blocker was never really that drafts are hard to work with. It was that Muster used the wrong command. draft create knows four fields - title, body, labels, assignee - while task create --draft takes the whole task surface and writes the same file: priority, type, milestone, acceptance criteria, dependencies, a parent. Both 1.48.0 and 1.50.1 have it. Every draft write now goes through one path, so the two cannot drift apart, and a note can hold everything the format allows one to.

What is still out of reach is a status, because a draft's status is Draft and the CLI rejects any other for one. The form says that where the status control would be rather than dropping the field silently.

Capture was the missing half: until now a note could be triaged but not made, so the inbox could only ever be as full as something else had filled it. It is the same form as a new task with one checkbox, because deciding whether a thought is ready to be a task is exactly what triage is for - it should not also decide how much you are allowed to write down.

Promotion now says what the note became. The CLI prints only 'Promoted draft DRAFT-1', so the new id is found by seeing which task id was not there a moment before; promotion is the only thing happening between the two readings and both are taken under the same store. The panel opens on it, so a note is captured, triaged and finished in one pass instead of being promoted and then hunted for on the board.

Clicking a note opens it in the panel, which needed the panel to look in the drafts as well as the tasks - a one-line change that makes the whole body readable instead of the two lines the inbox row has room for.

Verified in the browser against a scratch bench: a note captured from the application arriving on disk with status Draft, a priority and two acceptance criteria; a draft opening in the panel; a rewrite carrying priority, type, both labels and the criteria across; and promotion opening TASK-2. Nine Go tests cover the same against the real CLI. The earlier inbox checks all still pass, and axe-core reports no violations on the screen, the edit form or the confirmation.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A note is now as workable as a task, everywhere except the one place Backlog.md will not allow.

The cause was a wrong command rather than a missing feature: Muster captured through draft create, which knows four fields, when task create --draft takes the whole task surface and writes the same file. Every draft write goes through one path now, so a note can carry a priority, a type, a milestone and acceptance criteria. Only a status cannot cross, because a draft's status is Draft, and the form says so where the control would be.

Notes can also be made from the application at last - the same form as a task with one checkbox - and a note can be read whole in the panel, and promoting one opens the task it became, so capture, triage and finishing happen in one pass.

Verified in the browser against a scratch bench, reading the files on disk: a captured note with its priority and criteria, a rewrite carrying every field across, a draft in the panel, and promotion opening TASK-2. Nine Go tests against the real CLI, the earlier inbox checks still green, no axe-core violations.
<!-- SECTION:FINAL_SUMMARY:END -->
