---
id: TASK-90
title: Stop the draft release from making release-please forget what shipped
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 18:48'
updated_date: '2026-08-28 18:49'
labels: []
dependencies: []
priority: high
type: bug
ordinal: 90000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Cutting the release as a draft breaks release-please's accounting. A draft release records a tag name but GitHub does not create the tag until it is published, and release-please finds the previous release by its tag. In the same run that creates the draft it then computes the next release pull request, sees no tag for what it has just released, treats the whole history as unreleased, and opens a pull request for the next version whose changelog repeats every feature the project has ever had.

Observed on the first real release: merging the 1.0.0 pull request produced release v1.0.0 with all six artefacts attached and published — that half worked — and immediately a pull request for 1.1.0 listing every feat since the beginning, including the one released in 1.0.0.

The draft was chosen so nobody sees a version with nothing to download. That is worth less than release-please knowing what has shipped.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A release cut from a merged release pull request does not produce a pull request repeating already-released changes
- [x] #2 The artefacts still reach the release, and the reason the draft was abandoned is recorded where the next person will look
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
1. Confirm the mechanism from the published repository rather than from reasoning: the 1.1.0 branch's changelog repeats commits that are parents of the 1.0.0 release commit.
2. Go back to a published release: draft false, drop the publish step, keep the upload.
3. Keep the version check before the upload — that part is orthogonal and worth keeping.
4. Rewrite the workflow comment and the README so the tried-and-abandoned draft is recorded with its reason, not silently removed.
5. Update the guard that asserted the draft ordering to assert what now matters.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Confirmed from the published repository rather than by reasoning. Fetched it over HTTPS and read the 1.1.0 branch: its changelog lists every feat in the project's history, including feat(release) d58818d, which is a parent of the 1.0.0 release commit. So release-please genuinely treated the whole history as unreleased rather than double-counting one boundary commit.

Also confirmed the release half worked: v1.0.0 is published, not a draft, with all six assets attached — the AppImage among them, built on the runner at 79 MB against 126 MB here, so the runner bundles less. Nothing had been downloaded.

The mechanism, exactly: the action creates the releases and then computes the next release pull request in the same run. At that moment the release it has just made is a draft, so its tag does not exist, so it is not found, so nothing has shipped. The draft only becomes tagged minutes later when the artefacts job publishes it — too late for a decision already made. A later push does repair the pull request, but the wrong one sits there until then, and merging it would cut a release whose changelog repeats everything.

Reverted to a public release. The window it was meant to close is a few minutes of a version carrying notes and no downloads; that is worth much less than release-please knowing what shipped. The abandoned approach is recorded in the workflow header, in the README and in the test, because it is the obvious thing to reach for and it looks like it should work.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Cutting the release as a draft left it untagged until the artefacts job published it, and release-please finds the previous release by its tag — so in the same run it concluded nothing had ever shipped and opened a 1.1.0 pull request whose changelog repeated the entire history, verified by reading that branch in the published repository. The release is public again from the moment it is cut; the artefacts job still verifies the version and attaches everything with checksums. The guard that asserted the draft ordering now asserts the opposite and explains why, fault-injected by turning the draft back on.
<!-- SECTION:FINAL_SUMMARY:END -->
