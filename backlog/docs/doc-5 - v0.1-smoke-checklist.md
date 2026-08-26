---
id: doc-5
title: v0.1 smoke checklist
type: guide
created_date: '2026-08-26 19:14'
updated_date: '2026-08-26 19:16'
---
What to check before believing a build works. Nine steps, a few minutes, and every one of them has failed at least once during development.

Run it against a built artefact, not against `wails3 task dev`: the point is to test what someone else would install.

## Before you start

```sh
wails3 task build EXTRA_TAGS=gtk3   # drop EXTRA_TAGS where WebKitGTK 6.0 exists
./bin/muster
```

You need at least two Backlog.md projects registered, and they should **not** be projects you care about — steps 6 and 7 write to disk. Make throwaway ones:

```sh
mkdir -p /tmp/smoke/alpha /tmp/smoke/beta
(cd /tmp/smoke/alpha && backlog init Alpha --defaults --no-git --integration-mode none && backlog task create "Alpha task one")
(cd /tmp/smoke/beta  && backlog init Beta  --defaults --no-git --integration-mode none && backlog task create "Beta task one")
```

Give Beta a status Alpha does not have, so step 7 has something to refuse:

```sh
sed -i 's/^statuses: .*/statuses: ["To Do", "In Progress", "In Review", "Done"]/' /tmp/smoke/beta/backlog/config.yml
```

Point Muster at them by writing `~/.config/muster/projects.yml`.

## The checklist

**1. It opens.** The window appears and stays. No blank screen, no error dialog.

**2. Both projects are in the roll,** each with its own colour and a task count that matches what is on disk.

**3. The columns are the union.** Beta's extra status has a column; Alpha's cards are absent from it rather than the column being missing.

**4. Cards say where they came from.** Every card carries a project colour and its id. Press `g`: cards clump by project.

**5. The panel opens.** Click a card. Title, status, priority, type, milestone, dates and acceptance criteria with their state. `Esc` closes it and the board is where you left it.

A task the CLI has just created has an empty body, so use one with a real description to see the markdown, the acceptance-criteria checkboxes and the dependency links. One of this repository's own tasks does nicely.

**6. A status change reaches the disk.** Drag a card to another column, then read the file:

```sh
grep '^status:' "/tmp/smoke/alpha/backlog/tasks/task-1 - Alpha-task-one.md"
```

It must say the new status. Then press `]` on a focused card and check again — the keyboard path writes too.

**7. An impossible move is refused.** Drag an Alpha card into Beta's extra status. The card must stay where it was and a message must name the project, the status and what to do about it. Nothing on disk changes.

**8. Changes made outside show up.** With the window open, run:

```sh
(cd /tmp/smoke/alpha && backlog task edit TASK-1 --add-label smoke)
```

The label appears within a second or so, without touching the window. Remove it again and it disappears.

**9. Nothing was written that should not have been.**

```sh
git -C <any real project you registered> status --short
```

Muster writes tasks through the CLI and never touches a project's `config.yml`. If anything else changed, stop and find out why.

**10. Nothing reached the network.** Unplug it, or watch with `ss -tnp`, and use the application. Muster is local-first: fonts and icons are bundled and rendered markdown has its remote resources stripped. A dependency upgrade is the likely cause if this ever fails.

## What is not covered

The MVP has no list view, no search, no analytics, no documents viewer and no hotkey capture. Those arrive in later milestones and this checklist grows with them.
