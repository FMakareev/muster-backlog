---
id: TASK-82
title: >-
  Register an MCP command the client can spawn, not one only the container can
  see
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 02:44'
updated_date: '2026-08-28 02:56'
labels: []
milestone: m-4
dependencies: []
priority: high
type: bug
ordinal: 82000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found by installing the .deb inside a distrobox container, exporting the application to the host, and connecting Claude Code from Muster's own connector.

mcpCommand in internal/app/agents.go resolves the server beside the running binary. Inside the container that is correct: os.Executable is /usr/local/bin/muster, /usr/local/bin/muster-mcp sits beside it, and os.Stat confirms it. So the connector writes /usr/local/bin/muster-mcp.

But a distrobox container shares $HOME with the host. The file the connector edits is therefore the host's ~/.claude.json, while the path it writes is only meaningful inside the container. Claude Code runs on the host, where /usr/local/bin holds no muster-mcp, and the server fails to start with ENOENT: no such file or directory, posix_spawn. Nothing in the client says where the path came from or why it is wrong.

This is the failure the mcpCommand comment already names - writing a command that does not exist into another program's configuration - reached by a route the check cannot see. os.Stat answers for the filesystem Muster is looking at, not for the one the client will spawn from, and when Muster runs in a container those are different filesystems that happen to share the config file.

The same shape applies to any containerised install that shares $HOME: a toolbox container, a podman container run with the home directory bound, a devcontainer. Flatpak is the mirror image, and README already covers it: there the client is sandboxed and cannot see a host path. Here Muster is contained and the client is not.

A container is detectable from inside - /run/.containerenv for podman, /run/host and the CONTAINER_ID environment variable for distrobox and toolbox - so the connector can know it is in one before it writes anything.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Connecting a client while Muster runs in a container registers a command that resolves in the environment the client will run in
- [x] #2 When no such command can be produced, the connector refuses and says why, rather than registering a path that fails with ENOENT
- [x] #3 The plan shown before connecting states which environment the registered command will run in, so the person can see it before it is written
- [x] #4 An install that is not in a container registers the same path it registers today
- [x] #5 Tests cover the containerised case and the ordinary one, without requiring a container to run them
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
This one could be verified against the real thing rather than a model of it: this repository is developed inside the very distrobox container the report describes, so the failing branch is the branch the code takes here by default.

Confirmed before writing anything. /run/.containerenv and /run/host are present, CONTAINER_ID=devbox and container=podman are set, and /usr/local/bin holds both binaries from the .deb installed in here - the same package the host does not have.

The fix rests on what the container's own tooling already provides. distrobox-export writes a small script into ~/.local/bin, and reading the one on this machine settled the design: called from the host with CONTAINER_ID unset it runs distrobox-enter into the container; called from a different container it hops back out through distrobox-host-exec; called from its own it execs the binary directly. One path that resolves from every side, in the directory the container shares with the host - which is the same sharing that makes the client's configuration file reachable from in here at all, and therefore the same sharing that made the container's own paths meaningless out there.

So: detect the container, and inside one register the copy in the shared home rather than the one beside the binary. A plain copy of the server there is accepted too, since it is a static binary and what matters is the path being on the shared home, not what put it there. With nothing there at all it refuses and prints the distrobox-export line to run, because writing a path only the container can see is the exact failure being fixed and it looks like it worked.

Detection uses the marks every runtime leaves - /run/.containerenv for podman and therefore for distrobox and toolbox, /.dockerenv for docker, /run/host for the host mount, and the CONTAINER_ID and container variables. It is a function variable, so both branches are testable without being run in a container, and one test checks the detection against the marks actually present on whatever machine runs it.

The plan states the environment before anything is written. That was the missing half of the original failure: the path was wrong in a way nothing said out loud, and the client that failed on it could not say where the path came from.

Verified in the running application, in this container: the plan now shows /home/fmakareev/.local/bin/muster-mcp rather than /usr/local/bin/muster-mcp, with the line about which environment it resolves in above the client's own note. That exact command was then spawned the way a client spawns it and answered with the five projects and the refusal for an unregistered path.

An install outside a container is untouched: the ordinary branch is the same code it was, and its test pins that the environment line is empty there, since there is nothing to say.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Muster now detects that it is running in a container and registers the copy of the server in the shared home directory, which resolves whether the client runs inside the container or outside it, rather than a path that is real inside and meaningless out.

The design came from reading the wrapper distrobox-export already writes: from the host it enters the container, from another container it hops back out, from its own it execs directly. One path, every side. With nothing in the shared home it refuses and prints the export command, because writing a path only the container can see is the failure being fixed and it looks like it succeeded.

The plan says which environment the command will run in before anything is written - the half that was missing when this went wrong, since the client that failed could not say where the path came from.

Verified against the real case: this repository is developed inside that same distrobox container, so the plan was driven in the running application and the registered command spawned as a client spawns it.
<!-- SECTION:FINAL_SUMMARY:END -->
