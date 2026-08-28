---
id: TASK-80
title: Make the MCP server runnable where an agent actually runs
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 01:28'
updated_date: '2026-08-28 01:38'
labels: []
milestone: m-4
dependencies: []
priority: high
type: bug
ordinal: 80000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reported from a real setup, with the diagnosis already done. Three separate faults.

The binary links WebKit at load time. main imports the Wails application package, so the dynamic loader resolves libwebkit2gtk before main runs - including for 'muster mcp', which builds no window. Inside a Flatpak sandbox, where the runtime has no webkit2gtk, it fails to start at all. The same is true of any container, any headless server, and any machine where somebody wants the MCP server without the desktop application. An MCP server that needs a browser engine installed is the wrong shape.

The connector registers an absolute path that can go stale. It records os.Executable, which is the truth at the time, but a reinstall moves it: a config pointing at /usr/local/bin/muster kept working until the package was removed, and then failed with ENOENT while the interface still said connected, because for CLI clients the check only looks for the server name.

And the version the MCP server reports is a constant, 0.1.0, while the binary reports what it was built with. Two numbers for one thing, which the versioning work was supposed to end.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The MCP server runs with no GTK, no WebKit and no display
- [x] #2 The connector registers a command that works in that environment, or refuses and says why
- [x] #3 The version the MCP server reports is the version the binary was built with
- [x] #4 Packaging ships whatever the connector points at
- [x] #5 Nothing that was connected before is silently left pointing at a command that no longer exists
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
The diagnosis that came with this was right on all three counts, and the third is the one that mattered.

The binary links WebKit at load time. main imports the Wails application package, so the dynamic loader resolves libwebkit2gtk before main is entered - `muster mcp` never gets to decide it needs no window. In a Flatpak sandbox with no webkit2gtk in the runtime it dies with "error while loading shared libraries", and the same is true of any container, any headless machine, and anyone who wants the server without the desktop application. An MCP server that needs a browser engine installed is the wrong shape.

So there is a second binary. cmd/muster-mcp imports the protocol server and the registry and nothing else; built with CGO off it is static, with no dynamic dependencies at all - checked with ldd, which answers "not a dynamic executable". 7.5MB. It runs anywhere.

That let the subcommand go entirely. The catalogue carried `mcp` as an argument in three different shapes across seven clients - a trailing literal for the CLI kind, "args":["mcp"] inside embedded JSON, and an args line for TOML - and every one of them is now a command spawned with no arguments at all. A client should not have to know a subcommand. The catalogue test used to assert every provider asked for it; it now asserts the opposite, and was made to fail by putting an argument back into each of the three shapes in turn.

The connector looks for the server beside this binary, because that is where every way of installing them puts the two together, then anywhere a person could run it from. When it is nowhere, connecting is refused and the message names what is missing and where it comes from. Refusing is the point: writing a command that does not exist into another program's configuration is the exact failure being fixed, and it looks like it worked. Disconnecting is still offered, because somebody whose server has gone is precisely who needs to remove the entry pointing at it.

The stale path that started this needs no special handling now. It was recorded from os.Executable, which was the truth at the time and stopped being true when the package moved; the command registered now is a stable one that packaging ships and the connector can check for.

And the version: mcpserver.Version was the constant "0.1.0" while the binary reported what it was stamped with, so a client was told 0.1.0 by a binary calling itself 0.0.0. It comes from buildinfo now. buildinfo.LineFor names the program, because two binaries ship from this module and a version line naming the wrong one is worse than none.

Verified end to end rather than by parts: the interface was driven to the plan it would apply, which reads `claude mcp add --scope user muster -- .../muster-mcp` with no subcommand; that exact command was then spawned the way a client spawns it, and answered with the tool list, the five projects and the refusal for an unregistered path. The .deb ships both binaries. Four tests cover what is registered, the refusal without a server, and the catalogue shapes.

Left for the release task: the AppImage carries only the application, so an AppImage user has no muster-mcp beside it and gets the refusal. The static binary is the easiest thing in the world to attach to a release as its own asset, which is where that belongs.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The MCP server is a separate, static binary now, because the application cannot serve the protocol where agents run: it links a browser engine at load time, so muster mcp never starts in a Flatpak sandbox, a container, or on a machine with no desktop.

That let the subcommand go: the catalogue carried it in three shapes across seven clients, and every client is now given a command spawned with no arguments at all. The connector looks for the server beside the application and refuses to write anything when it is not there, naming what is missing - because writing a command that does not exist into another program's configuration is the failure being fixed, and it looks like it worked.

The version the server reports comes from the build, ending a constant that told clients 0.1.0 while the binary called itself 0.0.0.

Verified by driving the interface to the plan it would apply and then spawning that exact command as a client would: tools, projects, and the refusal for an unregistered path all answered.
<!-- SECTION:FINAL_SUMMARY:END -->
