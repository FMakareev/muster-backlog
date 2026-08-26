---
id: TASK-55
title: Initialise a Backlog.md project in any folder from the UI
status: To Do
assignee: []
created_date: '2026-08-26 15:28'
labels: []
milestone: m-3
dependencies:
  - TASK-24
  - TASK-15
priority: high
type: feature
ordinal: 55000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Adding a project should not require dropping into a terminal. The backlog init command is fully non-interactive - every prompt it asks has a corresponding flag, plus a defaults switch - so the UI can present the same choices as a form instead of emulating a dialogue. This replaces any notion of a special personal backlog: any folder can become a project, including one that is not a git repository.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A folder can be picked and initialised as a Backlog.md project from the UI
- [ ] #2 The form covers the choices backlog init exposes: project name, backlog directory, config location, task prefix, zero-padded ids, git integration, agent instructions and integration mode
- [ ] #3 Each option shows what it does and what the default is, so the form is usable without reading CLI help
- [ ] #4 Initialisation runs through the backlog CLI and its output is surfaced on failure
- [ ] #5 A folder that already contains a backlog is detected and offered for registration instead of initialisation
- [ ] #6 A folder that is not a git repository is supported through the no-git path
- [ ] #7 The newly initialised project appears in the registry and on the board without a restart
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
