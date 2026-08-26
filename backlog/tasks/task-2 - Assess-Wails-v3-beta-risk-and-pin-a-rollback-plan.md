---
id: TASK-2
title: Assess Wails v3 beta risk and pin a rollback plan
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 14:59'
updated_date: '2026-08-26 19:39'
labels: []
milestone: m-0
dependencies: []
documentation:
  - backlog/decisions/decision-1 - Build-the-desktop-shell-on-Wails-v3-beta.md
priority: high
type: spike
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The specification (section 8, item 1) flags Wails v3.0.0-beta.8 as a deliberate decision rather than inertia. Establish what beta actually costs here: API churn between betas, Linux WebKit runtime quirks, packaging maturity, and what a migration away would take. The outcome is a written decision, not code.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ADR records the chosen version, why beta is acceptable, and the concrete risks accepted
- [x] #2 Upgrade policy is written down: when beta bumps are taken and how they are verified
- [x] #3 Rollback plan names the fallback shell and estimates the work to switch
- [x] #4 Findings that change the specification are reflected back into it
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Establish the factual position of Wails v3: how long it has been in beta, the cadence of beta releases, whether breaking changes occur between betas, and what the project itself says about reaching stable.
2. Record the concrete risks already met in this repository rather than only theoretical ones - the GTK 4 / WebKitGTK 6.0 requirement found in TASK-1 is the first.
3. Define an upgrade policy: when a beta bump is taken and how it is verified, given that the whole application sits on this dependency.
4. Name the fallback shell and estimate honestly what switching would cost, based on how much of the codebase is actually shell-coupled.
5. Reflect anything that changes the specification back into it, and hand the ADR text to TASK-12 which owns the decision log.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Measured position of Wails v3, August 2026.

Release cadence: beta.6 on 9 Aug through beta.14 on 26 Aug - nine releases in seventeen days. The specification pinned beta.8, published 12 Aug, so this repository was already six releases behind on the day it was scaffolded. Whatever policy is chosen, it has to assume a bump roughly every two to three days is available.

The project's own position: each release carries 'the API is stable, but you may still encounter issues before the final 3.0 release', and the documentation splits the surface explicitly. Stable: core application APIs, window management, menu system, event system, file dialogs, service bindings. Unstable and subject to change before GA: advanced window options, platform-specific features, experimental features. Everything Muster needs sits in the stable half - a window, bound services, events, and file dialogs for the Projects screen. That is the single most reassuring fact found, and it also defines a rule: stay out of the unstable half. No GA date is committed anywhere.

Defect found and fixed. Wails ships two parallel version streams - the Go module and the @wailsio/runtime npm package - and they are not synchronised: the Go module is at beta.14 while npm stops at beta.13. The scaffold in TASK-1 took @wailsio/runtime from npm 'latest' and got beta.13 against a Go side on beta.8, so the repository shipped a five-release mismatch across the IPC boundary. Aligned npm down to 3.0.0-beta.8 to match the Go module and the CLI; rebuilt and ran clean. Version alignment across the two streams is now a thing that has to be checked on every bump, and the npm side can lag, so the newest usable version is the newest one present in both.

Concrete risk already realised: the default Linux build targets GTK 4 and WebKitGTK 6.0, which Ubuntu 24.04 LTS does not ship. Building there needs the gtk3 tag. Wails supports both paths, but this is the kind of platform detail that can shift between betas and must be re-checked on upgrade.

Rollback cost, measured rather than guessed. Shell coupling today: one Go file (main.go, 47 lines, 8 call sites into the application package), the vite plugin, and 4 generated binding files. Our entire Go surface is 69 lines. More importantly the architecture keeps it that way by design - the parser, the aggregated store, the fsnotify watcher, the backlog CLI adapter, the hotkey handler and the MCP server are all plain Go with no shell dependency. The realistic fallback is Wails v2: same language, same architecture, mature and stable. Switching means rewriting window setup and bindings, not the product.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Settled the Wails v3 beta question deliberately, on measurements rather than on inherited habit. Recorded as decision-1.

Findings: nine beta releases in seventeen days, so a bump is available every two to three days; Wails publishes an explicit stable/unstable API split and everything Muster needs - window, bound services, events, file dialogs - sits in the stable half; no GA date is committed. Accepted the beta with three guardrails: stay out of the unstable half of the API, bump deliberately as a single commit moving Go module, CLI and npm runtime together, and treat Wails v2 as the fallback.

The spike found and fixed a live defect rather than only assessing risk: Wails ships two unsynchronised version streams, and the scaffold had taken @wailsio/runtime from npm 'latest', shipping beta.13 against a Go side on beta.8 - a five-release mismatch across the IPC boundary. Aligned npm to beta.8, rebuilt clean, and confirmed the window still opens; the rule that the newest usable version is the newest present in both streams is now part of the upgrade policy.

Rollback cost measured rather than estimated: shell coupling is one Go file at 47 lines with 8 call sites, plus the Vite plugin and generated bindings, against a total Go surface of 69 lines - and the architecture keeps the parser, store, watcher, CLI adapter, hotkey and MCP server shell-independent by design.

Reflected back into the specification: section 8 no longer lists the beta as an open question, and the name-display question is closed by the format contract's finding that only 20% of filenames round-trip to their title.
<!-- SECTION:FINAL_SUMMARY:END -->
