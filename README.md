# Muster

A local-first desktop task manager over all your [Backlog.md](https://github.com/MrLesk/Backlog.md) projects at once.

> **Status: pre-alpha.** This is a project skeleton. It builds and opens a window with a placeholder board; it does not read your backlogs yet. See [the roadmap](backlog/docs/doc-2%20-%20Roadmap-to-1.0.md).

Muster reads Backlog.md markdown directly and writes only through the `backlog` CLI. It adds no fields, labels or sidecar files of its own — the format stays entirely Backlog.md's to define.

## Prerequisites

| Requirement | Version used |
| :-- | :-- |
| [Go](https://go.dev) | 1.25 or newer |
| [Node.js](https://nodejs.org) | 24 or newer |
| [pnpm](https://pnpm.io) | 11 or newer |
| [Wails v3 CLI](https://v3.wails.io) | v3.0.0-beta.8 |
| [Backlog.md CLI](https://github.com/MrLesk/Backlog.md) | 1.48.0 (runtime, not build) |

Install the Wails CLI with:

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.8
```

### Linux system libraries

The default Linux build targets **GTK 4** and **WebKitGTK 6.0**:

```sh
# Fedora
sudo dnf install gtk4-devel webkitgtk6.0-devel
# Debian / Ubuntu 24.10+
sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev
```

Distributions that do not ship WebKitGTK 6.0 — **Ubuntu 24.04 LTS** among them — can build against GTK 3 and WebKit2GTK 4.1 instead by passing the `gtk3` build tag:

```sh
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
wails3 task build EXTRA_TAGS=gtk3
```

## Build and run

```sh
wails3 task build     # build to bin/muster
wails3 task run       # build and run
wails3 task dev       # run with frontend hot reload
```

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

The [specification](backlog/docs/doc-1%20-%20Muster-Specification-v0.1.md) and the [roadmap](backlog/docs/doc-2%20-%20Roadmap-to-1.0.md) explain what is being built and in what order.

## Licence

MIT — see [LICENSE](LICENSE).
