---
id: TASK-68
title: 'Edit the rest of a task: dependencies, references and ordering'
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-27 00:38'
updated_date: '2026-08-27 01:13'
labels: []
milestone: m-5
dependencies: []
priority: high
type: feature
ordinal: 68000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The panel edits status, priority, assignee, milestone, labels, title, body sections and acceptance criteria. What it cannot touch is everything that relates a task to something else: dependencies, references, documentation links, the parent, and the ordinal that decides manual order.

Dependencies are the one that bites. 317 of 875 live tasks carry them, better than a third, and the analytics screen already reports which tasks are blocked and by what - while the relationship it reports can only be created or removed in a terminal. 1.50.1 added --clear-deps, --clear-refs and --clear-docs, so removal is now as expressible as addition.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Dependencies can be added and removed from the panel, against tasks in the same project
- [x] #2 References, documentation links and modified files can be edited
- [ ] #3 A task's parent can be set and cleared
- [x] #4 Manual order can be changed, and the board and list agree on the result
- [x] #5 A dependency that would not resolve is refused with the reason, before it is written
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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Measure the flags first. task edit sets dependencies with --dep and clears them with --clear-deps; references have --ref to replace, --add-ref and --remove-ref to adjust, --clear-refs to empty; documentation has --doc and --clear-docs; modified files have --modified-file; manual order has --ordinal. The CLI validates dependencies itself and refuses an id it cannot find, with a message naming it.
2. There is no --parent on task edit - only on task create - so a task's parent cannot be changed after the fact through the CLI at all. That acceptance criterion cannot be met without writing to the file directly, which this project does not do. Record it rather than quietly checking it, and put the question to the user.
3. Backend: one wrapper per relationship, each following the existing pattern of write-then-reload. Dependencies are validated against the store before the call as well, so the common mistake is caught with a better message than the CLI's and without running a command.
4. Manual order is worth doing as dragging rather than as a number field, because that is what manual order is. The kanban's move-card event carries the id of the card it was dropped in front of, which is enough: the new ordinal is the midpoint between its neighbours, and where there is no gap the column is restacked at multiples of 1000 - the same allocation the CLI uses. The corpus already contains two hand-set midpoint ordinals from the CLI's own web UI, so this is the expected shape rather than an invention.
5. Panel: a section for what a task relates to, editable in place, next to the dependencies it already displays.
6. Tests against the real CLI for each write, and a browser pass that drags a card within a column and reads the ordinals off disk.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Acceptance criterion 3, setting and clearing a task's parent, is left unchecked because Backlog.md cannot do it. task create takes -p; task edit has no --parent at all and answers 'unknown option'. The only ways to change a parent would be writing the file directly, which this project does not do, or recreating the task, which loses its id and its history. Recorded rather than quietly checked; the criterion needs dropping or Backlog.md needs the flag.

Two CLI facts came out of the measurement. modified_files can be set and never emptied: there is no --clear-modified-files, and --modified-file "" exits 0 having changed nothing - a silent no-op. So the interface refuses to empty that one list and says why, rather than offering a control that appears to work. And the CLI validates dependencies itself, refusing the whole edit and naming the ids it could not find, so nothing partial is written; Muster checks first anyway, because the commonest mistake is pointing at a task in another project and saying so without running a command is both faster and clearer.

Manual order is dragging rather than a number field, because that is what manual order is. The board's move event carries the id of the card the dragged one was dropped in front of, which is enough: the new ordinal is the midpoint between its neighbours. Ordinals are neither unique nor mandatory, so where there is no room the column is restacked at multiples of 1000 first - the same allocation the CLI uses, and the corpus already contains two hand-set midpoints written by Backlog.md's own web interface. A card dropped into another column is two changes and both are made, status then position, or it would land wherever its old ordinal happened to put it.

While replacing the read-only dependency block I removed the subtasks section from the panel by accident; eslint caught it as an unused variable, and it is restored and covered by a check. Worth recording separately: that check reported a failure for twenty minutes that was not one - innerText applies text-transform, so a heading rendered in capitals never matched a case-sensitive assertion. The section had been there the whole time.

Verified in the browser against a scratch bench, reading the files after each write: dependencies written and shown back as links, an unresolvable id refused with the reason and nothing written, references written, and a card dragged to the top of its column taking ordinal 500 above the 1000 that was there - with the board agreeing afterwards. Eight Go tests, and no axe-core violations.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A task's relationships are editable: what it depends on, its references, its documentation links and the files it touched, all in the panel, and manual order by dragging a card within its column.

Dependencies are the reason: 317 of 875 live tasks carry them, and the analytics screen reported which tasks were blocked while the relationship it reported could only be changed in a terminal. One that does not resolve is refused with the reason before anything is written, because ids resolve inside their own project only.

Ordering is computed from where the card landed - the midpoint between its new neighbours, with the column restacked at multiples of 1000 when the gaps have run out, which is the allocation Backlog.md uses itself.

One criterion is not met and is left unchecked: a task's parent cannot be changed, because task edit has no --parent and this project does not write task files directly. One CLI limitation surfaced with it - the files-touched list can be set but never emptied - and the interface says so rather than offering a control that does nothing.

Verified in the browser with the files read after every write, plus eight Go tests against the real CLI and a clean accessibility pass.
<!-- SECTION:FINAL_SUMMARY:END -->
