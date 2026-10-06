#!/usr/bin/env bash

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

die() { printf 'rename-task: %s\n' "$1" >&2; exit 1; }

NAME="${1:-}"
DESC="${2:-}"
[ -n "$NAME" ] && [ -n "$DESC" ] \
  || die "usage: scripts/rename-task.sh <short-kebab-name> \"one-line description of the diff against main\""

case "$NAME" in
  task/*) die "pass the SHORT name, not the branch: '${NAME#task/}' rather than '$NAME'" ;;
esac
grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$' <<< "$NAME" \
  || die "name must be lower-case kebab (a-z, 0-9, single hyphens): got '$NAME'"

OLD="$(git branch --show-current)"
case "$OLD" in
  task/*) ;;
  *) die "current branch '$OLD' is not a task/* branch" ;;
esac
NEW="task/$NAME"

if [ "$OLD" != "$NEW" ]; then
  git show-ref --verify --quiet "refs/heads/$NEW" && die "branch $NEW already exists"
  git branch -m "$OLD" "$NEW"
fi
git config "branch.$NEW.description" "$DESC"

if [ "$OLD" != "$NEW" ] && git ls-remote --exit-code --heads origin "$OLD" >/dev/null 2>&1; then
  git push -q -u origin "$NEW"
  git push -q origin --delete "$OLD"
fi

BASE=main
git show-ref --verify --quiet refs/remotes/origin/main && BASE=origin/main

printf '\033[1m%s\033[0m  (was %s)\n  %s\n\ndiff against %s — the name and description must state THIS:\n' "$NEW" "$OLD" "$DESC" "$BASE"
git diff --stat "$BASE...HEAD"
