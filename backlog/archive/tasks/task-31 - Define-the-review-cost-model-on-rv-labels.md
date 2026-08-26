---
id: TASK-31
title: Define the review cost model on rv labels
status: To Do
assignee: []
created_date: '2026-08-26 15:02'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-2
dependencies:
  - TASK-24
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 31000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Specification section 4: review cost is expressed as a label - rv:s equals one point, rv:m three, rv:l six, and an unlabelled task defaults to three. The labels field is empty in all 640 tasks, so it is free to use and requires no new format. This task fixes the model and makes cost readable and editable in the application.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Cost mapping is implemented with the unlabelled default applied consistently
- [ ] #2 A task carrying contradictory rv labels resolves deterministically and is reported
- [ ] #3 Review cost is visible on the card and in the task panel
- [ ] #4 Cost can be set from the task panel through the CLI write adapter
- [ ] #5 The cost model and the meaning of each level are documented for contributors
<!-- AC:END -->
