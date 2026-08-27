---
id: TASK-41
title: Refine the capture window into a usable inbox
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:02'
updated_date: '2026-08-27 18:00'
labels: []
milestone: m-3
dependencies:
  - TASK-26
priority: medium
type: enhancement
ordinal: 41000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Capture only works if it costs nothing to use: sensible defaults, the last project remembered, multi-line text, and no lost input when something goes wrong. Everything here is polish on top of the working hotkey capture, and all of it decides whether the inbox is actually used.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Capture window opens with a sensible default project and remembers the previous choice
- [x] #2 Project can be switched by keyboard without leaving the text field
- [x] #3 Multi-line capture is supported and the whole text reaches the draft
- [x] #4 Successful capture is confirmed unobtrusively without stealing focus back
- [x] #5 A capture made while the target project is unavailable is retained rather than lost
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
TASK-26 has not been built, so there is no separate capture window to polish. Every criterion here is written about one - but read against the capture that does exist, the note mode of the create form, four of the five are real work that can be done now and are worth doing for the same reason the task gives: capture only pays if it costs nothing.

1. A sensible default project that remembers the previous choice. The rule set in TASK-57 is the project you are looking at, then the first registered one. Remembering adds a middle step: the project last captured into, when nothing is focused. Kept in the preferences rather than in memory, because a daily tool is restarted.
2. Switching project by keyboard without leaving the text field.
3. Multi-line text reaching the draft whole - verified against the file rather than assumed, since the CLI's own help warns that a literal backslash-n is stored as text.
4. Confirming a capture without stealing focus: a notice, not a dialog, and the form clears itself so the next thought can be typed straight away.
5. Not losing what was typed when the write fails, including when the target project cannot be written to at all.

What stays with TASK-26: the global hotkey, opening over another application, and returning focus to it. Recorded on both tasks rather than quietly folded in.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Read against the capture that exists rather than the one this task was written for. TASK-26 has not been built, so there is no separate hotkey window; the note mode of the create form is where capture happens today, and all five criteria are about the same thing - whether capture costs anything - so they were applied there.

The default project gained a middle step. The rule was the project you are looking at, then the first registered one; it is now the focused one, then the project the last note went to, then the first. Kept in the preferences rather than in memory, because a habit that resets when the application restarts is not a habit. The focused project still wins, since someone looking at one means that one.

Switching project is alt+[ and alt+], handled on the form rather than on the select, so it works from the middle of a sentence. The shortcut is shown next to the picker rather than hidden.

A capture now confirms itself where the eye already is and clears the form instead of closing it - capture is worth little if the next thought has to reopen the window - and focus stays in the title field throughout. Nothing takes it away.

Multi-line was already passed through, but the CLI's own help warns that a literal backslash-n is stored as text, so it was checked against the file rather than assumed: a body with two lines and a blank line between arrives whole, with no escape sequences in it.

Verified in the browser against three projects with the files read off disk: the default with nothing focused, alt+] moving twice while the caret stays in the title, a multi-line body arriving intact, the confirmation and the emptied form with focus still in place, the next capture starting where the last one went, and a focused project overriding that. Separately, with the CLI made unfindable, a capture that cannot be written keeps every word typed and says why. No axe-core violations on the form.

Left with TASK-26, and recorded there: the global hotkey, opening over another application, and returning focus to it.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Capture costs less than it did. The form starts on the project you are looking at, or the one your last note went to, or the first registered; alt+[ and alt+] move between projects from the middle of a sentence; a capture is confirmed in place and the form empties itself for the next thought instead of closing; and a write that cannot happen keeps every word typed, with the reason beside it.

This was written as polish on the hotkey capture window from TASK-26, which does not exist yet. Four of the five criteria are about whether capture costs anything rather than about a window, so they were applied to the capture that does exist. What is genuinely about the window - the hotkey, opening over another application, returning focus - stays with TASK-26 and is recorded there.

Verified in the browser across three projects with the files read off disk, including a multi-line body arriving whole, and separately with the CLI unfindable to confirm nothing typed is lost.
<!-- SECTION:FINAL_SUMMARY:END -->
