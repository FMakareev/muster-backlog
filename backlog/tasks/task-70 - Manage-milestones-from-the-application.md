---
id: TASK-70
title: Manage milestones from the application
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-27 00:38'
updated_date: '2026-08-27 01:39'
labels: []
milestone: m-5
dependencies: []
priority: medium
type: feature
ordinal: 70000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Milestones are the axis this backlog is planned on - 49 across the nine projects, the board groups by them, cards and the projects screen show their progress - and there is no way to make one, rename one or retire one without a terminal.

The CLI has add, rename, remove and archive. rename updates the tasks that point at the milestone by default, and remove asks what should happen to them: clear, keep or reassign. Both of those choices have to be surfaced rather than decided quietly, because they rewrite other files.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A milestone can be added to a project from the application
- [x] #2 A milestone can be renamed, and what happens to the tasks pointing at it is shown before it is done
- [x] #3 A milestone can be archived or removed, with the choice of what becomes of its tasks made explicitly
- [x] #4 Milestone progress on the board and the projects screen follows immediately
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
Measured before designing, and the measurement changed the design. milestone archive and milestone remove both end with the file in archive/milestones - neither deletes anything. The difference between them is entirely in what happens to the tasks that named the milestone: archive leaves them pointing at it, remove clears them, keeps them or reassigns them. So the interface offers one act, retiring, and asks the question that actually differs - what becomes of the tasks - instead of making someone choose between two words that sound like they mean different things to the file.

The form says how many tasks are affected before the choice is made, and says that the file is archived either way. Reassigning to a milestone that does not exist is refused before anything runs, because the tasks would have nowhere to go.

Renaming keeps the id, which is what keeps the tasks attached: they reference m-0, not the title. The CLI updates referencing tasks by default and that default is left on - a milestone renamed while its tasks still name the old one is a rename that only half happened.

A real bug surfaced while testing this: Milestones() returned archived milestones alongside live ones, because it walked the whole scanned slice without looking at the class. Retiring one from the application made it obvious - it stayed on the list. Archived milestones are now out of the list, the board's grouping and the places a task can be assigned, which is what retiring one is supposed to mean.

Verified in the browser against a scratch bench with the files read after each write: a milestone added with its description, renamed with its tasks still attached, and retired with its one task reassigned to another - the file in archive/milestones afterwards, and the progress figures following immediately. Four Go tests against the real CLI cover all three task-handling answers. No axe-core violations across the screen and both states of the form.

One thing to note about the checks themselves: an assertion failed for a second time this session because innerText applies text-transform and the sentence was styled as an uppercase micro label. It was the wrong style for a sentence anyway, so it is a sentence now.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Milestones can be added, renamed and retired from the Projects screen.

Retiring is the interesting one. Backlog.md has two commands for it and both archive the file; what actually differs is the fate of the tasks that named the milestone, so the interface asks that question - leave them, clear them, or move them to another milestone - and says how many tasks are involved before anything happens. Renaming keeps the id, which is what keeps the tasks attached.

A bug fell out of building it: archived milestones were listed alongside live ones everywhere, because the list walked the whole scanned slice without looking at the class. Retiring one and watching it stay on screen is what showed it.

Verified in the browser with the files read after every write, four Go tests against the real CLI covering all three task-handling answers, and a clean accessibility pass.
<!-- SECTION:FINAL_SUMMARY:END -->
