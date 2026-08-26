---
id: TASK-18
title: Watch project directories with fsnotify
status: To Do
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-1
dependencies:
  - TASK-17
priority: high
type: feature
ordinal: 18000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Liveness is the reason this is an application and not a report: an agent changes a status and the screen follows without polling. Watch each project task directory, debounce the bursts that editors and the CLI produce, and reload only what changed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Each registered project task directory is watched for create, write, rename and delete
- [ ] #2 Rapid bursts are debounced into a single reload
- [ ] #3 Only the affected project is rescanned on a change
- [ ] #4 Watches survive directory recreation and editor atomic-save patterns
- [ ] #5 A project directory that disappears is reported without terminating the other watches
- [ ] #6 Watcher shuts down cleanly with no leaked goroutines or descriptors
<!-- AC:END -->
