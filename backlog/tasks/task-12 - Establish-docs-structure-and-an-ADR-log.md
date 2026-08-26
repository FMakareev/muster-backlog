---
id: TASK-12
title: Establish docs structure and an ADR log
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 16:06'
labels: []
milestone: m-0
dependencies: []
documentation:
  - backlog/docs/doc-1 - Muster-Specification-v0.1.md
priority: medium
type: docs
ordinal: 12000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Decisions taken now - the beta shell, the CLI-only write path, the review-cost model - will be questioned later by contributors and by me. An architecture decision record log keeps the reasoning attached to the repository instead of to a chat transcript. Seed it with the decisions already made.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/ layout is defined and documented, covering specification, roadmap and decisions
- [x] #2 ADR template and numbering convention are checked in
- [x] #3 ADR records the desktop shell choice
- [x] #4 ADR records the license and release automation choice
- [x] #5 ADR records that writes go only through the backlog CLI and reads go directly to markdown
- [x] #6 The v0.1 specification is published under docs/ and linked from the README
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Do not create a parallel top-level docs/ tree. Backlog.md already owns two native entities for this - documents and decisions - and the project's founding rule is to use what the format provides rather than invent alongside it. The layout is therefore backlog/docs/ for specifications and guides and backlog/decisions/ for architecture decision records.
2. Record the convention itself as a document: what belongs in each, the numbering scheme the CLI imposes, the status vocabulary, and when a decision is worth recording.
3. Write the three decisions the acceptance criteria name: the desktop shell (using the measurements from TASK-2), the licence and release automation, and the read-direct / write-through-CLI split.
4. Link the specification from the README.
5. Flag the CLI gap found while doing this: backlog decision create writes only frontmatter and an empty Context/Decision/Consequences skeleton, and neither the CLI nor the MCP server can fill the body. Filling it means editing the file, which the project's own instructions discourage.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Deliberately did not create a top-level docs/ tree. Backlog.md already owns documents and decisions as native entities, the project's founding rule is to use what the format provides rather than invent alongside it, and a second tree would be documentation that Muster's own documents viewer could not display. Layout is backlog/docs/ for specifications and guides, backlog/decisions/ for ADRs, root for README, CONTRIBUTING and LICENSE.

The ADR template is not a checked-in file: 'backlog decision create' generates the Context / Decision / Consequences skeleton, and numbering is decision-N assigned by the CLI in creation order. doc-4 documents both, along with the status vocabulary and the rule that an accepted decision is superseded by a new one rather than rewritten.

Tooling gap found and documented: 'backlog decision create' writes frontmatter and an empty skeleton, and nothing in 1.48.0 can fill the body - there is no decision update subcommand, and the MCP server exposes task, milestone and document tools but none for decisions (verified by listing its tools over stdio). Decision bodies are therefore written by editing the file after creation, touching only the body sections and never the CLI-owned frontmatter. This is the single place this repository does not go through the CLI; it is recorded in doc-4 and in decision-3 so it is a known exception rather than a silent one.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Established where documentation and decisions live and seeded the decision log.

Defined the layout as Backlog.md's own entities - backlog/docs/ and backlog/decisions/ - rather than a parallel docs/ tree, and documented the conventions in doc-4: what belongs where, the CLI-assigned decision-N numbering, the Context/Decision/Consequences template, the status vocabulary, and when a decision is worth recording at all. Wrote the three decisions the acceptance criteria name: decision-1 the desktop shell, built on the measurements from TASK-2; decision-2 the MIT licence and release-please automation, citing the dependency audit from TASK-3; decision-3 the read-direct / write-through-CLI split, citing the format contract from TASK-14. Added a Documentation section to the README indexing all four documents and the decision log.

Verified every relative link in README, CONTRIBUTING and doc-4 resolves to an existing file by walking them programmatically - zero broken. All three decisions read back through the CLI with status accepted.
<!-- SECTION:FINAL_SUMMARY:END -->
