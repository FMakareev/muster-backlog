---
id: TASK-49
title: Make good on the 1.0 the tag already claims
status: To Do
assignee: []
created_date: '2026-08-26 15:03'
updated_date: '2026-08-28 18:52'
labels: []
milestone: m-4
dependencies:
  - TASK-47
  - TASK-48
  - TASK-46
  - TASK-44
priority: high
type: task
ordinal: 49000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
v1.0.0 shipped on 2026-08-28 without meaning to: release-please cuts a package's first release as 1.0.0 unless the configuration says otherwise, and ours did not. The number was kept rather than withdrawn, so the promise a major version makes is now in force and the work that was going to be done before making it has not been done.

Nothing here changes because of that except the order. Version 1.0 is a promise about stability: the registry format, the label conventions and the write path stop changing without a major bump. What is left is to state that promise where someone relying on it would look, and to say plainly what is known not to work.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Registry format and the CLI write contract are declared stable and documented as such
- [ ] #2 Post-1.0 breaking-change policy is written down
- [ ] #3 The commitment to add no field of our own to the Backlog.md format is stated as a project guarantee
- [ ] #4 CHANGELOG covers everything since the first pre-release
- [x] #5 The 1.0 tag, release and artefacts are published through the release automation
- [ ] #6 Known limitations and explicit non-goals are stated in the release notes
<!-- AC:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @claude
created: 2026-08-28 18:52
---
Retitled after v1.0.0 shipped by accident on 2026-08-28. The tag exists and its artefacts were published by the automation, which is AC 5, so that one is checked. Every other criterion is now a claim the released version already makes and the project has not yet backed up. The versioning promise itself was rewritten for post-1.0 under TASK-91.
---
<!-- COMMENTS:END -->
