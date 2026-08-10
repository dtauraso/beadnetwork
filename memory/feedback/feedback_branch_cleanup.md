---
name: Delete task branches once merged
description: After a task branch fast-forwards or merges into main, delete it locally and on remote without re-asking
type: feedback
---

After a task branch lands on `main` (fast-forward or merge commit),
delete it locally and on remote in the same step. Don't ask again.

**Why:** matches CLAUDE.md's "avoid long-lived feature branches" posture
(the v0 `visual-editor` branch was the cautionary example). Commits,
screenshots, session-log entries, and memory files all reach `main`
via the merge, so an AI or human can grep the repo for any of them
without the branch ref. Named branch survival adds no recovery
value vs. the clutter cost.

**How to apply:** when a task branch is confirmed merged into `main`,
run `git branch -d <name>` and `git push origin --delete <name>`
without prompting. Still confirm before deleting an unmerged branch
or one with uncommitted divergence.
