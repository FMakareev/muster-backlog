---
id: decision-1
title: Build the desktop shell on Wails v3 beta
date: '2026-08-26 16:04'
status: accepted
---
## Context

Muster is a local-first desktop application with a Go backend - it parses markdown, watches the filesystem, shells out to the `backlog` CLI and registers a global hotkey, all of which are natural in Go and awkward elsewhere. The shell question is therefore narrow: what puts a webview around Go code.

Wails v3 was in beta at the time of this decision, which the specification flagged as something to settle deliberately rather than inherit by inertia from a previous project.

Measured in TASK-2, August 2026:

- Release cadence is fast: beta.6 on 9 August through beta.14 on 26 August, nine releases in seventeen days.
- The project states the API is stable and splits its surface explicitly. Stable: core application APIs, window management, menu system, event system, file dialogs, service bindings. Unstable before GA: advanced window options, platform-specific features, experimental features.
- No general-availability date is committed.
- The Go module and the `@wailsio/runtime` npm package are two parallel version streams that are not synchronised; npm can lag the Go module by a release.

## Decision

Build on **Wails v3**, pinned to `v3.0.0-beta.8` across the Go module, the CLI and the npm runtime.

Everything Muster needs sits in the half of the API that Wails calls stable: a window, bound services, events, and file dialogs for the Projects screen. **Deliberately stay out of the unstable half** - advanced window options, platform-specific features, experimental APIs. If a feature seems to require one, redesign around it or raise the beta risk again.

Upgrade policy:

- Beta bumps are taken deliberately, never automatically, and never as a side effect of another change.
- A bump is one commit that moves the Go module, the `wails3` CLI and `@wailsio/runtime` together. The newest usable version is the newest one present in **both** streams, since npm lags.
- Every bump is verified by a clean build, a launch that holds a window open, and the full lint and check suite - the same evidence required of the original scaffold.
- Alpha and release-candidate versions are not taken.

Fallback: **Wails v2** - same language, same architecture, mature and stable.

## Consequences

- The application inherits a dependency that is explicitly pre-1.0, on a project with no committed GA date. This is accepted knowingly.
- Version alignment across two streams becomes a recurring chore. It was already wrong once: the scaffold in TASK-1 resolved `@wailsio/runtime` from npm `latest` and shipped beta.13 against a Go side on beta.8 - a five-release mismatch across the IPC boundary, fixed by aligning npm down.
- The Linux build targets GTK 4 and WebKitGTK 6.0 by default. Ubuntu 24.04 LTS ships neither, so it builds with the `gtk3` tag against GTK 3 and WebKit2GTK 4.1. Both paths are documented; platform requirements may shift between betas and must be re-checked on upgrade.
- The cost of being wrong is bounded by architecture rather than by hope. Shell coupling today is one Go file - `main.go`, 47 lines, 8 call sites - plus the Vite plugin and the generated bindings. The parser, the aggregated store, the fsnotify watcher, the CLI adapter, the hotkey handler and the MCP server are plain Go with no shell dependency, and must stay that way. Switching shells means rewriting window setup and bindings, not the product.
- Pinning to beta.8 while beta.14 exists means starting six releases behind. The first application of the upgrade policy should be a deliberate bump to the newest version present in both streams.
