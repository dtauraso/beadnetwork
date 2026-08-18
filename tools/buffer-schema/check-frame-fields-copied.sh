#!/usr/bin/env bash

# PLACEMENT: tools/topology-vscode/Buffer/streamframe/node_stream_frame.go,runtopology/node_stream.go | every NodeStreamFrame field must be named by the adapter that fills it, or it streams a zero while everything compiles

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

python3 - <<'PY'
import re, sys

# The frame struct and the literal that fills it, field by field. A field added to the
# struct but not to the literal is not a compile error -- Go zero-fills it -- so the
# feature ships silently dead: composed, typed, packed, and empty on the wire. That has
# happened three times (LabelAnchor, and ChannelVectors twice), each time because a bulk
# substitution matched nothing and nothing said so.
PAIRS = [
    ("tools/topology-vscode/Buffer/streamframe/node_stream_frame.go", "NodeStreamFrame",
     "runtopology/node_stream.go"),
]

# Fields the adapter legitimately does not name, with the reason.
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
