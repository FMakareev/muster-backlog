---
some_other_key: value
---

Frontmatter that parses as YAML but carries no `id`. Intended handling: skip with a
diagnostic - `id` and `title` are the only two fields present in 100% of real entity files,
so their absence is the cheapest "this is not a Backlog entity" test.
