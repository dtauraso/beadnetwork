---
name: Verified work lands without per-merge sign-off
description: A task branch with EMPTY stop-checks output merges to main immediately; unmerged work is invisible to David because the editor opens the main checkout
type: feedback
---

David's instruction, 2026-07-29: "then change it so what I ask for happens?"

A task branch whose `bash scripts/stop-checks.sh` output is EMPTY merges to `main`
immediately, without asking. Force-pushing your own task branch after a rebase, and
deleting it once merged, are part of landing it — not separate permissions to request.

**Why:** the editor opens the MAIN CHECKOUT, which `check-main-checkout-on-main.sh` pins to
`main`. Work sitting on a task branch is therefore INVISIBLE to David no matter how
finished and verified it is. Asking "may I merge?" leaves him looking at old code — he
reported "I don't see a difference" while three verified branches sat unmerged, and asked
why nothing was landing. The ask does not protect him from anything; it just withholds the
result.

**Still requires sign-off:** force-pushing `main` or a branch you do not own, rewriting
published history, removing a dependency, deleting someone else's branch.

**After merging, say what to reload:** webview change → reopen the file; extension-host
change → "Developer: Reload Window" (see [[feedback_two_process_editor_reload]]).

Related: [[feedback_branch_cleanup]] (delete merged branches without re-asking, same
spirit), [[feedback_finish_calibrated_work]], [[feedback_no_deferrals]].
