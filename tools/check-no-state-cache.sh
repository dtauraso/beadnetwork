#!/usr/bin/env bash
# check-no-state-cache.sh — forbid synced live-state snapshot files. Run from repo root:
# bash tools/check-no-state-cache.sh
#
# WHY THIS EXISTS (drift-checklist item #2 — "context contamination"): a repo guard can't
# see the conversation, but CLAUDE.md's session-handoff section bans the MECHANISM that
# reintroduces stale state into new turns — "Do not recreate a handoff.md or a
# continuation-prompt template ... not in a synced snapshot." Live task state is DERIVED
# (branch descriptions + tools/next.sh + memory/ + MODEL.md), never cached in a hand-synced
# doc that drifts. This guard fails if such a snapshot file reappears, enforcing that ban
# statically. (docs/planning/visual-editor/session-log.md is the friction LOG, not a state
# cache, and is explicitly allowed.)
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Filenames whose EXISTENCE is a banned synced-state cache. Case-insensitive basename match.
# Kept narrow and explicit to avoid false positives (session-log.md is NOT here).
banned_regex='^(handoff|session-handoff|session_state|continuation-prompt|continuation_prompt|next-session|handoff-notes)\.md$'

hits="$(find . \
  -type d \( -name node_modules -o -name .git -o -name out -o -name handoff-archive \) -prune -o \
  -type f -print 2>/dev/null \
  | while IFS= read -r p; do
      b="$(basename "$p")"
      shopt -s nocasematch
      if [[ "$b" =~ $banned_regex ]]; then echo "$p"; fi
      shopt -u nocasematch
    done)"

if [[ -n "$hits" ]]; then
  echo "STATE CACHE FORBIDDEN: a synced live-state snapshot file exists (CLAUDE.md bans handoff/continuation caches):"
  printf '%s\n' "$hits" | sed 's/^/  /'
  echo "  Live state is DERIVED (branch descriptions + tools/next.sh + memory/ + MODEL.md) — delete the snapshot."
  exit 1
fi

exit 0
