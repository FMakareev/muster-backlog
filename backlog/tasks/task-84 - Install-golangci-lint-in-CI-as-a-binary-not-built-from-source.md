---
id: TASK-84
title: 'Install golangci-lint in CI as a binary, not built from source'
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 13:54'
updated_date: '2026-08-28 13:59'
labels: []
dependencies: []
priority: high
type: bug
ordinal: 84000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
CI fails before it lints anything: `go install golangci-lint@v2.13.1` is compiled with the Go version CI takes from go.mod (1.25.0), and golangci-lint v2.13.0 raised its own go directive to 1.26.0. The install exits with 'requires go >= 1.26.0 (running go 1.25.0; GOTOOLCHAIN=local)'.

Building the linter from source ties the version this project can pin to the Go floor the project declares, so the same failure returns every time golangci-lint moves ahead of go.mod. Installing the published binary breaks that coupling and skips a slow compile. The README has the same defect: it tells a person on Go 1.25 to run the exact command that fails.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 CI installs the pinned golangci-lint without compiling it, and the pinned version stays v2.13.1 with go.mod still declaring 1.25.0
- [x] #2 The downloaded binary is verified against the checksums published for that release before it is used
- [x] #3 README and CONTRIBUTING tell a person on the documented minimum Go the install command that actually works
- [x] #4 A test fails if golangci-lint goes back to being built from source in CI
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
1. Reproduce: confirm golangci-lint v2.13.x declares go 1.26.0 while go.mod declares 1.25.0, so `go install` under CI's toolchain fails before compiling this project.
2. Replace the `go install` of golangci-lint in .github/workflows/ci.yml with a download of the published linux-amd64 release archive, pinned to GOLANGCI_VERSION, verified against that release's checksums file, unpacked into the same ~/go/bin the existing cache covers.
3. Fix the same defect in README.md and CONTRIBUTING.md: the documented install command must work on the documented minimum Go.
4. Add a regression test in ci_test.go asserting the linter is not built from source in CI, naming why.
5. Verify: run the new test, fault-inject by putting `go install golangci-lint` back, run the whole suite and `wails3 task lint`.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Measured first. golangci-lint's own go directive: v2.12.0 and v2.11.0 say go 1.25.0, v2.13.0 and v2.13.1 say go 1.26.0. go.mod says 1.25.0 and CI takes its Go from go.mod, so `go install` of the pinned linter cannot succeed there. The Wails CLI's directive is go 1.25.0, so that one still installs from source and was left alone.

Rejected: pinning back to v2.12.2. It fixes today and returns the next time golangci-lint moves ahead of this module's floor, and it would leave contributors on Go 1.26 running an older linter than the one they would install by hand. Rejected: raising go.mod to 1.26.0 — a lint tool should not set a project's minimum Go.

Taken: download the published linux-amd64 archive for GOLANGCI_VERSION, verify it against that release's checksums file, install into the ~/go/bin the existing cache already covers. The released binary is built with go1.27.0, which is the point: it has no relationship with this module's Go at all. It also removes a multi-minute compile from every cache miss.

Ran the step's exact commands locally: checksum verified, binary reports 2.13.1. Fault-injected both halves — an archive replaced under the pinned name is rejected, and a checksums file with nothing to verify exits non-zero rather than passing silently.

Docs carried the same defect: the README told a person on the documented minimum Go to run the command that fails, and the lint section said @latest against a pinned table. Both replaced.

AC 1 evidence: the workflow now downloads rather than compiles, GOLANGCI_VERSION is still v2.13.1 and go.mod still declares 1.25.0; the step's commands were run verbatim on this machine, which is the closest verification available without a runner. The run itself is only provable on the next push.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
CI installed golangci-lint with `go install`, which compiles it with the Go taken from go.mod. v2.13.0 raised the linter's own go directive to 1.26.0 while this module declares 1.25.0, so the pipeline died on 'requires go >= 1.26.0' before compiling a line of this project. The workflow now downloads the pinned linux-amd64 release and verifies it against that release's published checksums, which decouples the linter from this module's Go floor for good and drops a slow compile. README and CONTRIBUTING gave the same broken command and now give the working one. TestTheLinterIsNotBuiltFromSource fails if the linter goes back to being compiled, or if the download stops being checksum-verified; both were fault-injected. wails3 task lint: 0 issues; full suite green with MUSTER_REQUIRE_BACKLOG_CLI=1.
<!-- SECTION:FINAL_SUMMARY:END -->
