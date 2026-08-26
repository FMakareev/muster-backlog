---
id: TASK-9
title: Add contributor and community documents
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 16:08'
labels: []
milestone: m-0
dependencies:
  - TASK-5
priority: medium
type: docs
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
An open repository needs the documents a stranger looks for before opening a pull request: how to set up, how to commit, how work is tracked, how to report a vulnerability, and what behaviour is expected. Backlog.md-based task tracking is unusual enough that contributors must be told the backlog lives in the repository, not in GitHub Issues.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 CONTRIBUTING.md covers environment setup, hook installation, lint and test commands, commit convention and pull request flow
- [x] #2 CONTRIBUTING.md explains that tasks live in backlog/ and are managed with the backlog CLI
- [x] #3 CODE_OF_CONDUCT.md is present with a working contact address
- [x] #4 SECURITY.md states supported versions and a private reporting channel
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Expand CONTRIBUTING.md, which currently covers only the commit convention, into the full contributor path: prerequisites, one-command setup, hook installation, lint and test commands, the pull request flow, and how work is tracked.
2. Make the Backlog.md tracking explicit and early. A contributor arriving from GitHub will look for Issues and find none; if that is not explained in the first screen they will file one anyway or leave.
3. Add CODE_OF_CONDUCT.md based on Contributor Covenant 2.1.
4. Add SECURITY.md with a supported-versions statement honest about pre-1.0 and a private reporting channel.
5. Contact addresses: use GitHub's private vulnerability reporting and the maintainer's GitHub profile rather than publishing a personal email address into a public repository unasked. Flag this for the user to override if they want an email there.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Did not publish a personal email address. The acceptance criterion asked for a working contact address in the Code of Conduct; both it and SECURITY.md route through GitHub instead - private vulnerability reporting via the Security tab, and a private message to the maintainer's GitHub profile for conduct reports. Publishing someone's email into a repository that becomes permanently public is not something to do without being asked, and the GitHub routes are working channels. If an email is wanted there, say which and it is a one-line change in both files.

SECURITY.md states an explicit threat model rather than boilerplate, because this application's shape makes the usual web scope wrong. In scope: argument injection into the backlog CLI subprocess, path traversal outside registered projects, anything triggered by rendering untrusted markdown from task bodies, and data destruction beyond what the CLI itself would produce. Out of scope: a malicious backlog binary already on PATH, content in a folder the user deliberately registered, and anything needing privileged local access. The supported-versions table says plainly that there is no release yet and only the default branch is supported.

CONTRIBUTING leads with how work is tracked, not with setup. A contributor arriving from GitHub will look for Issues, find none, and either file one anyway or leave - so the absence has to be explained before anything else. It also states that changes adding a field, label convention or sidecar to the Backlog.md format will be declined, since that is the project's defining constraint and a contributor should learn it before writing code rather than in review.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the documents a stranger looks for before contributing.

Expanded CONTRIBUTING.md from the commit-message section alone into the full path: how work is tracked in Backlog.md rather than GitHub Issues, setup, the pre-push checks, hook installation with the measured time budget, the pull request flow, where to report problems, and the format constraint that gets a change declined. Added CODE_OF_CONDUCT.md (Contributor Covenant 2.1) and SECURITY.md with a pre-1.0 supported-versions statement and a threat model specific to a local-first desktop application that shells out to a CLI.

Verified by walking every relative link in the five root and docs files programmatically - zero broken - and by running the commands the documents tell contributors to run: 'wails3 task lint' green, 'go test -tags gtk3 ./...' clean.

Contact routing deliberately goes through GitHub rather than a published email address; recorded in the notes with the reasoning and how to change it.
<!-- SECTION:FINAL_SUMMARY:END -->
