---
id: TASK-17
title: Aggregate all projects into one in-memory store
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-1
dependencies:
  - TASK-16
  - TASK-15
priority: high
type: feature
ordinal: 17000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The multi-repo view needs one queryable collection across every registered project, held in memory - 640 files are milliseconds to load and no database is warranted. Task identity must be qualified by project, because task IDs collide across repositories.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Store holds tasks from every registered project keyed by a project-qualified identity
- [x] #2 Store answers queries by status, project, milestone, priority, type and label
- [x] #3 Cross-project ID collisions are handled without shadowing
- [x] #4 Reload of a single project updates the store without rescanning the others
- [x] #5 Loading the full corpus at startup stays within a documented time budget
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/store holds one aggregated, in-memory view across every registered project, guarded by a read-write mutex because the fsnotify watcher in TASK-18 will reload projects while the UI reads.
2. Identity is a Ref of project path plus the entity Key that already carries kind, class and id. Ids collide across projects in 200 of 351 cases and one id names two different tasks inside a single project, so nothing may be keyed on id alone.
3. Query is one struct of optional filters combined with AND, values compared case-insensitively because enum case is not normalised on disk - priority is written lowercase against a capitalised config.
4. Reload replaces exactly one project's slice and touches nothing else, so a file change costs one project rescan rather than a full corpus reload.
5. Expose each project's declared status list as raw material. The union algorithm that turns them into board columns is TASK-27's and does not belong here.
6. Measure load time over the real nine-project corpus and write the figure down, since the criterion is a documented budget rather than a feeling.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Identity is a Ref of project path plus the backlog.Key that already carries kind, class and id. Nothing is keyed on id alone anywhere: 200 of 351 distinct task ids appear in more than one project, and in the author's storefront one id names two different tasks within a single project because archiving is a soft delete that lets ids be reused. A test asserts that the same id in two projects yields two items with distinct refs and that each is retrievable by its own ref.

Filters compare case-insensitively throughout. This is not politeness: priority is written lowercase on disk against a capitalised config list, so a case-sensitive filter for High would silently match nothing at all - the worst kind of bug, since the UI would simply look empty.

Reload replaces exactly one project's state and scans outside the lock, so a slow disk never blocks a reader. A test asserts the untouched project's LoadedAt is unchanged, which is stronger than asserting its task count is unchanged - a rescan producing the same count would slip past that.

Concurrency is real, not theoretical: the fsnotify watcher in TASK-18 will reload while the UI reads. Guarded by a read-write mutex and covered by a test running eight concurrent readers against four concurrent reloaders, verified under the race detector.

Deliberately not implemented here: the union of status lists. The store exposes each project's declared list as raw material, and turning that into ordered board columns is TASK-27's, with its own rules for disjoint lists. Putting it here would have been the easy scope creep.

Measured over the nine real projects and 884 tasks: full load 105 ms, reloading every project one by one 100 ms - about 11 ms each, which is what a single file change actually costs - and a filtered query over the whole corpus 274 microseconds. Budget set at 2 seconds and enforced by a test, so a regression fails the build rather than being noticed later.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added internal/store: one in-memory view across every registered project, safe for concurrent use.

Entities are keyed by a Ref of project path plus kind, class and id, so ids that collide across projects - and the one that collides inside a project - never shadow each other. Query is a struct of optional filters over project, status, milestone, priority, type, label, assignee, kind, class and free text, combined with AND and compared case-insensitively. Reload rescans exactly one project. A project that fails to load is kept and reported rather than dropped, and its diagnostics surface alongside the parser's.

Verified by 14 tests at 85.3% coverage, including a concurrency test of eight readers against four reloaders that passes under the race detector, and a test that a reload leaves the other project's load timestamp untouched.

Against the nine real projects: 9 projects and 884 tasks loaded in 105 ms, a single-project reload about 11 ms, a filtered query over the whole corpus 274 microseconds. The documented startup budget is 2 seconds and a test enforces it, so a regression fails the build. The figures are in the README.

The union of status lists is deliberately not here - the store hands over each project's declared list and TASK-27 owns turning them into columns.
<!-- SECTION:FINAL_SUMMARY:END -->
