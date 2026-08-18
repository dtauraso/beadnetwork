#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: topology/nodes/*/edges/*/target.json | two committed edges may not target the same target+targetHandle (fan-in)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
NODES_DIR="$REPO_ROOT/topology/nodes"

if [ ! -d "$NODES_DIR" ]; then
  echo "check-no-fan-in: MISCONFIGURED — $NODES_DIR not found; refusing a vacuous pass." >&2
  echo "  This guard exists to catch fan-in in the committed production topology. A real" >&2
  echo "  incident: this dir moved from topology/edges to topology/nodes/<source>/edges and" >&2
  echo "  the guard silently passed for several commits scanning a dir that no longer" >&2
  echo "  existed. If topology/nodes/ legitimately moved again, update NODES_DIR deliberately." >&2
  exit 1
fi

edge_count=$(find "$NODES_DIR" -mindepth 3 -maxdepth 3 -type d -path '*/edges/*' | wc -l | tr -d ' ')
if [ "$edge_count" -eq 0 ]; then
  echo "check-no-fan-in: MISCONFIGURED — 0 edge dirs found under $NODES_DIR/*/edges/*/." >&2
  echo "  The scan must actually see real edges; refusing a vacuous pass. If the committed" >&2
  echo "  topology legitimately has no edges yet, allowlist that state deliberately instead" >&2
  echo "  of letting this guard read as clean by accident." >&2
  exit 1
fi

report=$(python3 - "$NODES_DIR" <<'PY'
import json, glob, os, sys, collections
nodes_dir = sys.argv[1]
seen = collections.defaultdict(list)
def read(edge_dir, name):
    try:
        return json.load(open(os.path.join(edge_dir, name)))
    except Exception:
        return None

for edge_dir in sorted(glob.glob(os.path.join(nodes_dir, "*", "edges", "*"))):
    if not os.path.isdir(edge_dir):
        continue
    key = (read(edge_dir, "target.json"), read(edge_dir, "target-handle.json"))
    seen[key].append(os.path.basename(edge_dir))
for (target, handle), labels in sorted(seen.items()):
    if len(labels) > 1:
        print(f"fan-in: edges {labels} all target input port {target}.{handle}")
PY
)

if [ -n "$report" ]; then
  echo "check-no-fan-in: an input port takes exactly one edge (fan-in was removed from the model)."
  echo "Use distinct input ports for multiple sources into one node. Offending ports:"
  echo "$report"
  exit 1
fi

exit 0
