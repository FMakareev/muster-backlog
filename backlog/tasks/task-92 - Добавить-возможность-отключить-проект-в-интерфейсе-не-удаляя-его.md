---
id: TASK-92
title: 'Reach hiding from the project roll, where a person looks for it'
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 00:51'
updated_date: '2026-08-30 03:27'
labels: []
dependencies: []
priority: medium
type: feature
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Hiding a project already works. It shipped in 509f943 on 2026-08-26: `hidden` in projects.yml, `ProjectState.Visible()` in the store, and a hide button on the Projects screen. A hidden project stays registered and loaded, and stays out of the board, the lists, search and the figures.

Two days later the same person asked for the feature as though it did not exist. That is the defect: the feature is unreachable from where a person stands when they want it.

Where they stand is the project roll. It is the one part of the interface that is always on screen, it lists every project, and clicking a project there focuses it. Nothing in it mentions hiding. Hiding lives on a separate screen behind `p`, as lowercase text at the end of a row, past the two arrow buttons. The only prose that names it is the Projects nav tooltip in screens.ts, which the Shell renders as `title=` — a hover tooltip, in a keyboard-driven tool, in a file whose own comment says a shortcut belongs in the interface rather than in a tooltip.

Bringing a hidden project back has the same problem in reverse: once it leaves the roll there is nothing in the roll to say it exists, so the way back is only through the Projects screen.

The semantics do not change. Hidden stays hidden everywhere, search included: one rule is worth more than an exception that has to be explained.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A project can be hidden from the project roll, without opening another screen
- [x] #2 A hidden project can be brought back from the project roll: the roll says how many are hidden and offers them
- [x] #3 Both controls carry an accessible name and are reachable by keyboard; neither appears only on hover
- [x] #4 Hiding semantics are unchanged: a hidden project stays registered and loaded, and stays out of the board, the lists, search and the figures
- [x] #5 A test fails if the roll stops offering hiding or stops offering the way back
- [x] #6 README says where hiding lives
- [x] #7 Rewriting an entry never writes a name that only repeats the folder name, so hiding a project cannot freeze its display name
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Stop the registry from freezing a folder name. Add registry.Override(name, path) — displayName inverse, returning "" when the name only repeats the folder — and route AddProject and SaveProject through it, so no caller can freeze a name whatever it sends. Test first: it fails today.
2. Put the write in one place. Add setProjectHidden(project, hidden) to board.ts on top of applyWrite, and have the Projects screen call it instead of its own copy. Two surfaces offering hiding must not disagree about what the rest of the entry is.
3. Offer hiding in the roll. Each row gets a hide button beside the count: a real button, always in the accessibility tree with a name of its own, drawn only on hover or keyboard focus so a permanent control does not take width from the name on the surface the application exists to show.
4. Offer the way back in the roll. When any project is hidden the roll ends with an expandable "N hidden", listing each with show. That is what tells a person hiding exists at all, and it is the only path back that does not cross a screen boundary.
5. Test the roll by mounting it. vitest gains the svelte plugin and the browser resolve condition — measured, no new dependency — and the test asserts the hide control, its accessible name, and the way back. Fault-inject both.
6. README: say where hiding lives, in the same breath as the roll.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The frozen name was found by reproducing it, not by reading: hiding a project through the service wrote `name: one` into a registry that had never carried one. The fix went into the registry package rather than into the two controls, because there are now two of them. registry.Override is displayName's inverse and both AddProject and SaveProject pass through it. The test was red before it and green after.

The roll is tested by mounting it. vitest gained the svelte plugin and the browser resolve condition and no new dependency - measured with a throwaway spike before the approach was committed to. Five tests, and all three guards were broken on purpose to watch them fail: without the focus release one test fails, without the hide control three, without the way back one.

Hiding the project you are focused on used to leave the board narrowed to a project no longer on it - nothing on screen, and nothing saying why. The roll now lets go of the focus first.

What could not be verified here: there is no browser in this environment. Chrome is not running for the devtools bridge and playwright has no chromium installed, so the roll was never clicked in a running window. The server build was made and started against a throwaway registry to prove it boots, and the two things a mounted DOM cannot answer were checked in the built stylesheet instead - that Tailwind really generated `.group-hover\:opacity-100:is(:where(.group):hover *)` and `.focus-visible\:opacity-100:focus-visible`, without which the control would have been invisible for good. The real registry was not touched.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Hiding a project was already built and already correct; it was unreachable. The roll - the one surface that is always on screen and lists every project - said nothing about it, so the feature was asked for again two days after it shipped.

Each row in the roll now carries its own hide control, named for its project, and the foot of the roll counts what has been put away and gives each of them back. The space the control occupies is reserved whether or not it is drawn, so the roll does not move under the cursor, and the button is in the page for anything that is not a mouse. Hiding the project you are focused on releases the focus first, which it did not before: the board narrowed to a project no longer on it and showed nothing without saying why. Both surfaces now write through one function, because the registry entry is rewritten whole.

A separate defect was found by reproducing it on the way: hiding wrote the folder name into projects.yml as an explicit override, freezing it. registry.Override now sits on the path every write takes, so no control can do it.

Verified: five new tests that mount the roll in a real DOM and click it, with the svelte plugin added to vitest and no new dependency; all three guards broken on purpose to watch them fail. The name fix was red before it and green after. Full suite green, golangci-lint 0 issues, svelte-check 0 errors. There is no browser in this environment, so the roll was not clicked in a running window - the two things a mounted DOM cannot answer were checked in the built stylesheet, that Tailwind really emitted the group-hover and focus-visible rules the control depends on.
<!-- SECTION:FINAL_SUMMARY:END -->
