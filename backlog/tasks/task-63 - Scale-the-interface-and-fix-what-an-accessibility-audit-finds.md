---
id: TASK-63
title: Scale the interface and fix what an accessibility audit finds
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 21:06'
updated_date: '2026-08-26 21:14'
labels: []
milestone: m-2
dependencies: []
priority: high
type: enhancement
ordinal: 63000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The interface is dense by design, which is right for a board over nine projects and wrong for a large screen or for eyes that need more. A scale setting makes that a choice rather than a constraint. While the type is being looked at, the rest of the accessibility surface deserves the same treatment: contrast, focus order, roles, labels and keyboard reachability across every screen, checked with a tool rather than by eye.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A preference scales the whole interface, remembered across restarts
- [x] #2 Scaling changes type, spacing and controls together rather than only the font size
- [x] #3 An automated accessibility audit runs against every screen and its findings are recorded
- [x] #4 Contrast meets WCAG AA for text and interface elements, or the exception is written down with the reason
- [x] #5 Every interactive element is reachable and operable by keyboard with a visible focus indicator
- [x] #6 Controls carry accessible names, and screens carry landmarks and headings that describe them
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
Scaling is one root font size that everything derives from. Every size in the stylesheet was already in rem - type, spacing, the fixed chrome heights - so setting html font-size moves the whole interface together rather than leaving large text in small boxes. The value is clamped on load between 75 and 200 per cent, because a hand-edited 5 or 5000 would leave an interface nobody could use to fix the setting. Covered by a test over six values.

The audit was run with axe-core against the running application over nine screens and states - board, task panel, list, overview, documents, search, filters, the create form and preferences - at WCAG 2.0 A/AA, 2.1 A/AA and 2.2 AA.

It found 429 failing nodes across three rules, and every one of them was real:
- 390 contrast failures. The tertiary text colour was #5b6376 on the sunken surface, which is 3.11 to 1 against a 4.5 requirement. Readable to me at this desk and not to everyone. The two dimmer tiers are now #a7aec0 and #828b9d, both clearing 4.5 to 1 on all three surfaces.
- 35 target-size failures. Acceptance-criteria checkboxes and filter chips were under 24 by 24 with no room around them, which a dense interface will not give them on its own.
- 4 selects with no accessible name: the project picker on the create form and the three preference controls, all of which had a visible label that was not associated with the control.

The second run reports zero violations on all nine.

What the tool cannot answer is still owed: a screen reader has not been near this, and the audit says nothing about whether the reading order makes sense to someone who cannot see the layout.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added an interface scale preference and fixed everything an automated audit found.

Scale runs from 75 to 200 per cent, is clamped on load, and works by setting the root font size, which every other size derives from - so type, spacing, controls and chrome move together. Verified in the browser: choosing 150 per cent took the root size from 16px to 24px and the whole interface with it.

The audit ran axe-core against the running application across nine screens and states at WCAG 2.2 AA. It found 429 failing nodes under three rules - 390 contrast, 35 target size, 4 unnamed selects - and all of them were genuine. The most significant was the tertiary text colour at 3.11 to 1 against a 4.5 requirement, used in 390 places: legible at my desk and not to everyone. Both dim tiers were relightened and re-measured, checkboxes and chips were given room, and the selects were given names. The audit now reports zero violations on every screen.

Recorded in the notes: a screen reader has not been near this, and no automated tool can say whether the reading order makes sense to someone who cannot see the layout.
<!-- SECTION:FINAL_SUMMARY:END -->
