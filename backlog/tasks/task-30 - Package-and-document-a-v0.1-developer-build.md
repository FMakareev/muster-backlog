---
id: TASK-30
title: Package and document a v0.1 developer build
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:01'
updated_date: '2026-08-26 19:29'
labels: []
milestone: m-1
dependencies:
  - TASK-25
  - TASK-26
priority: medium
type: chore
ordinal: 30000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The MVP has to be runnable by someone other than its author for the open-source promise to mean anything, and by me without a development environment for the trial period. A local Linux build with documented steps is enough at this stage; distribution artefacts come with 1.0.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A documented command produces a runnable Linux artefact
- [x] #2 Runtime prerequisites including the backlog CLI version are documented
- [x] #3 First-run instructions cover creating the registry and adding a project
- [x] #4 A smoke checklist covering board, panel and writing a status change is documented and passes on the artefact
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The v0.1 developer build is the plain binary from 'wails3 task build', which is what the README documents. Packaging targets exist and work - 'wails3 task package' produced a .deb, an .rpm, a .pkg.tar.zst and an AppImage - but proper release artefacts belong to TASK-47 and are not claimed here.

Two things found while packaging, both for TASK-47:
- The AppImage is 78 MB because it bundles GTK 3. decision-2 flagged exactly this: GTK and WebKitGTK are LGPL and dynamic linking against the system is unproblematic, but bundling them changes the analysis and needs answering before a public release.
- The generated .desktop entry carries template metadata - Name=muster, Keywords=wails, Version=1.0 - which is what a user would see in their launcher.

Fixed the nfpm package metadata, which still described 'A muster application' from 'My Company' with wails.io as its homepage.

The checklist is doc-5, ten steps, and it was executed against a built artefact rather than written from memory. Nine of the ten were driven in a browser against a server-mode build of the same tree; the first, that the desktop binary opens a window and keeps it, was run against bin/muster directly. All pass: two projects in the roll with distinct colours and correct counts, columns as the union with Beta's extra status present, cards carrying their project, the panel opening and closing without disturbing the board, a drag then a keyboard press moving a task To Do to In Progress to Done with each step confirmed by reading the file, an impossible move refused with the reason while the file stayed unchanged, an external CLI edit appearing without interaction, both project configs byte-identical afterwards, and zero network requests.

Running it improved it. Step 5 asked for markdown in the panel, but a task the CLI has just created has an empty body, so the step now says to use a task that has one. And a tenth step was added for the network, since that is the property most likely to be broken by a dependency upgrade and the least likely to be noticed.

Follow-up after the user hit this: 'wails3 task package' failed for them with the gtk4 pkg-config errors, because every command in this repository needed EXTRA_TAGS=gtk3 passed by hand. I had been passing it all along and documented it in the README, which is not the same as it working. On a machine without GTK 4 the bare command has to work.

The Taskfile now derives EXTRA_TAGS from what the machine actually has: on Linux it probes pkg-config for gtk4 and webkitgtk-6.0, falls back to gtk3 when only GTK 3 and WebKit2GTK 4.1 are present, and emits nothing anywhere else so macOS and Windows are untouched. Passing EXTRA_TAGS explicitly still overrides it.

Verified bare, with no flags: build, test, lint, package and dev all work, and 'wails3 build DEV=true' - which is what dev mode runs internally - picks the tag up too, so 'wails3 task dev' starts and Vite comes up. The hardcoded -tags gtk3 in the test task and in the pre-push hook is gone; both now go through the task runner so the tag is detected rather than pinned to one machine's situation.

README, CONTRIBUTING and the smoke checklist updated: they now say to install the GTK 3 development packages and that the build works the tag out itself.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @claude
created: 2026-08-26 19:14
---
Adjusted the fourth criterion: it named hotkey capture, which was part of the MVP in the first draft of the plan but moved to m-3 when the plan was reworked. The checklist covers what the MVP actually contains.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Documented the v0.1 developer build and wrote a smoke checklist that was executed rather than imagined.

'wails3 task build' produces bin/muster; the README carries the prerequisites including the Backlog.md CLI version, the Linux system libraries for both the GTK 4 and GTK 3 paths, the first-run registry with a worked example, and now the packaging command and a link to the checklist.

The checklist is doc-5: ten steps covering the window opening, the roll, union columns, card identity and grouping, the panel, status writes by drag and by keyboard confirmed against the files, a refused impossible move, an external edit appearing live, project configs staying untouched, and no network traffic. All ten pass on a built artefact.

Executing it improved it twice: step 5 expected rendered markdown but a freshly created task has no body, and a tenth step was added for the network property, which a dependency upgrade could silently break. Packaging also works - deb, rpm, Arch and AppImage - and two findings were recorded for TASK-47: the AppImage bundles LGPL GTK libraries, which decision-2 said would need revisiting, and the generated desktop entry still carries template metadata.
<!-- SECTION:FINAL_SUMMARY:END -->
