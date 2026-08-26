---
id: TASK-51
title: Build cross-project search
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:27'
updated_date: '2026-08-26 20:47'
labels: []
milestone: m-2
dependencies:
  - TASK-17
  - TASK-20
priority: high
type: feature
ordinal: 51000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
With several projects loaded, finding a task by memory of its title is the most common navigation. Search runs over titles and bodies of tasks, drafts, documents and decisions across every registered project, and is reachable from anywhere by keyboard.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Search matches task titles and bodies across every registered project
- [x] #2 Documents, decisions and drafts are searchable alongside tasks
- [x] #3 Results show the owning project and the kind of item found
- [x] #4 Search opens by keyboard shortcut and selects a result without a mouse
- [x] #5 Selecting a result opens the corresponding panel or viewer
- [x] #6 Search over the full corpus returns within a documented time budget
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Search across every project, opened with / from anywhere.

It looks through tasks, drafts, milestones, documents and decisions of every registered project, matching titles, ids and bodies. Title and id hits rank above body hits, because someone typing a few words is usually reaching for something they remember by name; a body hit carries an excerpt so it can explain why it matched. Arrows move, Enter opens, Escape closes, and a slower earlier query can never overwrite a later one's results.

Verified against the nine real projects and 890 tasks: 'lefthook' found the task by title and Enter opened it in the panel; 'pkg-config', a phrase that appears only in a body, found the task with the surrounding text as an excerpt.
<!-- SECTION:FINAL_SUMMARY:END -->
