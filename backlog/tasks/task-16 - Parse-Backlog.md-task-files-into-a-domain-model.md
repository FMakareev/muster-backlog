---
id: TASK-16
title: Parse Backlog.md task files into a domain model
status: To Do
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 15:26'
labels: []
milestone: m-1
dependencies:
  - TASK-14
priority: high
type: feature
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A frontmatter and body parser over the tasks, drafts, milestones, documents and decisions of one project. It must preserve the full task body - description, acceptance criteria, implementation plan, notes - because fast access to that body is what a table view cannot give. Only fields Backlog.md itself defines are modelled; Muster adds none of its own.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Parser reads tasks, drafts, milestones, documents, decisions and completed directories of a project
- [ ] #2 Domain model exposes only fields Backlog.md defines: status, priority, milestone, dependencies, ordinal, type, labels, assignee and dates
- [ ] #3 Per-project config.yml is parsed for statuses, priorities, types, labels and task prefix
- [ ] #4 Full body is retained with description, acceptance criteria, implementation plan and notes distinguishable
- [ ] #5 Acceptance criteria are parsed as individual items with their checked state
- [ ] #6 Title comes from frontmatter, never from the file name
- [ ] #7 A malformed file is skipped with a diagnostic and does not abort the scan
- [ ] #8 Tests run against the pinned reference corpus
<!-- AC:END -->
