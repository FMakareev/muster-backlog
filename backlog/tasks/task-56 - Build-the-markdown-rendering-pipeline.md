---
id: TASK-56
title: Build the markdown rendering pipeline
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:29'
updated_date: '2026-08-26 17:08'
labels: []
milestone: m-1
dependencies:
  - TASK-20
priority: high
type: feature
ordinal: 56000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Task bodies, documents and decisions all render through one pipeline, so it is defined once rather than three times. It must handle what Backlog.md files actually contain: section markers, acceptance-criteria checklists, task cross-references, code blocks and Mermaid diagrams. Rendering is local and offline - no content leaves the machine.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Markdown renders with code highlighting and Mermaid diagram support
- [x] #2 Task cross-references such as TASK-42 render as links that navigate within the application
- [x] #3 Acceptance-criteria checklists render as checkable items reflecting their state on disk
- [x] #4 Backlog.md section markers are handled without leaking into the output
- [x] #5 Rendering is sandboxed against untrusted markdown and makes no network requests
- [x] #6 Relative links to files inside a project resolve
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Linters and formatters pass across Go and frontend
- [ ] #2 Automated tests cover the change and the suite is green
- [ ] #3 User-facing behaviour change is reflected in README or docs
- [ ] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. One pipeline for task bodies, documents and decisions: marked to parse, DOMPurify to sanitise, highlight.js for code, mermaid loaded only when a diagram is actually present so the common case does not carry it.
2. Sanitising is not optional. Task bodies come from files the application did not write, edited by agents; a script tag or an onerror attribute in a description must be inert. Everything goes through DOMPurify with a hook that also drops remote resource URLs, so rendering a task can never make a network request - a local-first tool must not let a task body phone home.
3. Task cross-references are rewritten into links the application handles itself, resolving inside the owning project only, because ids collide across projects.
4. Section markers are stripped before parsing rather than escaped after; they are HTML comments and would otherwise either vanish silently or leak depending on the sanitiser's mood.
5. Acceptance criteria render as real checkboxes reflecting the state on disk, disabled until writes go through the CLI.
6. Add vitest, which the frontend has been missing, and cover the pipeline including the hostile cases.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
One pipeline in src/lib/markdown.ts, used by the task panel, the documents viewer and anything else: marked to parse, DOMPurify to sanitise, highlight.js for code, mermaid imported on demand.

mermaid is 95 KB gzipped and most task bodies contain no diagram, so it is a dynamic import that only loads when a diagram is actually present. Vite splits it into its own chunk, confirmed in the build output.

Section markers are stripped before parsing rather than left to the sanitiser. They are HTML comments, and whether they vanish or leak would otherwise depend on the sanitiser's configuration rather than on a decision.

Task references are linked in the rendered HTML rather than in the markdown, so an id inside a code span or a fenced block stays code. Verified in the browser: TASK-14 and STORY-7 in prose became links, TASK-42 inside a Go snippet did not. The pattern matches any uppercase prefix rather than assuming TASK, because task_prefix is per-project configuration.

Two real defects found by looking at the rendered output rather than at the tests:
- Forbidding the input tag to block forms also killed the acceptance-criteria checkboxes, which are exactly what criterion three asks for. Inputs are now allowed but a hook removes any that is not a checkbox and forces every checkbox disabled - the state shown is the state on disk, and changing it has to go through the CLI.
- Every mermaid node rendered as an empty box. mermaid puts labels in a foreignObject as HTML, and the SVG-only sanitiser profile strips it. Fixed at the source by configuring mermaid with htmlLabels false so labels are SVG text, which keeps the sanitiser strict rather than widening it to let HTML through.

The no-network criterion is enforced structurally, not hoped for: a hook strips any src that is not local or a data URI, keeping the original in an inert data-blocked-src so a reader can see something was blocked. Verified by recording every request the page made while rendering a body containing a remote image - zero offsite requests.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the markdown pipeline every rendered body will use, and the vitest setup the frontend was missing.

renderMarkdown parses with marked, strips Backlog.md section markers, links task references outside code, and sanitises with DOMPurify under a hook that blocks remote resources and allows only disabled checkboxes. highlightCode and renderMermaid run afterwards on the mounted DOM; mermaid is imported on demand so the common case does not carry 95 KB it does not need. A Markdown.svelte component wraps all of it and reports task-reference clicks.

Verified twice over. 18 vitest tests cover the pipeline, weighted towards the hostile cases: script tags, event handlers, javascript: links, iframes, forms, and a remote image beacon. Then in a real browser against a body containing all of it at once: one mermaid diagram rendered with all five node labels, five highlighted code tokens, TASK-14 and STORY-7 linked while TASK-42 inside a code block was left alone, three checkboxes with the right checked states and all disabled, one image blocked, the external link targeting the system browser, no section markers leaking - and zero offsite network requests recorded while rendering.

Both defects that mattered were found by looking at the output rather than by reading the code: forbidding inputs had silently killed the acceptance-criteria checkboxes, and every mermaid node was rendering as an empty box because the sanitiser strips the foreignObject its labels live in. Also wired 'wails3 task test' and a frontend test job in the pre-push hook, so the frontend suite now runs where the Go one does.
<!-- SECTION:FINAL_SUMMARY:END -->
