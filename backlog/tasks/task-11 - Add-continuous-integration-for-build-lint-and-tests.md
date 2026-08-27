---
id: TASK-11
title: 'Add continuous integration for build, lint and tests'
status: In Progress
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-27 21:37'
labels: []
milestone: m-4
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
- [x] #1 Workflow runs on pull requests and on pushes to the default branch
- [x] #2 Pipeline runs Go and frontend linting, unit tests and a full desktop build
- [x] #3 Go and Node toolchain versions in CI match those documented for local development
- [x] #4 Dependency and build caching keeps a typical run within a documented time budget
- [ ] #5 Required status checks are configured so a red pipeline blocks merge
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The pipeline runs the same three commands a person runs before pushing - wails3 task lint, test, build - rather than a list of its own. A separate list drifts, and a pipeline checking something else is a pipeline that goes green while the machine it speaks for goes red.

One job, not three. Splitting lint, test and build would run them in parallel and pay the setup - apt, Go, Node, two Go tools compiled from source - three times over, and the setup is most of the wall clock.

The important finding: without the Backlog.md CLI, thirty-eight tests skip. Every write in this application goes through that CLI, so every test of a write runs the real thing against a real project on disk, and a skip is not a failure. A pipeline that did not install it would have reported green having tested none of the writes - the most valuable half of the suite gone, silently. So CI installs it, and that is not enough on its own: MUSTER_REQUIRE_BACKLOG_CLI turns the skip into a failure, and the test step sets it. The thirty-eight inline skips became one shared guard in internal/testcli, which skips on a machine without the toolchain - somebody reading the code should not need all of it - and fails where the environment says the CLI has to be there. Both directions were tested by running the suite with a PATH that has the Go toolchain and not the CLI.

Everything CI installs is pinned, and the Go and pnpm versions are not written twice: Go comes from go.mod through go-version-file, pnpm from the packageManager field, which package.json did not have and which pnpm/action-setup needs. The README said to install golangci-lint at @latest, which is not a version and would turn the pipeline red on somebody else's release day; it is pinned to v2.13.1 in both places now.

ubuntu-latest has no WebKitGTK 6.0, so CI takes the GTK 3 path the README already documents for Ubuntu 24.04. Nothing forces the tag - the build detects what is present and chooses, which is the same code a contributor exercises.

The build step is not finished by producing a file. The last step asks the binary what version it is and compares it against build/config.yml with the same expression the Taskfile uses, so a broken stamp fails the run rather than shipping a binary nobody can identify. That assertion was run locally against the real build, and against a wrong version to see it fail.

Measured for the budget, on one machine: lint 14s, a cold test run 40s, an incremental build 2s - about a minute of actual work. The rest of a run is setup, so all four slow things are cached: the Go modules and build cache, the pnpm store, and the two Go tools keyed on their pinned versions. Ten minutes warm is the stated budget, and the README says that a run consistently over it means a cache is not being hit.

Six tests cover the workflow itself, since it cannot be run before the repository exists: that it triggers on pull requests and on main, that it runs the three commands and installs the CLI and refuses the skip, that every version is pinned and matches the README, and that the slow things are cached and the job has a timeout. Four were made to fail first - letting the write tests skip, unpinning the linter, pinning it to a version the README does not document, and removing the tool cache. The first attempt at the pinning check passed on "latest" because the README happens to contain that word elsewhere; it requires something that looks like a version now.

Left open: #5, required status checks. That is a repository setting rather than a file, and it cannot be configured before the repository exists. It belongs to the first push, TASK-13, whose own criteria already include branch protection requiring these checks.
<!-- SECTION:NOTES:END -->
