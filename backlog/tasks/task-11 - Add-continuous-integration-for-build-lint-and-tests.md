---
id: TASK-11
title: 'Add continuous integration for build, lint and tests'
status: To Do
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 15:25'
labels: []
milestone: m-0
dependencies:
  - TASK-6
  - TASK-1
priority: high
type: chore
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
CI is the shared definition of green. It must run everything the hooks run plus the full build, so that a passing pull request means the desktop binary actually compiles on a clean machine.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Workflow runs on pull requests and on pushes to the default branch
- [ ] #2 Pipeline runs Go and frontend linting, unit tests and a full desktop build
- [ ] #3 Go and Node toolchain versions in CI match those documented for local development
- [ ] #4 Dependency and build caching keeps a typical run within a documented time budget
- [ ] #5 Required status checks are configured so a red pipeline blocks merge
<!-- AC:END -->
