#!/usr/bin/env bash

# Sourced by scripts/stop-checks.sh.

collect_changed_files() {
  local worktree_changed base committed_changed
  worktree_changed=$(git status --porcelain 2>/dev/null | awk '{print $NF}')

  base=""
  if git rev-parse --verify -q origin/main >/dev/null 2>&1; then
    base="origin/main"
  elif git rev-parse --verify -q main >/dev/null 2>&1; then
    base="main"
  fi
  committed_changed=""
  if [ -n "$base" ]; then
    committed_changed=$(git diff --name-only "$base"...HEAD 2>/dev/null || true)
  fi

  printf '%s\n%s\n' "$worktree_changed" "$committed_changed"
}

STOP_CHECK_STAMP=".beadnetwork-cache/last-stop-check"

# branch re-ran go build, tsc, the npm build, eslint and vitest on every turn
collect_changed_since_last_check() {
  local worktree_changed last committed_changed
  worktree_changed=$(git status --porcelain 2>/dev/null | awk '{print $NF}')

  last=""
  [ -f "$STOP_CHECK_STAMP" ] && last=$(cat "$STOP_CHECK_STAMP" 2>/dev/null)
  if [ -z "$last" ] || ! git rev-parse --verify -q "$last" >/dev/null 2>&1; then
    collect_changed_files
    return
  fi

  committed_changed=$(git diff --name-only "$last" HEAD 2>/dev/null || true)
  printf '%s\n%s\n' "$worktree_changed" "$committed_changed"
}

mark_stop_check_done() {
  mkdir -p "$(dirname "$STOP_CHECK_STAMP")" 2>/dev/null || return 0
  git rev-parse HEAD > "$STOP_CHECK_STAMP" 2>/dev/null || true
}
