---
id: TASK-30
title: Package and document a v0.1 developer build
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-1
dependencies:
  - TASK-25
  - TASK-26
priority: medium
type: chore
ordinal: 30000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The MVP has to be runnable by someone other than its author for the open-source promise to mean anything, and by me without a development environment for the trial period. A local Linux build with documented steps is enough at this stage; distribution artefacts come with 1.0.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A documented command produces a runnable Linux artefact
- [ ] #2 Runtime prerequisites including the backlog CLI version are documented
- [ ] #3 First-run instructions cover creating the registry and adding a project
- [ ] #4 A smoke checklist covering board, panel, status write and capture is documented and passes on the artefact
<!-- AC:END -->
