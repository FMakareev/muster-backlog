---
id: TASK-27
title: Reconcile differing status sets into unified board columns
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-1
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 27000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Statuses are per-project configuration in Backlog.md, and registered projects will not agree on them. The board therefore derives its columns from the union of every project status list rather than imposing one. Muster never edits another project config.yml to make the view simpler.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Board columns are the union of the status lists declared by the registered projects
- [x] #2 Column ordering is derived deterministically and documented, including how conflicting orders are resolved
- [x] #3 A project that does not declare a status shows as empty in that column rather than as an error
- [x] #4 A card cannot be dragged into a status its own project does not declare, and the reason is shown
- [x] #5 Adding or removing a status in a project config is picked up without restarting the application
- [x] #6 No code path writes to another project configuration file
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/board holds the column algorithm, in Go rather than in the frontend, so the ordering rules can be tested properly instead of being eyeballed in a browser.
2. Ordering is a weighted precedence vote. Each project's list contributes an ordering opinion for every pair of statuses it declares; the stronger direction wins, so eight projects agreeing outvote one that disagrees. Ties and cycles fall back to a documented, deterministic rule rather than to map iteration order.
3. Statuses match case-insensitively but keep the spelling first seen in registry order, because enum case is not normalised on disk.
4. A column records which projects declare it, which is what turns 'this project has nothing here' into an empty cell rather than an error, and what lets a drop be refused with a reason.
5. Nothing in this package or anywhere else opens another project's config for writing. The board adapts to what projects declare; it never asks them to agree.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The column algorithm lives in internal/board, in Go rather than in the frontend, so the ordering rules are tested rather than eyeballed in a browser.

Ordering is a weighted precedence vote. Every project that declares two statuses has an opinion about which comes first, counted once per project, and for each pair the stronger direction wins - so eight projects agreeing outvote one that does not, and the disagreement is recorded so the interface can explain an order that would otherwise look arbitrary. All pairs are counted, not only adjacent ones, so a list of four statuses contributes six opinions.

Two things had to be decided rather than left to chance, and both are documented in the code:
- Ties. When a pair is drawn, or several statuses are equally ready to be placed, the one declared by more projects goes first, then the one seen earlier in registry order, then alphabetically. Registry order is a user-visible choice, so deriving from it keeps the result predictable.
- Cycles. Three projects can disagree in a loop - A before B, B before C, C before A - and no order satisfies all three. The same tie-break cuts the loop. The alternative would be refusing to draw a board.

A test runs Build fifty times on the same input and asserts the order never changes, because nothing here may depend on Go's map iteration order.

Statuses match case-insensitively but keep the spelling of the first project to declare them, since enum case is not normalised on disk.

The measured position is easier than feared: across the nine real projects there are only two distinct lists and one is a superset of the other, so today's union orders itself with zero conflicts. That is a coincidence of this corpus rather than a property of the format, so disjoint lists and cycles are tested explicitly.

For the last criterion, no code path anywhere writes a project config. Verified two ways: every file-writing call in the Go source is inside a test, and the adapter can only invoke the task, draft and init subcommands. A test then exercises every write the application can perform against a real project and asserts the config file is byte-identical afterwards and was not even rewritten with the same content.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Board columns are now the union of what every project declares, resolved in internal/board and exposed through the service.

Ordering is a weighted vote across projects with documented tie-break and cycle rules; a column records which projects declare it, which is what makes a project's absence an empty cell rather than an error, and what lets a move be refused with a reason.

Verified at three levels. Ten Go tests cover the algorithm on the shape of the real corpus, on disjoint lists, on a majority outvoting a dissenter, on a three-way cycle, and on stability across fifty runs. Against the nine real projects over the running bridge the layout comes back as To Do, In Progress, In Review, Done with In Review declared by one project and zero conflicts. Then in a browser on a two-project bench with deliberately different status lists: dragging a card into a status its project does not declare is refused before anything moves, with the notice 'Alpha has no In Review status, so TASK-2 cannot move there. Add it to that project's config.yml if it belongs there', and the card stays where it was.

The no-writing criterion was checked rather than asserted: every file-writing call in the Go source is inside a test, the adapter can only run task, draft and init subcommands, and a test exercises every write the application offers and then asserts the project's config.yml is byte-identical and was not even rewritten with the same content.
<!-- SECTION:FINAL_SUMMARY:END -->
