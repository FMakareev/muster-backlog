---
id: TASK-12
title: Establish docs structure and an ADR log
status: To Do
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-0
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: medium
type: docs
ordinal: 12000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Decisions taken now - the beta shell, the CLI-only write path, the review-cost model - will be questioned later by contributors and by me. An architecture decision record log keeps the reasoning attached to the repository instead of to a chat transcript. Seed it with the decisions already made.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 docs/ layout is defined and documented, covering specification, roadmap and decisions
- [ ] #2 ADR template and numbering convention are checked in
- [ ] #3 ADR records the desktop shell choice
- [ ] #4 ADR records the license and release automation choice
- [ ] #5 ADR records that writes go only through the backlog CLI and reads go directly to markdown
- [ ] #6 The v0.1 specification is published under docs/ and linked from the README
<!-- AC:END -->
