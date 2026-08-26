---
id: TASK-2
title: Assess Wails v3 beta risk and pin a rollback plan
status: To Do
assignee: []
created_date: '2026-08-26 14:59'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-0
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: spike
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The specification (section 8, item 1) flags Wails v3.0.0-beta.8 as a deliberate decision rather than inertia. Establish what beta actually costs here: API churn between betas, Linux WebKit runtime quirks, packaging maturity, and what a migration away would take. The outcome is a written decision, not code.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 ADR records the chosen version, why beta is acceptable, and the concrete risks accepted
- [ ] #2 Upgrade policy is written down: when beta bumps are taken and how they are verified
- [ ] #3 Rollback plan names the fallback shell and estimates the work to switch
- [ ] #4 Findings that change the specification are reflected back into it
<!-- AC:END -->
