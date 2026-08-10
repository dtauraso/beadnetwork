#!/usr/bin/env bash
#
# PLACEMENT: nodes/Wiring/move_dispatch.go,nodes/Wiring/node_geometry.go | a composer struct (MoveDispatch, nodeGeometry) stays THIN: new state belongs in a named sub-struct, not a new loose field
# check-composer-fields.sh — keep this package's COMPOSER structs thin, not regrown
# god-objects. Run from repo root: bash tools/network/check-composer-fields.sh
#
# WHY THIS EXISTS (audit god-module decouple): MoveDispatch was a 27-field god-object
# owning ~7 responsibilities (row tables, seeds, stream wiring, UI state, mover registry,
# layout quantize, dispatch). It was decomposed into single-responsibility owner types
# (rowTables/geomSeeds/streamWiring/uiState/moverRegistry/layoutQuantizer), leaving
# MoveDispatch a thin composer that holds one pointer/value per owner and delegates.
# nodeGeometry was the same shape a layer down — 46 flat fields (channels, clock copy,
# dedicated stream, UI bytes, tilt indices, pair counters, out-wire arrays, neighbour
# tables, scene flags) — and was decomposed the same way into nodeMessaging/nodeClocks/
# nodeStream/nodeUI/nodeTilt/pairReadout/nodeOuts/neighborTopology/sceneFlags/nodeBeads
# (node_geometry_parts.go).
#
# The regrowth failure mode is silent: someone adds a NEW field directly to the composer
# instead of to the owning sub-type, and the god-object creeps back a field at a time.
# This guard caps each composer's field count so that can't happen unnoticed — a genuinely
# new responsibility must land as (or on) an owner type, not as loose composer state.
#
# It is NOT a line-count guard: the point is single-responsibility ownership, not file size.
#
# Adding a composer here: one COMPOSERS row, "<StructDecl>|<max fields>|<owner hint>".
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

# One row per capped composer:
#   <the exact struct declaration line>|<MAX_FIELDS>|<where new state should go instead>
#
# MoveDispatch: owner sub-objects + tr/persist/msgTap/ctx = 12 fields today, AT its cap of
# 12 (the cap this guard has always carried) — the next field there must land on an owner.
# nodeGeometry: id/geom/persistRoot/selfKind/quantOffset/tr + 10 owner sub-objects = 16
# fields today, headroom 20. A little room for legitimate coordination state, but well
# below god-object territory.
COMPOSERS=(
  "type MoveDispatch struct {|12|the responsible owner type (rowtables.RowTables/geomseeds.GeomSeeds/streamWiring/uiState/moverRegistry/layoutQuantizer)"
  "type nodeGeometry struct {|20|the responsible owner type (nodeMessaging/nodeClocks/nodeStream/nodeUI/nodeTilt/pairReadout/nodeOuts/neighborTopology/sceneFlags/nodeBeads, node_geometry_parts.go)"
)

fail=0
for row in "${COMPOSERS[@]}"; do
  decl="${row%%|*}"
  rest="${row#*|}"
  max="${rest%%|*}"
  owners="${rest#*|}"
  name=$(printf '%s' "$decl" | awk '{print $2}')

  # Locate the file containing this struct definition by scanning, rather than hardcoding
  # a path — a same-package file split can move the struct without this guard noticing
  # (memory/feedback_guards_hardcoding_single_file_break_on_split.md).
  # `|| true`: with set -e, a grep that matches nothing would abort the script on this
  # assignment and exit non-zero WITHOUT the MISCONFIGURED report below — a failure whose
  # reason nobody can read. Let it produce an empty FILE and fall into the report instead.
  FILE=$( (grep -rl "^${decl}\$" nodes/Wiring/ --include="*.go" || true) | grep -v _test | head -1 || true)
  if [[ -z "${FILE:-}" || ! -f "$FILE" ]]; then
    echo "check-composer-fields: MISCONFIGURED — could not locate '${decl}' under nodes/Wiring/*.go; refusing vacuous pass" >&2
    fail=1
    continue
  fi

  # Extract the struct body and count FIELD-bearing lines (skip the `type ... {` / closing
  # `}` lines, blank lines, and comment-only lines).
  body=$(awk -v decl="$decl" 'index($0, decl)==1 && !f {f=1;next} f&&/^\}/{f=0} f' "$FILE")
  if [[ -z "$body" ]]; then
    echo "check-composer-fields: MISCONFIGURED — could not locate the $name struct body in $FILE" >&2
    fail=1
    continue
  fi

  count=$(printf '%s\n' "$body" \
    | grep -vE '^[[:space:]]*//' \
    | grep -vE '^[[:space:]]*$' \
    | grep -c .)

  if [[ "$count" -gt "$max" ]]; then
    echo "$name has $count fields (cap $max) — it is regrowing into a god-object."
    echo "  Add new state to $owners,"
    echo "  or introduce a new owner — not a loose $name field. If a new owner legitimately"
    echo "  raises the composer count, bump this composer's cap deliberately with a note."
    fail=1
  fi
done

exit "$fail"
