---
id: TASK-45
title: Expose cross-project context over an MCP server
status: To Do
assignee: []
created_date: '2026-08-26 15:03'
updated_date: '2026-08-26 15:28'
labels: []
milestone: m-4
dependencies:
  - TASK-17
  - TASK-24
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: low
type: feature
ordinal: 45000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
With MCP an agent talks to Muster instead of one CLI per repository, and gains context no single project has: what is in flight elsewhere, which milestone is active, where the dependencies point. Built on the Go MCP SDK. Strictly optional - the application is complete without it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 MCP server exposes the aggregated task set, the registry, milestones, documents and decisions as read tools
- [ ] #2 Tools report per-project counts by status so an agent can see current load
- [ ] #3 Any write operation exposed goes through the same CLI adapter as the UI
- [ ] #4 The server can be enabled and disabled independently of the desktop window
- [ ] #5 Access is confined to the registered projects
- [ ] #6 Connecting an agent to the server is documented end to end
<!-- AC:END -->
