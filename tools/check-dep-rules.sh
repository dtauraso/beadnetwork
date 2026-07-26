#!/usr/bin/env bash
# check-dep-rules.sh — fail if a node-kind package imports a SIBLING node-kind package.
# Run from repo root: bash tools/check-dep-rules.sh
#
# WHY THIS EXISTS (audit-integrate-into-repo-systems): the blast-radius audit found that
# node packages are singly-owned and DO NOT import each other today — every kind depends
# only on the shared spine (nodes/Wiring, nodes/gatecommon, nodes/wire). That is a leanness
# win worth locking: a cross-kind import would couple two kinds so a change to one must load
# the other, widening blast radius. This guard keeps the property true for free (zero AI
# tokens).
#
# Ported from the Uncle-Bob raid keeper: dependency-checker's rule MODEL
# (~/Downloads/unclebob-repos/dependency-checker + empire-2025/dependency-checker.edn) —
# an allowed-dependency map plus fail-on-illegal-edge. dependency-checker itself is Clojure
# and does not run on Go; this is the portable rule model expressed as a native Go-repo guard.
#
# The rule (allowed-dependencies): every package under nodes/<Kind>/ may import ONLY
#   nodes/Wiring, nodes/gatecommon, nodes/wire   (plus stdlib / external — not our concern here)
# and NEVER a sibling nodes/<OtherKind>/. Wiring, gatecommon, and wire (the node-facing leaf
# extracted from Wiring by task/wiring-decompose — clock/ports/PacedWire/LayoutHolder/
# RowEvent/Register) are the shared spine and are exempt (they are depended-UPON, not kinds).
#
# Exit 0 clean (empty), exit 1 with a report — matches scripts/stop-checks.sh guard-loop
# contract (auto-discovered via tools/check-*.sh glob).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

NODES_DIR="$REPO_ROOT/nodes"
if [ ! -d "$NODES_DIR" ]; then
  # No node packages in this checkout — nothing to enforce.
  exit 0
fi

MODULE="github.com/dtauraso/wirefold"

# Shared spine: allowed as an import target from any kind, and not itself a "kind".
is_spine() { [ "$1" = "Wiring" ] || [ "$1" = "gatecommon" ] || [ "$1" = "wire" ]; }

fail=0
for dir in "$NODES_DIR"/*/; do
  kind="$(basename "$dir")"
  is_spine "$kind" && continue

  # Every internal import this kind's .go files make, reduced to the package name under nodes/.
  imported="$(grep -rhoE "\"$MODULE/nodes/[A-Za-z0-9_]+\"" "$dir" --include="*.go" 2>/dev/null \
      | sed -E "s#\"$MODULE/nodes/([A-Za-z0-9_]+)\"#\1#" | sort -u || true)"

  while IFS= read -r dep; do
    [ -z "$dep" ] && continue
    [ "$dep" = "$kind" ] && continue      # self (subpackage) — fine
    is_spine "$dep" && continue           # shared spine — allowed
    echo "ILLEGAL DEP: nodes/$kind imports sibling nodes/$dep — kinds must couple only through the shared spine (Wiring/gatecommon)"
    fail=1
  done <<< "$imported"
done

exit $fail
