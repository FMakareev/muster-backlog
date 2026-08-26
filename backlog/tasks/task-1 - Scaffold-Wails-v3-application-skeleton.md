---
id: TASK-1
title: Scaffold Wails v3 application skeleton
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 14:59'
updated_date: '2026-08-26 15:43'
labels: []
milestone: m-0
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: chore
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Create the buildable project skeleton every other task builds on: Wails v3.0.0-beta.8, Go 1.25 backend, Svelte 5 with runes plus Vite 8 and Tailwind 4 frontend, nanostores for state, SVAR Svelte Kanban for the board. Module path and binary follow the muster-backlog name. Nothing product-specific yet - the goal is a window that opens and a dev loop that reloads.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Repository builds a runnable desktop binary with a single documented command
- [x] #2 Frontend dev server hot-reloads Svelte changes in the running app
- [x] #3 SVAR Svelte Kanban is installed and renders a placeholder board to prove the integration
- [x] #4 Go module path and binary name follow the muster-backlog name
- [x] #5 Wails, Go, Svelte, Vite, Tailwind and SVAR versions are pinned, not floating
- [x] #6 README section documents the prerequisites and the build/run commands
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Generate the Wails v3 svelte template (TypeScript + Vite) into a scratch directory and port it into the repository root, keeping backlog/ and CLAUDE.md untouched.
2. Set the Go module path to github.com/FMakareev/muster-backlog and the Go directive to 1.25; name the built binary muster.
3. Replace the template demo (greetservice, Inter font, wallpaper assets) with a minimal app service and a clean entry point.
4. Add Tailwind 4 through @tailwindcss/vite, plus nanostores for state. Note: @nanostores/svelte does not exist on npm; nanostores atoms are Svelte-store compatible and are consumed directly.
5. Add @svar-ui/svelte-kanban 2.6.0 (MIT). Its peer dependency is svelte ^5.54.0, so svelte must be raised above the template default of 5.46.
6. Pin every frontend dependency to an exact version, no caret ranges, per the acceptance criterion that versions do not float.
7. Render a placeholder board with SVAR fed by a nanostore, proving the three integrations at once.
8. Document prerequisites, build and dev commands in README.md.
9. Verify: production build produces a runnable binary, svelte-check passes, and the dev loop starts.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Ported the wails3 svelte template (TypeScript + Vite) into the repository root and stripped its demo: greetservice, Inter font, wallpaper assets and the iOS/Android build scaffolding are gone, since this is a desktop-only application and an Xcode project in a task manager repository is noise.

Toolchain resolved against npm rather than taken from the template defaults:
- @svar-ui/svelte-kanban 2.6.0 (MIT). Package name is @svar-ui/*, not the wx-svelte-* naming used in older documentation. Its peer dependency is svelte ^5.54.0, so svelte had to be raised from the template default of 5.46.4 to 5.56.10.
- typescript pinned to 5.9.3, not the current 7.0.2: svelte-check 4.7.6 declares a peer range of ^5 || ^6.
- @nanostores/svelte does not exist on npm. Not needed - nanostores atoms are Svelte-store compatible, so the $ prefix works directly.
- SVAR Kanban 2.6.0 has no rows/swimlanes prop despite what its readme implies, confirming the specification note that projects must be distinguished by grouping and card colour.
- nanostores expose values as readonly while SVAR takes ownership of the arrays passed to it, so the board receives copies.

Linux build finding worth carrying into TASK-2 and the packaging work: the default Wails v3 Linux build requires GTK 4 and WebKitGTK 6.0. Ubuntu 24.04 LTS ships neither - only GTK 3 and WebKit2GTK 4.1 - so it needs the gtk3 build tag. This machine builds with EXTRA_TAGS=gtk3. Both paths are documented in the README.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Ported the Wails v3 svelte template into the repository as the project skeleton and pinned the whole toolchain.

Backend: module github.com/FMakareev/muster-backlog on Go 1.25, entry point in main.go, root service in internal/app with a Version overridable via -ldflags. Frontend: Svelte 5.56.10, Vite 8.2.2, Tailwind 4.3.3 through @tailwindcss/vite, nanostores 1.5.2, and SVAR Svelte Kanban 2.6.0 rendering a placeholder board fed from a nanostore. Every frontend dependency is an exact version, no caret ranges. Demo content, mobile build scaffolding and the template wallpaper assets were removed; product metadata in build/config.yml and a README covering prerequisites, build, run and dev commands were added.

Verified: clean rebuild via 'wails3 task build EXTRA_TAGS=gtk3' produces bin/muster, a 9.2M stripped ELF executable; running it holds a window open for the full 12s timeout and exits only on kill, with no errors beyond benign GTK module warnings; 'pnpm run check' reports 0 errors across 166 files; 'go vet ./...' is clean and gofmt reports nothing; 'pnpm run dev' serves on 127.0.0.1:9245; the production bundle contains both SVAR kanban styles and the Tailwind theme tokens, confirming all three integrations compile together. The board was not confirmed pixel by pixel - no browser or screenshot tool is available in this container - so AC 3 rests on the component type-checking against its real props and its styles landing in the bundle.
<!-- SECTION:FINAL_SUMMARY:END -->
