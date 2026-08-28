---
id: TASK-89
title: Keep the frontend build from deleting the tracked placeholder
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 15:48'
updated_date: '2026-08-28 15:49'
labels: []
dependencies: []
priority: high
type: bug
ordinal: 89000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The frontend build empties frontend/dist before writing, which removes frontend/dist/.gitkeep — the tracked file that makes `//go:embed all:frontend/dist` resolve on a clone with no build history.

Seen after building a fresh clone: git reports the placeholder deleted. The tree is dirty after every build, and the first person who commits that deletion puts the repository back to a state where nothing that compiles package main works until the frontend has been built.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The placeholder is still there after a production frontend build, and git reports nothing deleted
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Confirm the build removes it.
2. Add a small vite plugin that writes the placeholder back after the bundle closes, with the reason in a comment.
3. Verify by building and checking both the file and git status.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Confirmed before fixing: removed the placeholder, ran the production build, and git reported it deleted. The build empties dist, so the placeholder had to be written back after the bundle closes rather than merely created once.

A vite plugin does it, in the frontend's own config, so it holds however the build is invoked — through the Taskfile, through wails3, or by hand in frontend/. apply: 'build' keeps it out of the dev server. After the fix the same build leaves the directory with assets, index.html and .gitkeep, and git reports nothing changed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The production frontend build empties dist, which deleted frontend/dist/.gitkeep — the tracked file that makes the go:embed pattern resolve on a clone with no build history. Added a small vite plugin that writes it back on closeBundle, with the reason in the config. Verified by building with the placeholder removed (git reported it deleted) and building again with the plugin (directory intact, git reports nothing).
<!-- SECTION:FINAL_SUMMARY:END -->
