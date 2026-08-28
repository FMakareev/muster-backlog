---
id: TASK-47
title: Produce release artefacts and installation paths
status: In Progress
assignee:
  - '@claude'
created_date: '2026-08-26 15:03'
updated_date: '2026-08-28 16:23'
labels: []
milestone: m-4
dependencies:
  - TASK-78
priority: high
type: chore
ordinal: 47000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A 1.0 that only builds from source reaches nobody. Release automation must attach usable artefacts, and installation has to be documented for a user who has never seen a Wails application.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Tagged releases attach built artefacts produced by CI rather than by a local machine
- [ ] #2 Artefacts are verifiable through published checksums
- [x] #3 Installation and upgrade instructions are documented per supported platform
- [x] #4 Supported platforms and their prerequisites are stated explicitly, including what is untested
- [ ] #5 A released artefact passes the smoke checklist on a clean machine
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Build the artefacts in a second job of release-please.yml rather than a separate workflow: a release created with GITHUB_TOKEN raises no release or tag event, so a workflow keyed on those would never run. The job hangs off needs + the release_created output.
2. Cut the release as a draft, attach the artefacts, and only then clear the draft flag, so nothing is ever public without its downloads and a failed packaging run leaves nothing to retract.
3. Package with the same command a person runs — wails3 task package — from a checkout of the tag, so the version stamped in every artefact comes from the release commit.
4. Check the built binary reports the version being released before anything is uploaded, and smoke the AppImage by extracting it rather than mounting it, since a runner has no FUSE.
5. Name every artefact after the version, publish SHA256SUMS beside them, attach muster-mcp as its own asset for agents that run outside the desktop.
6. Document installation, upgrade and verification per artefact in the README, and state what is untested.
7. Guard the shape in tests: the artefacts job must exist, depend on the release, publish only after uploading, and pin the same toolchain as CI.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Built in a second job of release-please.yml rather than a workflow of its own, and that is forced rather than chosen: a release created with GITHUB_TOKEN raises no release event and its tag raises no push event, so a workflow keyed on either would sit there and never run, with nothing to say why.

The release is cut as a draft and published by the last step, after the artefacts are attached. Two consequences worth recording: a draft release records a tag name but GitHub does not create the tag until it is published, so the job checks out github.sha — the release commit that triggered the run — rather than the tag, which would not resolve; and a packaging failure now leaves a draft to re-run instead of a public release to retract.

Measured what packaging actually produces before designing around it: .deb, .rpm, .pkg.tar.zst and a 126 MB AppImage, plus both binaries. All four are attached, muster-mcp goes up as its own asset for agents that run where the desktop application cannot be installed, and SHA256SUMS beside them.

The AppImage is smoked by extracting it rather than running it: a runner has no FUSE, so a direct run would fail for a reason unrelated to the AppImage. Verified locally that --appimage-extract works and the inner AppRun reports the version.

Every shell step was dry-run verbatim against the real bin/ output: the version check passes for muster, muster-mcp and the AppImage; the renaming produces the five expected files; SHA256SUMS holds plain names. The documented verification command was checked the way a person would use it — with only one artefact downloaded, 'sha256sum --ignore-missing -c SHA256SUMS' says OK and exits 0.

Guards: TestTheReleaseWorkflowAttachesArtefactsBeforePublishing fails if the artefacts job stops hanging off the release, if nothing is uploaded, if the release is published before the upload, if no checksums are published, or if the release is created public; TestTheReleaseBuildUsesTheToolchainThePipelineChecked fails if the two workflows drift apart. All fault-injected.

AC 1, 2 and 5 stay open deliberately: they are claims about a release run, and only the first real release can close them. Everything they depend on is in place and verified as far as it can be without a runner.
<!-- SECTION:NOTES:END -->
