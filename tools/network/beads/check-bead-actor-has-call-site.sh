#!/usr/bin/env bash



# PLACEMENT: nodes/wire/beadchain/bead_actor.go,nodes/wire/beadchain/bead_wake_group.go | must have a production call site outside nodes/wire and _test.go













set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"





readonly SYMBOLS=(
  "beadchain.NewBead("
  "beadchain.NewBeadWakeGroup("
  ".StartDrag("
  ".BroadcastGeometry("
  ".EndDrag("
)



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
  echo "  The chain-bead actor (nodes/wire/beadchain/bead_actor.go, bead_wake_group.go) has drifted back to a validated-in-isolation primitive with nothing driving it in the live editor."
  echo "  Fix: wire it back into nodes/Wiring's node drag path (see bead_chain.go's history), or if it is being deliberately retired, delete the primitive and its tests together instead of leaving them dead."
fi

exit $fail
