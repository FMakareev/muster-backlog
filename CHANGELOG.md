# Changelog

Every entry here is derived from the commit history rather than written by
hand: [release-please](https://github.com/googleapis/release-please) reads the
[Conventional Commits](https://www.conventionalcommits.org/) on the default
branch, opens a release pull request, and writes everything below the
Unreleased heading when that pull request is merged. Edit a commit message,
not this file.

The format is [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
versioning is [Semantic Versioning](https://semver.org/spec/v2.0.0.html), read
the way the section below says while the major version is still zero.

## Versioning before 1.0

While the version starts with `0.`, SemVer allows anything to change in any
release. That is not much of a promise, so here is the one this project keeps
instead.

- **A minor bump — `0.1.0` to `0.2.0` — may break something.** It is where a
  breaking change goes while the major version is zero. What counts as
  breaking is listed below, and every one of them appears in the release notes
  under its own heading.
- **A patch bump — `0.1.0` to `0.1.1` — does not.** Fixes and additions that
  take nothing away.

Breaking, before 1.0, means any of these:

- The registry file `projects.yml` stops being readable by the previous
  version, or a field in it changes meaning.
- The preferences file `settings.yml` loses a setting or changes what one
  does.
- The minimum Backlog.md CLI version goes up. Muster writes only through that
  CLI, so requiring a newer one is a requirement on the person running it.
- A command-line interface changes: the `mcp` subcommand, or the tools it
  serves over the Model Context Protocol, since an agent's configuration
  points at them.

Two things are deliberately **not** breaking changes, at any version:

- Anything about the Backlog.md files themselves. Muster reads and writes that
  format and does not own it; it adds no field of its own, so there is nothing
  there for it to break. See the guarantee in the README.
- The layout of the interface. Where a button lives is not an interface anyone
  can depend on programmatically.

At 1.0 the first list becomes a major-version promise instead, which is what
1.0 is for.

## Unreleased

Nothing released yet. Everything so far is on the way to the first tag; see
[the roadmap](backlog/docs/doc-2%20-%20Roadmap-to-1.0.md) for what is still
missing.
