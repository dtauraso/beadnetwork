#!/usr/bin/env bash

# PLACEMENT: tools/topology-vscode/Buffer/streamframe/node_stream_frame.go,runtopology/node_stream.go | every NodeStreamFrame field must be named by the adapter that fills it, or it streams a zero while everything compiles

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

python3 - <<'PY'
import re, sys

# BOTH hops. A node's values pass through two field-by-field literals on their way to the
# wire -- the geometry actor fills NodeFrameInput, runtopology copies it into
# NodeStreamFrame -- and a field dropped at EITHER hop streams a zero just as silently.
# Checking only the second hop is what let LabelAnchor go quiet at the first: the labels
# projected to the scene origin and vanished, with the build green.
PAIRS = [
    ("nodes/Wiring/nodeactor/nodeframe/node_frame_input.go", "NodeFrameInput",
     "nodes/Wiring/nodeactor/node_geometry_stream.go"),
    ("tools/topology-vscode/Buffer/streamframe/node_stream_frame.go", "NodeStreamFrame",
     "runtopology/node_stream.go"),
]

EXEMPT = {
    "NodeStreamFrame": {
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
