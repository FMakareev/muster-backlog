---
id: TASK-91
title: Rewrite the versioning promise now that 1.0 has shipped
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 18:51'
updated_date: '2026-08-28 18:53'
labels: []
dependencies: []
priority: high
type: docs
ordinal: 91000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
v1.0.0 is released and the project is keeping it, so everything written for life before 1.0 is now wrong.

CHANGELOG.md carries a section headed 'Versioning before 1.0' describing what a minor bump may break while the major version is zero, and an 'Unreleased' section that says 'Nothing released yet' directly below the generated 1.0.0 entry. release-please's config still carries bump-minor-pre-major and bump-patch-for-minor-pre-major, which govern nothing once a release has a major version. A test asserts the pre-1.0 wording is present.

The promise the prose already anticipated — 'at 1.0 the first list becomes a major-version promise instead' — now has to be made.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The versioning section states what a major, minor and patch release mean now, and the list of what counts as breaking survives as the major-version promise
- [x] #2 No part of the changelog claims nothing has been released
- [x] #3 The release configuration carries no setting that only applied before 1.0
- [x] #4 The tests assert the promise that is now in force rather than the one that expired
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Rewrite the two hand-written sections of CHANGELOG.md: a post-1.0 versioning promise, and no 'Unreleased' section, since release-please writes the entries and there is now something released.
2. Fix the sentence describing where release-please writes, which named the Unreleased heading.
3. Drop bump-minor-pre-major and bump-patch-for-minor-pre-major from release-please-config.json.
4. Move the test assertions from the pre-1.0 wording to the post-1.0 one, and add a guard against the pre-1.0 settings coming back.
5. Reframe TASK-49, whose 1.0 acceptance criteria are now claims the tag has already made.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The pre-1.0 prose had already written its own successor — 'at 1.0 the first list becomes a major-version promise instead' — so the list of what counts as breaking carried over unchanged and only the bump semantics were rewritten.

Two things were wrong beyond the wording. The 'Unreleased' section said 'Nothing released yet' immediately below the generated 1.0.0 entry, because release-please writes new sections under the title and never touched it. And the sentence describing where release-please writes named the Unreleased heading, which was never where it wrote.

Added a short account of how 1.0 was reached, because a project whose changelog jumps from nothing to 1.0.0 with a milestone still open invites the question, and the answer is release-please's default first version.

Dropped bump-minor-pre-major and bump-patch-for-minor-pre-major: both govern nothing once the major version is non-zero, and a dead setting in the file that looks like where policy lives will be read as policy.

Both new guards fault-injected: restoring the pre-1.0 heading and re-adding a pre-1.0 setting each fail with the reason.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
v1.0.0 shipped, so the changelog's 'Versioning before 1.0' section, its 'Nothing released yet' notice sitting under the 1.0.0 entry, and release-please's two pre-1.0 bump settings all described a state the project had left. Rewrote the section as the major-version promise the old text said it would become, carried the list of what counts as breaking over unchanged, recorded how 1.0 was actually reached, removed the settings that no longer govern anything, and corrected the README pointer, which sent readers to the top of a file where the policy is at the foot. The tests now assert the promise in force and fail if either the old wording or the old settings come back; both fault-injected. wails3 task lint: 0 issues; suite green.
<!-- SECTION:FINAL_SUMMARY:END -->
