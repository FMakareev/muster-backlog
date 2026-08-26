---
id: TASK-52
title: Build the cross-project analytics dashboard
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:27'
updated_date: '2026-08-26 20:48'
labels: []
milestone: m-2
dependencies:
  - TASK-17
  - TASK-20
priority: medium
type: feature
ordinal: 52000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The backlog CLI overview command already defines a useful vocabulary from purely native data: status and priority breakdowns, average task age, stale tasks, tasks blocked on dependencies, and recent activity. This is the same view, but over every registered project at once and drillable, which is precisely what no single-repository tool can offer.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Dashboard shows status and priority breakdowns per project and in total
- [x] #2 Blocked tasks waiting on unmet dependencies are listed with the blocking task
- [x] #3 Stale tasks are surfaced with a configurable age threshold
- [x] #4 Average task age and recent activity are shown per project
- [x] #5 Milestone progress is shown per project
- [x] #6 Every figure drills through to the underlying task list
- [x] #7 All metrics derive from native Backlog.md data with no field of our own
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
The vocabulary is the one 'backlog overview' already established from native data - status and priority breakdowns, average age, stale work, blocked work - because a second way to describe the same backlog would only be a second thing to reconcile. What is new is that it spans every project at once and that every figure leads back to the tasks behind it.

Two definitions had to be chosen rather than assumed. Finished means the last status a project declares, which is the only definition the format offers. And average age counts open work only: a task closed a year ago is not an old task, it is a done one.

The first run on real data immediately showed the symptom this project started from. Section 1 of the specification recorded that 95 of 162 To Do tasks carried no priority. Across all nine projects now: 411 of 860 open tasks, 48 per cent.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The cross-project overview, opened with s.

Headline figures for everything at once - open tasks, how many carry no priority, average age of open work, how many wait on something unfinished, how many are untouched too long - then a per-project table with status breakdowns, then the blocked and stale tasks themselves, each row opening the task.

Verified against the nine real projects: 860 open tasks, 411 of them without a priority, average age 15 days, 106 blocked, 45 stale, and a nine-row per-project table. The blocked list names what each task waits on, and only unfinished dependencies count - a dependency that is done does not block.

Covered by tests on a fixture whose numbers are known exactly: counts, the stale threshold at three different values, blocked tasks distinguishing finished dependencies from open ones, and average age excluding finished work.

Verification found a real inconsistency: the project roll counted archived and completed tasks while every other surface counted live ones, so the roll said 62 where the overview said 51. Both now say the same, checked project by project.
<!-- SECTION:FINAL_SUMMARY:END -->
