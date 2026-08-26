---
id: TASK-19
title: Bridge backend state and events to the frontend
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 16:46'
labels: []
milestone: m-1
dependencies:
  - TASK-17
  - TASK-18
priority: high
type: feature
ordinal: 19000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Svelte layer needs the aggregated store and a live change feed. Define the binding surface between Go and the frontend once, so screens are added later without renegotiating the contract each time.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Frontend can request the full aggregated task set and the project registry
- [x] #2 Backend emits an event on store change and the frontend updates without a reload
- [x] #3 Shared types are generated rather than hand-mirrored on both sides
- [x] #4 Backend errors surface to the frontend as structured, displayable messages
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/app gains a BoardService bound into Wails: it owns the registry, the store and the watcher, and is the only thing the frontend talks to.
2. Wire the lifecycle to Wails' own hooks - ServiceStartup loads the registry, scans every project and starts the watcher; ServiceShutdown closes it. Doing this through the service interface rather than in main means shutdown is guaranteed to run.
3. Expose the domain types directly rather than hand-written DTOs, so the generated TypeScript comes from the single source of truth. A mirrored type is a type that will drift.
4. Errors do not travel as Go errors, which serialise to bare strings. Every failure becomes a structured Problem with a kind, a title, a detail and the path it concerns, so the UI can render it rather than print it.
5. The watcher's callback reloads exactly one project and emits a typed event carrying which project changed, so the frontend can refresh without polling and without a full reload.
6. Prove the chain in a test at the service level: write a file, assert the service reports the new task and that a change event was published.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
BoardService in internal/app is the frontend's only entry point. It owns the registry, the store and the watcher, and hooks into Wails' own ServiceStartup and ServiceShutdown rather than being wired in main, so shutdown is guaranteed to run and the watcher cannot be left holding descriptors.

Startup deliberately never returns an error. Wails aborts the whole application if a service's startup fails, which would leave a user with a window that refuses to appear because their registry has a typo. Instead every failure becomes a structured Problem the window can render.

Errors do not cross the bridge as Go errors, which serialise to bare strings and leave the UI with nothing to lay out. A Problem carries a kind, a title, a detail and the path concerned. The kinds matter: no_registry is first run rather than a fault, so the UI offers onboarding instead of an error, while registry, project and file are progressively narrower failures.

Types are generated, not mirrored. The domain entity is exposed directly and Wails generates the TypeScript from it, including a typed event map - the generated eventdata.d.ts declares muster:project:changed as carrying ProjectChanged. A hand-written DTO on the frontend would only be a type waiting to drift.

The registry path is a field rather than a call to registry.DefaultPath(). adrg/xdg caches its environment variables at process start, so a test cannot relocate the config home; injecting the path makes the service testable without touching global state, and production still resolves the default.

Verified against the real corpus rather than fixtures alone. Built a server-mode binary, wrote a registry over all nine real projects, and called the service over the HTTP bridge: 9 projects and 884 tasks returned, storefront correctly reported with layout .backlog, each project carrying its own status list, zero problems. Then connected a WebSocket client to the events endpoint and wrote a task file into TREELINE: muster:project:changed arrived carrying that project's path, and again on removal. Confirmed afterwards that TREELINE is clean and no other repository was written to.

Note for packaging: 'wails3 task build:server' fails on this machine for the same GTK reason as the desktop build; server mode needs the gtk3 tag too, which the task does not pass through. Worth raising in TASK-47.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added BoardService: the bridge between the aggregated store and the frontend, plus the Svelte side that consumes it.

The service exposes the registry, the full task set with filters, one task by ref, filter values and per-project status lists; it loads everything at startup, starts the watcher, and closes it on shutdown through Wails' service lifecycle. Failures become structured Problems with a kind, title, detail and path rather than bare error strings, and a missing registry is classified as first run rather than as a fault. The frontend mirrors backend state into nanostores, subscribes to the change event and refreshes without polling; types come from the generated bindings on both sides.

Verified against the nine real projects rather than fixtures. Over the HTTP bridge the service returned 9 projects and 884 tasks with storefront correctly detected as a .backlog layout and every project carrying its own status list, with zero problems. A WebSocket client on the events endpoint received muster:project:changed carrying the project path when a task file was written into TREELINE, and again when it was removed - proving the whole chain from an agent's write to a frontend refresh. The generated eventdata.d.ts types that event, which is the generated-not-mirrored criterion demonstrated rather than asserted.

Go tests cover the service at 83%: projects and tasks exposed, ids colliding across projects arriving distinguishable, one task fetched by ref with its full body, first run reported as such, a broken project and a skipped file each becoming a renderable problem, a disk change reaching the service, and a registry edit picked up by Reload. Lint and the full suite are green, and svelte-check passes on the frontend.
<!-- SECTION:FINAL_SUMMARY:END -->
