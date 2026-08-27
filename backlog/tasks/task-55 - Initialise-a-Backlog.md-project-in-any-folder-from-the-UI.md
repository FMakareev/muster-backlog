---
id: TASK-55
title: Initialise a Backlog.md project in any folder from the UI
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:28'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-3
dependencies:
  - TASK-24
  - TASK-15
priority: high
type: feature
ordinal: 55000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Adding a project should not require dropping into a terminal. The backlog init command is fully non-interactive - every prompt it asks has a corresponding flag, plus a defaults switch - so the UI can present the same choices as a form instead of emulating a dialogue. This replaces any notion of a special personal backlog: any folder can become a project, including one that is not a git repository.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A folder can be picked and initialised as a Backlog.md project from the UI
- [x] #2 The form covers the choices backlog init exposes: project name, backlog directory, config location, task prefix, zero-padded ids, git integration, agent instructions and integration mode
- [x] #3 Each option shows what it does and what the default is, so the form is usable without reading CLI help
- [x] #4 Initialisation runs through the backlog CLI and its output is surfaced on failure
- [x] #5 A folder that already contains a backlog is detected and offered for registration instead of initialisation
- [x] #6 A folder that is not a git repository is supported through the no-git path
- [x] #7 The newly initialised project appears in the registry and on the board without a restart
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
Runs on top of TASK-44's Projects screen; the two are one piece of work split in half.

1. The CLI wrapper already exists and is already non-interactive: Init passes --defaults and sets only what the form fills in, and it refuses the one combination 1.48.0 rejects (agent instructions with AI integration off) before running rather than after.
2. Inspect the folder first and let that decide what is offered: a folder that already holds a backlog is offered for registration, one that does not is offered initialisation, one that is not a git repository gets the no-git path chosen for it rather than presented as a question it will fail on.
3. The form carries the eight choices init exposes - name, backlog directory, config location, task prefix, zero-padded ids, git, agent instructions, integration mode - each with what it does and what the default is written next to it, so the form is usable without reading CLI help.
4. Failure surfaces the CLI's own output rather than a generic message: init writes files, and a half-written folder needs the real reason.
5. On success the folder is registered and reloaded, so it is on the board without a restart.
6. Tests over the option mapping, and the whole path driven through the browser against scratch folders: a plain folder, one that is not a git repository, and one that already holds a backlog.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The folder is inspected before anything is written, and what it turns out to be decides what is offered rather than what is asked. A folder that already holds a backlog is offered for registration and cannot be initialised over. One that is not a git repository has the no-git path chosen for it, since backlog init would refuse otherwise and the question has only one possible answer; a folder that is a repository gets it as a real choice instead.

The task description said the CLI is fully non-interactive because every prompt has a flag plus a defaults switch. That is not quite true, and the gap is a trap. The project name is a positional argument, not a flag, and --defaults does not answer it: 1.48.0 prompts for it anyway, and with no terminal on the other end the command exits 0 having created nothing at all. A silent success leaving an empty folder. So the name is always passed, defaulting to the folder's own name, and the result is checked rather than trusted - if the CLI reports success and there is still no backlog, that is what the interface says, instead of the registration step reporting the wrong problem a moment later.

Failure surfaces what the CLI printed rather than a summary of it, stderr first and then stdout, because init writes files and a half-written folder needs the real reason. The one combination the CLI refuses - agent instruction files with AI integration off - is prevented in the form and refused in the wrapper before the process starts, so nothing is written before the refusal.

Verified through the browser against scratch folders: a plain folder initialised with a task prefix and zero-padded ids that both reached the written config, a git repository told apart from a plain one, a folder that already holds a backlog offered for registration, a path that does not exist explained with no way to save it, and the new project on the board without a restart. Four Go tests cover the same paths against the real CLI, including the empty-name trap and the CLI's own words reaching the interface on failure.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Any folder can become a Backlog.md project from the interface, and adding a project no longer means opening a terminal.

The folder is inspected first and what it is decides what is offered: registration for one that already holds a backlog, initialisation for one that does not, and the no-git path chosen automatically for a folder that is not a repository - a choice, not a question, where the answer could only be one thing. The form carries the eight choices backlog init exposes, each with what it does and what its default is written beside it.

The most useful thing this found is that the CLI is not as non-interactive as the task assumed: the project name is positional and --defaults does not answer it, so without a name init prompts, finds no terminal, and exits 0 having created nothing. The name is always passed now, and the result is verified rather than trusted.

Verified in the browser against scratch folders - a plain folder, a git repository and one that already holds a backlog - with the written config read back to confirm the form's choices reached the CLI, plus four Go tests against the real CLI covering the same paths and the failure messages.
<!-- SECTION:FINAL_SUMMARY:END -->
