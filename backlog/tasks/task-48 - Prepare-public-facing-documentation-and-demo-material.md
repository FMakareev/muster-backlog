---
id: TASK-48
title: Prepare public documentation and demo material
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:03'
updated_date: '2026-08-27 21:57'
labels: []
milestone: m-4
dependencies:
  - TASK-44
  - TASK-52
  - TASK-50
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: medium
type: docs
ordinal: 48000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The niche is narrow but real: backlog browser is single-repo by design, and the VSCode extension switches between backlog folders rather than combining them. Nobody discovers that from a paragraph - the multi-project board and the analytics view have to be shown.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 README shows the board, the list view and the analytics dashboard in a way that renders on GitHub
- [x] #2 A usage guide covers registry, board, list, search, inbox, documents and analytics
- [x] #3 The relationship to Backlog.md is stated: reads markdown directly, writes only through the CLI, adds no field of its own
- [x] #4 Differences from backlog browser and from the VSCode extension are stated plainly and fairly
- [x] #5 Screenshots contain no private project data
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The screenshots are of five invented projects - Atlas, Beacon, Cinder, Drift and Ember - about work nobody is doing. That is the fifth acceptance criterion and it also shaped the corpus: it is built by a script kept beside the images, so anyone can rebuild it and retake them, and it is shaped to show what the product is for rather than to look busy. Ids collide across the five, one project declares a status the others do not, there are subtasks, milestones, drafts waiting and dependencies.

Three things came out of taking them, each a real defect rather than a photography problem.

The status bar showed the full registry path, permanently, on every screen. In real use that reads /home/<name>/.config/muster/projects.yml, so every screenshot anybody ever posts carries their username. It is abbreviated to ~ now. The first attempt at that was a bug and a bad one: abbreviating BoardService.RegistryPath itself would have sent registry.Add, Update, Remove and LoadFrom to a literal ~ that no filesystem expands, breaking every registry write. Display and use are separate methods now, and the reason is written above both. The Projects screen had the same problem for each project's folder, fixed the same way - ProjectView.DisplayPath beside Path, with Path still the identity every filter, selection and write is keyed on.

The figures showed 0001-01-01 for any task nobody has edited. A task with no updated_date parses to the zero time, and the list view had already been fixed for exactly this while the figures had not. The backend was right - it already falls back to the creation date when deciding staleness - so the fix was to show the date the rule actually used, and to say "no date" when there is none.

And the corpus needed age before the figures said anything. A backlog created five minutes ago reports zero blocked, zero stale, average age zero, which demonstrates nothing. Dependencies went in through the CLI, which has a flag. Dates did not - no command sets a creation date - so they are rewritten in the corpus files by the build script, with a seeded random so the same corpus comes out each time. That is only defensible because the corpus is disposable scaffolding for a screenshot.

The comparison is measured rather than asserted. `backlog browser` started in one of the five projects returns that project's tasks from its API and none of the others, and defaults to port 6420 in every project - both checked. The section says plainly that this is not a flaw but a single-project tool being one, and that with a single repository it is the simpler answer and needs nothing installed beyond the CLI. The VSCode extension is named, linked and described as switching between backlog folders rather than combining them, which is what the README already said.

The images are 1440 pixels wide at 880KB for nine, against 2.1MB at twice the density. GitHub renders a README image at about 880 pixels, so the larger set bought nothing but weight.

The guide is doc-7, written through the CLI as the conventions require, covering the registry, board, list, search, task panel, inbox, documents and figures, and preferences. doc-4's index of documents had gone three documents out of date and now lists all seven.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The README shows the board, the list and the figures, all of five invented projects, and says plainly how Muster relates to Backlog.md itself, to `backlog browser` and to the VSCode extension - with the browser comparison measured rather than asserted. A full guide is doc-7.

Taking the screenshots found three defects. The status bar showed the full registry path permanently, which in real use puts a username in every screenshot anyone posts; the Projects screen did the same per project. Both are abbreviated now, through display-only accessors kept separate from the paths every write is keyed on - the first attempt abbreviated the working path and would have broken every registry write. And the figures showed 0001-01-01 for any task nobody had edited, where the backend had already been falling back to the creation date correctly.

The corpus is built by a script kept with the images, so the screenshots can be retaken.
<!-- SECTION:FINAL_SUMMARY:END -->
