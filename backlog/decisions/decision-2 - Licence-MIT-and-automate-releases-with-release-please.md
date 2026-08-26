---
id: decision-2
title: Licence MIT and automate releases with release-please
date: '2026-08-26 16:04'
status: accepted
---
## Context

Muster is a layer on top of Backlog.md, which is MIT, and its board is built on SVAR Svelte Kanban, which has an MIT edition. Neither imposes a licence on work built above it, so the choice was open.

Versioning and the change log had to be decided at the same time, because the mechanism that produces them constrains how commits are written - and commit format is hard to retrofit once there is history.

## Decision

**MIT**, copyright 2026 FMakareev.

Releases are automated with **release-please**, driven by Conventional Commits: it opens a release pull request, derives the version bump from commit types, writes `CHANGELOG.md` and tags the release. Versions follow SemVer.

Before 1.0 a breaking change bumps the minor version rather than the major one. After 1.0 it bumps the major version, and the registry format and the CLI write contract stop changing without one.

Apache-2.0 was considered for its explicit patent grant and rejected: the per-file header requirement and NOTICE handling are friction that a single-maintainer desktop tool does not need, and MIT matches what the ecosystem below already uses.

## Consequences

- Commit messages become a build input rather than a matter of taste, which is why `commitlint` enforces a closed vocabulary of types and scopes and the `commit-msg` hook runs it. A sloppy message is now a broken build, not an untidy log.
- Nothing that ships carries a copyleft obligation. Verified in TASK-3 against resolved dependencies rather than declared ones: three third-party Go modules link into the binary (Wails MIT, adrg/xdg MIT, godbus/dbus BSD-2-Clause), and the frontend production tree is 46 MIT, 3 Apache-2.0 and 1 ISC. The two MPL-2.0 packages are build-time only.
- GTK and WebKitGTK are LGPL and dynamically linked from the host, which is compatible with shipping an MIT binary. This has to be revisited if packaging ever bundles those libraries instead of linking the system ones.
- The pre-1.0 minor-bump rule is documented in CONTRIBUTING but must also be configured in release-please, or the two will disagree.
