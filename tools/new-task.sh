#!/usr/bin/env bash

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

die() { printf 'new-task: %s\n' "$1" >&2; exit 1; }

NAME="${1:-}"
DESC="${2:-}"
[ -n "$NAME" ] || die "usage: tools/new-task.sh <short-kebab-name> [\"one-line description\"]"

case "$NAME" in
  task/*) die "pass the SHORT name, not the branch: '${NAME#task/}' rather than '$NAME'" ;;
esac
printf '%s' "$NAME" | grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$' \
  || die "name must be lower-case kebab (a-z, 0-9, single hyphens): got '$NAME'"

BRANCH="task/$NAME"

git show-ref --verify --quiet "refs/heads/$BRANCH" && die "branch $BRANCH already exists"

[ -z "$(git status --porcelain)" ] || die "working tree is dirty — commit or park (tools/wip.sh) before starting a new task"

BASE=main
if git show-ref --verify --quiet refs/remotes/origin/main; then
  git fetch -q origin main 2>/dev/null || true
  BASE=origin/main
fi

git checkout -q -b "$BRANCH" "$BASE"
[ -n "$DESC" ] && git config "branch.$BRANCH.description" "$DESC"

printf '\033[1m%s\033[0m created from %s, checked out\n' "$BRANCH" "$BASE"
[ -n "$DESC" ] || printf '  no description set — tools/next.sh will show this branch unlabelled.\n  set one: git config branch.%s.description "..."\n' "$BRANCH"
printf '\nwhen it is done: verify, merge to main, then\n  git branch -d %s\n' "$BRANCH"
