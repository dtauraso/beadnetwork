---
paths:
  - "docs/planning/**"
---

# Planning docs are branch-local

Planning docs (anything under `docs/planning/`) are authored on the
task branch where the work happens and do not ride the merge to main. Each new planning
doc starts with frontmatter naming its originating branch:

```
---
branch: task/<short-name>
---
```

Before merging a task branch, run `scripts/strip-branch-local-docs.sh task/<branch>` to
remove all docs tagged with that branch. The script is the source of truth — no judgment
per file required at merge time.

**A planning doc without a `branch:` tag is a bug, not a grandfathered exception.** The rule
used to end "this rule is forward-only; existing untagged docs stay until individually
judged", and what that bought was 26 v0-era pages sitting on main describing a model the
code had left — including pages named `was.html`. Nothing ever judged them, because
"until individually judged" names no one and no moment. They were deleted instead, and the
grandfather clause with them.

So: `docs/planning/` on `main` is EMPTY between changes. A plan lives on its branch, tagged,
and leaves with the merge. If you find an untagged doc there, it escaped a merge — delete
it, the same as the script would have.
