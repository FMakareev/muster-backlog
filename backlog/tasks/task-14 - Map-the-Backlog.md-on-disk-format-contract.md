---
id: TASK-14
title: Map the Backlog.md on-disk format contract
status: Done
assignee:
  - '@claude'
created_date: '2026-08-26 15:00'
updated_date: '2026-08-27 17:16'
labels: []
milestone: m-1
dependencies: []
documentation:
  - backlog/docs/doc-3 - Backlog.md-Format-Contract.md
priority: high
type: spike
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reads go straight to markdown, so the parser depends on a format owned by another project - and Muster is committed to never extending that format. Before writing the parser, establish exactly what the contract is against the real corpus of roughly 640 task files. Cover frontmatter fields, body section markers, acceptance-criteria markup, id and ordinal semantics, the per-project config.yml schema, and the layout of tasks, drafts, milestones, docs, decisions and completed directories.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every frontmatter field in use across the real corpus is catalogued with its type and optionality
- [x] #2 Body section and acceptance-criteria markers are documented with real examples
- [x] #3 The per-project config.yml schema is documented, including statuses, priorities, types and task prefix
- [x] #4 Directory layout including drafts, milestones, docs, decisions and completed is documented
- [x] #5 Format variations and malformed files found in the corpus are listed with the intended handling
- [x] #6 A pinned reference corpus of sample files is committed for parser tests
- [x] #7 The Backlog.md CLI version the contract was derived from is recorded
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Corpus scan complete (9 projects, 1021 markdown files: 852 tasks + 30 archived tasks, 58 milestones, 64 decisions, 17 docs, 0 drafts, 0 completed). All files UTF-8, LF-only, NFC, no BOM, all start with a '---' frontmatter block that parses as a YAML mapping - zero parse failures. Contract also cross-checked against the backlog.md 1.48.0 compiled binary (section marker constants, frontmatter serialiser field order, sanitizeFilename, config key parser).

Pinned reference corpus written to testdata/backlog-format/ (44 files in five numbered groups plus a README). Samples are either byte-exact 1.48.0 CLI output from a throwaway probe project, or real corpus files with structure preserved and prose/URLs/handles replaced by neutral placeholders; verified free of business names, personal handles and home paths. Covers minimal/full/dependency/AC/DOD/comment/subtask tasks, draft, milestone, doc, decision, five config.yml variants, a .backlog layout with an id colliding between tasks/ and archive/tasks/, and nine variant or malformed cases.

Findings written to doc-3 'Backlog.md Format Contract' (specification, ~500 lines, sections 0-8). Contract derived from backlog.md CLI 1.48.0. Section 8 lists 15 divergences between what 1.48.0 supports and what the corpus contains, plus the seven top parser hazards.

Independently re-verified the headline claims before accepting the spike, rather than taking the research at face value.

Confirmed: CLI version 1.48.0; 882 task files (852 active + 30 archived) matching a separate directory count; exactly two distinct status lists across the nine projects - eight on ["To Do", "In Progress", "Done"] and one (this repository) on the four-status list; cross-project id collision is real and pervasive (191 ids appear in more than one project by my own count, against the 200 reported); no project uses the root backlog.config.yml discovery path, so that third config location is CLI-supported but unexercised here; the reference corpus contains no home paths, emails or absolute paths.

Corrected two errors found in the document:
- The task frontmatter table gave title serialisation as bare 331 / quoted 602 / double-quoted 58 / folded 30. Those are the figures for all 1021 entity files, not the 882 tasks the table is measured over. Recounted for tasks only: bare 312, single-quoted 553, double-quoted 0, folded 17. The 30 folded titles are 17 tasks plus 13 decisions - corrected in the variations section too, since a parser author sizing the hazard from the task table would have been reading a number from a different population.
- A sentence read 'Only five keys are universal' and then listed seven. Now states seven.

Added .gitattributes marking testdata/backlog-format as -text. Without it, git line-ending normalisation would silently rewrite the CRLF fixture and the case it exists to cover would stop being covered. Verified the fixture is still CRLF in both index and worktree.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Mapped the Backlog.md on-disk format empirically against 1021 entity files across nine real projects on CLI 1.48.0, and committed a fixture corpus to test the parser against.

Produced doc-3 'Backlog.md Format Contract' (504 lines, sections matching the seven acceptance criteria): a quantified frontmatter catalogue (18 keys observed, 7 universal, with serialisation and optionality per key), the seven literal section markers in their two naming conventions, the per-project config.yml schema including the two distinct status lists found, the directory layout, the format variations with intended handling, and a section on where the 1.48.0 binary diverges from what this corpus happens to contain. Committed testdata/backlog-format/ - 44 sanitised files covering minimal and fully-populated tasks, subtasks, drafts, milestones, docs, decisions, four config variants, the .backlog layout with a colliding archived id, and every hazard case.

Findings that change the parser design: ids are unique only within one directory of one project, so the key must be (project, directory-class, id); folded block-scalar titles rule out any line-oriented frontmatter reader; filename slugs are lossy and one-way, with only 20% round-tripping, so titles must come from frontmatter; config lives in three possible locations; and priority enums are stored lowercase against a capitalised config.

Verified by re-deriving the headline numbers independently - CLI version, file counts, status lists, id collisions, config locations and corpus sanitisation all confirmed - and corrected two factual errors found in the process, recorded in the notes.
<!-- SECTION:FINAL_SUMMARY:END -->
