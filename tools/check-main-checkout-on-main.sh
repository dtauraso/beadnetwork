#!/usr/bin/env bash
set -euo pipefail

# check-main-checkout-on-main.sh — fail when the MAIN checkout is not on `main`.
#
# This repo runs several concurrent sessions (multiple chats, subagents) against one
# clone. Under the worktree-per-change rule (CLAUDE.md, tools/new-task.sh) every change
# gets its OWN worktree, which leaves exactly one job for the main checkout: hold `main`.
# Nothing else should ever be checked out there.
#
# WHAT THIS CATCHES, from a real incident on 2026-07-28: the main checkout had been
# switched to a task branch by one session. Another session ran `git merge --ff-only
# <its-branch>` intending to merge to main — and silently merged into whatever branch the
# shared tree happened to be on instead. `git push origin main` then succeeded as a no-op
# and was reported as success. Recovering it took a cherry-pick onto main, a rebase of the
# other branch to strip the commits back out, and an autostash around a third party's
# uncommitted edits.
#
# The single fact that would have prevented all of it: the main checkout was not on main.
# Every session's verify now says so, loudly, before any merge is attempted.
#
# WHY A GUARD AND NOT ETIQUETTE: `git checkout` in a shared tree is invisible to everyone
# else — no other session can see that its footing moved. A rule in prose is only read by
# whoever remembers to read it; this fires for whoever is about to be hurt by it.
#
# Detached HEAD counts as a failure too: a bisect or a rebase left mid-flight in the main
# checkout is exactly as dangerous to a concurrent merge as a task branch is.
#
# Exit 0 when the main checkout is on main, exit 1 (naming the branch) otherwise.

# Resolve the MAIN checkout: every worktree of a repo shares one .git dir, and for a
# standard clone the main checkout is its parent. Deliberately NOT the caller's own
# worktree — this guard is about the shared tree, and it must report the same answer no
# matter which worktree runs it.
COMMON_DIR="$(git rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)"
if [ -z "$COMMON_DIR" ]; then
  echo "check-main-checkout-on-main: MISCONFIGURED — not inside a git repository." >&2
  exit 1
fi
MAIN_ROOT="$(dirname "$COMMON_DIR")"

# A bare or unusually-laid-out repo has no main checkout to police. Say so rather than
# inventing a verdict.
if [ ! -d "$MAIN_ROOT" ]; then
  echo "check-main-checkout-on-main: no main checkout at '$MAIN_ROOT' — skipping." >&2
  exit 0
fi

BRANCH="$(git -C "$MAIN_ROOT" rev-parse --abbrev-ref HEAD 2>/dev/null || echo '?')"

if [ "$BRANCH" = "main" ]; then
  exit 0
fi

{
  if [ "$BRANCH" = "HEAD" ]; then
    printf 'check-main-checkout-on-main: the main checkout (%s) is on a DETACHED HEAD.\n' "$MAIN_ROOT"
  else
    printf 'check-main-checkout-on-main: the main checkout (%s) is on %s, not main.\n' "$MAIN_ROOT" "$BRANCH"
  fi
  printf '\n'
  printf 'That tree is shared by every session. While it is off main, a `git merge` run\n'
  printf 'there lands on %s instead of main, and `git push origin main` succeeds as a\n' "$BRANCH"
  printf 'no-op — which is exactly how commits went to the wrong branch on 2026-07-28.\n'
  printf '\n'
  printf 'Fix:\n'
  printf '  git -C %s checkout main\n' "$MAIN_ROOT"
  printf '\n'
  printf 'If you have work in progress there, it belongs in its own worktree:\n'
  printf '  tools/new-task.sh <short-kebab-name> "one-line description"\n'
} >&2
exit 1
