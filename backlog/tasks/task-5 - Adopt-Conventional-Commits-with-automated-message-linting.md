---
id: TASK-5
title: Adopt Conventional Commits with automated message linting
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 14:59'
updated_date: '2026-08-26 15:50'
labels: []
milestone: m-0
dependencies: []
priority: high
type: chore
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Release automation derives versions and the changelog from commit messages, so message shape is a build input, not a style preference. Define the allowed types and scopes for this repository and enforce them mechanically.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Commit convention including allowed types and scopes is documented in CONTRIBUTING
- [x] #2 commitlint configuration enforces the convention and rejects a malformed message
- [x] #3 Breaking-change notation is documented with an example
- [x] #4 A malformed and a well-formed message are both verified against the linter
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Turn the repository into a pnpm workspace so repository-level dev tooling has a home without a second, separately-installed package.json: root package.json plus pnpm-workspace.yaml with frontend as a member. One 'pnpm install' at the root then covers everything.
2. Add @commitlint/cli with config-conventional at the root, and a commitlint.config.js that declares the type and scope vocabulary this repository actually uses, rather than accepting anything conventional-shaped.
3. Document the convention in CONTRIBUTING.md: allowed types, allowed scopes, breaking-change notation with an example, and the fact that release-please derives versions from these messages.
4. Verify by feeding both a malformed and a well-formed message through the linter and confirming it rejects and accepts respectively. Hook wiring itself belongs to TASK-7.
5. Confirm the Wails build still works through the workspace layout.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Made the repository a pnpm workspace (pnpm-workspace.yaml plus a root package.json) rather than adding a second, separately-installed package.json. Repository-level tooling like commitlint now has a home and one 'pnpm install' at the root covers root and frontend both; without this, contributors would have to install twice and would silently skip whichever one they forgot. Verified the Wails build still works through the workspace layout - clean rebuild produces bin/muster unchanged.

commitlint 21.2.2 with config-conventional. Both the type and scope vocabularies are closed enum rules rather than the permissive defaults: an unrecognised type or scope is a hard failure. The scope list is derived from the product's actual surfaces as laid out in the specification (parser, store, watcher, cli, board, list, task, inbox, docs-view, analytics, search, projects, mcp, ui) plus app, deps and release. scope-empty is disabled so a repository-wide chore can legitimately have no scope, and body-max-line-length is off so BREAKING CHANGE footers can be written as prose.

Note for TASK-8: the pre-1.0 rule that a breaking change bumps minor rather than major is documented in CONTRIBUTING but has to be configured in release-please as well, otherwise the two disagree.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Adopted Conventional Commits with mechanical enforcement.

Added commitlint.config.js with closed type and scope enums, a root package.json and pnpm-workspace.yaml so repository-level tooling installs alongside the frontend in one step, and CONTRIBUTING.md documenting the types with their release effect, the scope vocabulary, breaking-change notation with a worked example, and how release-please consumes all of it.

Verified by piping messages through the linter: 'added some stuff' is rejected for empty type and subject; 'feat(frontend): add a thing' is rejected with 'scope must be one of [...]'; 'wip(board): add a thing' is rejected with 'type must be one of [...]'; 'feat(board): render swimlanes for each registered project' passes; 'chore: tidy the repository' passes with no scope; and a 'feat(cli)!:' message with a BREAKING CHANGE footer passes. Hook wiring is TASK-7. Confirmed the Wails build is unaffected by the workspace change with a clean rebuild.
<!-- SECTION:FINAL_SUMMARY:END -->
