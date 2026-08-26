---
id: TASK-24
title: Build the backlog CLI write adapter
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-1
dependencies:
  - TASK-17
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 24000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Specification section 5: every write goes through the backlog CLI, never through a writer of our own, because the file format is alive and would have to be chased at each CLI release. This adapter is the single choke point - it locates the binary, runs commands in the right working directory, and turns failures into something the UI can show.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Adapter resolves the backlog binary and verifies a supported version at startup
- [ ] #2 Commands run with the target project as the working directory
- [ ] #3 Non-zero exits and stderr are captured and surfaced as structured errors
- [ ] #4 Arguments are passed without shell interpolation
- [ ] #5 A missing or unsupported CLI is reported once, clearly, instead of failing per action
- [ ] #6 Every write is followed by a rescan so the store reflects the CLI result rather than an assumption
- [ ] #7 Concurrent writes to the same project are serialised
<!-- AC:END -->
