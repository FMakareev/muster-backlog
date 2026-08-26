---
id: TASK-19
title: Bridge backend state and events to the frontend
status: To Do
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-1
dependencies:
  - TASK-17
  - TASK-18
priority: high
type: feature
ordinal: 19000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Svelte layer needs the aggregated store and a live change feed. Define the binding surface between Go and the frontend once, so screens are added later without renegotiating the contract each time.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Frontend can request the full aggregated task set and the project registry
- [ ] #2 Backend emits an event on store change and the frontend updates without a reload
- [ ] #3 Shared types are generated rather than hand-mirrored on both sides
- [ ] #4 Backend errors surface to the frontend as structured, displayable messages
<!-- AC:END -->
