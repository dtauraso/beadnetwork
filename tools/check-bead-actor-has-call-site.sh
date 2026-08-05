#!/usr/bin/env bash
# check-bead-actor-has-call-site.sh — fail if the chain-bead actor primitive
# (nodes/wire/bead_actor.go, nodes/wire/bead_wake_group.go) has NO production call site.
#
# PLACEMENT: nodes/wire/bead_actor.go,nodes/wire/bead_wake_group.go | must have a production call site outside nodes/wire and _test.go
# Run from repo root: bash tools/check-bead-actor-has-call-site.sh
#
# WHY THIS EXISTS: the primitive was built and tested in isolation (bead_actor_test.go)
# for one whole commit with nothing in the running editor constructing a Bead or a
# BeadWakeGroup — `go build`, `go test`, and every other guard stayed green throughout,
# because "primitive exists and is tested" and "primitive is wired into the live path"
# are different facts and nothing else checks the second one. This guard closes that: it
# fails unless at least one symbol from the primitive is referenced from PRODUCTION Go
# source (not nodes/wire itself, not a _test.go file) — today that call site is
# nodes/Wiring/bead_chain.go's reconcileBeadChain/startBeadDrag/endBeadDrag.
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Symbols that only exist to be called from a real production wake/geometry/mode path —
# constructing a Bead or a BeadWakeGroup and never driving it would still trip this list,
# since StartDrag/BroadcastGeometry/EndDrag are the three verbs the lifecycle actually
# needs (PLAN.md's lifecycle section, now MODEL.md).
readonly SYMBOLS=(
  "wire.NewBead("
  "wire.NewBeadWakeGroup("
  ".StartDrag("
  ".BroadcastGeometry("
  ".EndDrag("
)

# Every non-test .go file OUTSIDE nodes/wire (the primitive's own package — a self-
# reference there proves nothing about a production caller elsewhere) is the corpus.
prod_files=()
while IFS= read -r f; do prod_files+=("$f"); done < <(
  find . -type f -name '*.go' \
    -not -path './nodes/wire/*' \
    -not -name '*_test.go' \
    -not -path '*/node_modules/*' \
    -not -path '*/.git/*' 2>/dev/null
)

if [[ ${#prod_files[@]} -eq 0 ]]; then
  echo "check-bead-actor-has-call-site: MISCONFIGURED — found 0 candidate .go files; refusing vacuous pass" >&2
  exit 1
fi

fail=1
for sym in "${SYMBOLS[@]}"; do
  if grep -qF "$sym" "${prod_files[@]}" 2>/dev/null; then
    fail=0
    break
  fi
done

if [[ "$fail" -eq 1 ]]; then
  echo "BEAD-ACTOR PRIMITIVE HAS NO PRODUCTION CALL SITE: none of ${SYMBOLS[*]} appear outside nodes/wire or a _test.go file."
  echo "  The chain-bead actor (nodes/wire/bead_actor.go, bead_wake_group.go) has drifted back to a validated-in-isolation primitive with nothing driving it in the live editor."
  echo "  Fix: wire it back into nodes/Wiring's node drag path (see bead_chain.go's history), or if it is being deliberately retired, delete the primitive and its tests together instead of leaving them dead."
fi

exit $fail
