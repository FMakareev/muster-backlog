---
id: TASK-58
title: Edit a task body from the panel
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 19:43'
updated_date: '2026-08-26 20:10'
labels: []
milestone: m-2
dependencies:
  - TASK-24
  - TASK-23
priority: high
type: feature
ordinal: 58000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The panel shows description, acceptance criteria, implementation plan and notes but cannot change any of them, so any real edit still means opening the file elsewhere. Editing them here is what makes the panel a working surface rather than a reader. Every write goes through the backlog CLI, which owns the section markup.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Description, implementation plan and notes can be edited and saved from the panel
- [x] #2 Acceptance criteria can be added, removed, reordered and ticked
- [x] #3 The title can be changed
- [x] #4 Editing shows the markdown source and saves it verbatim, without reformatting what was there
- [x] #5 A save that fails leaves the text on screen rather than discarding it
- [x] #6 An edit made elsewhere while the panel is open does not silently overwrite what is being typed
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Acceptance criteria are replaced wholesale rather than patched item by item. That makes adding, removing and reordering one operation, and it keeps the CLI's per-item index - which it renumbers on every insertion - out of the interface entirely. Ticking still uses that index, because that is what the CLI's check-ac takes.

Editing shows the markdown source and saves it verbatim. This is a file another tool wrote and will read again, so reformatting it on the way through would be a change nobody asked for.

The concurrent-edit case is handled by comparing what the file said when editing started against what it says now. When an agent writes to the same task mid-edit, the editor says so before the save replaces it, rather than discarding the other change silently.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The task panel edits what it shows: title, description, implementation plan, implementation notes and acceptance criteria.

Each section has its own edit control, shows the markdown source while editing and saves it verbatim through the backlog CLI. Criteria are replaced as a list, so adding, removing and reordering are one operation. A failed save leaves the text on screen. If the file changed underneath while someone was typing, the editor says so before overwriting.

Verified by driving each path separately against a bench and reading the file afterwards: the description, the plan, a three-item criteria replacement renumbered #1 to #3, and a rename all reached disk.
<!-- SECTION:FINAL_SUMMARY:END -->
