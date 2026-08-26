---
id: TASK-49
title: Cut the 1.0 release
status: To Do
assignee: []
created_date: '2026-08-26 15:03'
updated_date: '2026-08-26 15:28'
labels: []
milestone: m-4
dependencies:
  - TASK-47
  - TASK-48
  - TASK-46
  - TASK-44
priority: high
type: task
ordinal: 49000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Version 1.0 is a promise about stability: the registry format, the label conventions and the write path stop changing without a major bump. Take it deliberately, once the planner has been used long enough to know the model holds.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Registry format and the CLI write contract are declared stable and documented as such
- [ ] #2 Post-1.0 breaking-change policy is written down
- [ ] #3 The commitment to add no field of our own to the Backlog.md format is stated as a project guarantee
- [ ] #4 CHANGELOG covers everything since the first pre-release
- [ ] #5 The 1.0 tag, release and artefacts are published through the release automation
- [ ] #6 Known limitations and explicit non-goals are stated in the release notes
<!-- AC:END -->
