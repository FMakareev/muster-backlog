---
id: doc-1
title: Muster Specification v0.1
type: specification
created_date: '2026-08-26 14:57'
updated_date: '2026-08-26 19:45'
---
> Muster every backlog. One application over every project.

## 1. Problem

Measurements taken 2026-08-26 across working repositories:

| Project | To Do | In Progress | Done | Milestones |
| :-- | --: | --: | --: | --: |
| CarcassonneLike | 24 | 0 | 93 | 13 |
| Refloft | 30 | 1 | 134 | 14 |
| Jade Palace | 24 | 6 | 46 | 2 |
| Treeline | 26 | 3 | 12 | 7 |
| HorrorShop | 10 | 1 | 68 | 2 |
| WallDiggers | 48 | 2 | 112 | 11 |
| **total** | **162** | **13** | **465** | **49** |

All six run on Backlog.md CLI 1.48.0, roughly 640 task files. A seventh project has no backlog at all.

Every existing tool handles one backlog at a time:

- `backlog browser` is single-repo by design, and `default_port: 6420` is hardcoded in all six configs, so six instances fight over one port.
- The VSCode extension (`ysamlan/vscode-backlog-md`) is genuinely good and covers the day-to-day well: kanban with drag-and-drop, a sortable task list, drafts, documents, decisions, live file sync. Its multi-backlog support **switches** the active folder rather than combining folders — and it only exists inside VSCode.

So there is no view that answers "what is happening across everything I work on", and no place for tasks in folders that are not development repositories.

### Rejected alternative

Obsidian plus symlinks to `backlog/tasks` of the six repositories plus Bases (tested on a bench, Obsidian 1.13.7). It indexes, it groups, the numbers add up. Rejected: a table gives no fast access to the task body — description, acceptance criteria, implementation plan — and without that, 640 tasks read as a dump.

## 2. What this is

A local-first desktop task manager with Backlog.md as its backend and a thin layer of its own for working across several projects at once.

Ordinary task management surface: kanban board, list view, filters and search, analytics, a drafts inbox, and a reader for project documents and decisions. The difference from everything above is that all of it spans every registered project simultaneously, and that any folder can be turned into a project from the interface.

## 3. Non-goals

- **Introduces no format of its own.** If Backlog.md does not support a field natively, Muster does not invent it. This rules out custom labels, custom frontmatter and sidecar files — no exceptions, no clever conventions.
- **Does not write task files itself.** Every write goes through the `backlog` CLI (see §5).
- Not a server, not a team tool, not cloud, not sync. One machine, one person.
- Not a time tracker and not hour-based estimation.
- Does not replace `git`. Agents commit; Muster only shows state.
- Not an estimation or capacity planner. Work-in-progress limits are counts of native data, nothing more.

## 4. Model

### Project registry

`~/.config/muster/projects.yml` is the only configuration Muster owns. It holds where projects are and how they are displayed — nothing that belongs to Backlog.md:

```yaml
projects:
  - path: /var/mnt/mydata/Dev/TREELINE
    name: Treeline          # display name; defaults to the folder name
    color: "#7aa2f7"
  - path: /var/mnt/mydata/Dev/ref_canvas
    name: Refloft

wip_limits:                 # optional, advisory only
  In Progress: 3
```

Everything else — statuses, priorities, types, labels, task prefix — is read from each project's own `config.yml`, which stays the authority.

### Task

Exactly what lies in `<project>/backlog/tasks/*.md`, using only the fields Backlog.md defines: `status`, `priority`, `milestone`, `dependencies`, `ordinal`, `type`, `labels`, `assignee`, `created_date`, `updated_date`.

### Statuses

Statuses are per-project configuration and registered projects will not agree on them. The board therefore derives its columns from the **union** of every project's declared status list. A project that does not declare a status is simply empty in that column, and a card cannot be dragged into a status its own project does not declare.

Muster never edits another project's `config.yml`.

### Drafts as inbox

Backlog.md keeps drafts off the board by design. That is what makes capture cheap: anything can be thrown in without deciding whether it deserves tracking. Muster surfaces drafts as an inbox with a triage view, so the pile stays visible and gets emptied.

### Analytics

The `backlog overview` command already defines a useful vocabulary from purely native data: status and priority breakdowns, average task age, stale tasks, tasks blocked on unmet dependencies, recent activity, milestone progress. Muster shows the same figures across every registered project at once, drillable to the underlying tasks.

Work-in-progress limits are the one advisory signal on top: a count per column against a configured number, flagged but never enforced. It addresses the observed pattern of thirteen simultaneously active tasks, six of them in one project.

## 5. Architecture

Stack: **Wails v3.0.0-beta.8, Go 1.25, Svelte 5, Vite 8, Tailwind 4, nanostores**, with **SVAR Svelte Kanban** (MIT, native Svelte 5, virtualised) for the board so drag-and-drop, grouping and large-board performance are not hand-built. Hotkeys via `golang.design/x/hotkey`, MCP via the Go MCP SDK.

**Reads are direct.** A frontmatter parser of our own over `tasks/`, `drafts/`, `milestones/`, `docs/`, `decisions/`, `completed/`. 640 files take single-digit milliseconds. Everything lives in memory; no database is needed.

**Writes go only through the `backlog` CLI.** `task create|edit`, `draft create|promote`, `init`. The VSCode extension writes markdown directly and gets away with it, but the format is alive — bodies are marked up with sections (`<!-- SECTION:DESCRIPTION:BEGIN -->`, `<!-- AC:BEGIN -->`), `ordinal` holds manual ordering, id generation handles collisions, and there is a `backlog doctor` command precisely because that is not trivial. A writer of our own would have to chase every CLI release, and would be the first step toward owning the format. The board therefore settles on what a rescan confirms, not on where a card was dropped.

**Project creation goes through `backlog init`.** Every prompt that command asks has a corresponding flag — `--backlog-dir`, `--config-location`, `--task-prefix`, `--zero-padded-ids`, `--integration-mode`, `--agent-instructions`, `--no-git` — plus `--defaults`. So the UI presents a form over those flags rather than emulating a dialogue. Any folder can become a project, git repository or not.

**Nothing reaches the network.** Fonts, icons and every other asset are bundled, and rendered markdown has its remote resources stripped. This has to be checked rather than assumed: SVAR themes load an icon font and a typeface from cdn.svar.dev by default, which was found only by recording the requests a running window makes. The themes are mounted with fonts disabled and the icons are served from the bundle. Any new dependency needs the same check.

**Liveness via fsnotify** on each project's directories. An agent changes a status and the screen updates without polling.

**Capture via a global hotkey.** One window, one field: project plus text, sent to `backlog draft create`.

**MCP server (optional, last milestone).** An agent talks to Muster instead of one CLI per repository and sees context no single project has.

## 6. Screens

1. **Board.** Multi-project kanban: columns from the union of project statuses, cards coloured and groupable by project, drag-and-drop status changes, work-in-progress limits flagged.
2. **List.** Sortable, filterable table across every project, with a customisable column set. The mode for scanning many tasks, where the board is the mode for moving a few.
3. **Task.** A panel opened by click from anywhere: description, acceptance criteria with checkboxes, implementation plan, notes, dependencies, file links. Fast access to the full body is what a table alone cannot give.
4. **Inbox.** Drafts across every project, with triage: promote, edit, reassign, discard.
5. **Documents.** Documents and decision records of every project, rendered with the same markdown pipeline as task bodies, Mermaid included.
6. **Analytics.** Cross-project overview, drillable.
7. **Projects.** Add a folder, initialise a backlog in it, rename, recolour, reorder, hide.
8. **Capture.** A window on a global hotkey.
9. **Search.** Across tasks, drafts, documents and decisions of every project, by keyboard from anywhere.

## 7. MVP v0.1

The read surface plus the minimum of writing:

- project registry from yml (no editing UI)
- parser plus fsnotify
- unified board columns from the union of project statuses
- multi-project kanban with drag-and-drop status changes
- task panel with the full body
- markdown rendering with Mermaid
- priority and assignee changes through the CLI

Out of scope: list view, search, analytics, work-in-progress limits, documents viewer, inbox and capture, initialising projects from the UI, MCP.

The MVP answers one question: **does a standalone application over several projects at once beat the per-repository tooling already in use?** The VSCode extension covers single-repo work well, so the honest comparison is against it.

First use says *not yet*: the multi-project view works and performs, but a viewer that can move cards is not a task manager. See the [verdict](doc-6%20-%20MVP-v0.1-verdict.md). The hypothesis is neither refuted nor confirmed, and m-2 now carries what stands between the two.

## 8. Open questions

1. **SVAR Kanban has no swimlanes in the MIT edition.** Confirmed by reading the 2.6.0 component API: there is no rows prop, whatever the readme implies. Projects are distinguished by grouping and card colour instead. Confirm this still reads well at nine projects before building the rest of the board.
2. **Column ordering across disagreeing projects.** The union is well defined; the order is not, and it needs a deterministic, documented rule. Measured position is easier than feared: only two distinct status lists exist across the nine projects, and one is a superset of the other, so today's union orders itself. That is a coincidence of this corpus, not a property of the format — the rule must still handle disjoint lists.

### Settled since the first draft

- **Wails v3 beta** — settled in [decision-1](../decisions/decision-1%20-%20Build-the-desktop-shell-on-Wails-v3-beta.md): pinned to beta.8 across all three version streams, confined to the stable half of the API, with an upgrade policy and Wails v2 as the fallback.
- **Name display** — files are named `task-10 - Bike-rig-on-2D-physics-frame-and-wheels.md`, hyphens instead of spaces. Measured in the format contract: only 20% of filenames round-trip back to their title, and five collapse to nothing at all. The title always comes from frontmatter. No longer a question, just a rule.

## 9. Open source

Backlog.md is distributed under MIT (verified on 1.48.0) and SVAR Kanban has an MIT edition, so the layer on top is unencumbered. Muster is MIT.

The README makes one promise: **a local-first desktop task manager over all your Backlog.md projects at once.** The comparison to `backlog browser` and to the VSCode extension is stated plainly and fairly — both are good at one backlog; this is the one that spans them.

## 10. Decisions taken (2026-08-26)

- Stack fixed as Wails v3 beta per §5; beta risk tracked as a spike with a pinned version and rollback plan.
- Board and list built on SVAR Svelte Kanban (MIT) rather than hand-rolled drag-and-drop.
- Board columns are the union of project status lists; other projects' configuration is never modified.
- Writes go only through the `backlog` CLI.
- No field, label convention or sidecar of our own is ever added to the Backlog.md format. The review-cost labels and the review-budget planner from the first draft of this specification were dropped for exactly this reason; what survives is work-in-progress limits computed from native counts.
- No special "personal project" concept: any folder can be initialised as a project from the UI.
- License MIT; release automation via release-please driven by Conventional Commits.
- Repository name `muster-backlog`.
- Documentation and decisions live as Backlog.md documents and decisions rather than in a separate `docs/` tree — see [decision-3](../decisions/decision-3%20-%20Read-Backlog.md-markdown-directly-write-only-through-the-CLI.md) and the [conventions](doc-4%20-%20Documentation-and-decision-conventions.md).
- Publication to GitHub waits until the application actually delivers a multi-project board, not merely a tidy repository.
