---
id: TASK-16
title: Parse Backlog.md task files into a domain model
status: In Review
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-26 16:27'
labels: []
milestone: m-1
dependencies:
  - TASK-14
priority: high
type: feature
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A frontmatter and body parser over the tasks, drafts, milestones, documents and decisions of one project. It must preserve the full task body - description, acceptance criteria, implementation plan, notes - because fast access to that body is what a table view cannot give. Only fields Backlog.md itself defines are modelled; Muster adds none of its own.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Parser reads tasks, drafts, milestones, documents, decisions and completed directories of a project
- [x] #2 Domain model exposes only fields Backlog.md defines: status, priority, milestone, dependencies, ordinal, type, labels, assignee and dates
- [x] #3 Per-project config.yml is parsed for statuses, priorities, types, labels and task prefix
- [x] #4 Full body is retained with description, acceptance criteria, implementation plan and notes distinguishable
- [x] #5 Acceptance criteria are parsed as individual items with their checked state
- [x] #6 Title comes from frontmatter, never from the file name
- [x] #7 A malformed file is skipped with a diagnostic and does not abort the scan
- [x] #8 Tests run against the pinned reference corpus
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/backlog holds the domain model and the parser. The model carries only fields Backlog.md defines; the parser is driven entirely by the measured contract in doc-3, not by guesswork.
2. Frontmatter: normalise CRLF, then take the first ---...--- pair anchored at byte 0 and stop. Seven task files contain bare --- lines inside comment envelopes, so scanning the whole file for --- is wrong. Parse the block with a real YAML parser, because 30 files use folded block scalars for the title and a line reader silently drops them.
3. Sections: seven literal marker pairs in two naming conventions, all optional, matched on the marker rather than on offset from the heading. Apply the marker grammar only to task-like directories - a document can quote markers in prose, and doc-1 in this very repository does.
4. Acceptance criteria and definition-of-done share one grammar with a per-item index, falling back to positional numbering for the legacy index-less form the CLI still migrates.
5. Config: try YAML, fall back to a tolerant line reader, because 1.48.0 parses its own config with a hand-rolled line reader and therefore accepts files strict YAML would reject.
6. Scanning yields entities plus diagnostics. A file without frontmatter or without an id is not an entity - skip it with a diagnostic and carry on, never abort the project.
7. Test against the pinned corpus for every variation, and against all nine real projects as a smoke test to prove the parser survives a thousand real files.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Parser lives in internal/backlog: model.go for the domain types, parse.go for frontmatter and section markers, config.go for the project config, scan.go for walking a project.

Decisions taken straight from the measured contract rather than from intuition:
- Frontmatter is the first ---...--- pair anchored at byte 0 and nothing else. Seven task files in the author's projects hold bare --- lines inside comment envelopes, so a reader that scans the file for a fence takes half a comment for frontmatter.
- The block is parsed with a real YAML parser. Thirty files across four projects write the title as a folded block scalar, and a line reader returns an empty title for every one of them.
- The seven section markers are hardcoded literals in two naming conventions, and matching is on the marker alone - the whitespace between heading and marker is not uniform, so anything measured from the heading is wrong.
- The marker grammar is applied only to task-like directories. A document can quote the markers in prose, and doc-1 in this repository does exactly that.
- Present-but-empty is distinct from absent throughout: an AC block with nothing in it yields a non-nil empty slice, not a missing section.
- Config is parsed as YAML with a fallback to a tolerant line reader, because Backlog.md 1.48.0 reads its own config with a hand-rolled line reader and therefore accepts files strict YAML rejects.
- Ordering follows the CLI's comparator: ordinal-bearing first, then ordinal, then id compared numerically so TASK-9 precedes TASK-10 under any zero-padding scheme.

Cross-validation worth recording. The section counts this parser produces were compared against the independent grep-based measurements in doc-3, and every difference is exactly accounted for by edits made in this session: descriptions 883 against 882 (+1 task added), AC sections 877 against 876 (+1), implementation plans 390 against 378 (+12 - exactly the twelve tasks planned this session), notes 637 against 626 (+11), final summaries 573 against 562 (+11), and definition-of-done sections 8, comments 7 and DoD items 31 all matching exactly. Two independent measurements of the same corpus agreeing to the unit is much stronger evidence than either alone.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added internal/backlog: the domain model and the parser for Backlog.md projects, driven entirely by the format contract measured in TASK-14.

Reads tasks, drafts, milestones, documents, decisions, completed and the three archive directories, walking nested document subdirectories. The model carries only fields Backlog.md defines - including reporter, subtasks and onStatusChange, which are in the serialiser but appear in no file yet - and adds nothing of its own. Task bodies keep description, plan, notes and final summary as distinguishable sections, acceptance criteria and definition-of-done items parse with their checked state and index, and the comment envelope is read in both its current and its newer per-comment form. Titles come from frontmatter, never from filenames. A file that is not an entity becomes a diagnostic and the scan continues.

Verified at two levels. 40 unit tests against the pinned reference corpus cover every variation the contract names: folded titles, CRLF, a body-less task, an empty AC block, index-less legacy criteria, an unnormalised assignee, a document quoting task markers, the same id in tasks/ and archive/, and five kinds of non-entity file. Coverage is 84.1%.

Then against all nine real projects: 1027 entities parsed in 101 ms with zero diagnostics, every entity carrying an id and a single-line title, and every active task's status valid against its own project's declared list. Section counts were then compared against the independent measurements in doc-3 and agree exactly once this session's own twelve plans, eleven note-appends and eleven final summaries are accounted for.

'wails3 task lint' and 'go test -tags gtk3 ./...' both green.
<!-- SECTION:FINAL_SUMMARY:END -->
