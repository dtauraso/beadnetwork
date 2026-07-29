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
#
# Expressed directly as a single `find -iname` alternation (one process for the whole tree)
# rather than a shell loop spawning `basename` per file. Every alternative from the old
# `banned_regex` must appear here verbatim — dropping one would silently stop banning that
# filename. `-iname` gives the same case-insensitive match as the old `shopt -s nocasematch`.
# worktrees/ is pruned for BOTH reasons: tools/new-task.sh puts one full checkout of this
# repo there per open task, so without it this walks the whole source tree once per open
# task (cost scaling with in-flight work, not with this tree), and a hit there would be
# another branch's file, which this branch cannot fix.
#
# Pruned rather than enumerated with `git ls-files --others --exclude-standard`, unlike
# check-comment-vocab.sh: that flag pair hides GITIGNORED files, and a handoff cache someone
# added to .gitignore is exactly the case this guard has to catch.
hits="$(find . \
  -type d \( -name node_modules -o -name .git -o -name out -o -name handoff-archive -o -name worktrees \) -prune -o \
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
