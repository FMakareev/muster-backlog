---
id: TASK-3
title: Add MIT license
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 14:59'
updated_date: '2026-08-26 19:40'
labels: []
milestone: m-0
dependencies: []
priority: high
type: chore
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Backlog.md ships under MIT (verified on 1.48.0), so a layer on top is unencumbered. Publishing without a license file means nobody may legally use the code, so this blocks the public push.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 LICENSE file at the repository root contains the MIT text with the correct copyright holder and year
- [x] #2 License is declared in package metadata and stated in the README
- [x] #3 Third-party dependency licenses are reviewed for MIT compatibility and the result is recorded
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add an MIT LICENSE at the repository root with FMakareev as copyright holder and 2026 as the year, matching the git identity configured for this repository.
2. Declare the licence in package metadata: frontend/package.json license field, and state it in README and build/config.yml.
3. Audit the licences of every dependency actually resolved - Go module graph and the pnpm tree - and record the result in the task notes, since MIT compatibility is the acceptance criterion rather than a guess.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Dependency licence audit, run against what is actually resolved rather than what is declared.

Go - only three third-party modules are linked into the binary (go list -deps with the production and gtk3 tags), the rest of the module graph is Wails tooling that never ships:
- github.com/wailsapp/wails/v3 v3.0.0-beta.8 - MIT
- github.com/adrg/xdg v0.5.3 - MIT
- github.com/godbus/dbus/v5 v5.2.2 - BSD-2-Clause

Frontend production tree (pnpm licenses list --prod): 46 MIT, 3 Apache-2.0 (aria-query, axobject-query, typescript), 1 ISC (picocolors). All permissive and MIT-compatible.

Frontend dev-only tree adds MPL-2.0 (lightningcss and its linux-x64 binary, pulled in by Tailwind 4), BSD-3-Clause (source-map-js) and 0BSD (tslib). MPL-2.0 is file-level weak copyleft and these are build-time only - they are not distributed in the binary - so they impose no obligation on releases. Worth re-checking if lightningcss output ever ends up vendored.

Not a dependency but worth recording: GTK and WebKitGTK are LGPL and are dynamically linked from the host system, which is compatible with shipping an MIT binary. This becomes relevant when packaging bundles libraries rather than linking system ones (TASK-47).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the MIT LICENSE at the repository root, copyright 2026 FMakareev, matching the git identity configured here. Declared the licence in frontend/package.json, in build/config.yml product metadata and in the README.

Verified the third acceptance criterion by auditing resolved dependencies rather than declared ones: go list -deps under the production and gtk3 build tags shows only three third-party modules actually link into the binary (Wails MIT, adrg/xdg MIT, godbus/dbus BSD-2-Clause), and pnpm licenses list shows the frontend production tree is 46 MIT, 3 Apache-2.0 and 1 ISC. No copyleft licence reaches anything that ships; the two MPL-2.0 packages are build-time only. Full breakdown recorded in the implementation notes.
<!-- SECTION:FINAL_SUMMARY:END -->
