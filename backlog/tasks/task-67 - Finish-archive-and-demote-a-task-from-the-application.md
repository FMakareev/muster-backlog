---
id: TASK-67
title: 'Finish, archive and demote a task from the application'
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-27 00:38'
updated_date: '2026-08-27 00:45'
labels: []
milestone: m-5
dependencies: []
priority: high
type: feature
ordinal: 67000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Three commands Backlog.md has and Muster does not wrap: task complete, task archive and task demote. The gap shows in the files - across the nine reference projects, 591 of 875 live task files are already at their project's terminal status and none has ever been moved into completed/, which is empty in all nine. Finished work sits mixed in with the work.

Demote is the other half of the inbox: a note promoted a moment too early can go back. It is a dedicated command; task edit -s Draft is refused, which is what made this look impossible earlier.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A finished task can be moved into completed/ from the panel
- [x] #2 A task can be archived, with a confirmation, and leaves the board without being deleted
- [x] #3 A task can be sent back to the inbox as a draft
- [x] #4 Each action is offered only where it makes sense, and says what it will do before it does it
- [x] #5 Every one of them verifies its outcome rather than trusting the exit code
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
1. Wrap the three commands: task complete, task archive, task demote. Each verifies its outcome rather than trusting the exit code, the way the draft writes do - the class of silent success this CLI has shown before is not one to assume away.
2. Offer each only where it means something. Completing is for a task at its project's terminal status; the CLI refuses anything else, and a control that is going to be refused should not be there. Archiving and demoting apply to any live task.
3. Say what each will do before it does it, because all three move the file somewhere the board does not show: completed/, archive/tasks/ and drafts/ respectively. Archiving and demoting ask first; completing is the ordinary end of a finished task and does not.
4. Demoting sends a task back to the inbox, so the inbox has to follow immediately - it is the same store reload every other write uses.
5. Tests against the real CLI for each, including refusing to complete a task that is not finished, and a browser pass reading the files on disk afterwards.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Two things came out of this that were not in the plan.

The first is a bug that had been there all along and this change made obvious: a query naming no class returned every class. Archiving is a soft delete and completed/ is where finished work is filed, but the board, the list, the counts and the filters had been carrying archived tasks as though they were live - 30 of them across the author's projects. It only became visible when a completed task refused to leave the board. An unqualified query now means the live ones, and anything that wants the rest says so.

The second is what demote actually does. It renames the id to DRAFT-n and moves the file into drafts/, and leaves status and ordinal exactly as they were: a task demoted from In Progress arrives in the inbox still saying status: In Progress and carrying ordinal 1000. So the format contract's claim that a draft has status Draft and no ordinal holds only for drafts the CLI created, and anything reading drafts has to key on the directory rather than on the status. doc-3 now says so, and the test asserts the measured behaviour rather than the expected one.

Completing is offered only for a task at its project's own last declared status, which is the rule the CLI applies - verified against a project whose statuses end in Shipped, not Done, where a Shipped task completes and a To Do one is refused with the CLI's own words. Archiving and demoting ask first, because both move a file somewhere the board does not look; filing a finished task does not ask, because that is the ordinary end of a task. All three verify their outcome rather than trusting the exit code.

Verified in the browser against a scratch bench, reading the directories after each action: an archived task absent from the board and from the count, an unfinished task told what it is waiting for, a finished one filed into completed/ and gone from the board, an archive that asks and then lands in archive/tasks, and a demote that lands in drafts/ and appears in the inbox immediately. Eight Go tests, and no axe-core violations on the panel.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A task can now end somewhere from inside the application: filed into completed/ when it reaches its project's own last status, archived as a soft delete, or sent back to the inbox as a note.

The measured reason for the first: 591 of 875 live task files across the nine projects are already finished, and completed/ was empty in all of them, because nothing in the application could move one.

Two findings on the way. A query naming no class had been returning archived tasks as though they were live, everywhere - the board, the list, the counts, the filters - and only showed itself when a completed task refused to disappear; an unqualified query now means the live ones. And task demote leaves status and ordinal untouched, so a demoted note is a draft by where it lives rather than by what it says, which the format contract now records.

Verified in the browser against a scratch bench with the directories read after every action, plus eight Go tests against the real CLI and a clean accessibility pass.
<!-- SECTION:FINAL_SUMMARY:END -->
