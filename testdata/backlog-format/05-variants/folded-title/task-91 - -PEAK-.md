---
id: TASK-91
title: >-
  A title long enough that the YAML writer emits it as a folded block scalar instead of an
  inline quoted string, which a line-oriented frontmatter reader will read as an empty title
status: Done
assignee:
  - '@claude'
created_date: '2026-08-07 00:37'
updated_date: '2026-08-07 00:43'
labels: []
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Placeholder prose. 30 of 1021 files in the surveyed corpus use `title: >-`. The filename here
also shows the second hazard: this file was created by an older CLI whose slugifier dropped
non-ASCII characters, so only the ASCII fragment survived in the name.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Placeholder criterion
<!-- AC:END -->
