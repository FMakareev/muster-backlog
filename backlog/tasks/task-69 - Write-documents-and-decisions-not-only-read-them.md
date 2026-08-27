---
id: TASK-69
title: 'Write documents and decisions, not only read them'
status: To Do
assignee: []
created_date: '2026-08-27 00:38'
labels: []
milestone: m-5
dependencies: []
priority: medium
type: feature
ordinal: 69000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The documents viewer renders what is there and offers no way to add to it. backlog doc create, doc update and decision create all exist.

This project's own conventions say to write a decision when a choice closes off alternatives - and writing one means leaving the application that is showing you the decisions. Note that doc update replaces the whole body, so editing has to send the complete document rather than a patch.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A document can be created in any project, with its type chosen from what the CLI accepts
- [ ] #2 A document's body can be edited and saved
- [ ] #3 A decision can be created with its status
- [ ] #4 Editing sends the whole body, because that is what doc update takes, and nothing is silently truncated
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
