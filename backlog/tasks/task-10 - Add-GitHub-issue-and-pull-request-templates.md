---
id: TASK-10
title: Add GitHub issue and pull request templates
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-27 21:05'
labels: []
milestone: m-4
dependencies: []
priority: medium
type: chore
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reports arrive without the information needed to act on them unless the form asks for it. Templates must collect what actually matters here: operating system, Backlog.md CLI version, number of registered projects, and whether the problem reproduces with a single project.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Bug report template collects OS, application version, Backlog.md CLI version and reproduction steps
- [x] #2 Feature request template asks for the problem before the proposed solution
- [x] #3 Pull request template references the related task ID and a contributor checklist
- [x] #4 Issue template chooser points questions at Discussions rather than Issues
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The bug form asks for the two versions together, because Muster writes only through the backlog CLI and which CLI is involved is half of nearly every answer. It says where to get them - the copy button in Preferences, which exists for this form and was built in TASK-8 - rather than leaving a person to find out.

It also asks the three things this application's failures actually turn on. The session type, X11 or Wayland, because the tray and the global hotkey behave differently under each. How many projects are registered, because most of what can go wrong only happens when there is more than one: ids collide across projects and statuses differ between them. And whether it still happens with a single project, which halves the search - noting that hiding the others in the Projects screen is enough, so nobody unregisters anything to answer it. There is an optional field for the task's frontmatter, which for a write going wrong usually settles the question outright, since the answer is always whatever is on disk.

The feature form asks the problem first and the tests prove the order rather than trusting it: the problem is the first question after any prose, the proposed solution comes after it and is optional, and the problem is required. A request that arrives as a solution has thrown away the thing that would let someone find a better one, and often the answer is something the application already does, which only the problem statement reveals. The form also states the no-new-format rule and where such a request belongs instead, so nobody writes one out for nothing.

The chooser turns blank issues off - a blank issue is how both forms get skipped - and points questions and half-formed ideas at Discussions, and problems with the backlog CLI itself at Backlog.md.

CONTRIBUTING said flatly that there are no GitHub Issues in this project, which these templates would have contradicted on the day they landed. Reconciled rather than left: the work is still tracked in backlog/, and issues are the way in from outside - a report is read and, if it is going to be done, becomes a task, which is then what the pull request references.

Issue forms cannot be tried out before the repository exists, and a malformed one does not fail loudly: GitHub declines to render it and the reporter gets a blank box. So the structure is tested here - every field has a type GitHub renders and a label, dropdowns have options, ids are unique, and nothing puts required under attributes where it silently does nothing. Nine tests in all. Three were made to fail first: the problem asked after the solution, required moved where it does nothing, and blank issues switched back on. Each was caught.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Two issue forms, a chooser and a pull request template.

The bug form asks for the two versions together and says where the interface copies them from; it also asks the three things this application's failures turn on - the session type, how many projects are registered, and whether it reproduces with one. The feature form asks the problem before the solution, and the tests prove that order rather than trusting it.

The chooser turns blank issues off and sends questions to Discussions and Backlog.md problems upstream. CONTRIBUTING said there were no issues in this project at all, which these would have contradicted; it now says what is true - the work is tracked in backlog/, and an issue is the way in from outside.

None of this can be tried out before the repository exists, and a malformed form does not fail loudly - GitHub just declines to render it. Nine Go tests check the structure and the things each form exists to ask for, three of them made to fail first.
<!-- SECTION:FINAL_SUMMARY:END -->
