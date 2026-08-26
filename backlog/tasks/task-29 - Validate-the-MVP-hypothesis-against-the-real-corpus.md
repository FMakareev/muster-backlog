---
id: TASK-29
title: Validate the MVP against the real corpus
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 15:28'
labels: []
milestone: m-1
dependencies:
  - TASK-21
  - TASK-23
  - TASK-25
  - TASK-27
  - TASK-28
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: task
ordinal: 29000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The MVP exists to answer one question: does a standalone desktop app over several Backlog.md projects at once beat the per-repository tooling already in use - backlog browser and the VSCode extension, both of which handle one backlog at a time. Run it as the daily tool over the real projects, then write down the verdict before building further.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 MVP runs against the real projects as the primary tool for a documented trial period
- [ ] #2 Startup time, live update latency and memory use on the full corpus are measured and recorded
- [ ] #3 The verdict is written down with the evidence behind it, including what the per-repository tools still do better
- [ ] #4 Findings that change scope are reflected in the specification and roadmap before later milestones start
<!-- AC:END -->
