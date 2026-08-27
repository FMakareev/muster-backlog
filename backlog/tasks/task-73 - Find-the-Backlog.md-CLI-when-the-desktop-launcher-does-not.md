---
id: TASK-73
title: Find the Backlog.md CLI when the desktop launcher does not
status: Done
assignee:
  - '@claude'
created_date: '2026-08-27 17:17'
updated_date: '2026-08-27 17:26'
labels: []
milestone: m-5
dependencies: []
priority: high
type: bug
ordinal: 73000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reported from a packaged build: a deb installed in a container and exported to the host shows 'Changes cannot be saved - the backlog CLI was not found: install it and make sure backlog is on PATH', while backlog runs fine in a terminal on the same machine.

Measured cause: the CLI lives in ~/.local/share/pnpm/bin, which is on PATH only because a shell rc file puts it there. An application started from a desktop launcher inherits the session environment, not the shell's, so exec.LookPath finds nothing. This is not specific to pnpm - npm, bun and cargo global bins are all in the same position.

Looking on PATH is right and should stay first. What is missing is everything after it: the places this CLI is actually installed, a way to say where it is when discovery is wrong, and an error that names where it looked instead of repeating advice the person has already followed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The CLI is found when it is installed in a package manager's global bin and PATH does not mention it
- [x] #2 An explicit path can be set in preferences and wins over discovery
- [x] #3 When it still cannot be found, the message names every place that was searched
- [x] #4 A found binary is still checked for version, and a wrong one is refused with the reason
- [x] #5 Nothing is executed from a directory that is not the one it was found in
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
There were two causes, and the second was invisible until the first was fixed.

The one in the report: the CLI lives in ~/.local/share/pnpm/bin, which is on PATH only because a shell rc file puts it there, so exec.LookPath finds nothing when the application is started from a desktop launcher. Muster now looks on PATH first and then in the places package managers actually install into - PNPM_HOME, BUN_INSTALL, NPM_CONFIG_PREFIX, the XDG data directory, the pnpm, bun, npm-global, cargo and .local/bin directories under home, and the system ones.

The second: the binary pnpm installs is not a binary. It is a shell script ending in exec node, and node on this machine comes from nvm, which is on PATH for exactly the same reason and no other. So even pointed straight at the right file, the CLI failed with 'exec: node: not found'. Reproduced with env -i HOME=... PATH=/usr/bin:/bin. The CLI is now run with a PATH that includes the node version manager directories as well, which is what makes a found binary actually work.

Two consequences of that. Every candidate is tried rather than the first that exists, because a file can be there and still not run. And an explicit path from the preferences is used as given and its failure reported about that path, rather than quietly searching elsewhere and reporting something about a binary the person did not name.

The failure message now lists every directory searched and explains why a CLI that works in a terminal can be missing here - being told to install something you have already installed is no help.

Setting the path takes effect immediately rather than at the next start: someone typing it is looking at a banner saying writes are unavailable, and asking them to restart to find out whether they guessed right is a poor answer. The stale problem is dropped when the CLI resolves.

Verified by reproducing the report - the server started with env -i, a launcher's PATH and the author's own home, where it now finds and runs the CLI - and separately with a home containing nothing, where it reports every place it looked and then works once the path is set in Preferences. Seven Go tests cover discovery, the environment-variable override, the version check, the explicit path, passing over a binary that cannot run, and a wrapper that needs node.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A packaged build could not find the Backlog.md CLI although it ran in a terminal on the same machine. Two causes, one hiding behind the other.

The CLI is installed where only a shell rc file puts it on PATH, and a desktop launcher does not inherit that. Muster now searches the directories package managers install into as well. But the binary pnpm installs is a shell script that execs node - and node is installed the same way, so it failed even when found. The CLI is now run with a PATH that can find its own interpreter, which is what made the fix real rather than half of one.

Where it still cannot be found, the message names every directory searched and says why a terminal can find what this cannot, and Preferences takes an explicit path that applies immediately.

Verified by reproducing the reported environment with env -i and a launcher's PATH, and with seven Go tests.
<!-- SECTION:FINAL_SUMMARY:END -->
