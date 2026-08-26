---
id: TASK-21
title: Render the multi-project kanban board
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 15:26'
labels: []
milestone: m-1
dependencies:
  - TASK-20
priority: high
type: feature
ordinal: 21000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The central screen: one board over every registered project at once. This is what the VSCode extension cannot do - it switches between backlog folders rather than combining them. Built on SVAR Svelte Kanban (MIT, native Svelte 5, virtualised) so drag-and-drop, grouping and large-board performance come from the library instead of being hand-built.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 SVAR Svelte Kanban is integrated and themed consistently with the application
- [ ] #2 Board renders tasks from every registered project in the unified columns
- [ ] #3 Project of origin is visible on every card and tasks can be grouped by project
- [ ] #4 Cards show title from frontmatter, priority, type, labels, assignee and milestone
- [ ] #5 Board reflects file system changes live without user action
- [ ] #6 Board stays responsive with the full corpus of roughly 640 tasks loaded
- [ ] #7 Clicking a card opens the task panel
<!-- AC:END -->
