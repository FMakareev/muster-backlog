---
id: TASK-76
title: Connect an agent client in one click
status: Done
assignee:
  - '@claude'
created_date: '2026-08-27 18:17'
updated_date: '2026-08-27 18:30'
labels: []
milestone: m-5
dependencies: []
priority: high
type: feature
ordinal: 76000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The MCP server exists and the README explains how to point a client at it. That paragraph is the whole cost: it asks someone to find their client's config file or remember its command.

Refloft solves this and the module is worth taking rather than reinventing: a catalogue of clients, each described as either a command to run or a config file to write, with the plan shown in full before anything happens, a backup on every write, and disconnect by the same route as connect.

It is a port rather than a copy. Refloft serves MCP over HTTP, so every entry in its catalogue registers a URL; Muster serves over stdio, so every entry has to register a command and its arguments instead. The command must be the absolute path to this binary - a client will not have it on PATH, which is the same lesson the CLI discovery bug taught.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Preferences list the AI clients this machine has, and whether Muster is already connected to each
- [x] #2 Connecting shows exactly what will be run or written, in full, before it happens
- [x] #3 Connecting works for both kinds of client: those with their own command and those with a config file
- [x] #4 Disconnecting is offered by the same route, and every config write leaves a backup
- [x] #5 The registered command is this binary's own path, so a client can start it without PATH
- [x] #6 A client that is not installed is shown as such rather than offered
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
Ported from refloft rather than rewritten: internal/agents and internal/whichbin, with their tests. The mechanism is the valuable part and it is already careful - a catalogue of clients as data, two kinds of client, a plan shown in full before anything happens, a backup on every config write, disconnect by the same route.

It is a port because refloft serves MCP over HTTP and registers a URL, while Muster is spawned. Every {url} became {command}, and every entry in the catalogue was rewritten for stdio. Three of them needed their own shape rather than a substitution: Claude Desktop no longer needs mcp-remote at all, Zed takes a custom source with the command spelled out, and opencode calls a spawned server local and takes the whole command as one list.

The command registered is this binary's own absolute path, resolved through any symlink. A bare name would reproduce, in somebody else's program, exactly the bug that made a packaged build report the Backlog.md CLI missing - and it would be harder to diagnose there.

whichbin turned out to document the same problem the CLI discovery bug was about, more thoroughly than the code written for it, so backlogcli.Find now looks through whichbin and keeps its own list only for two things whichbin cannot serve: telling someone where the search went, and offering a second candidate when the first is a wrapper that cannot run.

Three things the port itself got wrong, all found by running it rather than reading it. The backup file was still called mcp.json.refloft-backup-… - another project's name, left in somebody else's directory, where it reads as a bug in whatever they were using; there is now a test that nothing in the catalogue or the code mentions where this came from. The ported tests ignored four error returns that this project's linters do not allow, so they are checked. And the scrollable panes showing the command and the diff were not reachable by keyboard, which axe caught: they are regions with a tab stop now, and Svelte's own rule is silenced with the reason written next to it.

Verified in the browser against a stub client on PATH and a real Cursor config file: the list with an installed client and an absent one, the exact command shown before anything ran, nothing written until the button, the client's own command run, the config file gaining our entry while keeping the one already in it, a backup left beside it, and disconnecting removing only ours. Plus a catalogue test over all nine entries and no axe-core violations.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Preferences list the AI clients this machine has and connect one in a click, by running that client's own command or writing its own configuration file - shown in full first, backed up, and undone by the same button.

Ported from refloft rather than rewritten, because the mechanism there is already careful. What changed is the whole reason for the port: refloft serves MCP over HTTP and registers a URL, Muster is spawned, so every catalogue entry now registers this binary's absolute path with the mcp subcommand. Three clients needed their own shape rather than a substitution.

Three defects came out of the porting, all found by running it: a backup file still named after refloft in other people's directories, four error returns the ported tests ignored, and scrollable panes that no keyboard could reach. All fixed, and the first has a test so it cannot come back.

Verified against a stub client and a real Cursor config, with a catalogue test over all nine entries and a clean accessibility pass.
<!-- SECTION:FINAL_SUMMARY:END -->
