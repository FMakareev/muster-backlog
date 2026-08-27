---
id: TASK-74
title: Read and write task comments
status: To Do
assignee: []
created_date: '2026-08-27 17:27'
labels: []
milestone: m-5
dependencies: []
priority: medium
type: feature
ordinal: 74000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The parser has handled the comments envelope since the format was mapped, and the panel shows what is there. Writing one means a terminal: task edit --comment appends, with --comment-author recording who said it.

Ten of 875 tasks carry comments today, which is either a fair measure of how useful they are or a measure of how awkward they have been to add. The cheap way to find out is to make adding one cost nothing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A comment can be added to a task from the panel
- [ ] #2 The author recorded is the person using the application, not a fixed string
- [ ] #3 New comments appear in the thread immediately, in the order the file holds them
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
- [ ] #5 Linters and formatters pass across Go and frontend
- [ ] #6 Automated tests cover the change and the suite is green
- [ ] #7 User-facing behaviour change is reflected in README or docs
- [ ] #8 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
