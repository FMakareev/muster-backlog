---
id: TASK-74
title: Read and write task comments
status: Done
assignee:
  - '@claude'
created_date: '2026-08-27 17:27'
updated_date: '2026-08-27 19:18'
labels: []
milestone: m-5
dependencies: []
priority: medium
type: feature
ordinal: 74000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The parser has handled the comments envelope since the format was mapped, and the panel shows what is there. Writing one means a terminal: task edit --comment appends, with --comment-author recording who said it.

Ten of 875 tasks carry comments today, which is either a fair measure of how useful they are or a measure of how awkward they have been to add. The cheap way to find out is to make adding one cost nothing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A comment can be added to a task from the panel
- [x] #2 The author recorded is the person using the application, not a fixed string
- [x] #3 New comments appear in the thread immediately, in the order the file holds them
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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. The CLI appends a comment with task edit --comment, and records who said it with --comment-author. Measured: without an author the author line is simply absent from the file, not filled with a placeholder, so an unsigned comment is a real state rather than a failure.
2. Who the person is has to come from somewhere真. Nothing in Backlog.md's config carries it - no default_assignee in any of the nine projects - but git does, and the corpus already agrees with it: the assignees in these files are @FMakareev and @claude, and git config user.name here is FMakareev. So: a preference if one is set, then the project's git identity, then nothing at all rather than a made-up name.
3. The application resolves that, not the interface: identity is not something a form should decide.
4. The panel gets a box under the comments it already shows, and Preferences show which name will be signed so it is not a surprise.
5. Tests against the real CLI for the round trip and for the unsigned case, and a browser pass reading the file afterwards.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Who signs a comment had no obvious source. Backlog.md's config carries nothing for it - no default_assignee in any of the nine projects - but git does, and the files already agree with git: the assignees here are @FMakareev and @claude, and git config user.name reports FMakareev. So the order is a preference if one is set, then whatever git answers for that folder, then nothing. Git is read rather than duplicated, and cached per project because a panel showing the author under every comment would otherwise start a subprocess on every render.

Not signed with something invented. Measured: task edit --comment without --comment-author leaves no author line at all rather than an empty one, so an unsigned comment is a state the format has, and writing one is the honest answer when nobody knows the name.

Two defects came out of writing it, both found by running the tests rather than reading them.

The parser could not read an unsigned comment. It required an author line to begin an entry, so a comment written without one was silently dropped - the file had it, the panel did not. Every comment in the corpus is signed, which is why it survived until the application started writing them. Now a date can begin an entry too, with a test.

And the tests themselves wrote into the developer's own configuration. SaveSettings resolved its path through the XDG package, which nothing could redirect, so a test that saved a preference put an author into the real ~/.config/muster/settings.yml. It did. The preferences path is injectable now, the way the registry's already was, every test gets its own beside the registry it wrote, and the file was put back.

Verified in the browser reading the file after each action: an existing comment shown, the name it will be signed with shown before it is used, a two-paragraph comment arriving whole and signed with that name, appearing after the one already there, and the box emptying for the next. Four Go tests cover the round trip, the preference, the git identity and the unsigned case, plus one in the parser. No axe-core violations.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Comments can be written from the panel, not only read.

They are signed with a name from the preferences, or with whatever git answers for that folder - which is where a person's name already is, and what these files already agree with. With neither, a comment is written unsigned, because that is a state the format has and a name is not something to invent.

Two defects surfaced while building it. The parser dropped any comment without an author line, which is exactly what the CLI writes when no author is given - the file had it and the panel did not. And the test helper wrote preferences into the developer's own configuration, which one test proved by leaving an author in it; the settings path is injectable now and the file was restored.

Verified in the browser against the file on disk, with five Go tests and a clean accessibility pass.
<!-- SECTION:FINAL_SUMMARY:END -->
