#!/usr/bin/env bash
# new-task.sh — start a change. One worktree per change, not one branch per change.
#
# WHY A WORKTREE AND NOT JUST A BRANCH: `git checkout` mutates the ONE working tree every
# session shares, so two agents (or an agent and you) working at the same time fight over
# it — this repo has already had a merge land on the wrong branch because the shared tree
# had been switched underneath it, and a rebase had to be autostashed around someone's
# uncommitted edits. A worktree gives each change its own directory and its own checked-out
# branch: `git checkout` is never needed, uncommitted work is never in anyone else's way,
# and stop-checks verifies the tree the work is actually in (see scripts/stop-checks.sh's
# worktree handling).
#
#   tools/new-task.sh <short-kebab-name> ["one-line description"]
#
# creates:
#   branch    task/<short-kebab-name>   from origin/main (falls back to main)
#   worktree  worktrees/<short-kebab-name>
#   the branch description tools/next.sh reads
#
# and prints the cd line. node_modules is NOT installed per worktree — stop-checks
# symlinks the main checkout's on first run, so TS checks work with no extra 136M.

set -euo pipefail

MAIN_ROOT="$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")"
cd "$MAIN_ROOT"

die() { printf 'new-task: %s\n' "$1" >&2; exit 1; }

NAME="${1:-}"
DESC="${2:-}"
[ -n "$NAME" ] || die "usage: tools/new-task.sh <short-kebab-name> [\"one-line description\"]"

# Kebab only: the name becomes a branch name AND a directory name, so anything needing
# quoting or escaping in either place is rejected up front rather than half-created.
case "$NAME" in
  task/*) die "pass the SHORT name, not the branch: '${NAME#task/}' rather than '$NAME'" ;;
esac
printf '%s' "$NAME" | grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$' \
  || die "name must be lower-case kebab (a-z, 0-9, single hyphens): got '$NAME'"

BRANCH="task/$NAME"
WT="worktrees/$NAME"

# Refuse to half-create. Either of these existing means a change by this name is already
# in flight, and silently reusing it is how two changes end up sharing one branch.
git show-ref --verify --quiet "refs/heads/$BRANCH" && die "branch $BRANCH already exists (worktree: $(git worktree list | grep -F "[$BRANCH]" | awk '{print $1}' || echo 'none'))"
[ -e "$WT" ] && die "$WT already exists"

# Base on origin/main when we have it — starting from a stale local main is how a branch
# silently begins life behind and picks up a conflict later.
BASE=main
if git show-ref --verify --quiet refs/remotes/origin/main; then
  git fetch -q origin main 2>/dev/null || true
  BASE=origin/main
fi

git worktree add -b "$BRANCH" "$WT" "$BASE" >/dev/null 2>&1
[ -n "$DESC" ] && git -C "$WT" config "branch.$BRANCH.description" "$DESC"

printf '\033[1m%s\033[0m created from %s\n' "$BRANCH" "$BASE"
[ -n "$DESC" ] || printf '  no description set — tools/next.sh will show this branch unlabelled.\n  set one: git -C %s config branch.%s.description "..."\n' "$WT" "$BRANCH"
printf '\n  cd %s/%s\n\n' "$MAIN_ROOT" "$WT"
printf 'when it is done: verify, merge to main, then\n  git worktree remove %s && git branch -d %s\n' "$WT" "$BRANCH"
