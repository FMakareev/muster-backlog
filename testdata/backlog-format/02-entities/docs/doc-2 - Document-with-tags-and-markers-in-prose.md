---
id: doc-2
title: Document with tags and markers quoted in prose
type: specification
created_date: '2026-08-15 16:52'
updated_date: '2026-08-26 15:30'
tags:
  - seo
  - audit
---
Documents have no section markers of their own: everything after the frontmatter is free
markdown. That matters, because a document may quote task markers such as
`<!-- SECTION:DESCRIPTION:BEGIN -->` or `<!-- AC:BEGIN -->` inside prose or a fenced block:

```markdown
## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 This is an example, not a real criterion
<!-- AC:END -->
```

A marker scanner that runs over every markdown file in a project instead of only over the task
directories will mis-parse this document.
