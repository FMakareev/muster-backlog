---
id: TASK-42
title: Build the drafts inbox and triage view
status: To Do
assignee: []
created_date: '2026-08-26 15:02'
updated_date: '2026-08-26 15:28'
labels: []
milestone: m-3
dependencies:
  - TASK-24
  - TASK-41
priority: medium
type: feature
ordinal: 42000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Backlog.md keeps drafts off the board by design, which is what makes capture cheap - and also what makes an unopened drafts directory the new inbox nobody reads. The triage view is where the inbox is emptied: promote, edit, reassign or discard.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Inbox lists drafts across every project with their capture date
- [ ] #2 A draft can be promoted to a task through the backlog CLI
- [ ] #3 A draft can be edited before promotion and reassigned to another project
- [ ] #4 A draft can be archived or discarded with confirmation
- [ ] #5 The view shows how long drafts have been waiting so a growing inbox is visible
- [ ] #6 Inbox depth is visible from the main navigation
<!-- AC:END -->
