#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: none | checks the repo-global git stash stack, not a set of source files

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../../.."

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
  printf 'Park work in progress with scripts/wip.sh instead — a commit on your own branch is\n'
  printf 'private to it, findable in git log, and undone with git reset --soft HEAD~1.\n'
} >&2
exit 1
