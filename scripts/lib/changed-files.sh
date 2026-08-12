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
