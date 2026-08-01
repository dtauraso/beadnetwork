#!/usr/bin/env bash
# new-task.sh — start a change: a task branch in this ONE checkout.
#
#   tools/new-task.sh <short-kebab-name> ["one-line description"]
#
# creates:
#   branch    task/<short-kebab-name>   from origin/main (falls back to main)
#   the branch description tools/next.sh reads
#
# and checks it out. There is one checkout and one working tree; concurrent
# sessions share it, so commit or otherwise land your work before switching away
# (see CLAUDE.md's no-`git stash` rule — the stash stack is shared too).

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

die() { printf 'new-task: %s\n' "$1" >&2; exit 1; }

NAME="${1:-}"
DESC="${2:-}"
[ -n "$NAME" ] || die "usage: tools/new-task.sh <short-kebab-name> [\"one-line description\"]"

# Kebab only: the name becomes a branch name, so anything needing quoting or
# escaping there is rejected up front rather than half-created.
case "$NAME" in
  task/*) die "pass the SHORT name, not the branch: '${NAME#task/}' rather than '$NAME'" ;;
esac
printf '%s' "$NAME" | grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$' \
  || die "name must be lower-case kebab (a-z, 0-9, single hyphens): got '$NAME'"

BRANCH="task/$NAME"

# Refuse to half-create. An existing branch by this name means a change is already
# in flight, and silently reusing it is how two changes end up sharing one branch.
git show-ref --verify --quiet "refs/heads/$BRANCH" && die "branch $BRANCH already exists"

# Working tree must be clean before switching — uncommitted edits would otherwise
# ride along onto the new branch silently. Land or commit them first.
[ -z "$(git status --porcelain)" ] || die "working tree is dirty — commit or park (tools/wip.sh) before starting a new task"

# Base on origin/main when we have it — starting from a stale local main is how a branch
# silently begins life behind and picks up a conflict later.
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
