---
id: TASK-75
title: Change many tasks at once
status: To Do
assignee: []
created_date: '2026-08-27 17:27'
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
- [ ] #1 Several tasks can be selected in the list
- [ ] #2 Status, priority, milestone and labels can be changed for the whole selection
- [ ] #3 The result says what was changed and names anything that failed
- [ ] #4 A selection spanning projects is handled, or refused with the reason
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
- [ ] #5 Linters and formatters pass across Go and frontend
- [ ] #6 Automated tests cover the change and the suite is green
- [ ] #7 User-facing behaviour change is reflected in README or docs
- [ ] #8 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
