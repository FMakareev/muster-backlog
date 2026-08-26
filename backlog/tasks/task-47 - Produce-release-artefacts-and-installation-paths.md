---
id: TASK-47
title: Produce release artefacts and installation paths
status: To Do
assignee: []
created_date: '2026-08-26 15:03'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-4
dependencies:
  - TASK-8
  - TASK-11
  - TASK-30
priority: high
type: chore
ordinal: 47000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A 1.0 that only builds from source reaches nobody. Release automation must attach usable artefacts, and installation has to be documented for a user who has never seen a Wails application.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Tagged releases attach built artefacts produced by CI rather than by a local machine
- [ ] #2 Artefacts are verifiable through published checksums
- [ ] #3 Installation and upgrade instructions are documented per supported platform
- [ ] #4 Supported platforms and their prerequisites are stated explicitly, including what is untested
- [ ] #5 A released artefact passes the smoke checklist on a clean machine
<!-- AC:END -->
