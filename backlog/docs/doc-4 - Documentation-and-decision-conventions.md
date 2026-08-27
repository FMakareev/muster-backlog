---
id: doc-4
title: Documentation and decision conventions
type: guide
created_date: '2026-08-26 16:05'
updated_date: '2026-08-27 21:54'
---
Where documentation and decisions live in this repository, and how to add to them.

## There is no top-level `docs/`

Backlog.md already defines two entities for this, and the founding rule of this project is to use what the format provides rather than invent alongside it. A second documentation tree would mean two places to look, two things to keep in sync, and — since Muster's own documents viewer reads Backlog.md projects — a body of documentation the product itself could not display.

```
backlog/
  docs/         specifications, guides, research write-ups
  decisions/    architecture decision records
  milestones/   milestone definitions
  tasks/        the work itself
```

Everything else — `README.md`, `CONTRIBUTING.md`, `LICENSE` — stays at the repository root, because that is where GitHub and a newcomer look for it.

## Documents

For anything durable that is not a decision and not a task: specifications, guides, research findings, format contracts.

```sh
backlog doc create "Title" -t specification   # readme | guide | specification | other
backlog doc update doc-4 --content "$(cat draft.md)"
backlog doc list --plain
```

`doc update --content` replaces the whole body, so edit a draft in a scratch file and write it in one go.

Current documents:

| Document | What it is |
| :-- | :-- |
| [Muster Specification v0.1](doc-1%20-%20Muster-Specification-v0.1.md) | what is being built, and explicitly what is not |
| [Roadmap to 1.0](doc-2%20-%20Roadmap-to-1.0.md) | the five milestones and why they are ordered that way |
| [Backlog.md Format Contract](doc-3%20-%20Backlog.md-Format-Contract.md) | the on-disk format, measured against 1021 real files |
| Documentation and decision conventions | this document |
| [v0.1 smoke checklist](doc-5%20-%20v0.1-smoke-checklist.md) | what a build has to survive before it counts as working |
| [MVP v0.1 verdict](doc-6%20-%20MVP-v0.1-verdict.md) | what the first build did against the real corpus |
| [Using Muster](doc-7%20-%20Using-Muster.md) | the user guide: every screen, with screenshots |

## Decisions

An architecture decision record captures a choice that was not obvious, together with what it costs. Write one when a decision closes off alternatives, when it will be questioned later, or when the reasoning lives only in someone's head.

Do **not** write one for a choice that is reversible in an afternoon, or for something the code already states plainly.

```sh
backlog decision create "Title in the imperative" -s accepted
```

Numbering is `decision-N`, assigned by the CLI in creation order — not by topic, not by date. Structure is fixed by the template: **Context**, **Decision**, **Consequences**. Consequences is the section that earns the document: record what the choice costs and what it forecloses, not only what it enables.

Status is one of `proposed`, `accepted`, `rejected`, `superseded`. A decision is never edited to say something different once accepted — write a new one and mark the old superseded, so the reasoning at the time survives.

Current decisions:

| Decision | Status |
| :-- | :-- |
| [Build the desktop shell on Wails v3 beta](../decisions/decision-1%20-%20Build-the-desktop-shell-on-Wails-v3-beta.md) | accepted |
| [Licence MIT and automate releases with release-please](../decisions/decision-2%20-%20Licence-MIT-and-automate-releases-with-release-please.md) | accepted |
| [Read Backlog.md markdown directly, write only through the CLI](../decisions/decision-3%20-%20Read-Backlog.md-markdown-directly-write-only-through-the-CLI.md) | accepted |

### A known gap in the tooling

`backlog decision create` writes frontmatter and an empty `Context` / `Decision` / `Consequences` skeleton, and nothing in Backlog.md 1.48.0 can fill the body — there is no `decision update` in the CLI, and the MCP server exposes task, milestone and document tools but none for decisions.

So decision bodies are written by editing the file after the CLI has created it. This is the single place where this repository does not go through the CLI, and it is a tooling limitation rather than a preference. The frontmatter the CLI owns is never touched by hand.

Everything else — tasks, drafts, milestones, documents — goes through the CLI, always.
