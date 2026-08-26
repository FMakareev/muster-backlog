---
id: TASK-66
title: Make a draft as workable as a task
status: To Do
assignee: []
created_date: '2026-08-26 23:40'
labels: []
milestone: m-3
dependencies: []
priority: high
type: feature
ordinal: 66000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Drafts went unused because there was nothing to do with one. Capture is cheap and everything after it was not: a draft could not be created from the application at all, could not carry a priority or a milestone, and could only be read as a title and a body.

What the CLI actually allows is more than the drafts inbox uses. task create --draft takes the whole task field surface - description, labels, assignee, priority, type, milestone, acceptance criteria, dependencies, references, ordinal - and writes a draft with it. What it still does not allow is editing one: task edit refuses a DRAFT id on 1.48.0 and on 1.50.1 alike, so rewriting stays capture-and-discard, and the fields carried across should be all of them rather than four.

A promoted draft is an ordinary task and the panel already edits those completely, so promotion is the other half of the answer: promote and open, rather than promote and go looking for it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A note can be captured into the inbox from the application, with the fields the CLI accepts for a draft
- [ ] #2 Rewriting a draft carries every field the CLI can set, not only the title and body
- [ ] #3 A draft opens in the task panel so its whole content can be read
- [ ] #4 Promoting a draft opens the resulting task, so the note can be finished in one go
- [ ] #5 The inbox says what a draft cannot hold and why, rather than silently dropping it
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
- [ ] #5 Linters and formatters pass across Go and frontend
- [ ] #6 Automated tests cover the change and the suite is green
- [ ] #7 User-facing behaviour change is reflected in README or docs
- [ ] #8 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->
