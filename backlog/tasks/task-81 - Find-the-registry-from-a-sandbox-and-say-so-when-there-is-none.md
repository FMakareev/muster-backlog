---
id: TASK-81
title: 'Find the registry from a sandbox, and say so when there is none'
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 02:03'
updated_date: '2026-08-28 02:11'
labels: []
milestone: m-4
dependencies: []
priority: high
type: bug
ordinal: 81000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Two faults found by running the MCP server from a client inside a Flatpak sandbox.

The registry is not found. registry.DefaultPath resolves through XDG_CONFIG_HOME, and a sandboxed client has that redirected - inside Obsidian's Flatpak it is ~/.var/app/md.obsidian.Obsidian/config, where no muster registry exists. The real one, with eleven projects in it, is at ~/.config/muster/projects.yml. The same is true of any sandboxed or containerised client, and there is no way to say where the registry is short of setting XDG_CONFIG_HOME for the whole process.

And when it is not found, nothing says so. New and Reload both swallow ErrNoRegistry and carry on with an empty store, which is right for the application on first run - there is no registry yet and the interface offers to make one - and wrong for an agent, which asks what projects there are and is told none. An empty answer and a missing file are not the same thing, and the difference is exactly what a person debugging this needs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A registry outside the resolved config directory is still found
- [x] #2 There is a way to say where the registry is without changing the whole environment
- [x] #3 Reading and writing agree on one path, so nothing ends up split between two
- [x] #4 An agent asking what projects exist is told the registry is missing, not that there are none
- [x] #5 An existing setup keeps reading the file it reads today
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Both faults were real and neither was particular to that machine.

The registry was looked for through XDG_CONFIG_HOME, and a sandboxed client has that redirected: inside Obsidian's Flatpak it is ~/.var/app/md.obsidian.Obsidian/config, which has never held a registry. The same happens in any container and under any other sandboxed client. DefaultPath now falls back to ~/.config/muster/projects.yml when the resolved path holds nothing, so the ordinary case needs no configuration at all, and MUSTER_REGISTRY names it outright for anything that cannot be handed a different environment. Read and write still agree, because both go through the same function: the fallback only applies when the resolved path has no file, so an existing setup keeps reading exactly what it reads today.

One thing had to change to make that work and is worth knowing: adrg/xdg computes ConfigHome once at package init, so reading it answers for the environment the process started with and ignores anything set later. The variable is read directly now when it is set and absolute, which is what the specification says, with the library's answer - including its ~/.config default - standing otherwise. Without this the test could not set the variable at all, and it failed against the real home directory.

The second fault is the one that misleads. New and Reload swallowed ErrNoRegistry and served an empty store. That is right for the application - first run has no registry and the interface offers to make one - and wrong for an agent, which asked what projects existed and was told none. An empty answer and a missing file are not the same thing and a client cannot tell them apart. Every tool that answers across projects now reports the missing registry instead, naming the file it looked for and the variable that points elsewhere. The server still starts, because a client that cannot connect cannot be told anything.

Verified through the protocol, not by unit test alone: with XDG_CONFIG_HOME redirected into a sandbox directory holding no registry, list_projects returned all five projects from the real home; with no registry anywhere it returned an error naming the path; and MUSTER_REGISTRY pointed a server with an empty home at a registry elsewhere. Four Go tests cover the resolution and one drives all four cross-project tools through a session against a missing registry.

A test had to change shape along the way. TestConnectingIsRefusedWithoutTheServer arranged an environment with no server binary, which stopped working the moment the package was installed on this machine: whichbin searches absolute directories like /usr/local/bin that no environment variable can hide. It replaces the command function instead, so it tests the refusal rather than the filesystem.

What is not solved, and cannot be from this side: the path to the binary. A Flatpak sandbox cannot see /usr/local/bin at all, so a client inside one has to reach the host itself - flatpak-spawn --host - or be given a copy inside the sandbox. The connector registers a path that is correct on the host, which is the only path it can know. The README says so plainly rather than leaving it to be discovered.

Also noted from the same report, not acted on: piping into the server and closing stdin immediately makes it exit with "server is closing: EOF" before writing anything. Clients hold stdin open so it does not affect them, but a smoke test written as a pipe would fail for that reason and not a real one.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The registry is found when XDG_CONFIG_HOME is redirected, which is what every sandboxed client does, and can be named outright with MUSTER_REGISTRY for anything that cannot be handed a different environment. The fallback applies only when the resolved path holds nothing, so an existing setup reads exactly what it read before and reading and writing still agree on one path.

And a missing registry is now said out loud. The server used to swallow it and serve an empty store - right for the application on first run, wrong for an agent, which asked what projects existed and was told none. Every tool that answers across projects now names the file it looked for and the variable that points elsewhere.

Verified through the protocol: redirected into a sandbox directory it found the real registry and listed five projects; with none anywhere it returned the error; and MUSTER_REGISTRY redirected a server whose home was empty.

The path to the binary is not solved and cannot be from this side - a Flatpak sandbox cannot see /usr/local/bin, so a client inside one needs flatpak-spawn --host or its own copy. The README says so.
<!-- SECTION:FINAL_SUMMARY:END -->
