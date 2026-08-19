---
name: feedback_verify_run_rewrites_edge_json
description: Running the sim (incl. stop-checks and the pre-push hook) rewrites topology edge JSONs with float noise; never `git add -A` after verifying
metadata:
  type: feedback
---

STALE past-tense of a since-superseded mechanism: drags (including this file's own
`deltaPolarPhi`/`deltaPolarTheta` churn) moved to index-only persistence in
`nodes/<id>/drag/edges/<label>.json` (gitignored), and the base delta itself
(`nodes/<id>/edges/<label>.json`'s `deltaIndexR/Phi/Theta`) is now an authored integer
step count, not a float the run recomputes. Kept as historical record of the underlying
lesson below, which still applies to any owner-writes-its-own-file churn.

Any run of the binary — including `scripts/stop-checks.sh` and the `.githooks/pre-push`
hook that wraps it — used to leave `topology/nodes/*/edges/*.json` modified. The edgeMover
persisted its polar delta on geometry recompute, and the round-trip changed the last
digit or two of `deltaPolarPhi`/`deltaPolarTheta`. Nothing moved; it was float noise.

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
`scripts/new-task.sh` will refuse to start the next task until it is discarded.

Related: [[feedback_verify_subagent_commits]], [[feedback_check_the_signal_the_check_emits]].
