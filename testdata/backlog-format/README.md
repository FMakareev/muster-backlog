# Backlog.md format reference corpus

Pinned sample files for the Muster parser's unit tests. Derived from **Backlog.md CLI 1.48.0**
and from a survey of 1021 real entity files across nine projects (852 tasks, 30 archived tasks,
58 milestones, 64 decisions, 17 documents). The full findings live in the project document
**"Backlog.md Format Contract"** (`backlog doc list`).

Muster **reads** these files and never writes them. Every write goes through the `backlog` CLI,
so this corpus exists to pin down what the reader must accept - not what it may produce.

## Provenance and sanitisation

Samples marked *pristine* were produced by running `backlog init` / `backlog task create` /
`backlog draft create` / `backlog milestone add` / `backlog doc create` / `backlog decision create`
against 1.48.0 in a throwaway project. They are byte-exact CLI output.

Every other sample is **shaped after a real file but sanitised**: identifiers, dates, ids,
quoting style, marker placement, whitespace and list serialisation are preserved exactly as
found, while prose bodies, URLs, file paths and people's handles were replaced with neutral
placeholders. Structure is the point of these files; the prose is not.

Provenance is given per sample below. Project names refer to: `muster` (this repository),
`wall_diggers`, `ref_canvas`, `CarcassonneLike`, `jade-palace`, `online_shopping_roguelike`,
`TREELINE`, `portfolio_website`, and one private commercial web project referred to here as
`storefront`, which is the only one using a `.backlog/` directory.

Anything labelled **SYNTHETIC** is not in the corpus at all. It covers behaviour the 1.48.0
binary demonstrably supports but that no project on this machine happens to exercise.

## 01-tasks - the normal shapes

| Path | Provenance | What it pins |
|---|---|---|
| `minimal/task-1 - Minimal-task.md` | pristine 1.48.0 | The floor: 8 frontmatter keys and a **completely empty body**. No headings, no markers. A parser must not require any section. |
| `minimal/task-103 - Description-and-acceptance-criteria-only.md` | `wall_diggers` | The corpus floor: 8 keys, Description + Acceptance Criteria only, no `milestone`, `priority`, `type` or `updated_date`. |
| `full/task-3 - Fully-populated-task.md` | pristine 1.48.0 | All seven body sections in canonical order and every optional frontmatter key the CLI will emit from flags. Note `--priority High` was written to disk as `priority: high`. |
| `full/task-061 - Every-frontmatter-key-observed.md` | `storefront` | The 14-key maximum observed in the corpus, including `references`, `documentation`, `modified_files` and `parent_task_id` together. |
| `dependencies/task-53 - Multiple-dependencies.md` | `muster` | `dependencies` as a populated block sequence, including a `TASK-N.NN` subtask id. |
| `dependencies/task-54 - Empty-lists-serialised-inline.md` | `muster` | The empty-list form. `assignee`, `labels` and `dependencies` are written inline as `[]`, never omitted, never as an empty block sequence. |
| `acceptance-criteria/task-016 - Mixed-checked-and-unchecked.md` | `jade-palace` | `- [x] #N` and `- [ ] #N` in one block, with `#`, `-`, `:` and `[` characters inside criterion text. |
| `acceptance-criteria/task-51 - Definition-of-done-block.md` | `muster` | A `<!-- DOD:BEGIN -->` block. Same item grammar as AC, separate marker and heading. Only 8 tasks in the corpus have one. |
| `comments/task-76 - Comments-section.md` | `wall_diggers` | `<!-- COMMENTS:BEGIN -->` with two `author:` / `created:` / `---` / body / `---` envelopes. The bare `---` lines inside the body are the reason a naive "split on `---`" frontmatter reader breaks. |
| `subtask/task-3.1 - A-subtask.md` | pristine 1.48.0 | `parent_task_id`, dotted id, own `ordinal`. The parent carries **no** `subtasks` key. |
| `subtask/task-073.02 - Zero-padded-subtask.md` | `storefront` | Zero-padded parent segment with an unpadded child segment, under `zero_padded_ids: 3`. |
| `non-ascii/task-2 - ...` | pristine 1.48.0 | Cyrillic filename produced by the current slugifier, showing which ASCII punctuation is deleted (`'`, `(`, `)`, `,`) versus replaced by a hyphen (`/`, `:`). |

## 02-entities - drafts, milestones, docs, decisions

| Path | Provenance | What it pins |
|---|---|---|
| `drafts/draft-1 - A-draft-idea.md` | pristine 1.48.0 | Drafts use a separate `DRAFT-N` id sequence, a `draft-N` filename prefix, `status: Draft`, and carry **no `ordinal`**. |
| `milestones/m-0 - sample-milestone.md` | pristine 1.48.0 | Two-key frontmatter (`id`, `title` only) and a free-form `## Description` with no markers. |
| `milestones/m-9 - m10---...` | `ref_canvas` | Milestone filenames are lowercased and truncated to 50 characters, so the slug is even less recoverable than a task's. Title is the one double-quoted scalar the milestone serialiser emits. |
| `docs/doc-1 - Sample-doc.md` | pristine 1.48.0 | `id`, `title`, `type`, `created_date`; body is free markdown with no markers. |
| `docs/doc-2 - Document-with-tags-and-markers-in-prose.md` | `muster` | Optional `tags` list, and task markers quoted inside prose and a fenced block. A marker scanner pointed at every `.md` file in a project mis-parses this. |
| `decisions/decision-1 - Sample-decision.md` | pristine 1.48.0 | `id`, `title`, `date` (not `created_date`), `status`; `## Context` / `## Decision` / `## Consequences`. |
| `decisions/decision-10 - Accepted-...md` | `online_shopping_roguelike` | `status: Accepted` capitalised. Decision status case is not normalised: the corpus holds `accepted` (30), `Accepted` (29) and `proposed` (5). |

## 03-config - config.yml variants

The board's columns are the union of these status lists, so this is the set that matters most.

| Path | Provenance | Status list |
|---|---|---|
| `statuses-default/config.yml` | 8 of 9 projects | `["To Do", "In Progress", "Done"]` |
| `statuses-with-in-review/config.yml` | `muster` | `["To Do", "In Progress", "In Review", "Done"]`, plus a populated `definition_of_done` block list |
| `zero-padded-and-editor/config.yml` | `storefront` | default list, but with `zero_padded_ids: 3`, `default_editor`, `default_port: 6421` and an empty `definition_of_done: []` |
| `statuses-custom-synthetic/config.yml` | **SYNTHETIC** | `["Backlog", "Ready", "Doing", "Blocked", "Review", "Shipped"]` - plus every remaining key the 1.48.0 parser understands: `types`, `priorities`, `hide_empty_columns`, `on_status_change`, `default_assignee`, `default_reporter`, a non-`task` `task_prefix` |
| `root-config-synthetic/backlog.config.yml` | **SYNTHETIC** | The project-root config location, carrying `backlog_directory` to relocate the data folder |

## 04-layout-dot-backlog - the `.backlog/` layout

A complete directory skeleton as `backlog init --backlog-dir .backlog` produces it: `tasks`,
`drafts`, `milestones`, `docs`, `decisions`, `completed`, `archive/{tasks,drafts,milestones}`.
Empty directories are normal - `init` creates all of them, and no project in the corpus had
anything in `completed/`.

It also pins the **id collision** case: `tasks/task-073` and `archive/tasks/task-073` are
different tasks holding the same id, because archiving is a soft delete and the CLI reuses
archived ids. (`.gitkeep` files stand in for the empty directories so git preserves them.)

## 05-variants - what a naive parser breaks on

| Path | Real or synthetic | Intended handling |
|---|---|---|
| `folded-title/task-91 - -PEAK-.md` | real (`wall_diggers`) | `title: >-` folded block scalar, 30 files in the corpus. **Parse frontmatter with a real YAML parser**, never line-by-line. |
| `empty-filename-slug/task-87 - -.md` | real (`wall_diggers`) | Filename slug collapsed to `-`. Take the title from frontmatter, never from the filename. Also the only corpus file carrying all six section kinds. |
| `empty-ac-block/task-94 - -.md` | real (`wall_diggers`) | An `AC:BEGIN`/`AC:END` pair with nothing between, and **no `ordinal` key**. Empty section, not an error. Missing ordinal sorts last. |
| `bare-email-assignee/task-140 - ...md` | real (`storefront`) | An assignee with no leading `@`. `assignee` is a list of free-form strings. |
| `crlf-line-endings/task-2 - ...md` | **SYNTHETIC** | No corpus file uses CRLF, but 1.48.0 normalises `\r\n` before parsing sections. Do the same. |
| `legacy-ac-without-index/task-30 - ...md` | **SYNTHETIC** | `- [x] text` with no `#N`. 1.48.0 still ships a migration path for it; fall back to positional numbering. |
| `no-body/task-1 - ...md` | real shape, pristine content | Frontmatter and nothing else. Every section is optional. |
| `not-a-task/README.md` | **SYNTHETIC** | No frontmatter. Skip with a diagnostic; do not fail the scan. |
| `not-a-task/notes.md` | **SYNTHETIC** | Valid YAML frontmatter with no `id`. Skip with a diagnostic. `id` and `title` are the only fields present in 100% of real entity files. |
