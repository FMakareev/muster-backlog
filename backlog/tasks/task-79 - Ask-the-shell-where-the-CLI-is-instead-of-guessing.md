---
id: TASK-79
title: 'Ask the shell where the CLI is, instead of guessing'
status: Done
assignee:
  - '@claude'
created_date: '2026-08-27 23:49'
updated_date: '2026-08-27 23:59'
labels: []
milestone: m-4
dependencies: []
priority: high
type: bug
ordinal: 79000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Muster could not find the backlog CLI on the author's own machine, and the message listed every directory it had searched without listing the one it was in: ~/.local/share/pnpm/bin.

Two separate faults. The list of install directories has ~/.local/share/pnpm but not its bin subdirectory, and PNPM_HOME the same way - pnpm keeps its own binaries directly in that folder and the ones it installs globally in bin underneath, so both have to be searched. That is a one-line fix and it fixes this machine.

The second is the shape of the problem. A finite list of guesses will always be missing somebody's directory, and every entry added is a guess that happened to be right once. exec.LookPath is already the equivalent of running which inside the process, and it fails for the reason the whole package exists: a desktop launcher does not read the shell profile where the PATH entry lives.

What does work is asking the owner's login shell, which reads exactly those files. Measured on this machine: `$SHELL -lic 'command -v backlog'` answers with the right path in about 1.5 seconds. That is slow enough to be a last resort and cheap enough to be worth it, because it needs no list and cannot go stale.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The CLI is found where pnpm installs it globally
- [x] #2 A command outside every known directory is still found, by asking the login shell
- [x] #3 Asking the shell cannot hang the application or run for long
- [x] #4 The shell is asked only after the cheap lookups fail, and the answer is not asked for twice
- [x] #5 A shell that answers with something that is not an executable path is ignored
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
Two faults, and the second is the one worth having.

The immediate one: pnpm keeps its own binaries directly in PNPM_HOME and the ones it installs globally in bin underneath it. Both lists - whichbin's and backlogcli's - had the first and not the second, so a CLI installed with `pnpm add -g` sat in ~/.local/share/pnpm/bin while the error listed ~/.local/share/pnpm. Both lists now have both, and a test covers a launcher's environment: nothing on PATH, no PNPM_HOME, no shell.

The second is the shape. The question asked was "can we not just run which?", and the answer is that exec.LookPath already is that, run inside this process, and it fails for the reason the package exists: a desktop launcher hands the process the system PATH, not the one a terminal shows. But the owner's shell does know, because it reads the profile where that entry was written. Measured here: `$SHELL -lic 'command -v backlog'` answers correctly in about 1.5 seconds. So that is now the last resort, after PATH and after the list, and it needs no list and cannot go stale.

Three things had to be right about it.

It cannot hang the application, and the first version did. exec.CommandContext kills the shell on timeout, but Output waits on the pipes, and anything a profile starts inherits them - so a profile with a background process held it for the full sixty seconds of the very test written to prove it would not. WaitDelay gives up on the pipes shortly after the kill, and on Unix the shell is put in its own process group so the kill reaches what the profile started rather than only the shell.

It cannot be shell syntax. The name goes in through an environment variable and the command is `command -v -- "$WHICHBIN_NAME"`, so nothing a caller passes is ever parsed. A test passes a name containing a command substitution and checks the command did not run.

It cannot be believed blindly. `command -v` answers with a bare word for a function, an alias or a builtin, and a profile prints banners and version-manager notices besides. Only an absolute path to something runnable is accepted, taken from the end of the output.

Asked once per name per process, because a second and a half is far too long to pay twice for an answer that will not change while the application runs.

Two test-isolation leaks surfaced, both real. XDG_DATA_HOME was left pointing at the developer's own ~/.local/share, so the search reached into it - invisible until pnpm/bin was added to the list and it started finding the real install. And the shell fallback reads the owner's real profile by design, which no fake HOME can isolate, so the tests that assert nothing is found now set SHELL to empty and say why.

The message says the shell was asked, and only when there was one to ask.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The CLI is found where pnpm actually puts it, and — more usefully — where anyone puts it.

pnpm keeps its own binaries in PNPM_HOME and the globally installed ones in bin underneath; both lists had only the first, so the error named every directory except the one holding it.

The general fix answers the question that prompted it. Running which inside the process is what exec.LookPath already does, and it fails because a launcher gives the process the system PATH. The owner's login shell does know, because it reads the profile where that PATH entry was written, so it is asked last: about 1.5 seconds, once per name, bounded so a hanging profile cannot take the application with it, with the name passed as data rather than syntax and only an absolute path to something runnable believed.

Verified against a launcher-like environment for both: pnpm's global bin, and a tool in a directory no list contains that only the shell knows about.
<!-- SECTION:FINAL_SUMMARY:END -->
