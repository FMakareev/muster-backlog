# Muster

**A local-first desktop task manager over all your [Backlog.md](https://github.com/MrLesk/Backlog.md) projects at once.**

> **Status: pre-alpha.** It reads and writes real Backlog.md projects: board, list, search, filters, documents, figures, task editing, and adding or initialising projects from the interface. Not packaged, not released, and not yet used by anyone but its author. Follow [the roadmap](backlog/docs/doc-2%20-%20Roadmap-to-1.0.md) for what is still missing.

## Why

Backlog.md keeps tasks as markdown inside the repository they belong to, which is the right place for them. The cost of that is one backlog per repository, and every tool for it works on one at a time: `backlog browser` is single-repo by design and pins the same port in every project, and the excellent [VSCode extension](https://github.com/ysamlan/vscode-backlog-md) switches between backlog folders rather than combining them.

So there is no answer to _what is happening across everything I work on_. Nine repositories, a thousand task files, and no single view of them.

Muster is that view: one board, one list, one search, one set of numbers, over every project you register.

## What it will be

An ordinary task manager — kanban with drag-and-drop, a sortable list, filters, cross-project search, a drafts inbox, a reader for project documents and decisions, and analytics. Nothing about that is novel. The only thing that is: all of it spans every registered project at once, and any folder on your disk can become one from inside the application.

The documents viewer writes as well as reads. A document can be created in any project with its type and body, and edited afterwards — `doc update --content` replaces a document wholesale, so the editor holds the whole file and sends it as it stands. A decision can be created with its status, and there it stops: Backlog.md writes a skeleton with Context, Decision and Consequences headings and has no command that fills them in, so the viewer says the decision itself is written in the file rather than pretending to be an editor for it.

The **Inbox** (`i`) is the drafts folder made visible. Backlog.md keeps drafts off the board by design, which is what makes capture cheap and also what makes an unread drafts folder the pile nobody looks at — so notes are listed oldest first, across every project, each saying how long it has waited, with the depth on the navigation itself. From there a note is promoted into a task, rewritten, moved to another project, or discarded into the archive.

Notes are captured from the same form that makes tasks — one checkbox says whether it goes to the board or to the inbox — and a note carries everything Backlog.md lets a draft hold: description, labels, assignee, priority, type, milestone and acceptance criteria. Only a status is out of reach, because a draft's status is Draft, and the form says so rather than dropping it silently. Clicking a note opens it in the task panel to be read whole; promoting one opens the task it became, so a note can be finished in a single pass.

A note opens in the task panel to be read whole, and nothing more: every editing control there writes through `task edit`, which refuses a `DRAFT-` id, so a note gets the two things that do work — promote it, or rewrite it in the inbox — and the panel says why. Promoting from there opens the resulting task, editable.

Backlog.md has no `draft edit` — `task edit` refuses a `DRAFT-` id outright — so rewriting a note captures a new one and archives the old, which is also the only way to move a draft between projects. Everything above survives that; the capture date does not, and the interface says so rather than hiding it.

A task's relationships are editable in the panel: what it depends on, the references and documents it points at, and the files it touched. Dependencies resolve inside their own project — ids collide freely across projects — so one that does not resolve is refused with the reason before anything is written. The files-touched list is the one exception to editing: Backlog.md can set it and has no command that empties it, so the interface says so instead of offering a button that does nothing. Manual order is dragging: a card dropped within its column takes the ordinal between its new neighbours, and where the gaps have run out the column is restacked at multiples of 1000, which is the allocation Backlog.md uses itself.

A task ends somewhere too. When one reaches its project's own last status it can be filed into `completed/`, where Backlog.md keeps finished work; any task can be archived, which is a soft delete rather than a delete; and any task can be sent back to the inbox as a note. Archived and completed tasks are off the board, the list, search and the figures — asking for nothing in particular means asking for the live ones.

Where Backlog.md has a relationship, Muster shows it without inventing a shape for it. Subtasks are the case that tempts a redesign: a parent card says how many of its subtasks are finished and a subtask card says whose it is, but the board stays flat and the links live in the task panel. Nesting cards inside cards would cost more legibility across nine projects than the relationship is worth.

## What it is not

- **Not a replacement for Backlog.md.** It is a view onto it.
- **Not a new format.** Muster reads Backlog.md markdown directly and writes only through the `backlog` CLI. It adds no field, no label convention and no sidecar file of its own — if Backlog.md does not support something natively, Muster does not store it. That is a hard rule, not an aspiration; it is why the review-budget planner from the first draft of the specification no longer exists.
- **Not a server, a team tool, or a sync service.** One machine, one person, no accounts, no network.
- **Not a time tracker**, and not an estimation tool.

## Prerequisites

| Requirement                                            | Version used                      |
| :----------------------------------------------------- | :-------------------------------- |
| [Go](https://go.dev)                                   | 1.25 or newer                     |
| [Node.js](https://nodejs.org)                          | 24 or newer                       |
| [pnpm](https://pnpm.io)                                | 11 or newer                       |
| [Wails v3 CLI](https://v3.wails.io)                    | v3.0.0-beta.8                     |
| [golangci-lint](https://golangci-lint.run)             | v2 (needed by the git hooks)      |
| [Backlog.md CLI](https://github.com/MrLesk/Backlog.md) | 1.48.0 minimum, 1.50.1 recommended (at runtime, not to build) |

1.48.0 is the floor because that is the version the [format contract](<backlog/docs/doc-3 - Backlog.md-Format-Contract.md>) was measured against; 1.50.1 writes byte-identical files but reports two failures that 1.48.0 swallows, so it is what the author runs. Muster never relies on an exit code for a write whose result it can check.

The two Go tools install themselves:

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.8
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

### Linux system libraries

The default Linux build targets **GTK 4** and **WebKitGTK 6.0**:

```sh
# Fedora
sudo dnf install gtk4-devel webkitgtk6.0-devel
# Debian / Ubuntu 24.10+
sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev
```

Distributions that do not ship WebKitGTK 6.0 — **Ubuntu 24.04 LTS** among them, where it only arrives in 24.10 — can build against GTK 3 and WebKit2GTK 4.1 instead:

```sh
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
```

Nothing else to do: the build detects what the machine has and picks the tag itself. `EXTRA_TAGS=gtk3` still works if you want to force it.

## Build and run

From a clean machine, after the prerequisites above:

```sh
git clone https://github.com/FMakareev/muster-backlog
cd muster-backlog
pnpm install                        # workspace dependencies and the git hooks
wails3 task build
./bin/muster
```

A window opens with a placeholder board. That is currently the whole of it — there is no registry to point at your projects yet.

Day to day:

```sh
wails3 task run       # build and run
wails3 task dev       # run with frontend hot reload
```

Linux packages, if you want them rather than the bare binary:

```sh
wails3 task package   # .deb, .rpm, .pkg.tar.zst and an AppImage into bin/
```

Before trusting a build, walk the [smoke checklist](backlog/docs/doc-5%20-%20v0.1-smoke-checklist.md). It takes a few minutes and every step in it has failed at least once.

## Configuration

Muster keeps one file of its own, at `$XDG_CONFIG_HOME/muster/projects.yml` — or `~/.config/muster/projects.yml` when `XDG_CONFIG_HOME` is unset. It records **where** projects are and how to display them, and nothing else:

```yaml
projects:
  - path: ~/Dev/treeline # a leading ~ is expanded
  - path: /var/mnt/data/refloft
    name: Refloft # optional; the folder name is used otherwise
    color: "#7aa2f7" # optional

wip_limits: # optional, advisory only
  In Progress: 3
```

Order in the file is the order on screen. How a project _works_ — its statuses, priorities, types, labels, task prefix — is read from that project's own `config.yml` and is never copied here, because a copy would go stale the moment the project changed.

You do not have to write this file by hand. The **Projects** screen (`p`) adds folders — typed, pasted, or chosen with the desktop's own directory picker — renames and recolours them, arranges the order, hides a project from the board without unregistering it, and turns a folder that has no backlog into one by running `backlog init` behind a form. Edits are written into the file in place: comments, key order and indentation survive, and only an entry that actually changes is rewritten. Unregistering a project removes the entry and leaves everything in the folder alone — nothing in Muster deletes a backlog.

A milestone can also be made wherever one is chosen — the create form, the task panel, the inbox — because the moment you find you need one is the moment you are assigning a task, not the moment you are on the Projects screen. A task's milestone can be changed and cleared from the panel like any other field.

Milestones are managed from the same screen: added, renamed, and retired. Retiring one is where the care goes, because whichever command does it the file ends up in `archive/milestones` — the real choice is what becomes of the tasks that named it, and leaving them alone, clearing them and moving them elsewhere are three different things to a plan. The form says how many tasks are affected and makes the choice explicit.

An entry may also carry `hidden: true`, which keeps a project registered and loaded but out of the board, the lists, search and the figures.

Muster finds a project's data directory the way the Backlog.md CLI does: `backlog/`, then `.backlog/`, then a custom path named by `backlog_directory` in a root `backlog.config.yml`. A registered folder that holds no Backlog.md project is shown as such rather than dropped, and never prevents the others from loading.

There is no registry file until you create one, and that is not an error — the application will offer to add the first project.

Everything is held in memory and reloaded from disk rather than cached in a database. Measured over nine real projects and 884 tasks: a full load takes about 100 ms, reloading a single project after a file change costs about 11 ms, and a filtered query over the whole corpus takes under a millisecond. The startup budget is 2 seconds, enforced by a test.

## Git hooks

Installed by `pnpm install`. They format and lint staged files before a commit, check the commit message against the convention, and run tests before a push. See [CONTRIBUTING.md](CONTRIBUTING.md).

## Lint and format

One command covers Go and the frontend:

```sh
wails3 task lint      # golangci-lint, ESLint, Prettier, svelte-check
wails3 task lint:fix  # apply every available autofix
```

`golangci-lint` is the only tool not installed by the steps above:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

Frontend-only checks, from `frontend/`:

```sh
pnpm run check        # svelte-check
pnpm run lint         # ESLint + Prettier
pnpm run build        # production frontend build
```

## Stack

Wails v3 (beta) with a Go backend and a Svelte 5 frontend: Vite 8, Tailwind 4, nanostores for state, and [SVAR Svelte Kanban](https://svar.dev/svelte/kanban/) (MIT) for the board. Every frontend dependency is pinned to an exact version.

## Project management

This repository tracks its own work in Backlog.md. Tasks live in [`backlog/`](backlog/) and are managed with the `backlog` CLI — not in GitHub Issues.

```sh
backlog board          # the kanban board for this repository
backlog task list --plain
backlog overview
```

## Documentation

Documentation lives as Backlog.md documents and decisions, not in a separate tree — see [Documentation and decision conventions](backlog/docs/doc-4%20-%20Documentation-and-decision-conventions.md) for why and how to add to it.

|                                                                                      |                                                       |
| :----------------------------------------------------------------------------------- | :---------------------------------------------------- |
| [Specification v0.1](backlog/docs/doc-1%20-%20Muster-Specification-v0.1.md)          | what is being built, and explicitly what is not       |
| [Roadmap to 1.0](backlog/docs/doc-2%20-%20Roadmap-to-1.0.md)                         | the five milestones and why they are ordered that way |
| [Backlog.md Format Contract](backlog/docs/doc-3%20-%20Backlog.md-Format-Contract.md) | the on-disk format, measured against 1021 real files  |
| [Decisions](backlog/decisions/)                                                      | architecture decision records                         |

## Contributing

Work is tracked as Backlog.md tasks in this repository, not as GitHub Issues. [CONTRIBUTING.md](CONTRIBUTING.md) covers setup, the commit convention, the hooks and the pull request flow; participation is governed by the [Code of Conduct](CODE_OF_CONDUCT.md). Security problems go through the private channel in [SECURITY.md](SECURITY.md), never a public report.

## Licence

MIT — see [LICENSE](LICENSE).
