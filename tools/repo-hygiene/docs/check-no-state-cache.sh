#!/usr/bin/env bash

# PLACEMENT: docs/**,*.md | no synced live-state snapshot doc; task state is derived from git/memory/MODEL.md

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

hits="$(find . \
  -type d \( -name node_modules -o -name .git -o -name out -o -name handoff-archive -o -path './.claude/worktrees' \) -prune -o \
  -type f \( \
       -iname 'handoff.md' \
    -o -iname 'session-handoff.md' \
    -o -iname 'session_state.md' \
    -o -iname 'continuation-prompt.md' \
    -o -iname 'continuation_prompt.md' \
    -o -iname 'next-session.md' \
    -o -iname 'handoff-notes.md' \
  \) -print 2>/dev/null)"

if [[ -n "$hits" ]]; then
  echo "STATE CACHE FORBIDDEN: a synced live-state snapshot file exists (CLAUDE.md bans handoff/continuation caches):"
  printf '%s\n' "$hits" | sed 's/^/  /'
  echo "  Live state is DERIVED (branch descriptions + tools/next.sh + memory/ + MODEL.md) — delete the snapshot."
  exit 1
fi

exit 0
