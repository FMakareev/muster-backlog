---
id: doc-3
title: Backlog.md Format Contract
type: specification
created_date: '2026-08-26 15:52'
updated_date: '2026-08-26 15:56'
---
Muster reads Backlog.md markdown directly and is committed to never extending the format. This document states what that format actually is, established empirically against every Backlog.md project on this machine and cross-checked against the CLI binary itself. The parser is built to this contract.

## 0. Method and scope

Nine projects were surveyed in full — every file read, every value aggregated. Nothing here is generalised from a sample.

| Project | Directory | Tasks | Archived | Milestones | Docs | Decisions |
|---|---|---:|---:|---:|---:|---:|
| storefront (private commercial web app) | `.backlog/` | 160 | 17 | 4 | 4 | 8 |
| ref_canvas | `backlog/` | 165 | 0 | 14 | 7 | 6 |
| wall_diggers | `backlog/` | 162 | 2 | 11 | 2 | 0 |
| CarcassonneLike | `backlog/` | 117 | 0 | 13 | 1 | 2 |
| online_shopping_roguelike | `backlog/` | 79 | 0 | 2 | 0 | 29 |
| jade-palace | `backlog/` | 76 | 0 | 2 | 0 | 18 |
| muster (this repo) | `backlog/` | 45 | 11 | 5 | 2 | 0 |
| TREELINE | `backlog/` | 41 | 0 | 7 | 1 | 1 |
| portfolio_website | `backlog/` | 7 | 0 | 0 | 0 | 0 |
| **total** | | **852** | **30** | **58** | **17** | **64** |

**1021 entity files, 3.4 MB.** Zero drafts and zero files in `completed/` anywhere.

Two independent sources were used:

1. **The corpus** — what these files actually contain.
2. **The 1.48.0 binary** — `backlog` ships as a compiled Bun executable whose bundled JavaScript is recoverable with `strings`. The section-marker constants, the frontmatter serialiser, `sanitizeFilename`, the acceptance-criteria parser and the `config.yml` key switch were all read directly out of it.

Where those two disagree, the gap is called out. That gap is where a parser written only against today's corpus breaks on tomorrow's file. Section 8 lists every instance.

**Corpus-wide invariants.** All 1021 files are UTF-8, NFC-normalised, LF-only (zero CRLF), have no BOM, end with a trailing newline, begin with `---` on line 1, and their frontmatter parses as a YAML **mapping**. There were **zero YAML parse failures and zero files without frontmatter**. The format is far cleaner in practice than a defensive parser would assume — which is precisely why the hazards below matter: they are rare, not absent.

---

## 1. Frontmatter field catalogue

### 1.1 Tasks — 16 distinct keys observed

Measured over 882 files (852 active + 30 archived).

| Key | YAML type | Present | Empty | Serialisation | Distinct values |
|---|---|---:|---:|---|---|
| `id` | string | 882 (100%) | 0 | bare | 351 |
| `title` | string | 882 (100%) | 0 | bare 312 / `'…'` 553 / `"…"` 0 / **`>-` folded 17** | 881 |
| `status` | string | 882 (100%) | 0 | bare | 3 |
| `assignee` | list of strings | 882 (100%) | **409** | `[]` when empty, block `- '@x'` when populated | 5 |
| `created_date` | string | 882 (100%) | 0 | always `'…'` single-quoted | 373 |
| `labels` | list of strings | 882 (100%) | **506** | `[]` / block | 80 label values |
| `dependencies` | list of strings | 882 (100%) | **555** | `[]` / block | 178 combinations |
| `ordinal` | integer | 876 (99.3%) | 0 | bare | 176 |
| `updated_date` | string | 731 (82.9%) | 0 | always `'…'` | 553 |
| `milestone` | string | 620 (70.3%) | 0 | bare | 14 |
| `priority` | string | 464 (52.6%) | 0 | bare | 3 |
| `type` | string | 370 (42.0%) | 0 | bare | 7 |
| `references` | list of strings | 135 (15.3%) | 0 | block only | 93 combinations |
| `documentation` | list of strings | 107 (12.1%) | 0 | block only | 19 combinations |
| `parent_task_id` | string | 91 (10.3%) | 0 | bare | 29 |
| `modified_files` | list of strings | 27 (3.1%) | 0 | block only | 22 combinations |

**Seven keys are universal**: `id`, `title`, `status`, `assignee`, `created_date`, `labels` and `dependencies`, all at 100%. Everything else is optional. The minimum observed is 8 keys, the maximum 14; there are **60 distinct key orderings**, though the order is deterministic given which keys are present.

**Canonical key order**, read out of the 1.48.0 serialiser:

```
id, title, status, assignee, [reporter], created_date, [updated_date], labels,
[milestone], dependencies, [references], [documentation], [modified_files],
[parent_task_id], [subtasks], [priority], [type], [ordinal], [onStatusChange]
```

Bracketed keys are omitted when empty. A parser must not depend on order — but it is stable enough that any deviation signals a hand-edited file.

### 1.2 Field semantics

**`id`** — `TASK-<n>` or `TASK-<parent>.<child>` for subtasks. Always uppercase in frontmatter, always lowercase in the filename. Zero-padding is a per-project config choice: 418 files use `TASK-<nn>`, 316 use `TASK-<nnn>` (padded), 57 use `TASK-<n>`, 91 are dotted subtask ids. Under `zero_padded_ids: 3` the *parent* segment is padded and the *child* segment is not — `TASK-073.02`.

Ids are **unique only within one directory of one project**. 200 of 351 distinct task ids collide across projects (`TASK-1` exists in six of the nine). One collides *within* a project: `TASK-073` in storefront exists both in `tasks/` and in `archive/tasks/` as two different tasks, because archiving is a soft delete and the CLI reuses archived ids. **Muster's task key must be (project, directory-class, id).**

**`title`** — the authoritative title. See §5.1: the filename is not.

**`status`** — free-form string validated against the project's `statuses` list at write time only. All 882 files hold a value from their own project's list, so no orphan statuses were found. Case is preserved as configured.

**`assignee`** — a list, never a scalar, present in every file. 409 of 882 are empty. No task in the corpus has more than one assignee, but the type is a list and must be read as one. Values: `@claude` (352), plus three spellings of the same human maintainer handle (68 / 37 / 15 occurrences, two of them differing only in case) and **one bare email address with no `@` prefix**. The `@` is a convention the CLI adds, not a guarantee, and the same person appears under four spellings and two cases.

**Dates** — `created_date`, `updated_date` (and `date` on decisions, `created_date`/`updated_date` on docs) are **always** `YYYY-MM-DD HH:MM`, always single-quoted, 1899 values of 1899 matching, no exceptions. They are **UTC** with minute precision: the binary writes `new Date().toISOString().slice(0,16).replace("T"," ")`. There is no timezone marker, no seconds, no date-only form. The `date_format: yyyy-mm-dd` config key governs *display*, not storage, and setting it does not change what is written. `updated_date` is absent on 151 tasks (never edited since creation); no file has `updated_date` earlier than `created_date`.

**`labels` and `dependencies`** — the empty/populated serialisation split is the important detail. Empty is written **inline** as `labels: []`; populated is written as a **block sequence**:

```yaml
labels: []
dependencies: []
```
```yaml
labels:
  - performance
  - frontend
dependencies:
  - TASK-23
  - TASK-128.01
```

The key is never omitted for these two, nor for `assignee`. `references`, `documentation` and `modified_files` behave in the opposite way: the key is omitted entirely when empty and only ever appears as a block sequence. 80 distinct labels, at most 3 on one task; at most 7 dependencies. All 442 dependency values point at a task that exists in the same project — zero dangling references corpus-wide.

**`milestone`** — an **ID, never a title**. All 620 values match `m-<n>`, and every one resolves to a milestone file in the same project. The CLI's `-m/--milestone` flag accepts a title and resolves it to the id before writing, so a title never reaches disk.

**`ordinal`** — an integer for manual ordering. Present on 876 of 882; **missing on six** (all in wall_diggers). Range 1000–174000. 874 of 876 are exact multiples of 1000 — the CLI allocates `(count+1) * 1000` — with two hand-set exceptions (`11500`, `11700`) from drag-and-drop reordering. Ordinals are **not unique**: storefront has three colliding pairs and wall_diggers one, so ordinal alone is not a sort key. The 1.48.0 comparator sorts tasks *with* an ordinal before tasks without, then by ordinal, then by id.

**`priority`** — always **lowercase** on disk: `high` (250), `medium` (167), `low` (47). The config and CLI declare `High`, `Medium`, `Low`; the CLI lowercases before writing (verified: `--priority High` produces `priority: high`). **Compare case-insensitively.**

**`type`** — `feature` 156, `task` 92, `bug` 61, `enhancement` 26, `chore` 22, `spike` 9, `docs` 4. Exactly the seven defaults; no project overrides `types`.

**`parent_task_id`** — present on all 91 dotted ids and nothing else. In all 91 the child id is the parent id plus `.` plus a segment, so parentage is derivable from the id alone — but the field is authoritative and no subtask lacks it.

**`references` / `documentation` / `modified_files`** — free-form string lists mixing kinds. `references`: 47 URLs, 147 repo-relative paths, 4 bare filenames. `documentation`: 107 repo-relative paths *and* 7 bare document ids (`doc-004`) — **both forms in the same field**, so a resolver must try id lookup and path lookup. `modified_files`: repo-relative paths from the project root.

### 1.3 Milestones, documents, decisions, drafts

**Milestones** — exactly two keys, `id` and `title`, in all 58 files, one single ordering. `title` is the only double-quoted scalar the milestone serialiser emits (`title: "${…}"`).

**Documents** — six keys, three orderings:

| Key | Present | Notes |
|---|---:|---|
| `id` | 17/17 | `doc-<n>`, padded in one project |
| `title` | 17/17 | |
| `type` | 17/17 | `specification` 12, `guide` 3, `other` 2 |
| `created_date` | 17/17 | |
| `updated_date` | 11/17 | |
| `tags` | 2/17 | list of strings, block sequence, omitted when empty |

**Decisions** — exactly four keys, `id`, `title`, `date`, `status`, in all 64 files, one ordering. Note `date`, not `created_date`. `status` is **not case-normalised**: `accepted` (30), `Accepted` (29), `proposed` (5) — a case-sensitive comparison splits the same state in two.

**Drafts** — none in the corpus. Verified against 1.48.0 directly: a draft is a task-shaped file with a **separate `DRAFT-<n>` id sequence**, `status: Draft`, and **no `ordinal`**:

```yaml
---
id: DRAFT-1
title: A draft idea
status: Draft
assignee: []
created_date: '2026-08-26 15:44'
labels: []
dependencies: []
---
```

Promotion renumbers it into the task sequence and moves the file to `tasks/`.

---

## 2. Body sections and acceptance-criteria markup

### 2.1 The marker set

Section content is delimited by HTML comments. There are **exactly seven** section kinds, and the marker strings are literal constants in the binary — case-sensitive, single space inside the comment:

| Heading | Begin marker | End marker | Task files |
|---|---|---|---:|
| `## Description` | `<!-- SECTION:DESCRIPTION:BEGIN -->` | `<!-- SECTION:DESCRIPTION:END -->` | 882 |
| `## Acceptance Criteria` | `<!-- AC:BEGIN -->` | `<!-- AC:END -->` | 876 |
| `## Definition of Done` | `<!-- DOD:BEGIN -->` | `<!-- DOD:END -->` | 8 |
| `## Implementation Plan` | `<!-- SECTION:PLAN:BEGIN -->` | `<!-- SECTION:PLAN:END -->` | 378 |
| `## Implementation Notes` | `<!-- SECTION:NOTES:BEGIN -->` | `<!-- SECTION:NOTES:END -->` | 626 |
| `## Comments` | `<!-- COMMENTS:BEGIN -->` | `<!-- COMMENTS:END -->` | 7 |
| `## Final Summary` | `<!-- SECTION:FINAL_SUMMARY:BEGIN -->` | `<!-- SECTION:FINAL_SUMMARY:END -->` | 562 |

**Two naming conventions coexist.** Description, Plan, Notes and Final Summary use the long `SECTION:<ID>:` prefix; Acceptance Criteria, Definition of Done and Comments use a short bare prefix. There is no rule to infer — these are seven literals to hardcode.

The order above is the canonical write order and is never violated in the corpus.

### 2.2 Pairing, presence and whitespace

Measured over all 882 task files:

- **Every marker is paired.** Zero unbalanced `BEGIN` without `END` or vice versa.
- **Every heading has its marker and every marker has its heading.** Zero mismatches in either direction, for all six checked pairs.
- **No section is mandatory.** A task created by 1.48.0 and never edited has frontmatter and an entirely empty body — no headings, no markers at all. The corpus never contains one because every task there received a description, but the parser must accept it.
- Only 6 tasks lack `## Description`; only 6 lack `## Acceptance Criteria`.
- **Whitespace between heading and marker is not uniform.** Description, Plan, Notes and Final Summary always have a blank line between the heading and `BEGIN` (882/882, 378/378, 626/626, 562/562). Acceptance Criteria and Definition of Done never do (876/876, 8/8). Match on the marker, not on offset from the heading.
- Zero tasks have content after the last `END` marker, and zero have content before the first heading.

Real example, verbatim:

```markdown
## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Short description.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 First criterion
- [ ] #2 Second criterion
<!-- AC:END -->
```

### 2.3 Acceptance-criteria and Definition-of-Done items

One grammar serves both. **3336 acceptance-criteria lines and 31 Definition-of-Done lines were checked and all 3367 conform, with zero exceptions:**

```
^- \[( |x)\] #(\d+) (.+)$
```

- checked state is `[x]`, unchecked is `[ ]` — lowercase `x`, single space. 2072 checked, 1264 unchecked.
- `#N` is a **1-based per-item index**, present on every item in the corpus. It is the handle `backlog task edit --check-ac N` uses, and it is renumbered on insertion and deletion, so it is stable only within one revision of one file. Do not persist it as an identity.
- item text follows a single space and may contain `#`, `-`, `:`, `[`, `]`, Cyrillic and emoji.
- one file has an `AC:BEGIN`/`AC:END` pair with **nothing between them** — an empty section, not an error.

### 2.4 The Comments envelope

Seven tasks carry one. Inside `<!-- COMMENTS:BEGIN -->`, each comment is `author:` / `created:` / `---` / body / `---`:

```markdown
## Comments

<!-- COMMENTS:BEGIN -->
author: @claude
created: 2026-07-27 21:40
---
Comment body text.
---
<!-- COMMENTS:END -->
```

`created` here is **not** quoted, unlike every frontmatter date. The bare `---` lines are the envelope delimiters — the reason a "split the file on `---`" frontmatter reader is wrong. 7 task files contain standalone `---` lines in the body and all 7 are comment envelopes.

1.48.0 also defines per-comment `<!-- COMMENT:BEGIN -->` / `<!-- COMMENT:END -->` markers (singular) which do not appear anywhere in the corpus — a newer envelope form the parser should tolerate.

### 2.5 Non-task bodies

**Milestones, documents and decisions have no markers at all.** Everything after frontmatter is free markdown.

- Milestones: a `## Description` heading with plain prose (`Milestone: <title>` by default).
- Decisions: `## Context`, `## Decision`, `## Consequences`, optionally `## Alternatives` — plain headings, no markers, and the parser must tolerate their absence.
- Documents: entirely free markdown.

This matters more than it looks. A document may **quote task markers in prose or a fenced block** — `backlog/docs/doc-1` in this repository contains the literal strings `<!-- SECTION:DESCRIPTION:BEGIN -->` and `<!-- AC:BEGIN -->` in a sentence about the format. A marker scanner pointed at every `.md` file in a project mis-parses it. **Scan for markers only in task and draft directories.** (Inside *task* bodies specifically, zero markers appear in fenced blocks — but that is luck, not a guarantee.)

---

## 3. Per-project `config.yml`

### 3.1 Where it lives

The data directory is one of three shapes, and 1.48.0 resolves them in this order:

1. `backlog/` at the project root — **8 of 9 projects**.
2. `.backlog/` at the project root — **1 of 9** (storefront).
3. A **custom project-relative path**, declared by `backlog_directory` in a root config.

The config file is one of:

- `<backlog-dir>/config.yml` — **all 9 projects**.
- `<backlog-dir>/config.yaml` — accepted, unused here.
- `backlog.config.yml` at the **project root** (`backlog init --config-location root`) — unused here. This is the only file that can carry `backlog_directory`.

Discovery therefore cannot assume `backlog/config.yml`. Muster must probe `backlog/`, `.backlog/`, and the root `backlog.config.yml` (following its `backlog_directory`).

### 3.2 The status lists — verbatim

**This is the single most consequential finding for the board.** Statuses are per-project configuration, projects do not agree, and the board's columns are the union of these lists.

```yaml
# ref_canvas, wall_diggers, storefront, CarcassonneLike,
# online_shopping_roguelike, jade-palace, TREELINE, portfolio_website  (8 of 9)
statuses: ["To Do", "In Progress", "Done"]
```

```yaml
# muster (this repository)  (1 of 9)
statuses: ["To Do", "In Progress", "In Review", "Done"]
```

`default_status: "To Do"` in all nine.

Only **two distinct lists** exist today, and their union is `["To Do", "In Progress", "In Review", "Done"]` — one list is a strict prefix-and-suffix superset of the other, so ordering the union is trivial *right now*. That is a coincidence of this corpus, not a property of the format. `statuses` is an arbitrary ordered list of arbitrary strings: nothing forbids `["Backlog", "Doing", "Shipped"]` alongside `["To Do", "Done"]`, and the union of two lists that share no elements has no natural order. The union algorithm must handle disjoint and conflicting orderings, and the synthetic sixth-status config in the reference corpus exists to test exactly that.

`hide_empty_columns` is set by no project, so every column in the union is rendered.

### 3.3 Full key schema

Read from the 1.48.0 config parser. **`config.yml` is parsed by a hand-rolled line reader, not a YAML parser**, so it accepts both inline `["a", "b"]` and block `- a` list forms and strips quotes with a regex.

| Key | Type | Default | Set by | Notes |
|---|---|---|---|---|
| `project_name` | string | — | 9/9 | display name, not the directory name |
| `default_status` | string | `To Do` | 9/9 | all `"To Do"` |
| `statuses` | list of strings | `["To Do","In Progress","Done"]` | 9/9 | see §3.2 |
| `labels` | list of strings | `[]` | 9/9 | all empty — a suggestion list, not a constraint; the corpus has 80 labels in use |
| `types` | list of strings | `bug, feature, enhancement, task, chore, docs, spike` | **0/9** | never overridden |
| `priorities` | list of strings | `High, Medium, Low` | **0/9** | declared capitalised, **written lowercase** |
| `definition_of_done` | list of strings | absent | 4/9 | empty `[]` in 3, four items in muster |
| `date_format` | string | `yyyy-mm-dd` | 9/9 | display only; storage is always `YYYY-MM-DD HH:MM` |
| `max_column_width` | int | — | 9/9 | all `20` |
| `default_editor` | string | unset | 3/9 | `"micro"` |
| `auto_open_browser` | bool | `true` | 9/9 | all `true` |
| `hide_empty_columns` | bool | unset | **0/9** | |
| `default_port` | int | `6420` | 9/9 | `6421` in storefront |
| `remote_operations` | bool | `true` | 9/9 | `false` in jade-palace |
| `auto_commit` | bool | `false` | 9/9 | all `false` |
| `filesystem_only` | bool | `false` | 9/9 | all `false` |
| `zero_padded_ids` | int | unset | 2/9 | `3` in storefront and jade-palace |
| `bypass_git_hooks` | bool | `false` | 9/9 | all `false` |
| `check_active_branches` | bool | `true` | 9/9 | all `true` |
| `active_branch_days` | int | `30` | 9/9 | all `30` |
| `task_prefix` | string | `task` | 9/9 | all `"task"`; read-only via `backlog config` |
| `default_assignee` | string | unset | **0/9** | |
| `default_reporter` | string | unset | **0/9** | implies a `reporter` frontmatter field |
| `on_status_change` | string | unset | **0/9** | also accepted as `onStatusChange`; implies an `onStatusChange` frontmatter field |
| `backlog_directory` | string | unset | **0/9** | root config only |

Key ordering in the file varies (`definition_of_done` appears both mid-file and at the end) and must not be relied on.

---

## 4. Directory layout

`backlog init` creates the complete skeleton and **leaves the unused directories empty** — verified against 1.48.0. All nine projects have all eleven directories.

```
<project>/
├─ backlog/  or  .backlog/  or  <custom path>/
│  ├─ config.yml
│  ├─ tasks/          task-<n> - <slug>.md          852 files
│  ├─ drafts/         draft-<n> - <slug>.md           0 files
│  ├─ milestones/     m-<n> - <slug>.md              58 files
│  ├─ docs/           doc-<n> - <slug>.md            17 files
│  ├─ decisions/      decision-<n> - <slug>.md       64 files
│  ├─ completed/      task-<n> - <slug>.md            0 files
│  └─ archive/
│     ├─ tasks/       task-<n> - <slug>.md           30 files
│     ├─ drafts/                                      0 files
│     └─ milestones/                                  0 files
└─ backlog.config.yml   (only with --config-location root)
```

**`completed/` vs `archive/`.** These are different states, not synonyms. The CLI's own help is explicit: *"archive is only for duplicate, canceled, or invalid tasks"*, while finished work goes to `completed/`. Archiving is documented as a soft delete and **archived ids can be reused by new tasks**. Empty in all nine projects; the parser must still read both. `docs/` also supports nested subdirectories via a `path` argument, though none are used here.

### 4.1 Filename scheme

Every one of the 1021 files matches `<prefix>-<number> - <slug>.md`. The separator is **space-hyphen-space**, and zero files lack it.

The number is the id's numeric part, matching the id's zero-padding. The **prefix is lowercased while the frontmatter id is uppercased**: `task-87 - ….md` holds `id: TASK-87`. 882 of 1021 files differ from their id in case alone; the 139 that match exactly are milestones, docs and decisions, whose ids are lowercase to begin with. Match case-insensitively.

Drafts use the `draft-` prefix with a `DRAFT-` id. `task_prefix` can change the task prefix project-wide, so `task-` must not be hardcoded — read it from config.

**The slug is a lossy, one-way derivation of the title.** `sanitizeFilename` in 1.48.0 is:

```js
title.replace(/[<>:"/\\|?*]/g, "-")
     .replace(/['(),!@#$%^&+=\[\]{};]/g, "")
     .replace(/\s+/g, "-")
     .replace(/-+/g, "-")
     .replace(/^-|-$/g, "")
```

Two different characters both become `-`; a third class vanishes entirely; runs collapse. **Only 203 of 1021 slugs (20%) round-trip to their title by replacing hyphens with spaces.** 813 are lossy and 5 carry no title information at all. Milestones are worse still: `buildMilestoneFilename` additionally **lowercases and truncates to 50 characters**, so `m-9 - m10---инструменты-изучения-toolbox,-тулбар-объекта.md` is all that survives of `M10 - Инструменты изучения (toolbox), тулбар объекта`.

> **Rule: take the title from frontmatter. Never parse it out of the filename.** The filename is useful only for the id prefix and for locating the file.

---

## 5. Variations and malformed files

Nothing in the corpus is broken beyond parsing. Every one of the 1021 files yields valid frontmatter and a coherent body. What follows is **real variation the parser must support**, then hazards that are structurally possible but absent here.

### 5.1 Real variation — must be supported

**V1. Folded block-scalar titles — 30 files across all entity types (17 tasks, 13 decisions).** When a title exceeds the YAML line width the writer emits a folded scalar instead of an inline string:

```yaml
id: TASK-91
title: >-
  Изучил подход к генерации карты в PEAK и сделай выводы о том что можно
  использовать в этой игре
status: Done
```

A line-oriented frontmatter reader sees `title:` with an empty value and drops the title. Affects 30 of 1021 files across four projects.
**Handling: parse frontmatter with a real YAML parser. This alone disqualifies the naive approach.**

**V2. Non-ASCII filenames — 912 of 1021 (89%).** Cyrillic dominates; also `«` `»` `–` `—` `™` `→` `↔` `☰` `✓` and a comma in one milestone name. All NFC-normalised, all valid UTF-8.
**Handling: treat paths as UTF-8 byte strings, never transliterate, do not assume NFC on other platforms (macOS returns NFD from some APIs).**

**V3. Filename slugs stripped to nothing — 5 files.** All in wall_diggers: `task-87 - -.md`, `task-88 - -.md`, `task-89 - -.md`, `task-90 - -.md`, `task-94 - -.md`, plus `task-91 - -PEAK-.md` where only an ASCII fragment survived. All six have fully populated Cyrillic titles in frontmatter. Their creation dates interleave with neighbouring files whose Cyrillic slugs are intact, so this is an older or alternative writer that dropped non-ASCII — not a version boundary and not something the file itself signals.
**Handling: a covered case of the §4.1 rule. Never derive a title from a filename.**

**V4. Same id in `tasks/` and `archive/tasks/` — 1 occurrence.** `TASK-073` in storefront names two different tasks with different titles. Archiving is a soft delete and ids get reused.
**Handling: key on (project, directory-class, id). An id is not globally unique, not project-unique, only directory-unique.**

**V5. Ids collide across projects — 200 of 351.** `TASK-1` exists in six projects.
**Handling: every id reference — `dependencies`, `parent_task_id`, `documentation` — resolves inside its own project only. Never resolve across projects.**

**V6. Missing `ordinal` — 6 files.**
**Handling: sort tasks with an ordinal before tasks without, then by ordinal, then by id — matching 1.48.0's own comparator.**

**V7. Colliding ordinals — 4 pairs.** Three in storefront, one in wall_diggers.
**Handling: ordinal is not a unique key. Break ties by id.**

**V8. Empty acceptance-criteria block — 1 file.** `<!-- AC:BEGIN -->` immediately followed by `<!-- AC:END -->`.
**Handling: an empty section, not an error. Report zero criteria, not "no section".**

**V9. Assignee without `@` — 1 file.** A bare email address.
**Handling: `assignee` is a list of opaque strings. Do not validate, do not normalise, do not assume a leading `@`.**

**V10. Case-inconsistent enum values.** `priority` is lowercase on disk against a capitalised config list. Decision `status` is `accepted` in 30 files and `Accepted` in 29. Assignee handles vary in case.
**Handling: compare enums case-insensitively; fold decision statuses for display.**

**V11. Mixed value kinds in one list field.** `documentation` holds both repo-relative paths (107) and bare ids (7).
**Handling: try both resolutions; render the raw string when neither resolves.**

**V12. Task markers quoted inside document prose.** See §2.5.
**Handling: apply the marker grammar only to files under `tasks/`, `drafts/`, `completed/` and `archive/`.**

**V13. Bare `---` lines inside task bodies — 7 files.** All comment-envelope delimiters.
**Handling: split frontmatter on the *first* `---`…`---` pair anchored at byte 0, then stop. Never scan the whole file for `---`.**

### 5.2 Structurally possible, absent from this corpus

Each is supported by 1.48.0 and would appear the moment someone uses the feature. All are covered by samples in the reference corpus.

| Case | Handling |
|---|---|
| A task with **no body at all** — 1.48.0's default for a task created and never edited | Every section is optional. Return an entity with empty sections, not an error. |
| **CRLF line endings** — 1.48.0 normalises `\r\n` before parsing sections | Normalise on read. Zero corpus files, but Windows will produce them. |
| **Acceptance criteria without `#N`** — 1.48.0 ships `migrateToStableFormat` for exactly this | Accept `- [x] text` and fall back to positional numbering. |
| **`reporter`, `subtasks`, `onStatusChange` frontmatter** — all three are in the 1.48.0 serialiser | Read them; never require them. |
| **Per-comment `<!-- COMMENT:BEGIN -->` markers** | Tolerate both envelope forms. |
| **Non-`task` `task_prefix`, custom `types`/`priorities`** | Read from config; never hardcode `task-`, the seven types or the three priorities. |
| **A non-entity file in an entity directory** (README, `.gitkeep`, editor backup) | **Skip with a diagnostic; never fail the project scan.** Test: no frontmatter, or frontmatter without `id`. |
| **`completed/` and nested `docs/` subdirectories** | Empty and unused here; walk them anyway. |

### 5.3 Genuinely broken — skip with a diagnostic

**None found.** Zero parse failures, zero missing frontmatter, zero unpaired markers, zero malformed AC lines, zero dangling dependencies, zero invalid dates, zero orphan statuses, zero orphan milestone references across 1021 files. Every deviation above is legitimate format variation.

The policy for the categories that *would* be broken — unreadable file, non-UTF-8 bytes, frontmatter that is not a YAML mapping, frontmatter without `id` or `title`, unpaired section markers — is the same: **skip the file, emit a diagnostic naming the path and the reason, keep scanning.** One bad file must never blank a project on the board.

---

## 6. The pinned reference corpus

`testdata/backlog-format/` — 44 files with a README documenting each sample's provenance and purpose.

- `01-tasks/` — minimal (both the 1.48.0 empty-body floor and the corpus 8-key floor), fully populated (all 7 sections, all optional keys), dependencies (populated and the inline-`[]` empty form), mixed checked/unchecked acceptance criteria, a Definition of Done block, a Comments envelope, subtasks (padded and unpadded), and a Cyrillic filename.
- `02-entities/` — draft, two milestones, two documents (one quoting task markers in prose), two decisions (both status cases).
- `03-config/` — five configs: both real status lists, the zero-padded variant, a **synthetic six-status config** exercising every remaining key, and a **synthetic root `backlog.config.yml`** with `backlog_directory`.
- `04-layout-dot-backlog/` — a complete `.backlog/` skeleton including empty directories, pinning the `TASK-073` id collision between `tasks/` and `archive/tasks/`.
- `05-variants/` — every case from §5: folded title, empty slug, empty AC block, missing ordinal, bare-email assignee, CRLF, index-less AC, no body, and two non-entity files.

Samples are either byte-exact 1.48.0 output from a throwaway probe project, or real files with **structure preserved exactly and prose, URLs, paths and handles replaced by neutral placeholders**. The corpus is verified free of business names, personal handles, credentials and home-directory paths, since this repository will be public.

---

## 7. CLI version

```
$ backlog --version
1.48.0
```

Package `backlog.md@1.48.0`, installed via pnpm, running the compiled `backlog.md-linux-x64` binary. Every claim in this document is against that version. The corpus was written over 2026-07-23 to 2026-08-26 by that version and its recent predecessors.

**Re-derive this contract on every Backlog.md minor release.** The two checks that matter most: the section-marker constants and the frontmatter serialiser field order, both readable straight out of the binary.

### 7.1 Re-derived against 1.50.1 (2026-08-26)

The contract holds unchanged. Rather than reading the constants back out of a newer binary, the serialiser was compared by what it writes: both versions initialised a project and wrote a task carrying every section and every optional key — description, two acceptance criteria, a definition-of-done item, plan, notes, final summary, labels, assignee, priority, type, ordinal, reference, documentation and modified file — plus a draft.

**The files are byte-identical**, as are `config.yml` and the eleven-directory skeleton. Nothing in sections 1 to 6 changes.

Four behavioural differences matter to a program driving the CLI, none of them to the format:

| | 1.48.0 | 1.50.1 |
| :-- | :-- | :-- |
| `draft promote` / `draft archive` with an id that does not resolve | prints nothing useful, **exits 0** | exits 1 |
| `task edit` on a `DRAFT-` id | "Task DRAFT-1 not found", exits 1 | same, with a note about branches |
| `backlog init` with no project name and `--defaults` | prompts, finds no terminal, **exits 0 having created nothing** | unchanged |
| Read commands | plain text only | stable JSON output available |

The first is why Muster checks that a promoted or discarded note actually left the inbox instead of trusting the exit code, and the third is why the project name is always passed to `init` rather than left to the CLI's own default.

**`task demote` is the only way back from a task to a draft**, and it does less than its name suggests. Measured on 1.50.1: it renames the id to `DRAFT-<n>` and moves the file into `drafts/`, and leaves `status` and `ordinal` exactly as they were. A task demoted from `In Progress` arrives in the inbox still saying `status: In Progress`, carrying `ordinal: 1000`. So the earlier claim in section 1 that a draft has `status: Draft` and no ordinal holds only for drafts the CLI created — a demoted one has neither property, and anything reading drafts must key on the directory rather than on the status.

**Still absent in 1.50.1: any way to edit a draft.** There is no `draft edit`, and `task edit` refuses a `DRAFT-` id. A draft can be created with the full task field surface through `task create --draft`, and it can be promoted or archived, but between those two points it cannot be changed. Anything calling itself draft editing is capture-and-discard.

---

## 8. Where the CLI and the corpus diverge

The gap between "what the format supports" and "what these 1021 files happen to contain". Building to the corpus alone means breaking on the first project that uses one of these.

| Supported by 1.48.0 | In the corpus | Consequence for the parser |
|---|---|---|
| `reporter` frontmatter field (from `default_reporter`) | never | Read it; do not require it. |
| `subtasks` frontmatter list on parents | never — parents carry no back-reference | Derive the tree from `parent_task_id`, but tolerate `subtasks` appearing. |
| `onStatusChange` frontmatter field | never | Ignore for reads; do not choke. |
| Per-comment `<!-- COMMENT:BEGIN/END -->` markers | never — only the `---` envelope | Tolerate both. |
| `Acceptance Criteria (Optional)` legacy heading; index-less AC items | never — all 3336 items carry `#N` | Accept both; fall back to positional numbering. |
| Custom `types` and `priorities` | never — all nine use the defaults | Read from config; never hardcode the seven types or three priorities. |
| Custom `task_prefix` | never — all nine use `task` | Read from config; never hardcode `task-`. |
| `hide_empty_columns`, `default_assignee`, `default_reporter` | never set | Support them; default sensibly. |
| Root `backlog.config.yml` and `backlog_directory` | never — all nine use `<dir>/config.yml` | **Discovery must probe all three locations**, or a project simply will not be found. |
| Custom `backlog_directory` path | never — only `backlog/` and `.backlog/` | Same. |
| `completed/` and `archive/drafts/`, `archive/milestones/` | always empty | Walk them; they will fill up. |
| Nested subdirectories under `docs/` | never | Walk recursively. |
| Drafts | zero across nine projects | The `DRAFT-<n>` sequence and `status: Draft` are verified against 1.48.0 directly, not against the corpus. |
| Priority declared `High`/`Medium`/`Low`, written `high`/`medium`/`low` | lowercase only | **Compare case-insensitively** — a direct config-to-file comparison fails today. |
| `date_format` config key | all `yyyy-mm-dd` | Display only. Storage is always UTC `YYYY-MM-DD HH:MM`. Do not use this key to parse dates. |

### The seven hazards that matter most

1. **Folded `title: >-`** — 30 files. Use a real YAML parser, never a line reader.
2. **Filename slugs are lossy and one-way** — 80% do not round-trip, 5 carry no title at all. Title comes from frontmatter.
3. **Ids are only directory-unique** — 200 collide across projects, 1 collides between `tasks/` and `archive/`. Key on (project, directory-class, id).
4. **Bare `---` inside comment bodies** — 7 files. Anchor frontmatter to the first `---`…`---` pair at byte 0.
5. **Seven literal marker strings in two naming conventions**, all optional, whitespace after the heading not uniform. Hardcode all seven; match on the marker.
6. **Enum case is not normalised** — `priority` lowercase against a capitalised config, decision `status` split 30/29 across two cases.
7. **Config discovery has three locations** — `backlog/`, `.backlog/`, and a root `backlog.config.yml` that can point anywhere. Probing only the first misses 1 of the 9 projects on this machine today.
