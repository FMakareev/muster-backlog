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

## Tasks

Work is tracked with [Backlog.md](https://github.com/MrLesk/Backlog.md) in [`backlog/`](backlog/), not in GitHub Issues. Reference the task in the commit body or the pull request description.

```sh
backlog task list --plain
backlog task view TASK-1 --plain
backlog board
```

## Development

See the [README](README.md) for prerequisites, build, run and lint commands.

```sh
pnpm install          # installs the workspace and the git hooks
wails3 task build     # build
wails3 task lint      # every linter and formatter check
```

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
