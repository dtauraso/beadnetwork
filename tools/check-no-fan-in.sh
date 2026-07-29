#!/usr/bin/env bash
set -euo pipefail

# Fan-in is removed from the model: an input port takes exactly ONE incident edge (multiple
# sources into one node use DISTINCT input ports — as every production node does, e.g. a
# gate's FromLeft/FromRight). The loader enforces this at parse (validateNoFanIn, topo_spec.go)
# so a fan-in topology fails at load. This guard is the STATIC repo-side complement: it
# fails the build if the committed production topology (adjacency layout —
# topology/nodes/<source>/edges/*.json, an edge lives under its source node, no top-level
# topology/edges/ dir) has two edges targeting the same target+targetHandle, so a fan-in
# diagram can't be committed and only discovered when someone runs it.
#
# Exit 0 clean, exit 1 with a report.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
NODES_DIR="$REPO_ROOT/topology/nodes"

if [ ! -d "$NODES_DIR" ]; then
  # No topology tree in this checkout — nothing to enforce.
  exit 0
fi

report=$(python3 - "$NODES_DIR" <<'PY'
import json, glob, os, sys, collections
nodes_dir = sys.argv[1]
seen = collections.defaultdict(list)
for f in sorted(glob.glob(os.path.join(nodes_dir, "*", "edges", "*.json"))):
    try:
        d = json.load(open(f))
    except Exception as ex:
        print(f"unreadable edge file {os.path.basename(f)}: {ex}")
        continue
    key = (d.get("target"), d.get("targetHandle"))
    seen[key].append(d.get("label", os.path.basename(f)))
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
