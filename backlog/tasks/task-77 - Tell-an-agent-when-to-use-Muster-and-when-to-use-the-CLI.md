---
id: TASK-77
title: Tell an agent when to use Muster and when to use the CLI
status: Done
assignee:
  - '@claude'
created_date: '2026-08-27 19:04'
updated_date: '2026-08-27 19:07'
labels: []
milestone: m-5
dependencies: []
priority: high
type: enhancement
ordinal: 77000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
An agent connected to Muster and working inside a project has two overlapping ways to do the same thing, and nothing tells it which is which. The boundary is not subtle - the CLI is for acting inside one project and has the whole command surface, while Muster is for seeing across all of them and is the only thing that can - but it has to be said where an agent will read it.

Three places say it, and none of them is a project's CLAUDE.md: this server will usually be running outside any project, so it has to carry its own boundary rather than rely on repositories to state it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The server sends instructions at connection saying what it is for and what it is not
- [x] #2 Every tool's description bounds itself, so a tool read on its own does not mislead
- [x] #3 Read tools are annotated read-only and write tools are annotated as changing things
- [x] #4 What a client receives is checked, not just what the code sets
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
Three places, and the reason each is needed rather than one being enough.

The instructions the server sends at connection are where the boundary belongs, because it is one idea and not ten: this answers across projects, and inside a project the backlog CLI is the authority. It also says where to start, since ids and statuses are per-project and nothing else makes sense before list_projects.

Every write tool repeats a shortened form of it. A tool description is read on its own, out of order, by something deciding between ten of them - relying on the instructions having been read first would be relying on luck.

Annotations say whether a tool changes anything. The six reads are marked read-only and idempotent so a client can stop asking permission for questions; the four writes are not, and the two that overwrite - set_field and set_section - are marked destructive, because they forget what was there. None is idempotent: calling create_task twice makes two tasks.

Checked from the client's side rather than by reading the code that sets it, which is the only way to know an instruction is actually delivered: a test drives a real client and asserts on what it received, and the built binary was spawned the way a client spawns it and printed its instructions and every annotation.

Nothing was added to any project's CLAUDE.md, deliberately. This server usually runs outside any project, so it carries its own boundary rather than relying on repositories to state it.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
An agent connected to Muster while working inside a project had two overlapping ways to do the same thing and nothing telling it which. Now the server says so itself, in three places that cover the three ways an agent meets it.

At connection it sends what it is for and what it is not: across projects, not as a substitute for the backlog CLI inside one. Every write tool repeats a short form of that, since a description is read on its own. And every tool is annotated with whether it changes anything - the six reads read-only, the four writes not, and the two that overwrite marked destructive.

Verified from the client's side both ways: a test that drives a real client and asserts on what arrived, and the built binary spawned as a client spawns it, printing its instructions and every annotation.
<!-- SECTION:FINAL_SUMMARY:END -->
