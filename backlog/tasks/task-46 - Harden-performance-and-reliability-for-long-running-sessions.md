---
id: TASK-46
title: Harden performance and reliability for long-running sessions
status: To Do
assignee: []
created_date: '2026-08-26 15:03'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-4
dependencies:
  - TASK-29
priority: medium
type: enhancement
ordinal: 46000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The application is meant to stay open all day over seven repositories with agents writing into them continuously. Whatever leaks, degrades or drifts over hours will not show up in a short trial, so it needs deliberate exercise before a 1.0 promise.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Memory and descriptor use stay flat over a documented long-running session
- [ ] #2 Sustained external write bursts do not produce runaway rescans
- [ ] #3 A repository operation that rewrites many files at once, such as a branch switch, is handled without corrupting store state
- [ ] #4 Startup and live-update latency on the full corpus are measured against the recorded MVP baseline
- [ ] #5 Regressions in these figures are caught by a repeatable check
<!-- AC:END -->
