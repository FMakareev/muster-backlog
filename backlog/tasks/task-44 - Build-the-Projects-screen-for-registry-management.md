---
id: TASK-44
title: Build the Projects screen
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:03'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-3
dependencies:
  - TASK-24
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: high
type: feature
ordinal: 44000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The registry stops being a hand-edited file. Projects are added, removed, reordered and recoloured from the application, and the screen is where a folder is turned into a Backlog.md project in the first place.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A project can be added by folder path and removed from the registry in the UI
- [x] #2 Display name, colour and ordering are editable per project
- [x] #3 Each project shows its task counts, milestone progress and the Backlog.md configuration it declares
- [x] #4 A folder without a backlog offers initialisation from this screen
- [x] #5 Registry edits are written back preserving comments where possible
- [x] #6 An invalid or unreadable path is rejected with an explanation before it is saved
- [x] #7 A project can be temporarily hidden from the board without being removed from the registry
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
The registry is Muster's own file, and a person may well have annotated it, so rewriting it from a struct is not acceptable: measured on a probe, a struct round trip loses every comment, reorders keys alphabetically and changes the list indentation.

1. Registry writer. Edit the YAML through goccy's comment-preserving parser and print the file back. Measured on the same probe: replacing or appending one sequence element with a node built by ValueToNode keeps the top-level comments, the other entries' inline comments, the key order and the indentation; only the entry actually edited loses its own trailing comment, which is the honest limit of 'where possible'. Two things it must not do, both found on the probe: splicing a key parsed from a separate snippet into an existing mapping prints a truncated key, and replacing the whole sequence at once flattens the list indentation. So every operation works one entry at a time.
2. Identity is the path, not the index: paths are unique by construction and an index goes stale the moment the file is edited outside the application. Position is passed explicitly only for reordering.
3. Writes are atomic - temp file and rename - and a file that will not parse is refused rather than overwritten.
4. Hidden projects. A hidden entry is still loaded, so the Projects screen can show its counts and unhiding is instant, but every query-facing surface skips it. One helper over the store's project order rather than a check in each of Query, Search, Entities, Analytics and the status counts.
5. Service methods: add, update, remove, move, hide, and a folder inspection that says whether a path exists, is a git repository, and already holds a Backlog.md project. Each write reloads and emits, like every other write in the application.
6. The screen: registered projects in file order, each with its counts, milestone progress and the configuration it declares; name and colour editable in place; reorder, hide and remove; and adding a folder, which is where TASK-55 takes over.
7. Tests over the writer against a file with comments, and the screen driven through the browser against scratch folders - never the real registry.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The registry stays a file a person can read and write; this edits it in place rather than replacing it.

Measured on a probe before committing to an approach: marshalling the struct back loses every comment, sorts the keys alphabetically and flattens the list indentation. Printing goccy's comment-preserving tree keeps all three. Two limits of that library shaped the code, both found by trying it rather than by reading about it - a key parsed from a separate snippet and spliced into an existing mapping prints truncated (hidden: true came out as en: true), and replacing the whole sequence at once flattens the indentation. So every operation touches one element, and an element that changes is rebuilt whole. Only the entry actually being edited loses a trailing comment of its own, which is the honest limit of preserving them.

A third defect the probe did not catch and a test did: appending a block entry to a list written inline, which is what projects: [] looks like on an empty registry, produced YAML the parser rejected - and it went to disk. Inline lists are now rebuilt as block lists, and nothing is written that cannot be read back, checked on the bytes before the file is replaced. Writes are atomic and a registry that will not parse is refused rather than overwritten, because an unparseable registry is usually one someone is in the middle of editing.

Identity is the path, never the index: an index goes stale the moment the file is touched outside Muster. Staticcheck then found a real bug behind that: resolution overwrites Entry.Path with the expanded absolute path, so the form a person wrote - ~/Dev/thing - was already gone by the time an edit was written back, and saving would have quietly rewritten their file. Project now carries Written alongside, and a test covers it.

Hiding is a display choice rather than an unregistering: the project stays in the file, stays loaded, and is left out of the board, the lists, search, the figures and the problem list. Loading it anyway is deliberate - the Projects screen still shows what it holds and unhiding costs nothing. One predicate over the store's project order rather than a check in each of seven places.

Two smaller things this surfaced. The status strip counted hidden projects, disagreeing with the screen above it. And placeholders had become indistinguishable from values after the dim tiers were lightened for contrast - the colour field showed #7aa2f7 to three projects that had none set. Placeholders are italic now.

Verified in the browser against scratch folders, never the real registry: adding, renaming, recolouring, reordering, hiding and unhiding, unregistering with a confirmation and with the folder left untouched, and the file still carrying its header comment, its inline comment and its wip_limits section after all of it. axe-core reports no violations on the screen or on any state of the add form; two arrow buttons were 7px wide before it said so.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The registry is no longer a file you have to open a text editor to change.

The Projects screen adds folders, renames and recolours them, arranges the order, hides a project from the board without unregistering it, and shows what each one holds - task and draft counts, milestone progress, the layout and statuses it declares. Unregistering removes the entry and leaves the folder alone.

Edits are written through the YAML tree rather than marshalled from a struct, so comments, key order and indentation survive and only an entry that actually changes is rewritten. Writes are atomic, a registry that will not parse is refused rather than overwritten, and nothing is written that cannot be read back - a check that earned itself immediately by catching a file this code would otherwise have corrupted.

Verified in the browser against scratch folders through every operation, with the resulting file read back each time; nine Go tests over the writer and eight over the service; and an axe-core audit that found two undersized targets and now reports none.
<!-- SECTION:FINAL_SUMMARY:END -->
