---
paths:
  - "docs/planning/**"
---

# Planning docs are branch-local

Planning docs (anything under `docs/planning/` except `session-log.md`) are authored on the
task branch where the work happens and do not ride the merge to main. Each new planning
doc starts with frontmatter naming its originating branch:

```
---
branch: task/<short-name>
---
```

Before merging a task branch, run `tools/strip-branch-local-docs.sh task/<branch>` to
remove all docs tagged with that branch. The script is the source of truth — no judgment
per file required at merge time.

This rule is forward-only. Existing untagged docs stay until individually judged.

`session-log.md` is exempt — it is the durable history/friction record and rides to main.
