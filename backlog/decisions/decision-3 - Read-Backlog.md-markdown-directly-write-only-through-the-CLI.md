---
id: decision-3
title: 'Read Backlog.md markdown directly, write only through the CLI'
date: '2026-08-26 16:04'
status: accepted
---
## Context

Muster does not own the format it displays. Backlog.md defines it, and the founding rule of this project is that no field, label convention or sidecar file of ours is ever added to it.

That leaves the question of mechanism in both directions. The VSCode extension the author already uses reads and writes the markdown directly and gets away with it, so writing directly is demonstrably possible.

What TASK-14 measured across 1021 entity files on CLI 1.48.0 bears on this:

- Task bodies are marked up with seven literal HTML-comment section markers in two naming conventions, all optional.
- `ordinal` carries manual ordering; ids are unique only within one directory of one project and are reused after archiving.
- Enum values are not normalised - `priority` is stored lowercase against a capitalised config.
- The 1.48.0 serialiser can emit three keys that appear nowhere in the corpus, and still ships a migration for an older acceptance-criteria format.
- There is a `backlog doctor` command specifically because id generation is not trivial.

## Decision

**Reads go directly to the markdown.** A parser of our own over `tasks/`, `drafts/`, `milestones/`, `docs/`, `decisions/` and `completed/`, everything held in memory. The corpus is a thousand files and parses in milliseconds; no database is warranted.

**Writes go only through the `backlog` CLI.** `task create|edit`, `draft create|promote`, `init`. Every mutation in the application funnels through a single adapter, and the board settles on what a rescan confirms rather than on what the UI assumed.

No second writer is ever added. If the CLI cannot express something, that is a limit to work within or to raise upstream, not a reason to write the file ourselves.

## Consequences

- Muster can never corrupt a backlog in a way the CLI would not have produced itself. That is the whole point: these are the author's real repositories, edited concurrently by agents.
- The format stays entirely Backlog.md's to define. A CLI release that changes the on-disk shape changes what we read, but never leaves us writing a stale dialect.
- Writes cost a process spawn. Acceptable for the interaction rates involved - moving a card, setting a priority - and the reason drag-and-drop settles after a rescan instead of optimistically.
- The application gains a runtime dependency that must be located, version-checked and reported clearly when missing.
- Reads carry the burden instead: the parser must tolerate every variation the corpus contains, because it cannot delegate. TASK-14 catalogues them and `testdata/backlog-format/` pins them as fixtures.
- Known gap: `backlog decision create` writes only frontmatter and an empty skeleton, and neither the CLI nor the MCP server can fill the body. These decision records were therefore written by editing the files after the CLI created them. This is the one place the write rule is not honoured, because no CLI path exists.
