---
id: TASK-87
title: Stop release-please from rewriting build/config.yml as plain YAML
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 15:41'
updated_date: '2026-08-28 15:42'
labels: []
dependencies: []
priority: high
type: bug
ordinal: 87000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
release-please lists build/config.yml as a bare string in extra-files. For a .yml file that is not a line edit: release-please builds a CompositeUpdater of GenericYaml('$.version') and the annotation-based Generic updater, and the YAML half parses the file and re-serialises it.

Measured against the real file with release-please's own updaters: the result is 12 lines where there were 39. Every comment is gone, including the x-release-please-version annotation the second updater needs. The top-level 'version: 3' — the Wails config schema version, not the application's — is replaced by the release version. info.version keeps 0.0.0, unbumped, and loses its quotes, so the Taskfile's sed finds nothing and every binary and package in that release is stamped with an empty version.

The release tests caught it: three of them fail on such a file.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 release-please changes exactly one line of build/config.yml — the annotated version — leaving comments, quotes and the schema version untouched
- [x] #2 The behaviour is verified against release-please's own updater, not inferred from its documentation
- [x] #3 A test fails if the file goes back to being updated in a way that rewrites it
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
1. Reproduce with release-please's own updaters against the real build/config.yml, both the bare-string composite and the explicit generic one.
2. Change extra-files to the object form {type: generic, path: build/config.yml}, which is the annotation-driven line edit.
3. Guard it in release_test.go: a bare string for that file must fail, naming what a bare string does.
4. Verify the generic updater produces a one-line diff on the real file.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Measured with release-please's own updaters rather than reading its documentation. Installed the package and ran both code paths against the real build/config.yml.

Bare string — what was configured — resolves to CompositeUpdater(GenericYaml('$.version'), Generic). The YAML half parses and re-serialises: 39 lines become 12. Every comment gone including the x-release-please-version annotation; top-level 'version: 3' overwritten with the release number; info.version left at 0.0.0 and unquoted, so the Taskfile sed finds nothing and stamps an empty version; fileAssociations appears as null.

Explicit {type: generic}: one line changes, quotes and annotation intact, everything else byte-identical.

Also simulated the release commit — the generic updater's output plus a bumped manifest — and ran the release tests against it: green. So the next release PR passes rather than failing the way this one did.

The bare-string case failed with a JSON decode error at first, which does not tell anybody what is wrong; extra-files is now decoded raw so both shapes get a message that names the problem.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
release-please listed build/config.yml as a bare string in extra-files, which for a .yml file pairs the annotation-driven line edit with a YAML round-trip — and the round-trip wins the file. Measured against release-please's own updater: 39 lines to 12, every comment gone including the annotation the line edit needs, the Wails schema version overwritten with the release number, and info.version left unbumped and unquoted so the Taskfile would stamp an empty version into every binary and package of that release. Changed the entry to {type: generic, path: build/config.yml}, which changes exactly one line. Verified by running both updaters on the real file, and by running the release tests against a simulated release commit. TestTheReleaseConfigBumpsTheFileTheBuildReads now rejects a bare string and any updater other than generic, both fault-injected.
<!-- SECTION:FINAL_SUMMARY:END -->
