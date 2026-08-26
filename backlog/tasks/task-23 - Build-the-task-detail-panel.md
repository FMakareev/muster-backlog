---
id: TASK-23
title: Build the task detail panel
status: To Do
assignee: []
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 15:29'
labels: []
milestone: m-1
dependencies:
  - TASK-21
  - TASK-56
priority: high
type: feature
ordinal: 23000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The key screen. Its absence is precisely what disqualified the Obsidian plus Bases alternative: a table without fast access to the task body reads as a dump. The panel opens from anywhere and shows the whole task - description, acceptance criteria, implementation plan, notes, dependencies and file references.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Panel opens from a board card and closes without losing board state
- [ ] #2 Description, acceptance criteria, implementation plan and notes render as formatted markdown
- [ ] #3 Acceptance criteria display their checked state
- [ ] #4 Dependencies render as links that navigate to the referenced task
- [ ] #5 Project, milestone, priority, type, labels, assignee and updated date are visible
- [ ] #6 Panel content updates live when the underlying file changes
- [ ] #7 Panel opens and closes by keyboard
<!-- AC:END -->
