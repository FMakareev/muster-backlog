---
id: TASK-20
title: Build the application shell and state layer
status: Done
assignee: []
created_date: '2026-08-26 15:00'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-1
dependencies:
  - TASK-19
priority: high
type: feature
ordinal: 20000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The frontend skeleton every screen mounts into: layout, navigation between board and future screens, Tailwind theme and tokens, and nanostores wiring fed by the backend bridge. Visual decisions taken here set the tone of the whole product, so they are made deliberately rather than accumulated.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Application shell renders navigation and a content region
- [x] #2 nanostores hold the task set and registry and update from backend events
- [x] #3 Tailwind theme tokens for colour, spacing and typography are defined in one place
- [x] #4 Loading, empty and error states are handled at the shell level
- [x] #5 Keyboard navigation works for the primary layout without a mouse
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Design direction, decided rather than defaulted.

Typography was chosen from the data, not from taste. The obvious pick for a dense dark tool would have been Inter with a JetBrains-style mono, but 912 of the 1021 entity filenames in the reference corpus are non-ASCII with Cyrillic dominating, and the first candidate considered - Archivo - turns out to have no Cyrillic subset at all. A UI face without Cyrillic would fall back glyph by glyph on most task titles in this user's own projects. Golos Text is a Cyrillic-first grotesque and answers that directly; IBM Plex Mono carries ids, counts and paths and covers Cyrillic too. Both are SIL OFL and both are bundled as woff2 in the repository, 208 KB total, because a local-first application must not fetch fonts from a network it may not have.

The palette has no accent hue, deliberately. Every colour in this interface belongs to a project, and with nine projects on screen a brand hue would be a tenth colour competing with nine that carry meaning. Selection and focus are expressed with brightness and weight instead, which also keeps the focus ring readable against any project colour. Projects that pin no colour in the registry get a stable one derived from their path by FNV-1a over a fixed twelve-colour palette tuned for the dark ground - stable across restarts and unaffected by adding or removing other projects.

Unbuilt screens are listed in the nav and disabled rather than hidden, each saying what it is for and that it is not built yet. The shape of the product is useful information and a nav that quietly grows tells a reader nothing.

Two real defects were found by looking at the running application rather than at the code:
- The status strip reported 884 tasks while the board was narrowed to one project's 56. The narrowing now lives in one derived store that both the board and the strip read, so they cannot disagree, and the strip says '56 of 884' when narrowed.
- First run was being counted as '1 problem' in the strip, contradicting the design intent that a missing registry is an invitation rather than a fault. Both the strip and the banner now exclude it; the empty state already handles it.

Also caught by looking: SVAR ships its own theme, which read as a different application bolted into the window. Its variables are mapped onto the shell tokens. And the board is set readonly until writes go through the CLI in TASK-25 - without it, dragging a card would move it in memory only and show a state the files on disk do not have.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Built the application shell: a top bar with screen navigation, a permanent project roll, a content region, and a status strip, over a token system defined once in app.css.

The roll is the shell's signature - every registered project, always visible, each with its own colour and a right-aligned live count. Clicking or pressing Enter narrows everything to that project. Theme tokens cover colour, the type scale and the fixed chrome heights; nanostores hold the registry, tasks and problems and refresh from the backend change event with no polling.

Verified by driving the running application in a browser against the nine real projects, not by reading the code. The shell renders 884 tasks across nine projects with Cyrillic and Latin titles set in the same face; pressing b keeps Board active and pressing l does nothing because that screen is not built; Tab reaches the project roll with a visible 2px focus ring; Enter narrows the board to Muster's 56 tasks and marks it pressed. The first-run state was checked by pointing the binary at a config home with no registry: it names the exact file to create, shows a working example, and says the window picks it up on save. The desktop binary builds and holds a window open.

Verification found two defects that reading the code would not have: the status strip disagreeing with the board about how many tasks were shown, and first run being counted as a problem. Both fixed. Lint, svelte-check and the Go suite are green.
<!-- SECTION:FINAL_SUMMARY:END -->
