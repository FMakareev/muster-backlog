---
id: TASK-29
title: Validate the MVP against the real corpus
status: In Progress
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 19:45'
labels: []
milestone: m-1
dependencies:
  - TASK-21
  - TASK-23
  - TASK-25
  - TASK-27
  - TASK-28
documentation:
  - backlog/docs/doc-6 - MVP-v0.1-verdict.md
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
- [x] #2 Startup time, live update latency and memory use on the full corpus are measured and recorded
- [x] #3 The verdict is written down with the evidence behind it, including what the per-repository tools still do better
- [x] #4 Findings that change scope are reflected in the specification and roadmap before later milestones start
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Measurements recorded in doc-6: 1027 entities parsed in 101 ms with zero diagnostics, full load of nine projects 105 ms, single-project reload about 11 ms, filtered query 274 microseconds, 52 to 57 cards in the DOM against 884 loaded, full re-sort 95 ms, panel open 90 ms, 18 MB heap, zero network requests. Nothing there is a constraint - the premise that a thousand markdown files can live in memory and be re-read on every change is correct.

The verdict is not yet, in the author's words: not close to the extension, but already quite convenient. That is two findings. The multi-project view is worth having and no existing tool does it. But the MVP is a viewer that can move cards rather than a task manager: no task creation, no editing of a body, no search, no filters, milestones shown as bare ids that read like task ids, subtasks invisible, and reading uncomfortable on a wide screen because the panel is a column against the right edge.

Seven tasks came out of it. Four were already planned - search, filters - and three are new: creating tasks (TASK-57), editing a body (TASK-58), legible and groupable milestones (TASK-59), plus a centred reading mode (TASK-60) and subtasks on the board (TASK-61). m-2 was renamed from 'Views, filters and analytics' to 'Daily use: editing, views and filters' because that is now what it is for, and both the specification and the roadmap say so.

The first acceptance criterion is deliberately left unchecked. It asks for the MVP to be the primary tool over a documented trial period, and what has happened is a first use of perhaps an hour. A first impression of one's own work is the least reliable evidence available, and checking that criterion off would be claiming something that did not happen. The verdict stays provisional until there is a real stretch of days.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @claude
created: 2026-08-26 19:45
---
Leaving this In Progress rather than closing it. Everything except the trial period is done and written up in doc-6, but the criterion that matters most needs days of real use, which is yours to give rather than mine to manufacture. Revisit after m-2.
---
<!-- COMMENTS:END -->
