---
id: TASK-48
title: Prepare public documentation and demo material
status: To Do
assignee: []
created_date: '2026-08-26 15:03'
updated_date: '2026-08-26 15:28'
labels: []
milestone: m-4
dependencies:
  - TASK-44
  - TASK-52
  - TASK-50
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: medium
type: docs
ordinal: 48000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The niche is narrow but real: backlog browser is single-repo by design, and the VSCode extension switches between backlog folders rather than combining them. Nobody discovers that from a paragraph - the multi-project board and the analytics view have to be shown.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 README shows the board, the list view and the analytics dashboard in a way that renders on GitHub
- [ ] #2 A usage guide covers registry, board, list, search, inbox, documents and analytics
- [ ] #3 The relationship to Backlog.md is stated: reads markdown directly, writes only through the CLI, adds no field of its own
- [ ] #4 Differences from backlog browser and from the VSCode extension are stated plainly and fairly
- [ ] #5 Screenshots contain no private project data
<!-- AC:END -->
