---
id: TASK-75
title: Change many tasks at once
status: Done
assignee:
  - '@claude'
created_date: '2026-08-27 17:27'
updated_date: '2026-08-27 19:49'
labels: []
milestone: m-5
dependencies: []
priority: medium
type: feature
ordinal: 75000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
With 875 live tasks across nine projects, some changes are only sensible in bulk: giving a milestone to everything that shares a label, retiring a label, moving a set of tasks to another status.

Every write still goes through the CLI one task at a time, because that is the only writer there is. What bulk means here is choosing the set once and describing the change once, and being told plainly what happened - including which ones failed, since a run of twenty writes can partly fail and pretending otherwise would be worse than not offering it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Several tasks can be selected in the list
- [x] #2 Status, priority, milestone and labels can be changed for the whole selection
- [x] #3 The result says what was changed and names anything that failed
- [x] #4 A selection spanning projects is handled, or refused with the reason
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
- [x] #5 Linters and formatters pass across Go and frontend
- [x] #6 Automated tests cover the change and the suite is green
- [x] #7 User-facing behaviour change is reflected in README or docs
- [x] #8 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Measured first, on a two-project bench.

One backlog task edit takes one task id - the usage line says [taskId], singular - so bulk here means choosing the set once and describing the change once, not one command. But one call can carry status, priority, milestone and labels together, and it validates everything before writing anything: an edit with a good status and a bad priority left the file untouched and exited 1. So it is one CLI call per task rather than four, and each task either takes the whole change or none of it.

The trap is the milestone. -m m-99 does not fail. It writes milestone: m-99 into the file, and the milestone list then invents an entry m-99: m-99 with no file behind it. A title works the same way: -m 'Totally made up' writes that string. Milestone ids are per project, so applying one across projects would silently plant dangling references in every project but one.

1. backlogcli: EditTask(dir, id, TaskChange) building one task edit from whatever is set. Status, priority and milestone are pointers - nil leaves the field alone, a set empty milestone means --clear-milestone. Labels are add and remove lists, which combine in one call.
2. app: ChangeMany(BulkChange) - group by project, one edit per task, collect a failure per task with the CLI's own words, reload each touched project once at the end rather than after every write, and emit progress since 280ms per task means a forty-task run takes eleven seconds.
3. Refuse a milestone for a selection spanning projects, in the service and not only in the form, for the reason above. Status and priority are offered as the intersection of what every selected project configures, so the form cannot offer a value that half the selection will reject.
4. List: a checkbox column, shift-click for a range, a header box for everything shown. The selection is held as refs but only ever acted on through what is currently visible, so changing a filter cannot leave invisible tasks in the set.
5. A bar above the table with the controls and, afterwards, what happened: how many changed and every failure named.
6. Go tests over a real bench for the mixed case, the cross-project case and the milestone refusal; browser verification reading the files; README.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
One `backlog task edit` takes one task id, so bulk here is the choosing rather than the command. But one call carries status, priority, milestone and labels together and validates all of them before writing anything - an edit with a good status and a bad priority exits 1 and leaves the file untouched - so it is one call per task and each task takes the whole change or none of it. `--add-label` and `--remove-label` combine, and neither disturbs a label it does not name.

The milestone is the trap, and it is why the cross-project case is a refusal rather than a preference. `task edit -m m-99` does not fail. It writes `milestone: m-99` into the file, exits 0, and `milestone list` then shows an entry `m-99: m-99` with no file behind it. A title works the same way: `-m "Totally made up"` writes that string. Milestone ids belong to one project, so applying one across a selection would plant a dangling reference in every project but one and report success for each. The whole run is refused before any of it starts, in the service and not only in the form, and the refusal names the projects. Status, priority and labels cross projects perfectly well and are not refused.

Nothing is offered that half the selection would reject: status and priority are the intersection of what every chosen project configures, not the union.

Two smaller measured limits. A priority cannot be cleared - `--priority ""` exits 0 and changes nothing, the same shape as `--modified-file ""` - so the form offers no way to clear one rather than a control that reports success and does nothing. And Backlog.md lower-cases the priority in the file whatever the configured spelling is: High goes in, `priority: high` comes out. A verification assertion that expected the spelling back failed on that, and the assertion was wrong, not the code.

Timing shaped two decisions. An edit takes about 280ms, so forty tasks is eleven seconds: the button counts through the run from an event, rather than sitting silent. And the project is re-read once per project at the end instead of once per task, because twenty rescans to learn the same thing would be most of the run.

The selection is held as refs but only ever acted on through what is currently visible, so narrowing a filter cannot leave invisible tasks in the set. Shift only ever adds to a range - a shift-click that unticked whatever it crossed would destroy a selection that took work to build. Removing a label is a row of the labels actually present on the chosen tasks rather than a text field, because retiring a label is the case that wants it.

A NUL byte reached the source while writing the "leave this field alone" sentinel, invisible in the file and working by accident. It is written as an explicit
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Tasks are ticked in the list - shift extends a range, the header box takes everything shown - and one form describes the change once: status, priority, milestone, labels to add and labels to remove.

Every write still goes through the CLI one task at a time, so the result reports how many took the change and names every one that did not in the words the CLI gave. One call per task carries the whole change, which is what makes each task take all of it or none: the CLI validates every field before writing anything.

Nothing is offered that half the selection would reject - status and priority are the intersection of what every chosen project configures. A milestone across projects is refused outright, and that is measured rather than cautious: task edit -m accepts an id the project does not have, writes it and reports success, so across a selection it would plant a dangling reference in every project but one.

Verified in a browser against a two-project bench reading the files off disk, thirteen checks including a run made to partly fail by deleting a task file under it; five Go tests at the service; axe-core clean in both selection states.
<!-- SECTION:FINAL_SUMMARY:END -->
