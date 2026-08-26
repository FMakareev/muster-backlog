---
id: TASK-30
title: Acceptance criteria written without the per-item index
status: Done
assignee:
  - '@claude'
created_date: '2026-07-23 15:50'
updated_date: '2026-07-24 09:12'
labels: []
dependencies: []
ordinal: 30000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Placeholder prose. SYNTHETIC VARIANT. Every acceptance-criteria line in the surveyed corpus
carries a `#N` index, but backlog 1.48.0 still ships a `migrateToStableFormat` path for
index-less items, so a reader must tolerate their absence and fall back to positional
numbering.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] A criterion with no index
- [ ] Another criterion with no index
<!-- AC:END -->
