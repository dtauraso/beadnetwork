#!/usr/bin/env bash

# PLACEMENT: src/Node/node_state.go,src/runtopology/node_stream.go | every NodeState field must be named by the adapter that fills it, or it writes a zero into the node file while everything compiles

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

python3 - <<'PY'
import re, sys

PAIRS = [
    ("src/Node/nodeactor/nodeframe/node_frame_input.go", "NodeFrameInput",
     "src/Node/nodeactor/node_geometry_stream.go"),
    ("src/Node/node_state.go", "NodeState",
     "src/runtopology/node_stream.go"),
]

EXEMPT = {
    "NodeState": {
        "NodeRow": "the adapter passes it as f.NodeRow under a different literal key",
    },
}

fail = False
for path, struct, adapter_path in PAIRS:
    src = open(path, encoding="utf-8").read()
    m = re.search(r"type " + struct + r" struct \{(.*?)\n\}", src, re.S)
    if not m:
        print(f"check-frame-fields-copied: MISCONFIGURED — {struct} not found in {path} "
              f"(renamed?); refusing vacuous pass", file=sys.stderr)
        sys.exit(1)

    fields = []
    for line in m.group(1).split("\n"):
        line = line.split("//")[0].strip()
        if not line or line.startswith("}"):
            continue
        mm = re.match(r"^([A-Za-z0-9_]+(?:\s*,\s*[A-Za-z0-9_]+)*)\s+[\w\[\]\*\.]+$", line)
        if mm:
            fields.extend(n.strip() for n in mm.group(1).split(","))
    if not fields:
        print(f"check-frame-fields-copied: MISCONFIGURED — parsed 0 fields out of {struct}; "
              f"the struct format changed and this guard would check nothing", file=sys.stderr)
        sys.exit(1)

    try:
        adapter = open(adapter_path, encoding="utf-8").read()
    except OSError:
        print(f"check-frame-fields-copied: MISCONFIGURED — adapter {adapter_path} not found "
              f"(moved?); refusing vacuous pass", file=sys.stderr)
        sys.exit(1)
    named = set(re.findall(r"([A-Za-z0-9_]+)\s*:", adapter))

    for f in fields:
        if f in named or f in EXEMPT.get(struct, {}):
            continue
        fail = True
        print(f"FRAME FIELD NEVER FILLED: {struct}.{f} is not named in {adapter_path}.")
        print(f"  Go zero-fills it, so the build passes, the types are right, and the field")
        print(f"  streams empty. Set it in the {adapter_path} literal, or exempt it here with")
        print(f"  a reason if the adapter fills it some other way.")

if fail:
    sys.exit(1)
print("check-frame-fields-copied: clean (every frame field is named by its adapter)")
PY
