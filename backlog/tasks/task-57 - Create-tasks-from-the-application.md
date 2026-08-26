---
id: TASK-57
title: Create tasks from the application
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 19:43'
updated_date: '2026-08-26 20:10'
labels: []
milestone: m-2
dependencies:
  - TASK-24
priority: high
type: feature
ordinal: 57000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A task manager that cannot create a task is a viewer. Today every task has to be created in a terminal, which breaks the loop the application exists to close: see the board, decide what is missing, add it. Creation goes through the backlog CLI like every other write, and the form offers what that project configures rather than a fixed list.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A task can be created from the board and from the project roll, into a chosen project
- [x] #2 The form covers title, description, status, priority, type, milestone, assignee, labels and acceptance criteria
- [x] #3 Status, priority and type choices come from the target project's own configuration
- [x] #4 The task is written by the backlog CLI and appears on the board once the rescan confirms it
- [x] #5 A failed creation explains why and loses nothing the person typed
- [x] #6 The form opens and submits by keyboard
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Tasks can be created from the application, which is what stops it being a viewer.

A form opened with n or from the top bar takes title, description, acceptance criteria one per line, status, priority, type, milestone, assignee and labels. Status, priority and type choices come from the target project's own configuration, and milestones from that project's milestones, so nothing is offered that the project cannot represent. The task is written by the backlog CLI, the id it assigns comes back rather than being predicted, and the panel opens on the new task once the rescan confirms it.

Verified in a browser against a bench: pressing n, filling the form and pressing ctrl+enter produced 'task-3 - Made-from-the-application.md' on disk with the description and both criteria in it, and the panel opened on that task. A failed creation leaves the form exactly as it was - losing what someone just typed because a command failed is the least forgivable thing a form can do.
<!-- SECTION:FINAL_SUMMARY:END -->
