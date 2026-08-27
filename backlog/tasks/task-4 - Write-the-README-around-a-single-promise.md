---
id: TASK-4
title: Write the README around a single promise
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 14:59'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-0
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: docs
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Specification section 9: the README must make exactly one promise - a multi-repo overview of Backlog.md with planning against review capacity. Everything else follows from it. A newcomer must understand in the first screen what this is, who it is for, and how it differs from backlog browser and from beads.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 First paragraph states the single promise without listing features
- [x] #2 README explains the relationship to Backlog.md: reads markdown directly, writes only through the CLI, introduces no format of its own
- [x] #3 Install and first-run instructions are runnable start to finish on a clean machine
- [x] #4 Project status is stated honestly, including that this is pre-1.0 and single-user
- [x] #5 Links to specification, roadmap, contributing and license resolve
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Rewrite the README so the opening states one promise - a desktop task manager over all your Backlog.md projects at once - and nothing else. The current opening does that already but immediately hands over to a prerequisites table; the structure needs to earn the reader's attention before asking for their toolchain.
2. State the relationship to Backlog.md early and plainly, since it is the thing a reader most needs to trust: reads markdown directly, writes only through the CLI, adds no format of its own.
3. Say honestly what this is not: pre-1.0, single-user, no server, no sync, and today a placeholder board that reads no real backlog.
4. Compare fairly with backlog browser and the VSCode extension. Both are good at one backlog; overstating the gap would be the easiest way to lose a reader who already uses them.
5. Verify the install and first-run path is actually runnable end to end on a clean machine rather than assumed - trace every command and every link.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Restructured so the reader is given a reason before being asked for a toolchain. The opening states the single promise; a Why section explains the gap it fills, naming backlog browser and the VSCode extension fairly - the extension is described as excellent and single-workspace, because overstating the gap is the fastest way to lose a reader who already uses it happily. A What it is not section states the format rule as a hard constraint and cites the review-budget planner being deleted as evidence that it is enforced rather than aspirational.

Found a real gap while verifying: golangci-lint was documented only in the Lint section, but the pre-commit hook runs it, and pnpm install installs the hooks. A contributor following the README in order would have had a broken first commit. Moved it into the prerequisites table alongside the Wails CLI.

Verified the install path by executing it rather than reading it: deleted node_modules, frontend/node_modules, bin, frontend/dist and frontend/bindings to simulate a fresh checkout, then ran the documented sequence. pnpm install pulled dependencies and installed all three hooks through the prepare script; the build regenerated the bindings and produced bin/muster; the binary held a window open for the full timeout. The regenerated bindings are byte-identical to the committed ones, so committing them is safe and the build is reproducible.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Rewrote the README around the single promise from specification section 9.

The opening states one thing - a local-first desktop task manager over all your Backlog.md projects at once - followed by Why (the gap: every existing tool works on one backlog at a time), What it will be (an ordinary task manager whose only novelty is spanning every project), and What it is not (no format of its own, no server, no sync, not a time tracker). The relationship to Backlog.md is stated as a hard rule with evidence. Status is stated honestly in the second line: pre-alpha, a placeholder board that reads no backlog yet.

Verified the install path by executing it from a simulated clean checkout - dependencies, build artefacts and generated bindings all deleted first - and the documented sequence ran start to finish, ending in a window that stayed open. All 12 local links resolve. 'wails3 task lint' green.

Verification caught a genuine defect: golangci-lint was documented only under linting, yet the pre-commit hook needs it and pnpm install installs that hook, so a contributor following the README in order would have hit a failing first commit. It is now a prerequisite.
<!-- SECTION:FINAL_SUMMARY:END -->
