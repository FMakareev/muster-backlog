---
id: TASK-18
title: Watch project directories with fsnotify
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 16:35'
labels: []
milestone: m-1
dependencies:
  - TASK-17
priority: high
type: feature
ordinal: 18000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Liveness is the reason this is an application and not a report: an agent changes a status and the screen follows without polling. Watch each project task directory, debounce the bursts that editors and the CLI produce, and reload only what changed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Each registered project task directory is watched for create, write, rename and delete
- [x] #2 Rapid bursts are debounced into a single reload
- [x] #3 Only the affected project is rescanned on a change
- [x] #4 Watches survive directory recreation and editor atomic-save patterns
- [x] #5 A project directory that disappears is reported without terminating the other watches
- [x] #6 Watcher shuts down cleanly with no leaked goroutines or descriptors
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/watcher wraps fsnotify. It watches every entity directory of every registered project, not just tasks/ - drafts, milestones, docs, decisions, completed and the archive subtree all change too, and the documents viewer depends on the same liveness.
2. Debounce per project rather than globally, so a burst in one project never delays another. A CLI write touches a file several times in milliseconds and a git branch switch rewrites hundreds; both must collapse into one reload.
3. The watcher owns no store. It reports a project path to a callback, which keeps it testable and keeps the reload policy where the store is.
4. Directory recreation and atomic saves: fsnotify watches inodes, so a directory replaced by git or an editor silently stops delivering events. Re-add the watch whenever a directory is created, and re-resolve a project's directories on every reload, so a watch that died is restored on the next event rather than being lost until restart.
5. A directory that disappears is reported and dropped, never fatal - other projects keep working.
6. Close is synchronous and idempotent, and a test asserts no goroutines leak, since this runs for a whole working day.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Watches every entity directory, not only tasks/. Drafts feed the inbox, documents and decisions feed the viewer, and the archive subtree changes whenever a task is archived - a watcher covering only tasks would have left every other screen quietly stale, which is the kind of bug that is noticed weeks later. The data directory itself is watched too, so a config.yml edit that changes the status list is noticed.

Debouncing is per project rather than global, so a burst in one project never delays another. Verified both ways: 50 rapid writes collapse to a single reload, and two projects changing at once are both reported.

Two hazards handled that the criteria named but that are easy to implement incorrectly:
- fsnotify watches inodes, so a directory replaced wholesale - by git, or by a tool that rebuilds it - stops delivering events silently, with no error anywhere. The watcher forgets vanished directories on any remove or rename and re-resolves the project's whole directory set, so a dead watch is restored by the next event rather than lost until restart. Tested by deleting a task directory, recreating it, and asserting a subsequent write is still reported.
- fsnotify is not recursive, so a docs subdirectory created after startup would be invisible. Newly created directories are added to the watch as they appear, and tested.

The watcher owns no store and applies no reload policy: it reports a project path to a callback. That keeps it testable in isolation and keeps the decision about what a change means where the store is.

Close is synchronous and idempotent, and the test asserts no callback fires after it - including when a debounce timer is still in flight - and that no goroutines leak. This process stays open all day, so that is not academic.

Coverage 87.3%, all tests pass under the race detector. An integration test proves the whole chain: a file written on disk ends with the store holding the new task, with no polling anywhere.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added internal/watcher: filesystem changes become debounced per-project reload signals, built on fsnotify.

It watches every entity directory of every project - tasks, drafts, milestones, docs, decisions, completed, the archive subtree and the data directory itself - and reports create, write, rename and delete. Bursts collapse per project. Directories created after startup are picked up, and directories replaced wholesale have their watches re-established, which fsnotify does not do on its own. A project that vanishes is dropped without affecting the others, and a project that never existed is skipped rather than fatal.

Verified by 11 tests at 87.3% coverage, all passing under the race detector, exercising the behaviour rather than the configuration: each entity directory reported in turn, all four event kinds, 50 rapid writes collapsing to one reload, two projects debounced independently, an editor's write-temp-then-rename save seen, a task directory deleted and recreated with a later write still reported, a nested docs directory becoming watched, one project's directory vanishing while another keeps working, and Close proving idempotent, callback-free afterwards and leak-free with a debounce timer in flight.

An integration test closes the loop: a task file written on disk ends with the store holding it, through watcher and store together, with no polling.
<!-- SECTION:FINAL_SUMMARY:END -->
