#!/usr/bin/env bash
# PreToolUse hook on Bash: a MUTATING git command must name the tree it acts on.
#
# The drift this exists to stop: an agent cd's into the main checkout to inspect
# something (legitimate — "which branch is main on?", "what is uncommitted there?"),
# and the shell's cwd PERSISTS across Bash calls. The next `git commit`/`git rebase`
# then lands in the main checkout instead of the task worktree, silently, because
# nothing in the command says where it runs. This repo has already lost a merge to
# exactly that (CLAUDE.md, "One worktree per change").
#
# Why this is a hook and not a tools/check-*.sh: the stop-time guards inspect REPO
# STATE, and this drift leaves no trace in repo state. The tree was clean, the main
# checkout was on main, every guard passed — the bad part was which directory the
# shell happened to be in, which is gone by the time anything could check it. A
# PreToolUse hook is the only thing that sees the command before it runs.
#
# The rule: location must be EXPLICIT, never ambient. A mutating git command either
# already starts with `cd <path> &&`, or this hook prepends one. Read-only git
# (status, log, diff, rev-parse, show, worktree list, fetch) is untouched — inspecting
# the main checkout is a thing we genuinely need to do, and that need is what creates
# the drift in the first place. Constraining only the mutating half is what lets the
# hook be strict without being in the way.
#
# Rather than blocking, it REWRITES: PreToolUse returns hookSpecificOutput.updatedInput,
# so the command just runs, in the right tree. Blocking would cost a round trip to say
# something the hook already knows how to fix.
#
# Ambiguity is the one case it will not guess at. ZERO task worktrees means there is no
# worktree to redirect to (work on main is the caller's call) — allowed untouched. TWO
# OR MORE means the hook cannot know which one this command belongs to — denied, with
# the list, so the caller names it. Guessing there would be the same class of silent
# wrong-tree write the hook exists to prevent.

set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# The MAIN checkout, not wherever this script happens to sit. This file is tracked, so
# it exists inside every worktree too — resolving the prefix from its own path would
# make the `worktrees/` test compare a worktree against itself, match nothing, and let
# every bare git command through. A guard that cannot fire reads exactly like a passing
# one, which is why this is derived instead of assumed: `worktree list` always names the
# main worktree first, whichever tree you ask from.
REPO_ROOT="$(git -C "$HERE" worktree list --porcelain 2>/dev/null | awk '/^worktree /{print $2; exit}')"
[ -n "$REPO_ROOT" ] || exit 0

payload="$(cat)"
cmd="$(printf '%s' "$payload" | jq -r '.tool_input.command // empty')"
[ -n "$cmd" ] || exit 0

# Mutating git verbs only. `stash` is here because CLAUDE.md bans it outright (the
# stash stack is repo-GLOBAL, so one worktree can pop another's work) — being forced
# to name a tree is a useful speed bump on the way to not using it at all.
if ! printf '%s' "$cmd" | grep -qE '\bgit[[:space:]]+(-[^[:space:]]+[[:space:]]+)*(commit|push|rebase|merge|add|reset|checkout|switch|cherry-pick|revert|stash|clean|am|apply|restore|tag)\b'; then
  exit 0
fi

# Already explicit — the caller named the tree, so respect it verbatim. This is also
# what makes the hook idempotent: a rewritten command re-entering the hook is a no-op.
if printf '%s' "$cmd" | grep -qE '^[[:space:]]*cd[[:space:]]+[^[:space:]]+[[:space:]]*&&'; then
  exit 0
fi

# `git -C <path>` is the other way to be explicit, and it is the stronger one — it
# survives a compound command that cd's again halfway through. Respect it too.
if printf '%s' "$cmd" | grep -qE '\bgit[[:space:]]+-C[[:space:]]'; then
  exit 0
fi

worktrees=()
while IFS= read -r line; do
  [ -n "$line" ] && worktrees+=("$line")
done < <(git -C "$REPO_ROOT" worktree list --porcelain 2>/dev/null \
  | awk '/^worktree /{print $2}' \
  | grep -F "$REPO_ROOT/worktrees/" || true)

case "${#worktrees[@]}" in
  0)
    # No task worktree exists, so there is nowhere to redirect to and nothing for the
    # main checkout to be confused WITH. Let it through.
    exit 0
    ;;
  1)
    target="${worktrees[0]}"
    jq -cn --arg c "cd $(printf '%q' "$target") && $cmd" --arg t "$target" '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        updatedInput: { command: $c },
        additionalContext: ("Mutating git runs in the task worktree, not the main checkout — this command was prefixed with a cd to " + $t + " because it named no tree. Name the tree yourself (cd <path> && ... , or git -C <path>) and this hook stays out of the way.")
      }
    }'
    exit 0
    ;;
  *)
    list=""
    for w in "${worktrees[@]}"; do list+="  $w"$'\n'; done
    jq -cn --arg l "$list" '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: ("More than one task worktree exists, so this hook cannot know which tree your mutating git command belongs to — and guessing is the exact failure it exists to prevent. Re-run naming the tree (cd <worktree> && ... , or git -C <worktree>):\n" + $l)
      }
    }'
    exit 0
    ;;
esac
