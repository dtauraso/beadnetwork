#!/usr/bin/env bash
set -euo pipefail

# check-no-stash.sh — fail when the stash stack is non-empty.
#
# The stash stack is REPO-GLOBAL: `git stash list` returns the same entries no matter
# which branch is checked out. With several sessions working concurrently in this one
# checkout, that makes it shared mutable state of the worst kind:
#
#   - `stash@{0}` means whatever the LAST session to push happened to save, so a `git stash
#     pop` from one session can restore another session's unrelated edits into your tree.
#   - An entry is invisible in `git log`, `git status` of the branch it came from, and
#     `tools/next.sh` — so work parked there is work nobody can find.
#   - `git rebase --autostash` uses the SAME stack, so a rebase silently participates.
#
# The alternative is strictly better and is what this repo requires: commit the WIP on your
# own branch (`tools/wip.sh`). A commit is private to your branch, visible in `git log`,
# survives a crash, and undoes with one `git reset --soft HEAD~1`.
#
# WHY A GUARD AND NOT PROSE: git has no pre-stash hook, so the stash CALL cannot be
# intercepted. What can be checked is the resulting state — an entry sitting on the stack —
# and that is the thing that actually hurts another session. This guard therefore fails on
# the residue, not the act.
#
# Recovering existing entries WITHOUT losing them:
#   git stash list                      # see them
#   git stash branch <name> stash@{N}   # restore entry N onto a new branch (non-destructive)
#   git stash drop stash@{N}            # only once you are sure it is not needed
#
# Exit 0 when the stack is empty, exit 1 (listing entries) otherwise.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

entries="$(git stash list 2>/dev/null || true)"
[ -z "$entries" ] && exit 0

n="$(printf '%s\n' "$entries" | wc -l | tr -d ' ')"
{
  printf 'check-no-stash: %s stash entr%s on the repo-global stack:\n\n' "$n" "$([ "$n" = 1 ] && echo y || echo ies)"
  printf '%s\n' "$entries" | sed 's/^/  /'
  printf '\n'
  printf 'The stash stack is shared by every concurrent session in this checkout, so these\n'
  printf 'are visible and poppable from any of them — stash@{0} is whoever pushed last.\n'
  printf '\n'
  printf 'Move each one onto a branch (non-destructive), then drop it:\n'
  printf '  git stash branch wip-<name> stash@{0}\n'
  printf '  git stash drop stash@{0}          # once you are sure\n'
  printf '\n'
  printf 'Park work in progress with tools/wip.sh instead — a commit on your own branch is\n'
  printf 'private to it, findable in git log, and undone with git reset --soft HEAD~1.\n'
} >&2
exit 1
