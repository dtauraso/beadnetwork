#!/usr/bin/env bash
set -euo pipefail

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

[ "$branch" = "main" ] && die "refusing to park WIP on main — start a change with scripts/new-task.sh <name>"

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

git add -A
git commit -q -m "$MSG"
printf 'wip: parked on %s as %s\n' "$branch" "$(git rev-parse --short HEAD)"
printf '  resume:  scripts/wip.sh --undo\n'
