---
id: TASK-8
title: 'Set up SemVer versioning, CHANGELOG and release-please'
status: In Progress
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-27 20:57'
labels: []
milestone: m-4
dependencies:
  - TASK-5
priority: high
type: chore
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Versioning and the change log must be derived from history, not maintained by hand. release-please reads Conventional Commits, opens a release pull request, bumps the version, writes CHANGELOG.md and tags the release. Pre-1.0 semantics need stating explicitly so users know what a minor bump means today.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 CHANGELOG.md exists in Keep a Changelog form with an Unreleased section
- [ ] #2 release-please workflow opens a release pull request from conventional commits on the default branch
- [ ] #3 Merging that pull request produces a SemVer tag, a GitHub release and an updated CHANGELOG.md
- [ ] #4 The version reported by the built binary matches the released tag
- [x] #5 Pre-1.0 versioning policy is documented, including what counts as breaking before 1.0
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Done and verified: the changelog and the versioning policy in it, and the version reaching the binary and the interface. Not done: the two criteria that need a GitHub repository to exist. They are named at the end.

The binary could not say what it was. No ldflags, no flag, nothing in the interface - and `-buildvcs=false` in every build path, so not even a commit. Every bug report would have started with a guess.

The version now has exactly one home, build/config.yml, which wails already reads for every manifest and which release-please bumps. Three things read it: the build stamps the binary from it through VERSION_LDFLAGS, the .deb takes its version from it, and `muster --version` prints it with the Go version and the platform, because the two questions after "which version?" are "built with what?" and "running where?".

A build with no stamp answers "dev", or "dev+<commit>" when the toolchain recorded one - never a number nobody released. buildinfo is its own package because two callers need it and neither can import the other: the command and the board service.

Two defects turned up in the plumbing.

Nothing was setting PRODUCT_VERSION. nfpm.yaml asks for it from the environment and no task ever put it there, so every .deb, .rpm and Arch package was built with whatever an unset variable expands to. The three packaging tasks set it now, from the same file the binary is stamped from.

And no build path passed ldflags for a version on any platform - not linux, darwin, windows or the server build. All four do now, in both their dev and production branches.

The interface showed no version of anything. CLIVersion had existed on the service since the CLI adapter was written and nothing had ever called it. Preferences now shows both numbers and copies them together with the user agent, because a person copying a version string by hand copies it wrong - and because the bug report template being written next asks for exactly these two.

release-please is configured with the go strategy rather than simple: simple writes a version.txt of its own, which would have been a second place the version lives. The go strategy touches only the changelog, leaving build/config.yml as the single source, updated through extra-files by the generic updater, which finds the line by an x-release-please-version annotation in the file itself.

The pre-1.0 policy is written at the top of CHANGELOG.md and the configuration is what makes it true rather than aspirational: bump-minor-pre-major on and bump-patch-for-minor-pre-major off means a feature and a breaking change both go to a minor, and only a fix goes to a patch. That is exactly the promise the file makes - a minor may break, a patch does not - and a test asserts both settings so the words and the behaviour cannot drift apart. The policy also names what counts as breaking here: the registry file, the settings file, the minimum Backlog.md CLI version, and the mcp subcommand and its tools. And what does not: anything about the Backlog.md format, which this project does not own and adds nothing to.

Verified. The version chain read from the file and back out of the binary, both ways: config.yml at 0.1.0 gave `muster 0.1.0`, back at 0.0.0 gave `muster 0.0.0`. A server build stamped 0.4.2 showed "Muster 0.4.2" and "Backlog.md CLI 1.50.1" in Preferences, put both on the clipboard, and confirmed the copy. Eight tests, and the four that guard the release chain were each made to fail first: stripping the annotation, letting the manifest drift from the file the build reads, and breaking the line the Taskfile's regex depends on.

Left open, because a GitHub repository does not exist yet:

- #2, that the workflow opens a release pull request. The workflow is written and committed and cannot be run until there is a remote.
- #3, that merging it produces a tag, a release and an updated changelog.
- #4, that the version the binary reports matches the released tag. Every link but the last is proven - release-please writes the tag and bumps build/config.yml in the same commit, and a build from that commit reports what build/config.yml says - but there is no tag to compare against yet.

All three close on the first push, which is TASK-13. Worth noting that TASK-13 depends on this task: the two cannot both go first, so this one is written now and finished by the push rather than before it.
<!-- SECTION:NOTES:END -->
