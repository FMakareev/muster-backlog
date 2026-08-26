---
id: doc-2
title: Roadmap to 1.0
type: guide
created_date: '2026-08-26 15:04'
updated_date: '2026-08-26 19:45'
---
Five milestones from an empty repository to a public 1.0, 45 tasks. Each ends where stopping would still leave something usable. This document explains the ordering; the `backlog` CLI holds the tasks themselves.

Scope source: [Muster Specification v0.1](doc-1%20-%20Muster-Specification-v0.1.md).

## m-0 — Repository and OSS foundation (13 tasks)

Everything that must be true before the first public push, plus the buildable skeleton the product is written into.

The skeleton (TASK-1) comes first because linters, hooks and CI have nothing to run without it. The Wails beta decision (TASK-2) is taken deliberately as an ADR. Conventional Commits (TASK-5) precede release automation (TASK-8), because release-please derives versions and the changelog from commit messages — message shape is a build input, not a style preference. The public push (TASK-13) is last: history becomes permanent at that moment.

Exit condition: a stranger can clone, build, run the linters and open a pull request that CI judges.

## m-1 — MVP v0.1: multi-project board (16 tasks)

The read surface plus the minimum of writing. Specification section 7.

The chain is linear by necessity — format contract, parser, store, watcher, bridge, shell, board — because each layer is meaningless without the one below it. Two tasks sit slightly outside that chain and matter more than their size suggests:

- **TASK-27, unified board columns.** Statuses are per-project configuration and projects will not agree. Taking the union rather than imposing a common set is what keeps the promise of never editing another project's `config.yml`.
- **TASK-24, the CLI write adapter.** The single choke point for every write. Everything that mutates a task in any later milestone goes through it, which is what keeps the format entirely Backlog.md's to define.

TASK-29 is the real end of the milestone: a written verdict on whether a standalone multi-project application beats the per-repository tooling already in use. The VSCode extension covers single-repo work well, so that is the honest comparison. It gates m-2 onward.

Exit condition: the MVP is the daily tool over the real projects, with the verdict written down.

## m-2 — Daily use: editing, views and filters (11 tasks)

The rest of an ordinary task manager, all of it cross-project.

This milestone grew after the MVP was first used. The [verdict](doc-6%20-%20MVP-v0.1-verdict.md) is that the multi-project view works and performs, but that the MVP is a viewer which can move cards rather than a task manager: it cannot create a task, cannot edit a body, has no search, no filters, and shows milestones as bare ids. Those are now here, because they are what stands between this and being the tool actually reached for.

The list view (TASK-50) is the mode for scanning many tasks where the board is the mode for moving a few. Search (TASK-51) becomes the main navigation once several projects are loaded. Analytics (TASK-52) mirrors the vocabulary `backlog overview` already defines from native data, but across every project and drillable. The documents viewer (TASK-53) makes the application the single place a project is read from, not just tracked in.

Work-in-progress limits (TASK-54) are what remains of the review-capacity idea from the first draft of the specification: a count of native data against a configured number, flagged and never enforced. No field of our own, no planner.

Exit condition: nothing about daily task work requires dropping into the CLI or into VSCode. The verdict is then revisited against the same question, with a real trial rather than a first impression.

## m-3 — Project onboarding and inbox (5 tasks)

Turning the tool from personal to installable, and closing the capture loop.

TASK-55 is the milestone's point: initialising a Backlog.md project in any folder from the interface. `backlog init` is fully non-interactive — every prompt has a flag — so this is a form over those flags, not an emulated dialogue. It also removes any need for a special "personal" project: a folder that is not a git repository is just a folder that gets `--no-git`.

Capture (TASK-26, TASK-41) and the inbox triage view (TASK-42) close the loop that drafts were designed for: capture stays cheap because drafts stay off the board, and the triage view is what stops the drafts directory from becoming the pile nobody opens.

Exit condition: a new project is added, initialised and captured into without leaving the application.

## m-4 — Packaging and 1.0 release (5 tasks)

What turns a working tool into something another person installs.

Long-session hardening (TASK-46) exercises what a short trial cannot: the application stays open all day while agents write into several repositories. The MCP server (TASK-45) is deliberately last and lowest priority — it is the one feature the application is complete without.

1.0 (TASK-49) is a promise about stability: the registry format and the CLI write contract stop changing without a major bump, and the commitment to add no field of our own to the Backlog.md format is stated as a project guarantee rather than an implementation detail.

Exit condition: a stranger installs a published artefact and it works.

## Rules that hold across milestones

- **No field of our own, ever.** If Backlog.md does not support it natively, Muster does not store it. This killed the review-cost labels and the review-budget planner in the first draft of this plan, and it will kill the next clever convention too.
- **All writes go through the `backlog` CLI.** Nothing in this roadmap adds a second writer.
- **Other projects' configuration is read, never written.** The board adapts to what projects declare instead of asking them to agree.
- Decisions are recorded as ADRs before the code that depends on them.
- Milestones after m-1 are provisional until TASK-29 answers the hypothesis.
