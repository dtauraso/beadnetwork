#!/usr/bin/env bash

# PLACEMENT: src/Input/dispatch/move_dispatch_api.go,src/Node/Wiring/nodeactor/node_geometry.go | a composer struct (MoveDispatch, nodeactor.NodeGeometry) stays THIN: new state belongs in a named sub-struct, not a new loose field

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

COMPOSERS=(
  "type MoveDispatch struct {|12|the responsible owner type (rowtables.RowTables/geomseeds.GeomSeeds/streamWiring/uiState/moverRegistry/nodeMover)"
  "type NodeGeometry struct {|22|the responsible owner type (owners.Messaging/Clocks/Stream/UI/Tilt/Readout/Outs/Topology/Flags/Beads/Deltas/ChannelVectors, src/Node/Wiring/nodeactor/owners/)"
)

fail=0
for row in "${COMPOSERS[@]}"; do
  decl="${row%%|*}"
  rest="${row#*|}"
  max="${rest%%|*}"
  owners="${rest#*|}"
  name=$(printf '%s' "$decl" | awk '{print $2}')

  FILE=$( (grep -rl "^${decl}\$" src/Input/ src/Node/Wiring/ --include="*.go" || true) | grep -v _test | head -1 || true)
  if [[ -z "${FILE:-}" || ! -f "$FILE" ]]; then
    echo "check-composer-fields: MISCONFIGURED — could not locate '${decl}' under src/Input/ or src/Node/Wiring/**/*.go; refusing vacuous pass" >&2
    fail=1
    continue
  fi

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
