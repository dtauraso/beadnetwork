#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: nodes/**/*.go | NewPacedWire must have exactly one non-test production call site, passing DwellTicksPerBead

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

CANONICAL_SPEED="DwellTicksPerBead"

cd "$REPO_ROOT"

CALLS=$(git ls-files -z --cached --others --exclude-standard '*.go' \
  | xargs -0 grep -n "NewPacedWire(" -- \
  | grep -v "_test.go" \
  | grep -v "func NewPacedWire(" \
  || true)

COUNT=$(printf '%s' "$CALLS" | grep -c . || true)

if [[ "$COUNT" -eq 0 ]]; then
  echo "uniform-pulse-speed: EMPTY — no non-test NewPacedWire call site found." >&2
  echo "  The constructor was renamed or removed; refusing a vacuous pass." >&2
  echo "  Update CANONICAL_SPEED / the grep in $0 to match the new shape." >&2
  exit 1
fi

HITS=0

if [[ "$COUNT" -ne 1 ]]; then
  echo "uniform-pulse-speed: expected exactly 1 non-test NewPacedWire call site, found $COUNT:"
  printf '%s\n' "$CALLS" | sed 's/^/  /'
  echo ""
  echo "  Uniform pulse speed is structural ONLY while production builds wires in one place."
  echo "  A second call site makes it a convention instead. Remove the speed parameter from"
  echo "  the production constructor rather than adding another caller."
  HITS=$((HITS + 1))
fi

if ! printf '%s' "$CALLS" | grep -q "$CANONICAL_SPEED"; then
  echo "uniform-pulse-speed: the production call site does not pass $CANONICAL_SPEED:"
  printf '%s\n' "$CALLS" | sed 's/^/  /'
  echo ""
  echo "  DwellTicksPerBead (nodes/wire/lattice/bead_lattice.go) is ticks per bead-step. A different"
  echo "  constant here (e.g. a raw PulseSpeedWuPerMs/PulseSpeedWuPerTick) would silently"
  echo "  desync this wire's timing from the bead-step count the source node's chain is"
  echo "  laid out on."
  HITS=$((HITS + 1))
fi

if [[ $HITS -eq 0 ]]; then
  echo "uniform-pulse-speed: clean"
  exit 0
fi

echo ""
echo "uniform-pulse-speed: $HITS violation(s) found"
exit 1
