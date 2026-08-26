---
id: TASK-61
title: Show subtasks on the board
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 19:43'
updated_date: '2026-08-26 21:47'
labels: []
milestone: m-2
dependencies: []
priority: low
type: enhancement
ordinal: 61000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Backlog.md tasks can have subtasks through parent_task_id, and 91 files in the reference corpus use them. On the board a subtask is an ordinary card with a dotted id, so the relationship is invisible and a parent looks no different from anything else. Worth showing, but not at the cost of making the board harder to read - a nested board is a worse board.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A card shows whether it has subtasks and how many are done
- [x] #2 A subtask card shows which task it belongs to
- [x] #3 Opening a parent lists its subtasks as links; opening a subtask links back
- [x] #4 The board remains readable at the full corpus size with this on
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Measure the corpus first, since the format offers two signals for one relationship. Across 712 task files in nine projects: 92 subtasks, all with both a dotted id and parent_task_id, never one without the other; the subtasks list field is populated in no file at all; nesting never goes past one level; every parent resolves by plain case-insensitive id equality, and 6 of the 92 have a parent in a different directory than their own.
2. Backend: resolve the relationship in Go, in one pass over the store, keyed by ref. Read parent_task_id and derive the forward direction from the back-links, because the subtasks field is never written. Search every class for the parent. Reuse the project's last declared status as the finished rule rather than restating it.
3. Expose it on the existing task view as a small optional block - parent ref, parent title, done and total - so a card can show both directions without a request per card, and nothing is added to the payload of the tasks that have no family.
4. Add a service call for the panel that returns a parent's subtasks in full, made once when a task is opened rather than for every card on the board.
5. Frontend: a count badge on a parent card, a parent chip on a subtask card, and both directions as links in the panel. The board stays flat - nesting it would make it harder to read, which is the one thing this must not cost.
6. Tests over the resolution rules, and measure the board at full corpus size with it on.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Measured before designing, because the format carries two signals for this one relationship and they are not equally useful. Over the author's nine projects: every subtask has both a dotted id and parent_task_id and neither ever appears without the other; the subtasks list is populated in no file at all, so the forward direction has to be derived from the back-links; nesting never goes past one level; every parent resolves on plain case-insensitive id equality, with no need to normalise the zero-padding that some projects use - which is as well, since normalising would invent a match between TASK-7 and TASK-007 in a project that has both. Six links cross directories: a parent can be completed or archived while its subtasks are not, so resolution searches every class rather than the child's own. doc-3 had already predicted the shape of this - 'derive the tree from parent_task_id, but tolerate subtasks appearing'.

Resolution lives in the store as one pass over every project, not a lookup per task: the board asks for hundreds of cards at once and a query per card would be hundreds of scans of the same slice. The task view gained an optional family block - parent ref, parent title, done and total - which is nil for the tasks that have no relationship, so nothing was added to the payload of the 800-odd cards that are not part of one. The list itself is a separate call the panel makes once when a task is opened.

Done counts a subtask at its own project's last declared status, or one the CLI has moved into completed/. That is the same definition analytics uses, and the two now share one function rather than restating the rule. Archived subtasks are in neither half of the count, because archiving is a soft delete and a board that does not show them should not count them.

The board stays flat, which was the constraint in the description and is the one thing this could not be allowed to cost. A parent card carries a count, a subtask card carries its parent's id, and both directions are links in the panel.

The subtask chip is knowingly redundant with the dotted id in this corpus - TASK-007.05 already says TASK-007 to anyone who knows the convention. It stays because parent_task_id is the authoritative field rather than a convention, and because the chip carries the parent's title where the id cannot.

Verification: nine new Go tests over the resolution rules, and a cross-check of the running application against the files on disk - 30 parents and 95 links read independently out of the markdown by a separate script and compared to what the store resolved, with zero mismatches in either the links or the done counts. Driven through the browser: 5 parent cards and 2 subtask cards showing their chips, a parent listing its subtasks in the panel, a subtask linking back, and the back-link opening the parent (TASK-075.01 -> TASK-075). At full corpus size the board holds 33 cards in 491 DOM nodes, 15 nodes per card, 21 MB heap, no card nested inside another, and scrolling a full column takes 792 ms. Resolving all 95 links over 902 tasks takes 233 microseconds, against a 50 ms budget now enforced by the corpus test.

The counts here move between readings: the reference projects are live and worked on by agents while this was being written, so the corpus went from 712 task files with 92 subtask links when the plan was drafted to 902 tasks with 95 links when it was verified. The rules held identically at both readings, and reload was checked not to lose links across a rescan.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Subtasks are now visible on the board without the board becoming nested.

A parent card says how many of its subtasks are finished, a subtask card says whose it is, and the panel carries the links in both directions. The board itself stays flat: nesting cards inside cards would cost more legibility across nine projects than the relationship is worth, and that was the constraint the task was written around.

The relationship is derived from parent_task_id in one pass over the store, because the subtasks list that would give the forward direction is populated in none of the 92 real subtask files. Parents are looked for in every directory, since six of those links have a parent that was completed or archived while its subtasks were not. A subtask counts as done at its project's own last declared status or once the CLI has moved it into completed/ - the same rule analytics uses, now shared rather than restated - and archived subtasks count in neither half.

Verified three ways: nine Go tests over the resolution rules; an independent read of the markdown on disk compared against what the running store resolved, 30 parents and 95 links with zero mismatches; and the browser, where a parent listed its subtasks, a subtask linked back and the back-link opened the parent. At full size the board is 33 cards in 491 DOM nodes with nothing nested, and resolving all 95 links across 902 tasks takes 233 microseconds.
<!-- SECTION:FINAL_SUMMARY:END -->
