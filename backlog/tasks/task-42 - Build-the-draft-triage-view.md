---
id: TASK-42
title: Build the drafts inbox and triage view
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:02'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-3
dependencies:
  - TASK-24
  - TASK-41
priority: medium
type: feature
ordinal: 42000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Backlog.md keeps drafts off the board by design, which is what makes capture cheap - and also what makes an unopened drafts directory the new inbox nobody reads. The triage view is where the inbox is emptied: promote, edit, reassign or discard.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Inbox lists drafts across every project with their capture date
- [x] #2 A draft can be promoted to a task through the backlog CLI
- [x] #3 A draft can be edited before promotion and reassigned to another project
- [x] #4 A draft can be archived or discarded with confirmation
- [x] #5 The view shows how long drafts have been waiting so a growing inbox is visible
- [x] #6 Inbox depth is visible from the main navigation
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Built before TASK-26 and TASK-41, which it is recorded as depending on. The dependency is on the capture side and the triage side does not need it: drafts already exist on disk, draft create, promote and archive were already wrapped, and this closes the one navigation item that went nowhere.

There were no drafts anywhere in the nine reference projects, so nothing here is measured against a corpus. It is measured against the CLI instead, on a scratch project, and four things came out of that:

A draft is DRAFT-n with status Draft and a created_date, in drafts/. Promoting renames it to the next task id, sets the project's first status, moves the file into tasks/ and keeps the capture date. Archiving moves it into archive/drafts.

Draft ids are reused. After DRAFT-1 was promoted and DRAFT-2 archived, the next capture was DRAFT-1 again - the number is the next free one in drafts/, not a counter. So a draft id is not identity over time, the same hazard archived task ids have.

Archiving does not check for collisions. Two different notes that had both been DRAFT-1 ended up as two files with the same id in archive/drafts. The inbox only ever shows active drafts, so nothing here trips over it, but it is a fact worth having written down.

There is no draft edit at all. task edit DRAFT-1 answers 'Task DRAFT-1 not found' and exits 1. Since Muster writes only through the CLI, editing a draft has to be capture-and-discard: create the new note, then archive the old one. That is also the only way to move a draft between projects, since ids and files belong to one project. The order matters and is deliberate - create first, archive second. A failed create leaves the original; a failed archive leaves two notes, which is visible and fixable; the reverse could lose the note.

The cost is that the rewritten note is captured now, so its wait restarts. The wait is the number this screen exists to show, so that is said in the form rather than hidden, and the failure case says plainly that both copies are still there rather than reporting a clean failure for a note that was in fact written.

Verified in the browser against a scratch bench of four drafts at 45, 12, 6 and 0 days across two projects: listed oldest first with every wait shown, depth on the navigation before the screen is opened and following it afterwards, promotion producing a task file on disk, a rewrite that moved a note to another project with the original gone from the old one, a discard that asks first and archives rather than deletes, and drafts still absent from the board. Six Go tests cover the ordering, the undated case, and every write against the real CLI. axe-core reports no violations on the screen, the edit form or the confirmation.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The inbox is the drafts folder made visible, and the last navigation item that went nowhere now goes somewhere.

Notes are listed oldest first across every project, each saying how long it has waited, with the depth on the navigation itself so a growing pile is visible without opening it. From there a note is promoted into a task, rewritten, moved to another project, or discarded into the archive with a confirmation.

The awkward part is that Backlog.md has no draft edit - task edit refuses a DRAFT id outright - so rewriting is capture-and-discard through the CLI, which is also the only way to move a draft between projects. That restarts the wait, and since the wait is what this screen is for, the form says so rather than hiding it.

Verified in the browser against a scratch bench of four drafts at 45, 12, 6 and 0 days in two projects, checking the files on disk after every action, plus six Go tests against the real CLI and an axe-core audit with no violations.
<!-- SECTION:FINAL_SUMMARY:END -->
