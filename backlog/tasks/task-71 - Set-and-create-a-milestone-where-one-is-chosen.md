---
id: TASK-71
title: Set and create a milestone where one is chosen
status: Done
assignee:
  - '@claude'
created_date: '2026-08-27 03:38'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-5
dependencies: []
priority: high
type: feature
ordinal: 71000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Two gaps found in first use, both about the same thing.

A task's milestone can be set when the task is created and never afterwards: the panel shows it as read-only text, and SetMilestone has been wrapped in the CLI layer since the beginning with nothing calling it. Since a milestone is the axis this backlog is planned on, moving a task between them is not a rare act.

And a milestone can only be created on the Projects screen, which is the wrong place to be standing when you discover you need one - you are assigning a task and the milestone you want does not exist yet. Every picker should be able to make one.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A task's milestone can be changed and cleared from the panel
- [x] #2 A milestone can be created from any picker that offers one, without leaving what you were doing
- [x] #3 The new milestone is selected straight away and appears everywhere the others do
- [x] #4 One picker serves every place a milestone is chosen, so the three do not drift
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
- [ ] #5 Linters and formatters pass across Go and frontend
- [ ] #6 Automated tests cover the change and the suite is green
- [ ] #7 User-facing behaviour change is reflected in README or docs
- [ ] #8 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
SetMilestone had been wrapped in the CLI layer since the beginning and nothing ever called it, so a task's milestone could be set at creation and never afterwards. It is a control in the panel now, alongside status, priority and assignee, and clearing it goes through --clear-milestone, which is a different command behind the same control.

Creating one is offered wherever a milestone is chosen, because that is where you discover you need it: the create form, the panel and the inbox all use one component, so the three cannot drift apart and behaviour added to it arrives everywhere at once. The new milestone is selected as soon as it is written, and one refresh puts it on the board, the cards and the other pickers.

Two mistakes of mine are worth recording rather than hiding. The picker's sentinel value for 'make a new one' went into the file as a literal NUL byte where a space was meant. The rendered option value carried that NUL, Playwright could not select it, grep treated the file as binary and stayed silent about it, and a replacement that should have fixed it matched nothing and reported nothing, because that one was written without an assert. The sentinel is __new__ now, the file is clean, and the rest of the source tree was checked: no NUL bytes outside the font files, where they belong.

Verified in the browser against a scratch bench with the files read after each write: the panel setting a milestone, moving it to another, and clearing it; the picker creating one and putting the task in it straight away; and the new milestone appearing in the create form's picker and the inbox's, both of which offer creation as well. A Go test covers set, move and clear against the real CLI, and axe-core is clean across the panel, the creation state and the create form.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A task's milestone can be changed and cleared from the panel, and a milestone can be made from any picker that offers one.

Both gaps were the same shape: the milestone is the axis this backlog is planned on, and it could only be decided at creation, from a list that could only be added to on another screen. One component now serves the create form, the panel and the inbox, so the three cannot drift and a new milestone appears everywhere at once.

Verified in the browser with the files read after every write, a Go test over set, move and clear against the real CLI, and a clean accessibility pass. Recorded in the notes: a literal NUL byte in the component's sentinel value cost most of the debugging, and it hid partly because a replacement of mine ran without an assert and reported nothing when it matched nothing.
<!-- SECTION:FINAL_SUMMARY:END -->
