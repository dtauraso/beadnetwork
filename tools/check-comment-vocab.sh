#!/usr/bin/env bash
# check-comment-vocab.sh — fail if retired architecture vocabulary reappears in .go/.ts
# code comments. Run from repo root: bash tools/check-comment-vocab.sh
#
# WHY THIS EXISTS (audit-integrate-into-repo-systems): the priors-fit audit found stale
# "fan-in is safe" comments in paced_wire.go and ports.go that CONTRADICT MODEL.md (fan-in
# is rejected at parse — see tools/check-no-fan-in.sh and loader.go validateNoFanIn). The
# existing tracker prose that flagged them had itself drifted (cited ports.go:173, actual
# :223), which is exactly why a line-pointer in a doc is the wrong tool and a grep guard is
# the right one: prose rots, `grep | exit 1` cannot.
#
# Ported from Uncle-Bob raid exemplar ~/Downloads/unclebob-repos/empire-2025/scripts/
# check-ai-map-access.sh — the locked-token, subtract-allowlist, fail-on-unexpected pattern.
# Companion to check-dead-doc-tokens.sh, which does the same for CLAUDE.md/MODEL.md prose.
#
# Exit 0 clean (empty), exit 1 with a report — matches the guard-loop contract in
# scripts/stop-checks.sh (auto-discovered via tools/check-*.sh glob).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Retired vocabulary that must not reappear in code comments. Each token names a concept the
# current MODEL.md has removed; a comment asserting it re-teaches the wrong model to the next
# reader (human or AI). Extend this list as later audit rounds retire more vocabulary.
readonly DEAD_COMMENT_TOKENS=(
  "fan-in is safe"
  "fan-in safe"
  # fd-3 / SnapshotState bridge machinery was deleted: the Go->TS bridge is now one
  # dedicated inherited-stdio pipe per goroutine (WIREFOLD_STREAM_FDS), and stdio index 3
  # is a reserved, UNUSED slot. These name code that no longer exists, so any reappearance
  # in a comment is drift. (Verified dead-as-code 2026-07-24: no SnapshotState type, no
  # handleFd3 def, nothing reads/writes fd 3.) NOTE: bare "fd 3" is intentionally NOT
  # guarded — the accurate "stdio index 3 is a reserved, unused slot" comments must survive.
  "SnapshotState"
  "handleFd3"
  "fd-3 fallback"
  "fd3 fallback"
)

fail=0
for token in "${DEAD_COMMENT_TOKENS[@]}"; do
  # -F fixed string, -n line numbers; restrict to lines that look like comments (// or *).
  hits="$(grep -rnIF --include="*.go" --include="*.ts" --include="*.tsx" \
      --exclude-dir={node_modules,out,.git,handoff-archive,memory} \
      -- "$token" . 2>/dev/null \
      | grep -vF "tools/check-comment-vocab.sh" \
      | grep -E ':[[:space:]]*(//|\*|#)' || true)"
  if [ -n "$hits" ]; then
    echo "RETIRED COMMENT VOCAB: '$token' — remove or reword; it contradicts the current model:"
    printf '%s\n' "$hits"
    fail=1
  fi
done

exit $fail
