<!--
Work in this project is tracked with Backlog.md in backlog/, not with GitHub
Issues. A pull request implements one task; naming it is how a reviewer finds
the acceptance criteria they are supposed to be checking against.
-->

## Task

<!-- e.g. TASK-42. `backlog task view TASK-42 --plain` shows what it asks for. -->

TASK-

## What this does

<!-- What changes for someone using it. Not a list of the files you touched. -->

## Which acceptance criteria it satisfies

<!--
Name them by number, and say which are left. A pull request that satisfies
three of five criteria is fine; one that says so is much better than one that
leaves the reviewer to work it out.
-->

- [ ] #1
- [ ] #2

## How it was verified

<!--
Name what you ran and what it showed. "Tests pass" is not verification — the
question a reviewer is asking is what would have caught this being wrong.
Measured behaviour of the backlog CLI belongs here too: this project's rule is
that CLI behaviour is measured before it is designed against, not assumed.
-->

## Checklist

- [ ] `wails3 task lint` is clean
- [ ] `wails3 task test` is green
- [ ] The commits follow [Conventional Commits](../CONTRIBUTING.md#commit-messages) and use a scope from the closed list
- [ ] Nothing here stores anything Backlog.md does not support natively — no new field, no sidecar file, no label convention standing in for one
- [ ] Every write still goes through the `backlog` CLI
- [ ] User-visible behaviour is reflected in the README
- [ ] The task's own notes record anything measured or decided that the code does not say for itself
