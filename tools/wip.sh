#!/usr/bin/env bash
set -euo pipefail

# wip.sh — park work in progress as a commit on YOUR branch, instead of `git stash`.
#
# The stash stack is repo-global (see tools/check-no-stash.sh): every concurrent session
# in this one checkout sees and can pop the same entries, so it is shared mutable state.
# A commit is not — it belongs to the branch you are on.
#
#   tools/wip.sh ["message"]     commit everything (tracked + untracked) as a WIP commit
#   tools/wip.sh --undo          undo the last WIP commit, restoring it to the working tree
#
# --undo is `git reset --soft HEAD~1` with a check that HEAD actually IS a wip commit, so
# it cannot silently eat a real commit.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

WIP_PREFIX='wip:'

die() { printf 'wip: %s\n' "$1" >&2; exit 1; }

branch="$(git rev-parse --abbrev-ref HEAD)"
[ "$branch" = "HEAD" ] && die "detached HEAD — check out a branch before parking work"

if [ "${1:-}" = "--undo" ]; then
  subject="$(git log -1 --format=%s)"
  case "$subject" in
    "$WIP_PREFIX"*) ;;
    *) die "HEAD is not a wip commit (subject: '$subject') — refusing to reset a real commit" ;;
  esac
  git reset --soft HEAD~1
  printf 'wip: undone — changes are back in the working tree, staged.\n'
  exit 0
fi

# main is shared; a wip commit there is exactly the mess this whole workflow avoids.
[ "$branch" = "main" ] && die "refusing to park WIP on main — start a change with tools/new-task.sh <name>"

if [ -z "$(git status --porcelain)" ]; then
  printf 'wip: nothing to park (working tree clean).\n'
  exit 0
fi

MSG="${1:-}"
if [ -n "$MSG" ]; then
  MSG="$WIP_PREFIX $MSG"
else
  MSG="$WIP_PREFIX $branch"
fi

# -A so untracked files ride along too — the case where `git stash` (without -u) silently
# leaves them behind is a classic way to lose new work.
git add -A
git commit -q -m "$MSG"
printf 'wip: parked on %s as %s\n' "$branch" "$(git rev-parse --short HEAD)"
printf '  resume:  tools/wip.sh --undo\n'
