---
id: TASK-52
title: Build the cross-project analytics dashboard
status: To Do
assignee: []
created_date: '2026-08-26 15:27'
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
- [ ] #1 Dashboard shows status and priority breakdowns per project and in total
- [ ] #2 Blocked tasks waiting on unmet dependencies are listed with the blocking task
- [ ] #3 Stale tasks are surfaced with a configurable age threshold
- [ ] #4 Average task age and recent activity are shown per project
- [ ] #5 Milestone progress is shown per project
- [ ] #6 Every figure drills through to the underlying task list
- [ ] #7 All metrics derive from native Backlog.md data with no field of our own
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
