---
id: TASK-59
title: Make milestones legible and group the board by them
status: To Do
assignee: []
created_date: '2026-08-26 19:43'
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
- [ ] #1 A milestone is shown by its title, not by its id alone, wherever it appears
- [ ] #2 Milestone and task are visually distinguishable at a glance
- [ ] #3 The board can be grouped by milestone as well as by project
- [ ] #4 Milestone progress is visible per project
- [ ] #5 A task whose milestone does not resolve shows the raw value rather than nothing
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
