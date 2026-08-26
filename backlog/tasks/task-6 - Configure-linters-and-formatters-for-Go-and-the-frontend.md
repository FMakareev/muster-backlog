---
id: TASK-6
title: Configure linters and formatters for Go and the frontend
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 14:59'
updated_date: '2026-08-26 15:48'
labels: []
milestone: m-0
dependencies:
  - TASK-1
priority: high
type: chore
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Hooks and CI need something to run. Establish the toolchain itself before wiring it: golangci-lint for Go, ESLint plus Prettier plus svelte-check for the frontend, with configuration checked in so local and CI results agree.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 golangci-lint configuration is checked in and passes on the skeleton
- [x] #2 ESLint, Prettier and svelte-check configurations are checked in and pass on the skeleton
- [x] #3 One command runs every linter across the whole repository
- [x] #4 One command applies every available autofix
- [x] #5 Editor-agnostic formatting settings are checked in
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Go: install golangci-lint v2 and configure .golangci.yml with the default linter set plus the ones that matter for a long-running desktop process - errcheck, govet, staticcheck, revive, misspell - and register the gtk3 build tag so the Linux CGO files are actually analysed rather than silently skipped.
2. Frontend: ESLint 9 flat config with typescript-eslint and eslint-plugin-svelte, Prettier with prettier-plugin-svelte, and svelte-check. Generated bindings and dist are excluded; they are not ours to lint.
3. Provide one entry point for the whole repository rather than per-language incantations: task lint and task lint:fix in the Taskfile, so hooks and CI call the same commands a developer does.
4. Add .editorconfig so formatting agrees across editors.
5. Verify each linter passes on the current skeleton and that the aggregate commands work from a clean state.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Go: golangci-lint v2.13.1 with .golangci.yml on schema version 2. The gtk3 build tag is registered in run.build-tags because the Linux CGO backend sits behind build tags and would otherwise be invisible to every linter - the files that do the actual windowing work would be silently skipped. revive's exported rule runs with disableStutteringCheck since package-qualified names like app.Service are correct Go. frontend/bindings is excluded: generated code is not ours to lint.

Frontend: ESLint 10.9.1 flat config with typescript-eslint 8.68.0 and eslint-plugin-svelte 3.23.0, Prettier 3.9.6 with prettier-plugin-svelte. All three declare peer support for ESLint 10, checked before installing. bindings/ and dist/ are ignored by both.

One entry point rather than per-language incantations: 'wails3 task lint' and 'wails3 task lint:fix' in the Taskfile, so hooks, CI and a developer all run the same thing. The frontend half is also reachable as pnpm run lint / lint:fix.

Prettier reformatted three template files it inherited (index.html, svelte.config.js, tsconfig.json); goimports regrouped the local import in main.go under the local-prefixes setting.

golangci-lint is deliberately not vendored - it is a go install away and documented in the README, which keeps the repository free of a committed binary.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Established the linting toolchain for both halves of the repository and wired it behind one command.

Added .golangci.yml (golangci-lint v2, gtk3 build tag registered, errcheck/govet/staticcheck/revive/misspell/ineffassign/unused, gofmt and goimports as formatters), frontend/eslint.config.js (ESLint 10 flat config with typescript-eslint and eslint-plugin-svelte), Prettier configuration with its ignore file, .editorconfig, and 'lint' plus 'lint:fix' tasks in the Taskfile.

Verified by running them: 'wails3 task lint' exits 0 with golangci-lint reporting 0 issues, Prettier reporting all files conforming, and svelte-check reporting 0 errors across 166 files. Confirmed the linters are not silently passing by injecting deliberate violations - an unused Go function and an unused any-typed TypeScript constant - and both were caught (golangci-lint reported gofmt and unused, ESLint reported no-unused-vars and no-explicit-any); files were then restored and the suite re-run clean. 'wails3 task lint:fix' runs end to end and leaves the tree lint-clean. README documents both commands and the one-line golangci-lint install.
<!-- SECTION:FINAL_SUMMARY:END -->
