---
id: TASK-8
title: 'Set up SemVer versioning, CHANGELOG and release-please'
status: To Do
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 19:32'
labels: []
milestone: m-4
dependencies:
  - TASK-5
priority: high
type: chore
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Versioning and the change log must be derived from history, not maintained by hand. release-please reads Conventional Commits, opens a release pull request, bumps the version, writes CHANGELOG.md and tags the release. Pre-1.0 semantics need stating explicitly so users know what a minor bump means today.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 CHANGELOG.md exists in Keep a Changelog form with an Unreleased section
- [ ] #2 release-please workflow opens a release pull request from conventional commits on the default branch
- [ ] #3 Merging that pull request produces a SemVer tag, a GitHub release and an updated CHANGELOG.md
- [ ] #4 The version reported by the built binary matches the released tag
- [ ] #5 Pre-1.0 versioning policy is documented, including what counts as breaking before 1.0
<!-- AC:END -->
