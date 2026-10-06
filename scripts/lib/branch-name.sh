#!/usr/bin/env bash

BRANCH_NAME_STAMP=".beadnetwork-cache/branch-name-shown"

nudge_branch_name() {
  [ "$MODE" = "hook" ] || return 0
  [ -t 0 ] && return 0

  local hook_input branch head shown base desc stat
  hook_input=$(cat 2>/dev/null)
  grep -q '"hook_event_name" *: *"Stop"' <<< "$hook_input" || return 0
  grep -q '"stop_hook_active" *: *true' <<< "$hook_input" && return 0

  branch=$(git branch --show-current 2>/dev/null)
  case "$branch" in
    task/*) ;;
    *) return 0 ;;
  esac

  base=main
  git rev-parse --verify -q origin/main >/dev/null 2>&1 && base=origin/main
  stat=$(git diff --stat "$base...HEAD" 2>/dev/null)
  [ -n "$stat" ] || return 0

  head=$(git rev-parse HEAD 2>/dev/null)
  shown=""
  [ -f "$BRANCH_NAME_STAMP" ] && shown=$(cat "$BRANCH_NAME_STAMP" 2>/dev/null)
  [ "$shown" = "$branch $head" ] && return 0
  mkdir -p "$(dirname "$BRANCH_NAME_STAMP")" 2>/dev/null
  printf '%s %s\n' "$branch" "$head" > "$BRANCH_NAME_STAMP"

  desc=$(git config "branch.$branch.description" 2>/dev/null)
  emit_block "Branch name check (once per commit). Does the name and description still state this branch's diff against $base?

  $branch
  description: ${desc:-(none)}

$stat

If not: scripts/rename-task.sh <short-kebab-name> \"description\". If so, stop again — this commit will not be shown twice."
}
