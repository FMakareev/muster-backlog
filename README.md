# Muster

**A local-first desktop task manager over all your [Backlog.md](https://github.com/MrLesk/Backlog.md) projects at once.**

> **Status: pre-alpha.** It reads and writes real Backlog.md projects: board, list, search, filters, documents, figures, task editing, and adding or initialising projects from the interface. Not packaged, not released, and not yet used by anyone but its author. Follow [the roadmap](backlog/docs/doc-2%20-%20Roadmap-to-1.0.md) for what is still missing.

## Why

Backlog.md keeps tasks as markdown inside the repository they belong to, which is the right place for them. The cost of that is one backlog per repository, and every tool for it works on one at a time: `backlog browser` is single-repo by design and pins the same port in every project, and the excellent [VSCode extension](https://github.com/ysamlan/vscode-backlog-md) switches between backlog folders rather than combining them.

So there is no answer to _what is happening across everything I work on_. Nine repositories, a thousand task files, and no single view of them.

Muster is that view: one board, one list, one search, one set of numbers, over every project you register.

![The Muster board over five projects at once](.github/media/board.png)

_Five projects in one set of columns. The colour on a card says which project it came from, `In Review` exists because one of the five declares it, and three different tasks are called TASK-1 — ids collide across projects, which is why nothing here is addressed by id alone._

## What it will be

An ordinary task manager — kanban with drag-and-drop, a sortable list, filters, cross-project search, a drafts inbox, a reader for project documents and decisions, and analytics. Nothing about that is novel. The only thing that is: all of it spans every registered project at once, and any folder on your disk can become one from inside the application.

The documents viewer writes as well as reads. A document can be created in any project with its type and body, and edited afterwards — `doc update --content` replaces a document wholesale, so the editor holds the whole file and sends it as it stands. A decision can be created with its status, and there it stops: Backlog.md writes a skeleton with Context, Decision and Consequences headings and has no command that fills them in, so the viewer says the decision itself is written in the file rather than pretending to be an editor for it.

The **list** (`l`) is the mode for scanning rather than moving: sortable on every column, with the visible column set remembered, and subtasks kept under the task they belong to instead of sorting away from it. Ticking rows changes many at once.

![The task list across every project](.github/media/list.png)

The **figures** (`s`) are the ones the format already supports: how much is open, how much has no priority, how old it is, what is waiting on unfinished work, and what nobody has touched in a month — per project and across all of them.

![Figures across every project](.github/media/stats.png)

The **Inbox** (`i`) is the drafts folder made visible. Backlog.md keeps drafts off the board by design, which is what makes capture cheap and also what makes an unread drafts folder the pile nobody looks at — so notes are listed oldest first, across every project, each saying how long it has waited, with the depth on the navigation itself. From there a note is promoted into a task, rewritten, moved to another project, or discarded into the archive.

A task can be started from where it belongs: the plus on a board column, or a double-click on its empty space, opens the form on that column's status. The card itself is still refused — the board would add it in memory, showing a task the files do not have — so it appears when the CLI has written it.

Capture is meant to cost nothing. The form starts on the project you are looking at, or the one your last note went to, or the first registered; `alt+[` and `alt+]` move between projects without leaving the text; a capture is confirmed in place and the form clears itself for the next thought rather than closing; and nothing typed is lost when the write cannot happen — the text stays where it is with the reason beside it.

Notes are captured from the same form that makes tasks — one checkbox says whether it goes to the board or to the inbox — and a note carries everything Backlog.md lets a draft hold: description, labels, assignee, priority, type, milestone and acceptance criteria. Only a status is out of reach, because a draft's status is Draft, and the form says so rather than dropping it silently. Clicking a note opens it in the task panel to be read whole; promoting one opens the task it became, so a note can be finished in a single pass.

A note opens in the task panel to be read whole, and nothing more: every editing control there writes through `task edit`, which refuses a `DRAFT-` id, so a note gets the two things that do work — promote it, or rewrite it in the inbox — and the panel says why. Promoting from there opens the resulting task, editable.

Backlog.md has no `draft edit` — `task edit` refuses a `DRAFT-` id outright — so rewriting a note captures a new one and archives the old, which is also the only way to move a draft between projects. Everything above survives that; the capture date does not, and the interface says so rather than hiding it.

A task's relationships are editable in the panel: what it depends on, the references and documents it points at, and the files it touched. Dependencies resolve inside their own project — ids collide freely across projects — so one that does not resolve is refused with the reason before anything is written. The files-touched list is the one exception to editing: Backlog.md can set it and has no command that empties it, so the interface says so instead of offering a button that does nothing. Manual order is dragging: a card dropped within its column takes the ordinal between its new neighbours, and where the gaps have run out the column is restacked at multiples of 1000, which is the allocation Backlog.md uses itself.

Some changes are only worth making in bulk: giving a milestone to everything that shares a label, retiring a label, moving a set of tasks to a status. The **list** ticks tasks — shift extends a range, the box in the header takes everything shown — and one form then describes the change once: status, priority, milestone, labels to add and labels to remove. Every write still goes through the CLI one task at a time, because that is the only writer there is, so the result says how many took the change and names every one that did not, in the words the CLI gave.

Nothing is offered that half the selection would reject: status and priority are the intersection of what every chosen project configures. A milestone is not offered across projects at all, and that one is measured rather than cautious — `task edit -m` accepts an id the project does not have, writes it into the file and reports success, so the milestone list then shows an entry with no file behind it. Across a selection that would plant a dangling reference in every project but one, silently.

Comments are a conversation rather than a field, and the panel writes them as well as reads them. They are signed with whatever you set in Preferences, or with the name git already knows — and with neither, written unsigned, which is a state the format has rather than a name to invent.

A task ends somewhere too. When one reaches its project's own last status it can be filed into `completed/`, where Backlog.md keeps finished work; any task can be archived, which is a soft delete rather than a delete; and any task can be sent back to the inbox as a note. Archived and completed tasks are off the board, the list, search and the figures — asking for nothing in particular means asking for the live ones.

Where Backlog.md has a relationship, Muster shows it without inventing a shape for it. Subtasks are the case that tempts a redesign: a parent card says how many of its subtasks are finished and a subtask card says whose it is, but the board stays flat and the links live in the task panel. Nesting cards inside cards would cost more legibility across nine projects than the relationship is worth.

Everything above, screen by screen, is in [Using Muster](backlog/docs/doc-7%20-%20Using-Muster.md).

## How it relates to the other tools

Fairly, and with the differences measured rather than asserted.

**[Backlog.md](https://github.com/MrLesk/Backlog.md) itself** is the format and the CLI, and Muster is a view onto it — not a replacement and not a fork. It reads that markdown directly and writes only through the `backlog` CLI, so there is exactly one writer and it is not this one. It adds no field, no label convention and no sidecar file of its own: if Backlog.md does not support something natively, Muster does not store it.

**`backlog browser`** is the CLI's own web interface, and it is good at what it does. It serves the project it is run in: started in one of the five projects above, its API returns that project's tasks and none of the others. It also defaults to port 6420 in every project, so a second one asks for another port. That is not a flaw — it is a single-project tool being a single-project tool, and if one repository is all you have, it is the simpler answer and needs nothing installed beyond the CLI you already have.

**The [VSCode extension](https://github.com/ysamlan/vscode-backlog-md)** puts the same backlog inside the editor, which is where a lot of this work actually happens, and it switches between backlog folders rather than combining them into one view.

Muster is worth having when the answer to _what is happening across everything I work on_ has to come from more than one repository at a time. With a single project it is a heavier way to get less.

## What it is not

- **Not a replacement for Backlog.md.** It is a view onto it.
- **Not a new format.** Muster reads Backlog.md markdown directly and writes only through the `backlog` CLI. It adds no field, no label convention and no sidecar file of its own — if Backlog.md does not support something natively, Muster does not store it. That is a hard rule, not an aspiration; it is why the review-budget planner from the first draft of the specification no longer exists.
- **Not a server, a team tool, or a sync service.** One machine, one person, no accounts, no network.
- **Not a time tracker**, and not an estimation tool.

## Install

Every release carries built packages for **Linux on x86_64**, produced by CI from the release commit rather than from anybody's machine. Pick the one your distribution speaks; they all install the same two programs, `muster` and `muster-mcp`.

```sh
# Debian, Ubuntu and derivatives
sudo apt install ./muster_<version>_amd64.deb

# Fedora, RHEL, openSUSE
sudo dnf install ./muster-<version>.x86_64.rpm

# Arch and derivatives
sudo pacman -U ./muster-<version>-x86_64.pkg.tar.zst
```

Or take the AppImage, which needs no package manager and no root:

```sh
chmod +x muster-<version>-x86_64.AppImage
./muster-<version>-x86_64.AppImage
```

Upgrading is the same command with the newer file; the packages replace the installed version in place. Your projects are not touched by any of it — Muster stores nothing but its own registry and settings under `~/.config/muster`, and the tasks themselves live in your repositories, written only through the `backlog` CLI.

### Verify what you downloaded

Every release publishes `SHA256SUMS` beside the artefacts. From the folder you downloaded into:

```sh
sha256sum --ignore-missing -c SHA256SUMS
```

`--ignore-missing` because the file lists every artefact of the release and you will have downloaded one or two of them. Without it, the files you did not take are reported as failures.

### The MCP server on its own

`muster-mcp` is what an agent client spawns, and the packages install it. It is also published as a separate file, `muster-mcp-<version>-linux-amd64`, because an agent may run somewhere the desktop application cannot be installed at all — inside a Flatpak sandbox, inside a container, or on a machine with no desktop. It is statically linked and depends on nothing, so it runs wherever it is put:

```sh
chmod +x muster-mcp-<version>-linux-amd64
mv muster-mcp-<version>-linux-amd64 ~/.local/bin/muster-mcp
```

See [Talking to an agent](#talking-to-an-agent) for what to do with it.

### What is not built

**macOS and Windows.** Muster is a Wails application and the scaffolding for both is in the repository, but nothing builds them, nothing tests them, and no release carries them. Treat them as unknown rather than as unsupported.

**Linux on arm64.** Same: it may well build, and nobody has.

Building from source is covered below, and works anywhere the prerequisites do.

## Prerequisites

| Requirement                                            | Version used                      |
| :----------------------------------------------------- | :-------------------------------- |
| [Go](https://go.dev)                                   | 1.25 or newer                     |
| [Node.js](https://nodejs.org)                          | 24 or newer                       |
| [pnpm](https://pnpm.io)                                | 11 or newer                       |
| [Wails v3 CLI](https://v3.wails.io)                    | v3.0.0-beta.8                     |
| [golangci-lint](https://golangci-lint.run)             | v2.13.1 (needed by the git hooks) |
| [Backlog.md CLI](https://github.com/MrLesk/Backlog.md) | 1.48.0 minimum, 1.50.1 recommended (at runtime, not to build) |

Muster looks for the CLI on PATH first and then in the places package managers install it — pnpm, npm, bun, cargo, `~/.local/bin` — because an application started from a desktop launcher does not inherit a shell's PATH, and neither does the CLI it runs: the binary pnpm installs is a shell script ending in `exec node`, and node is usually installed the same way. Both are handled.

If none of that finds it, it asks your login shell — `$SHELL -lic 'command -v backlog'` — which reads the profile where the PATH entry was written in the first place. That costs about a second and a half, so it is the last thing tried and the answer is kept for the rest of the session; it is also what makes the search work for an install location nobody thought to list. If even the shell does not know, the message names every directory that was searched and says the shell was asked, and Preferences takes an explicit path.

1.48.0 is the floor because that is the version the [format contract](<backlog/docs/doc-3 - Backlog.md-Format-Contract.md>) was measured against; 1.50.1 writes byte-identical files but reports two failures that 1.48.0 swallows, so it is what the author runs. Muster never relies on an exit code for a write whose result it can check.

The Wails CLI installs itself:

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.8
```

`golangci-lint` is downloaded rather than built. `go install` compiles it with the Go you already have, and it asks for a newer one than this project does — v2.13.1 needs Go 1.26, Muster builds on 1.25 — so on the minimum Go in the table above that command fails before it starts, with `requires go >= 1.26.0`. Its own installer fetches the release binary and needs no Go at all:

```sh
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v2.13.1/install.sh \
  | sh -s -- -b "$(go env GOPATH)/bin" v2.13.1
```

On Go 1.26 or newer, `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1` works as well. CI takes the binary, so nothing there depends on which Go it was built with.

The Wails CLI is itself a Wails program, so building it runs `pkg-config` for GTK 4 and WebKitGTK 6.0. On a machine that has GTK 3 instead — Ubuntu 24.04 among them — that first line fails before anything of this project is involved, complaining that `gtk4` is not in the pkg-config search path. Install it with the tag instead:

```sh
go install -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.8
```

Only the CLI needs telling. Muster's own build detects what the machine has and picks the tag itself.

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

A window opens. On first run there is no registry, so it offers to add a project: point it at any folder — one that already has a `backlog/` directory is registered as it is, and one that does not can be initialised from there. Everything after that is the board over whatever you registered.

Day to day:

```sh
wails3 task run       # build and run
wails3 task dev       # run with frontend hot reload
```

The application's icon is `build/appicon.svg`. Every other size and format comes from it — the packaged PNG, the one embedded in the binary, the Windows `.ico`, the macOS `.icns` — so the mark cannot differ between platforms:

```sh
node build/icon/render.mjs          # needs a global Playwright; see the script
wails3 task common:generate:icons
```

Linux packages, if you want them rather than the bare binary:

```sh
wails3 task package   # .deb, .rpm, .pkg.tar.zst and an AppImage into bin/
```

Before trusting a build, walk the [smoke checklist](backlog/docs/doc-5%20-%20v0.1-smoke-checklist.md). It takes a few minutes and every step in it has failed at least once.

## Talking to an agent

**`muster-mcp`** is a [Model Context Protocol](https://modelcontextprotocol.io) server over stdio. It draws no window and needs none — a client spawns it and talks over a pipe, whether or not the desktop application is open.

It is a separate binary, and it has to be. The desktop application links a browser engine at load time, so `muster mcp` — which builds no window either — still cannot start inside a Flatpak sandbox, a container, or on a machine with no desktop, which is where agents run. `muster-mcp` links none of that: it is a static binary with no dynamic dependencies at all, and it ships beside the application in the package. `muster mcp` remains for a machine that has the desktop libraries anyway.

Backlog.md already ships its own MCP server, and it is per-project: an agent gets one repository at a time. This one answers across every project you have registered, which is the only reason it exists.

Preferences has a list of the AI clients this machine has, and connects one in a click: it runs that client's own command, or writes that client's own configuration file. It shows exactly what it will run or write before doing either, keeps a backup of any file it changes, and disconnects by the same button. Nine clients are described in a data file — `internal/agents/agents.json` — because clients change their syntax more often than this ships; `MUSTER_AGENTS_FILE` points at a replacement without a rebuild.

A client that runs in a sandbox needs two things to line up, and both are now handled or explained.

The registry is found even when `XDG_CONFIG_HOME` is redirected — inside Obsidian's Flatpak it points at `~/.var/app/md.obsidian.Obsidian/config`, which has never held a registry, and the server used to answer "no projects" rather than "no registry". It now falls back to `~/.config/muster/projects.yml`, says which file it looked for when there is none, and takes `MUSTER_REGISTRY` as an outright answer:

```sh
claude mcp add --scope user muster -e MUSTER_REGISTRY=~/.config/muster/projects.yml -- /usr/local/bin/muster-mcp
```

The path to the binary depends on which side of a boundary each program is on, and the two cases are mirror images.

**Muster in a container, the client outside it.** A distrobox or toolbox container shares your home directory, so the file the connector edits is the host's while `/usr/local/bin/muster-mcp` means nothing there — a real path, on the wrong filesystem. The connector detects the container and registers the copy in `~/.local/bin` instead, which both sides can see, and the plan says so before writing anything. If there is no copy there it refuses and tells you to make one:

```sh
distrobox-export --bin /usr/local/bin/muster-mcp
```

That wrapper runs the binary through the container from the host, hops back out from a different container, and execs it directly from its own — so one registered path works from anywhere.

**The client in a sandbox, Muster outside it.** This one cannot be solved from Muster's side: a Flatpak sandbox cannot see `/usr/local/bin` at all, so a client running inside one has to reach the host itself — `flatpak-spawn --host /usr/local/bin/muster-mcp` — or be given a copy inside the sandbox.

The rest of this section is what that button does, for anyone who would rather do it themselves.

Point a client at the binary:

```json
{
  "mcpServers": {
    "muster": { "command": "/usr/local/bin/muster-mcp" }
  }
}
```

For Claude Code, `claude mcp add --scope user muster -- /usr/local/bin/muster-mcp` does the same thing. No arguments: a client should not have to know a subcommand. Use the full path unless `muster-mcp` is on the PATH your client will have — which, for anything started from a desktop launcher, it very likely is not.

| Tool | Answers |
| :-- | :-- |
| `list_projects` | every registered project, its task count by status, and the statuses, priorities and types it declares |
| `list_tasks` | tasks across all projects or one, with the board's filters |
| `get_task` | one task in full: body, criteria, plan, notes, references |
| `search` | text across tasks, drafts, documents, decisions and milestones |
| `list_milestones` | milestones with their progress |
| `list_entities` | documents, decisions, drafts or milestones, with their bodies |
| `create_task` | writes a task, or captures a note into the inbox |
| `set_field` | status, priority, assignee or milestone, one task at a time |
| `set_label` | adds or removes a label |
| `set_section` | replaces a description, plan or notes |

Start with `list_projects`: ids and statuses are per-project, and nothing else makes sense without knowing which projects exist.

The server says where its boundary is, so an agent does not have to guess. It sends an instruction at connection — use this to see across projects, not as a substitute for the `backlog` CLI when you are working inside one, where that CLI is the authority and has the whole command surface. Each write tool repeats it, because a tool description is read on its own. And every tool says whether it changes anything, so a client can stop asking permission for questions and keep asking for changes.

Every write goes through the `backlog` CLI, exactly as the interface does — there is still one writer. Every tool that names a project resolves it against the registry and refuses anything else, so an agent cannot reach a folder you have not registered; asking for one lists what is registered instead. Reads work even when the CLI is missing, since an agent asking what is in flight has no use for it.

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

## Continuous integration

The pipeline runs the same three commands as the section above — `wails3 task lint`, `test`, `build` — on every pull request and on the default branch, rather than a separate list that can drift from them. It also installs the Backlog.md CLI and refuses to let the tests that need it skip: thirty-eight of them exercise real writes against real projects, and without the CLI they would quietly pass over the most valuable half of the suite.

Its **time budget is ten minutes** with the caches warm. Measured on one machine, the three commands themselves take about a minute together — fourteen seconds to lint, forty to test from a cold test cache, two to build incrementally — so most of a run is setup: the GTK 3 packages, the Go modules, the pnpm store, and two Go tools compiled from source. All four are cached; the tools are keyed on the versions pinned in the workflow, which are the versions in the table above. A run that consistently exceeds the budget means a cache is not being hit, and that is the thing to look at first.

## Versions and releases

`muster --version` answers, and so does Preferences, which shows the same number beside the Backlog.md CLI's and copies both for a bug report.

The version is written in exactly one place, `build/config.yml`. The build stamps the binary from it, the `.deb` takes its version from it, and [release-please](https://github.com/googleapis/release-please) bumps it: it reads the Conventional Commits on the default branch, keeps a release pull request open with the next version and the changelog entries it implies, and merging that pull request cuts the release. A build that was never stamped calls itself `dev` rather than claiming a number nobody released.

A second job then packages that same release commit, refuses to go on unless every binary reports the version being released, and attaches the artefacts with their checksums. The release itself is public from the moment it is cut, so for the few minutes that packaging takes it carries notes and no downloads.

Cutting it as a draft and publishing it afterwards would close that window, and it was done that way once. It does not work: a draft records a tag name but GitHub creates the tag only when the release is published, and release-please finds the previous release by its tag — so in the same run it decides nothing has ever shipped and opens a pull request for the next version with the whole history in its changelog. That is what happened to v1.0.0.

Both halves live in one workflow rather than two, and that is not tidiness: a release created with `GITHUB_TOKEN` raises no `release` event and its tag raises no `push` event, so a workflow keyed on either would sit there and never run.

What a version number promises while the major version is still zero — which releases may break something, and what counts as breaking — is written at the top of [CHANGELOG.md](CHANGELOG.md).

## Git hooks

Installed by `pnpm install`. They format and lint staged files before a commit, check the commit message against the convention, and run tests before a push. See [CONTRIBUTING.md](CONTRIBUTING.md).

## Lint and format

One command covers Go and the frontend:

```sh
wails3 task lint      # golangci-lint, ESLint, Prettier, svelte-check
wails3 task lint:fix  # apply every available autofix
```

`golangci-lint` is the only tool not installed by the steps above; [Prerequisites](#prerequisites) has the command.

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
