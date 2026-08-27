---
id: TASK-64
title: Upgrade the Backlog.md CLI floor to 1.50.1
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 23:40'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-3
dependencies: []
priority: high
type: chore
ordinal: 64000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Measured on 1.48.0 against 1.50.1 on scratch projects: draft promote and draft archive against an id that does not resolve exit 0 on 1.48.0 and exit 1 on 1.50.1. Muster reads a zero exit as success, so on 1.48.0 a stale draft id reports a clean success and does nothing. Whatever the floor ends up being, Muster should not depend on the exit code alone for a write whose outcome it can check.

Not fixed by the upgrade, and worth recording rather than rediscovering: there is still no draft edit in 1.50.1, and backlog init still prompts for a project name when none is passed even with --defaults, exiting 0 having created nothing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The CLI the author runs is 1.50.1 and the whole suite passes against it
- [x] #2 Draft writes verify their outcome rather than trusting the exit code, so a silent no-op is caught on any version
- [x] #3 The version floor Muster enforces is stated with the reason for it
- [x] #4 The format contract records what changed between the measured version and the current one
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
- [ ] #5 Linters and formatters pass across Go and frontend
- [ ] #6 Automated tests cover the change and the suite is green
- [ ] #7 User-facing behaviour change is reflected in README or docs
- [ ] #8 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The upgrade is done and the whole suite passes against 1.50.1, uncached.

The contract was re-derived rather than assumed, and by what the CLI writes rather than by reading constants out of a newer binary: both versions initialised a project and wrote a task carrying every section and every optional key, plus a draft. The files are byte-identical, as are config.yml and the eleven-directory skeleton. Nothing in the format contract changes, which is recorded in doc-3 section 7.1 along with the four behavioural differences that do matter to a program driving the CLI.

The one that justified the upgrade: on 1.48.0, draft promote and draft archive exit 0 when the id does not resolve, so a write that did nothing arrives as a success. 1.50.1 exits 1. Muster now checks that the note actually left the inbox instead of trusting either, which holds on both versions and cost a map lookup against a store that had already been re-read.

The floor stays at 1.48.0 with the reason written next to it: that is the version the format was measured on, and raising it would turn an unmeasured claim into a hard requirement on someone else's machine. 1.50.1 is recorded as recommended, and the code can ask whether it has it.

Two things the upgrade does not fix, now written down so they are not rediscovered: there is still no way to edit a draft in 1.50.1 - no draft edit, and task edit refuses a DRAFT id - and backlog init still prompts for a project name when none is passed even with --defaults, exiting 0 having created nothing.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Upgraded the author's CLI to 1.50.1 and made Muster stop trusting an exit code it can check.

The format contract was re-derived by comparing what the two versions write - a fully populated task, a draft, the config and the skeleton - and the files are byte-identical, so sections 1 to 6 of doc-3 stand unchanged. What did change is behaviour around failure: 1.48.0 exits 0 when draft promote or draft archive is given an id it cannot resolve, which is a write that did nothing reported as success. Muster now confirms the note left the inbox, which holds on either version.

The floor stays at 1.48.0 because that is what the format was measured against; 1.50.1 is recommended, and the reason is written where someone will read it. Also recorded: 1.50.1 still cannot edit a draft, and backlog init still prompts for a project name it was not given.
<!-- SECTION:FINAL_SUMMARY:END -->
