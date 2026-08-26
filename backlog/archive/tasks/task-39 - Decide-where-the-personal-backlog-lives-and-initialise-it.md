---
id: TASK-39
title: Decide where the personal backlog lives and initialise it
status: To Do
assignee: []
created_date: '2026-08-26 15:02'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-3
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: task
ordinal: 39000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Specification section 8, item 4. Either backlog init inside the Obsidian vault, which puts personal tasks into the knowledge base and gets autocommits through obsidian-git for free, or a separate directory under the local data path. The choice affects backup, sync and whether personal tasks are searchable alongside notes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Decision is recorded as an ADR with the backup and sync consequences stated
- [ ] #2 The personal backlog is initialised at the chosen location
- [ ] #3 Existing personal tasks scattered in the vault inbox are migrated into it
- [ ] #4 The personal project is registered and loads like any other project
<!-- AC:END -->
