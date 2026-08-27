---
id: TASK-15
title: Load the project registry
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-1
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The registry is the only configuration Muster owns: which folders hold Backlog.md projects, and how they are displayed. It holds no task data and no settings that belong to Backlog.md - each project keeps its own config.yml as the authority on statuses, priorities, types and labels.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Registry is read from the XDG config location with a documented fallback
- [x] #2 Each entry carries a folder path and an optional display name and colour
- [x] #3 Per-project Backlog.md configuration is read from that project, never duplicated into the registry
- [x] #4 A malformed or missing registry produces an actionable message instead of a crash
- [x] #5 A path that is not a Backlog.md project is reported as such and does not block the other projects
- [x] #6 Registry ordering determines project ordering in the UI
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add goccy/go-yaml rather than gopkg.in/yaml.v3: it is actively maintained (v1.19.2 against yaml.v3's last release in 2022) and its errors carry line and column, which the 'actionable message' criterion depends on. Use adrg/xdg, already linked through Wails, for the config location.
2. internal/registry parses ~/.config/muster/projects.yml into an ordered list of entries carrying path, optional display name and optional colour, plus optional advisory WIP limits. Nothing about statuses, priorities or types goes in here - that is the project's own config to declare.
3. internal/project implements Backlog.md data-directory discovery per the format contract section 3.1: probe backlog/, then .backlog/, then a root backlog.config.yml following its backlog_directory. Accept both config.yml and config.yaml.
4. Loading never fails as a whole because one entry is bad. Each entry resolves to either a located project or a typed reason, and the caller renders the reasons. A missing registry is a distinct, non-error state so the UI can offer onboarding.
5. Test against the committed reference corpus for the .backlog layout and against constructed temp directories for the rest, including the failure paths.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Split into two packages rather than one. internal/registry owns Muster's own file; internal/project owns Backlog.md data-directory discovery, which is a separate concern the parser and the CLI adapter will both need.

Dependencies: goccy/go-yaml v1.19.2 over gopkg.in/yaml.v3, which last shipped in 2022 - goccy is actively maintained and its errors carry line and column, which the actionable-message criterion depends on. adrg/xdg was already linked through Wails, so the config location costs nothing new.

Discovery follows the format contract section 3.1 exactly: probe backlog/, then .backlog/, then a root backlog.config.yml following its backlog_directory. Both config.yml and config.yaml are accepted. The one deliberate deviation from using YAML: backlog_directory is read with a line reader, because Backlog.md 1.48.0 parses its own config with a hand-rolled line reader and therefore accepts files a strict YAML parser would reject. Discovery has to be at least as permissive as the tool that wrote the file, or a project the CLI is perfectly happy with becomes invisible to us.

Added a check the acceptance criteria did not ask for: a backlog_directory pointing outside its own project is rejected. Without it, a registered folder could reach arbitrary paths through its own config file, and registering a folder is meant to be trust in that folder, not in everything above it. Covered by a test.

One bad entry never sinks the load. Each entry resolves to either a location or a typed reason, and a missing registry is ErrNoRegistry - a first-run state, not a failure - so the shell can offer onboarding instead of showing an error.

Coverage: 90.7% of internal/project, 88.2% of internal/registry.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added internal/registry, which reads Muster's own projects.yml, and internal/project, which locates a Backlog.md data directory the way the CLI does.

The registry carries path, optional display name and optional colour per entry, plus advisory WIP limits, and deliberately has no field for anything the project itself declares. Discovery probes backlog/, .backlog/ and a root backlog.config.yml with backlog_directory, accepting both config file spellings. Loading is failure-tolerant by design: each entry resolves to a location or a typed explanation, a missing registry returns ErrNoRegistry as a first-run state, and malformed YAML names the file and the line.

Verified by 15 tests, 90.7% coverage of internal/project and 88.2% of internal/registry, run as part of a green 'wails3 task lint' and 'go test -tags gtk3 ./...'. The hidden .backlog layout is tested against the real fixture committed by the format-contract spike rather than a synthetic one, since that layout exists in the author's own projects and probing only 'backlog' would make it invisible. Failure paths are covered explicitly: missing folder, folder that is a file, backlog directory without a config, root config without backlog_directory, duplicate entries, blank paths, and a backlog_directory escaping its project.

On the ordering criterion: the loader preserves file order and never sorts, proven by test. There is no UI yet to render it - the board is TASK-21 - so the criterion is satisfied at the layer that owns ordering.

The registry format, its location and its XDG fallback are documented in the README.
<!-- SECTION:FINAL_SUMMARY:END -->
