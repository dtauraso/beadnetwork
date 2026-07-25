#!/usr/bin/env bash
# check-movedispatch-composer.sh — keep MoveDispatch a thin COMPOSER, not a regrown
# god-object. Run from repo root: bash tools/check-movedispatch-composer.sh
#
# WHY THIS EXISTS (audit god-module decouple): MoveDispatch was a 27-field god-object
# owning ~7 responsibilities (row tables, seeds, stream wiring, UI state, mover registry,
# layout quantize, dispatch). It was decomposed into single-responsibility owner types
# (rowTables/geomSeeds/streamWiring/uiState/moverRegistry/layoutQuantizer), leaving
# MoveDispatch a thin composer that holds one pointer/value per owner and delegates. The
# regrowth failure mode is silent: someone adds a NEW field directly to MoveDispatch
# instead of to the owning sub-type, and the god-object creeps back a field at a time.
# This guard caps the struct's field count so that can't happen unnoticed — a genuinely
# new responsibility must land as (or on) an owner type, not as loose MoveDispatch state.
#
# It is NOT a line-count guard: the point is single-responsibility ownership, not file size.
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

FILE="nodes/Wiring/node_move.go"
if [[ ! -f "$FILE" ]]; then
  echo "check-movedispatch-composer: MISCONFIGURED — $FILE not found (renamed?); refusing vacuous pass" >&2
  exit 1
fi

# Cap: the composer holds 6 owner sub-objects + tr/persist/msgTap/ctx = 10 fields today.
# A little headroom for legitimate coordination state, but well below god-object territory.
readonly MAX_FIELDS=12

# Extract the MoveDispatch struct body and count FIELD-bearing lines (skip the
# `type ... {` / closing `}` lines, blank lines, and comment-only lines).
body=$(awk '/^type MoveDispatch struct \{/{f=1;next} f&&/^\}/{f=0} f' "$FILE")
if [[ -z "$body" ]]; then
  echo "check-movedispatch-composer: MISCONFIGURED — could not locate the MoveDispatch struct body in $FILE" >&2
  exit 1
fi

count=$(printf '%s\n' "$body" \
  | grep -vE '^[[:space:]]*//' \
  | grep -vE '^[[:space:]]*$' \
  | grep -c .)

if [[ "$count" -gt "$MAX_FIELDS" ]]; then
  echo "MoveDispatch has $count fields (cap $MAX_FIELDS) — it is regrowing into a god-object."
  echo "  Add new state to the responsible owner type (rowTables/geomSeeds/streamWiring/"
  echo "  uiState/moverRegistry/layoutQuantizer), or introduce a new owner — not a loose"
  echo "  MoveDispatch field. If a new owner legitimately raises the composer count, bump"
  echo "  MAX_FIELDS deliberately with a note."
  exit 1
fi

exit 0
