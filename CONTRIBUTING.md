# Contributing to Muster

## Commit messages

This repository uses [Conventional Commits](https://www.conventionalcommits.org/). Messages are a build input, not a style preference: `release-please` derives the version bump and the `CHANGELOG.md` entries directly from them, so a message outside the vocabulary below either lands in the wrong release section or in none at all.

```
<type>(<scope>): <subject>
```

`commitlint` enforces this, and the `commit-msg` hook runs it before a commit is created.

### Types

| Type | Use for | Release effect |
| :-- | :-- | :-- |
| `feat` | a user-visible capability | minor bump |
| `fix` | a user-visible defect repaired | patch bump |
| `perf` | a change made for speed or resource use | patch bump |
| `refactor` | behaviour preserved, structure changed | none |
| `docs` | documentation only | none |
| `test` | tests only | none |
| `build` | build system, toolchain or dependencies | none |
| `ci` | continuous integration configuration | none |
| `chore` | repository housekeeping with no product effect | none |
| `revert` | undoing a previous commit | depends on what is reverted |

### Scopes

A scope names the part of the product the change belongs to. It is optional — a repository-wide chore has no meaningful scope — but when present it must be one of:

`app`, `parser`, `store`, `watcher`, `cli`, `board`, `list`, `task`, `inbox`, `docs-view`, `analytics`, `search`, `projects`, `mcp`, `ui`, `deps`, `release`

The vocabulary is closed on purpose. If a change genuinely does not fit, add the scope to `commitlint.config.js` in the same commit rather than reaching for an approximation.

### Breaking changes

Mark a breaking change with `!` after the scope **and** a `BREAKING CHANGE:` footer explaining what breaks and what to do about it:

```
feat(cli)!: require Backlog.md 1.48 or newer

BREAKING CHANGE: older CLI versions lack the draft promote command, so
drafts captured by the inbox cannot be turned into tasks. Upgrade with
`bun add -g backlog.md` before updating Muster.
```

Before 1.0 a breaking change bumps the minor version rather than the major one. After 1.0 it bumps the major version — see the versioning policy.

### Examples

```
feat(board): group cards by their originating project
fix(parser): keep frontmatter titles containing a colon intact
docs: explain the Ubuntu 24.04 GTK 3 build path
build(deps): pin svelte to 5.56.10 for SVAR compatibility
```

## How work is tracked

**The work is not tracked in GitHub Issues.** It is tracked with [Backlog.md](https://github.com/MrLesk/Backlog.md) in the [`backlog/`](backlog/) directory of this repository — tasks are markdown files, versioned alongside the code they describe. Muster itself is a tool for working with exactly this format, so the project uses it on itself.

Issues are the way in from outside, not the list of work: a [bug report](../../issues/new?template=bug_report.yml) or a [feature request](../../issues/new?template=feature_request.yml) is read and, if it is going to be done, becomes a task in `backlog/` — which is then what the pull request references. Questions go to Discussions rather than Issues, because an issue tracker that fills with answered questions stops being a list of anything.

```sh
backlog task list --plain          # what is open
backlog task view TASK-1 --plain   # one task in full
backlog board                      # the kanban board
backlog overview                   # counts, blocked tasks, stale tasks
```

Every task carries acceptance criteria. Those are the definition of done — not a suggestion, and not something to reinterpret while implementing.

Never edit files under `backlog/` by hand. The CLI owns their metadata, filenames and relationships; a hand edit will look fine and then diverge. The single documented exception is the body of a decision record, which no CLI path can write — see [the conventions document](backlog/docs/doc-4%20-%20Documentation-and-decision-conventions.md).

If you want to work on something, say so on the task before starting, so two people do not implement the same acceptance criteria twice.

## Setting up

See the [README](README.md) for the full prerequisite table. In short: Go 1.25+, Node 24+, pnpm 11+, the Wails v3 CLI, and `golangci-lint`.

```sh
git clone <this repository>
cd muster-backlog
pnpm install                 # installs the workspace and the git hooks
wails3 task build
```

On Ubuntu 24.04 LTS and anything else without WebKitGTK 6.0, install the GTK 3 development packages instead; the build detects which you have and picks the tag itself. The README lists both sets.

## Before you push

```sh
wails3 task lint             # golangci-lint, ESLint, Prettier, svelte-check
wails3 task lint:fix         # apply every available autofix
go test -tags gtk3 ./...     # Go tests
```

The hooks run most of this for you, but running it yourself is faster than finding out from a red pull request.

## Git hooks

`pnpm install` installs them through the `prepare` script. To install them by hand:

```sh
lefthook install
```

| Hook | What runs | Typical cost |
| :-- | :-- | :-- |
| `pre-commit` | format and lint the staged files only | under 3 seconds |
| `commit-msg` | `commitlint` against the message | under 1 second |
| `pre-push` | `go test` and `svelte-check` | under 3 seconds |

`pre-commit` restages what the formatters fix, so a commit is never left half-formatted. The budget above is the point of the design: a hook slow enough to provoke `--no-verify` is a hook that does not run. If `pre-commit` ever exceeds it, move the offending check to CI rather than to a flag.

## Pull requests

1. Branch from the default branch. Name the branch after the task: `task-21-multi-project-board`.
2. Keep the change to one task. If you find unrelated work, note it on a task rather than folding it in — a reviewer who has to hold two changes in their head reviews neither well.
3. Reference the task ID in the description, and say which acceptance criteria the change satisfies.
4. Explain how you verified it. "Tests pass" is not verification; naming what you ran and what it showed is.
5. Make sure `wails3 task lint` and the tests are green before asking for review.

Changes that add a field, a label convention or a sidecar file to the Backlog.md format will be declined regardless of how useful they are. That constraint is the point of the project, not an oversight — see [decision-3](backlog/decisions/decision-3%20-%20Read-Backlog.md-markdown-directly-write-only-through-the-CLI.md).

## Reporting problems

Bugs and ideas belong on a Backlog.md task in this repository, not in a GitHub issue. Security problems go through the private channel described in [SECURITY.md](SECURITY.md) — never in public.

## Code of conduct

Participation is governed by the [Code of Conduct](CODE_OF_CONDUCT.md).
