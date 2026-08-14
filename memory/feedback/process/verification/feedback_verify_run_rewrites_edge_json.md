---
name: feedback_verify_run_rewrites_edge_json
description: Running the sim (incl. stop-checks and the pre-push hook) rewrites topology edge JSONs with float noise; never `git add -A` after verifying
metadata:
  type: feedback
---

Any run of the binary — including `scripts/stop-checks.sh` and the `.githooks/pre-push`
hook that wraps it — leaves `topology/nodes/*/edges/*.json` modified. The edgeMover
persists its polar delta on geometry recompute, and the round-trip changes the last
digit or two of `deltaPolarPhi`/`deltaPolarTheta`. Nothing moved; it is float noise.

**Why:** verification and persistence share the same live code path. The edgeMover
that recomputes geometry is the same one that owns
`topology/nodes/<source>/edges/<label>.json`, so exercising the network at all writes
those files. This is correct behaviour, not a bug to fix — the owner writes its own file
(`.claude/rules/persistence-ownership.md`).

**How to apply:** after running stop-checks, `git checkout -- topology/nodes` before
staging, and never `git add -A` on a tree that has been verified. It swept the churn
into two commits in one session, each needing a `--amend` + force-push to undo. Stage
the files the change actually touched, or discard the topology churn first. Note the
pre-push hook re-runs stop-checks, so the churn reappears in the working tree AFTER a
clean commit — the tree being dirty right after a successful push is expected, and
`tools/new-task.sh` will refuse to start the next task until it is discarded.

Related: [[feedback_verify_subagent_commits]], [[feedback_check_the_signal_the_check_emits]].
