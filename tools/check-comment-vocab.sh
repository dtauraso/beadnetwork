#!/usr/bin/env bash
# check-comment-vocab.sh — fail if retired architecture vocabulary reappears in .go/.ts
# code comments. Run from repo root: bash tools/check-comment-vocab.sh
#
# WHY THIS EXISTS (audit-integrate-into-repo-systems): the priors-fit audit found stale
# "fan-in is safe" comments in paced_wire.go and ports.go that CONTRADICT MODEL.md (fan-in
# is rejected at parse — see tools/check-no-fan-in.sh and topo_spec.go validateNoFanIn). The
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
  # vocab-drift audit round: stale "fd3/fd-3 <live channel>" framings that imply fd 3
  # carries the live render path. The live path is per-goroutine dedicated streams.
  "fd3 side channel"
  "fd-3 side channel"
  "fd-3 content buffer"
  "fd3 content buffer"
  "fd-3 SCENE"
  "fd3 SCENE"
  "fd-3 node frame"
  "fd-3 Node block"
  "fd3 binary side"
  "fd3 snapshot"
  # atomic/lock comment audit (2026-07-25): nodes/Buffer have zero live atomics/mutexes
  # (check-no-network-locks.sh); these tokens named a shared-state mechanism that never
  # existed post-refactor or was deleted, and their reappearance re-teaches the wrong
  # model (memory/feedback_no_atomics_are_defects.md).
  "atomically-published"
  "atomic-snapshot-backed"
  "the atomic held"
  # stale-audit round (2026-07-25): KindId doc drift — it was renumbered to a
  # STABLE per-kind id (SPEC.md kindId) and is never a sort-order index; this
  # phrase re-teaches the old (wrong) model.
  "alphabetically-sorted"
  # stale-audit round (2026-07-25): the leaf extraction moved Register from
  # Wiring to nodes/wire — "Wiring.Register" is a ghost symbol.
  "Wiring.Register"
  # stale-audit round (2026-07-25): pump.ts was deleted and no TS camera store exists
  # (MODEL.md: TS holds no domain state); these named a render path that no longer
  # exists. Bare "pump" is intentionally NOT guarded — too generic/false-positive-prone.
  "useCameraStore"
  "CameraFromStore"
  "pump.ts"
  # audit round (2026-07-25): port_geometry.go claimed to be a "Go mirror" of TS
  # geometry-helpers.ts, but the TS port-geometry functions were removed when Go took
  # over geometry — there is no counterpart to mirror. Ban the specific false claim
  # (not bare "geometry-helpers.ts", which is a real screen-coord file still referenced).
  "Go mirror of the port-to-port segment geometry"
  # stale-audit round (2026-07-28): WindowAndInhibit*Gate is retired vocabulary; the live
  # kinds are SelectRight/SelectLeft (memory/project_node_color_vocab.md). gate.go's
  # RunGate/RunGateAccept doc comments called it a "window-and-inhibit gate loop", which
  # re-teaches the old kind name.
  "window-and-inhibit gate loop"
)

fail=0

# One whole-repo walk for ALL tokens at once (-f reads patterns from a file/process-sub),
# instead of one walk per token — same -F fixed-string, -n line-number, --include/
# --exclude-dir semantics as the old per-token grep. The self-exclusion and
# comment-line-only post-filters are still applied, just once to the combined output
# rather than once per token, so the meaning is unchanged.
all_hits="$(grep -rnIF --include="*.go" --include="*.ts" --include="*.tsx" \
    --exclude-dir={node_modules,out,.git,handoff-archive,memory} \
    -f <(printf '%s\n' "${DEAD_COMMENT_TOKENS[@]}") -- . 2>/dev/null \
    | grep -vF "tools/check-comment-vocab.sh" \
    | grep -E ':[[:space:]]*(//|\*|#)' || true)"

for token in "${DEAD_COMMENT_TOKENS[@]}"; do
  # Attribute hits back to this specific token in-memory (no repo re-walk). Match against the
  # CONTENT field only (strip the "path:line:" prefix first) so a token substring that happens
  # to appear in a path/line-number wouldn't misattribute a hit to the wrong token.
  hits="$(printf '%s\n' "$all_hits" | awk -F: -v t="$token" '
    { content = $0; sub(/^[^:]*:[0-9]*:/, "", content); if (index(content, t) > 0) print }
  ' || true)"
  if [ -n "$hits" ]; then
    echo "RETIRED COMMENT VOCAB: '$token' — remove or reword; it contradicts the current model:"
    printf '%s\n' "$hits"
    fail=1
  fi
done

exit $fail
