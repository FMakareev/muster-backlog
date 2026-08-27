---
id: TASK-69
title: 'Write documents and decisions, not only read them'
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-27 00:38'
updated_date: '2026-08-27 01:32'
labels: []
milestone: m-5
dependencies: []
priority: medium
type: feature
ordinal: 69000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The documents viewer renders what is there and offers no way to add to it. backlog doc create, doc update and decision create all exist.

This project's own conventions say to write a decision when a choice closes off alternatives - and writing one means leaving the application that is showing you the decisions. Note that doc update replaces the whole body, so editing has to send the complete document rather than a patch.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A document can be created in any project, with its type chosen from what the CLI accepts
- [x] #2 A document's body can be edited and saved
- [x] #3 A decision can be created with its status
- [x] #4 Editing sends the whole body, because that is what doc update takes, and nothing is silently truncated
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
1. Measured first. doc create takes a title, a type and a path and nothing else - the document it makes is empty, and doc update --content fills it, so creating with content is two calls behind one action. doc update also takes a title, a type, a path and tags, and --content replaces the whole body, which is why editing sends the entire document rather than a patch.
2. decision create takes a title and a status and writes a skeleton with Context, Decision and Consequences headings. There is no decision update or decision edit at all, so a decision's body cannot be written through the CLI. Creating one from the application is worth having anyway - it is the act that needs to be cheap - but the interface has to say plainly that the body is filled in the file, and show the path, rather than pretending otherwise.
3. Backend: CreateDocument, UpdateDocument and CreateDecision, each verifying the result rather than the exit code, since this CLI has reported success for writes that did nothing.
4. Docs screen: new document and new decision, and an edit affordance on the document being read. The viewer already has the raw body on the entity, so editing shows exactly what is in the file.
5. Tests against the real CLI, and a browser pass that reads the written files off disk.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The two entities are less alike than they look, and the difference decided the shape of this.

A document is created empty - doc create takes a title, a type and a path, and no content at all - and filled by doc update --content, which replaces the whole body. So creating one with its text is two calls behind a single action, and editing one means holding the entire file in the editor and sending it as it stands. That is not a design choice; it is the only form the command takes, and the editor says so.

A decision is created with a title and a free-form status, and Backlog.md writes a skeleton with Context, Decision and Consequences headings. There is no decision update and no decision edit - both answer 'unknown command'. So the body of a decision cannot be written through the CLI at all. Creating one from the application is still worth having, because writing the decision down at the moment the choice is made is the act that has to be cheap, but the form and the viewer both say plainly that the decision itself is written in the file, and the viewer shows the path.

One thing the tests found rather than the browser: the CLI writes into docs/ and decisions/ and does not create them. The hand-made fixtures elsewhere in this suite carry only tasks/, so these tests initialise a project through the CLI itself, which is also closer to what a real project is.

Verified in the browser against a scratch bench, reading the files after each write: a document created with type specification and its markdown, opening straight away in the viewer; an existing document opened for editing showing exactly what was in the file, then replaced wholesale with a new title; and a decision created with status accepted and the three headings, with the viewer saying where it is written. Three Go tests against the real CLI, and no axe-core violations across the screen and its three states.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The documents viewer writes as well as reads: a document can be created with its type and body and edited afterwards, and a decision can be created with its status.

The two are shaped by what the CLI allows rather than by symmetry. doc create takes no content, so creating a document with text is two calls behind one action; doc update replaces a document wholesale, so editing holds the whole file. A decision is different in kind: it is created with a skeleton and there is no decision update at all, so the interface makes the act of recording one cheap and then says, in the form and in the viewer, that the decision itself is written in the file.

Verified in the browser with the files read after every write, three Go tests against the real CLI, and a clean accessibility pass over the screen, both forms and the editor.
<!-- SECTION:FINAL_SUMMARY:END -->
