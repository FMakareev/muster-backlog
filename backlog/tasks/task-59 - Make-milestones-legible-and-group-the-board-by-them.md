---
id: TASK-59
title: Make milestones legible and group the board by them
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 19:43'
updated_date: '2026-08-26 20:11'
labels: []
milestone: m-2
dependencies:
  - TASK-21
priority: high
type: feature
ordinal: 59000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A card shows its milestone as a bare id such as m-1, which reads exactly like a task id and tells nobody what the milestone is. Milestones are also entities in their own right with their own titles and progress, and none of that is visible. Planning by milestone is the main way this backlog is organised, so the board has to speak in milestone names and be groupable by them.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A milestone is shown by its title, not by its id alone, wherever it appears
- [x] #2 Milestone and task are visually distinguishable at a glance
- [x] #3 The board can be grouped by milestone as well as by project
- [x] #4 Milestone progress is visible per project
- [x] #5 A task whose milestone does not resolve shows the raw value rather than nothing
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
A milestone now reads as its title in a bordered chip with a diamond, so it is distinguishable at a glance from the task id beside it. A bare m-1 was the whole problem: it looks exactly like a task id and says nothing.

Grouping cycles through none, project and milestone with the same g key rather than being a second control, and the choice is remembered in the settings file. Tasks with no milestone sort last, so an unplanned pile does not push planned work down the column.

Milestone progress comes from counting tasks in the project's own terminal status - the last one it declares - since that is the only definition of finished the format offers.

A defect found while verifying: milestones were shown as ids anyway, because the lookup read the nanostore imperatively inside a helper. That captured whatever was loaded at first render and never updated. The list is now passed in, so a component calling it subscribes and re-renders when the milestones arrive. And the deeper cause was worse: refresh never fetched them at all, so the store was empty - an earlier edit to that function had not applied and I had not checked.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Milestones are legible and the board groups by them.

A card shows the milestone's title in a marked chip rather than a bare id that reads like a task id; the panel does the same; and grouping cycles none, project, milestone on the g key with the choice remembered. Milestone progress is counted against each project's own terminal status.

Verified in a browser: a card that read '◇ m-0' now reads '◇ Ship the first cut', and pressing g twice reaches 'Grouped by milestone'.

Verification found two defects behind one symptom. The lookup read its store imperatively, so it never updated when milestones arrived; and refresh was not fetching milestones at all, because an earlier edit to it had silently not applied. Both fixed.
<!-- SECTION:FINAL_SUMMARY:END -->
