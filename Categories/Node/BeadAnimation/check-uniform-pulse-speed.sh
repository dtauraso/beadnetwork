#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: Categories/Node/**/*.go | NewBeadLine must take no timing argument, so pulse speed cannot vary per line

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

cd "$REPO_ROOT"

DECL=$(git ls-files -z --cached --others --exclude-standard '*.go' \
  | xargs -0 grep -n "func NewBeadLine(" -- \
  | grep -v "_test.go" \
  || true)

if [[ -z "$DECL" ]]; then
  echo "uniform-pulse-speed: EMPTY — no NewBeadLine declaration found." >&2
  echo "  The constructor was renamed or removed; refusing a vacuous pass." >&2
  echo "  Update the grep in $0 to match the new shape." >&2
  exit 1
fi

PARAMS=$(printf '%s' "$DECL" | sed -n 's/.*func NewBeadLine(\([^)]*\)).*/\1/p')

if [[ -n "$PARAMS" ]]; then
  echo "uniform-pulse-speed: NewBeadLine takes an argument, so a run can be given its own speed:"
  printf '%s\n' "$DECL" | sed 's/^/  /'
  echo ""
  echo "  Pulse speed is uniform across every bead run. It is structural only while the"
  echo "  constructor cannot express a per-run value -- a parameter makes it a convention"
  echo "  that each call site has to honour. How FAST beads move is the animation"
  echo "  goroutine's sleep (one round per human-speed cycle); how MANY slots they visit"
  echo "  is the radial index. Neither belongs on the run."
  echo ""
  echo "uniform-pulse-speed: 1 violation(s) found"
  exit 1
fi

echo "uniform-pulse-speed: clean — NewBeadLine takes no timing argument"
