---
id: TASK-45
title: Expose cross-project context over an MCP server
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:03'
updated_date: '2026-08-27 18:08'
labels: []
milestone: m-4
dependencies:
  - TASK-17
  - TASK-24
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: low
type: feature
ordinal: 45000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
With MCP an agent talks to Muster instead of one CLI per repository, and gains context no single project has: what is in flight elsewhere, which milestone is active, where the dependencies point. Built on the Go MCP SDK. Strictly optional - the application is complete without it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 MCP server exposes the aggregated task set, the registry, milestones, documents and decisions as read tools
- [x] #2 Tools report per-project counts by status so an agent can see current load
- [x] #3 Any write operation exposed goes through the same CLI adapter as the UI
- [x] #4 The server can be enabled and disabled independently of the desktop window
- [x] #5 Access is confined to the registered projects
- [x] #6 Connecting an agent to the server is documented end to end
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
Backlog.md already ships an MCP server, and it is per-project stdio: an agent gets one repository at a time. The thing Muster has that no single project does is the aggregate - what is in flight elsewhere, which milestone is active, where a dependency points - so that is what this exposes, and it is the whole reason for building it rather than pointing people at the CLI's own.

1. A subcommand rather than a second binary: muster mcp speaks stdio and never constructs the desktop application, which is what makes it independent of the window.
2. Reads come from the same store the interface reads, built from the same registry. Writes go through the same backlogcli runner, so there is still exactly one writer.
3. Every tool that names a project resolves it against the registry and refuses anything else. An agent cannot reach a folder the person has not registered.
4. Read tools: the projects with their counts and configuration, tasks with the same filters the board has, one task in full, cross-project search, milestones with progress, drafts, documents and decisions, and per-project counts by status.
5. Write tools: create a task, capture a note, and the four one-field edits the board uses - status, priority, milestone, labels. Anything more is the panel's job.
6. Documented end to end in the README: the command, the client configuration, and what each tool answers.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Backlog.md ships its own MCP server and it is per-project stdio - an agent gets one repository at a time. That settled the scope: this one exists only for the aggregate, and every tool answers across the registry or refuses.

muster mcp is a subcommand rather than a second binary and never constructs the application, which is what makes it independent of the window and also what an MCP client expects: a process it spawns and talks to over a pipe. Reads come from the same store the interface reads; writes go through the same backlogcli runner, so there is still exactly one writer.

Access is confined by one function that every tool calls. A project is resolved by name or path against the registry and nothing else; anything unregistered is refused with the list of what is registered, so an agent can correct itself rather than guess again. Ids are qualified by project throughout, because they collide across projects freely.

Two decisions worth recording. Every call reloads the registry and the projects first: agents are long-lived and the files are written by other things, so a snapshot taken at connection would drift further from the truth the longer the session lasted. And a missing CLI does not stop the server - reads work, writes say why when they are called, and the runner is retried once in case it was installed since the agent connected.

Ten tools: the projects with their counts by status and declared configuration, tasks with the board's filters, one task in full, cross-project search, milestones with progress, documents and decisions and drafts with their bodies, and four writes - create a task or capture a note, set one field, set a label, replace a section.

Verified two ways. Six Go tests drive a real client over an in-memory transport against two real projects built by the CLI, covering the aggregate answer, the filters, an id that does not resolve in the wrong project, refusal of unregistered paths, writes landing on disk as asked, and every tool carrying a description and a schema. Then the built binary was spawned as a client would spawn it, speaking JSON-RPC over its stdin and stdout against the author's own registry: it reported eleven projects with their counts, listed its ten tools, and refused /etc while naming what is registered.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
muster mcp serves the aggregate over the Model Context Protocol: every registered project in one answer, which is the only thing this offers that Backlog.md's own per-project MCP server does not.

It is a subcommand that draws no window, so it works whether or not the desktop application is running. Reads come from the store the interface reads and writes go through the same CLI adapter, so there is still one writer. Every tool resolves the project it is given against the registry and refuses anything else, naming what is registered so an agent can correct itself.

Ten tools, each with a description written for something that has to choose between them without asking.

Verified with six Go tests driving a real MCP client against real projects, and by spawning the built binary the way a client does and talking JSON-RPC to it over stdio against the author's own eleven projects. Documented end to end in the README, including the client configuration.
<!-- SECTION:FINAL_SUMMARY:END -->
