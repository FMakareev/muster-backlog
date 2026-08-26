---
id: doc-6
title: MVP v0.1 verdict
type: specification
created_date: '2026-08-26 19:44'
updated_date: '2026-08-26 19:44'
---
> **Status: provisional.** The measurements and the first-use assessment are recorded. The part that is missing is a real trial — days of using it instead of the existing tools — and that cannot be substituted for.

## The question

Specification section 7 states what the MVP is for:

> Does a standalone application over several projects at once beat the per-repository tooling already in use?

The honest comparison is against the VSCode extension `ysamlan/vscode-backlog-md`, which is good at one backlog, not against `backlog browser`, which nobody was going to use six times over.

## What was measured

Against nine real projects and 884 tasks, on the developer machine:

| | |
| :-- | --: |
| Parsing 1027 entities from disk | 101 ms, zero diagnostics |
| Full load, all nine projects | 105 ms |
| Reload of one project after a change | ~11 ms |
| Filtered query over the whole corpus | 274 µs |
| Cards in the DOM against 884 loaded | 52–57 |
| Full re-sort of the board | 95 ms |
| Opening the task panel | 90 ms |
| JS heap | 18 MB |
| Network requests during use | zero |

Nothing here is a constraint. The premise that a thousand markdown files can be held in memory and re-read on every change is correct, and the startup budget of two seconds is enforced by a test that currently passes twenty times over.

## The first-use verdict

**Not yet.** In the author's words after using it: *not close to the extension, but already quite convenient.*

That is two findings, not one.

**The multi-project view is worth having.** Seeing 884 tasks from nine repositories in one set of columns, each card carrying its project's colour, is something no existing tool does and it reads well. The panel gives the fast access to a task body whose absence disqualified the Obsidian alternative. Moving a task writes through the CLI and the files agree afterwards. None of that is in doubt.

**But the MVP is a viewer that can move cards, not a task manager.** What it cannot do is the reason it does not yet replace the extension:

1. **No task creation.** A task manager that cannot create a task sends you to a terminal, which breaks the loop it exists to close. → TASK-57
2. **No editing of a task's body.** Description, acceptance criteria, plan and notes are readable and unchangeable, so any real edit still happens elsewhere. → TASK-58
3. **No search.** With 884 tasks, finding one by remembering its title is the most common navigation there is. → TASK-51
4. **No filters** by status or priority. → TASK-22
5. **Milestones are illegible.** A card shows `m-1`, which reads like a task id and says nothing. Milestones are the main axis this backlog is planned on and the board cannot group by them. → TASK-59
6. **Subtasks are invisible** — a subtask is just another card with a dotted id. Optional, and not at the cost of a harder-to-read board. → TASK-61
7. **Reading is uncomfortable on a wide screen**, because the panel is a column against the right edge. → TASK-60

## What the per-repository tools still do better

- **Editing.** The extension edits task bodies in a real editor with frontmatter autocomplete, task-id completion and hover previews. Muster edits four fields through selects.
- **Being where the work is.** The extension is already open, in the window where the code is. Muster is another window to reach for — which is also its point, but the cost is real.
- **Search and navigation** inside one project, which the extension has and Muster does not.

Muster's only advantage today is the one it was built for: seeing across projects at once. That advantage is real, and it is not yet enough on its own.

## Consequence for the plan

The hypothesis is **not refuted** — the multi-project view works and performs — but it is **not yet confirmed** either, because what would make it a daily replacement is the ordinary task-manager surface it does not have.

m-2 was "Views, filters and analytics". It is now **Daily use: editing, views and filters**, and it carries the seven items above. The verdict is revisited after it, against the same question and with a real trial.

## What is still owed

A trial. Not a demonstration and not a measurement: a stretch of days where this is the tool reached for instead of the extension, with what went wrong written down. Everything above is a first impression, and first impressions of one's own work are the least reliable evidence there is.
