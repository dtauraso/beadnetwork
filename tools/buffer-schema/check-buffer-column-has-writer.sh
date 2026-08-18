#!/usr/bin/env bash

# PLACEMENT: tools/topology-vscode/Buffer/bufschema/layout_node.go,tools/topology-vscode/Buffer/bufschema/layout_overlay.go,tools/topology-vscode/Buffer/bufschema/layout_panel.go,runtopology/node_stream.go,runtopology/view_stream.go | a Node/Overlay/Panel column must be named by the runtopology adapter that fills its row, or it silently streams zeros

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

python3 - <<'PY'
import re, sys

PAIRS = [
    ("tools/topology-vscode/Buffer/bufschema/layout_node.go",    "bufLayoutNode",    "runtopology/node_stream.go"),
    ("tools/topology-vscode/Buffer/bufschema/layout_overlay.go", "bufLayoutOverlay", "runtopology/view_stream.go"),
    ("tools/topology-vscode/Buffer/bufschema/layout_panel.go",   "bufLayoutPanel",   "runtopology/view_stream.go"),
]

DERIVED = [
    (re.compile(r"^RingM\d+$"),        "RingMatrix", "the whole matrix rides one [16]float32 field"),
    (re.compile(r"^Label(Off|Len)$"),  "Label",      "the offset and length are computed from the label bytes"),
]

def columns(path, struct):
    src = open(path, encoding="utf-8").read()
    m = re.search(r"type " + struct + r" struct \{(.*?)\n\}", src, re.S)
    if not m:
        print(f"check-buffer-column-has-writer: MISCONFIGURED — {struct} not found in {path} "
              f"(renamed?); refusing vacuous pass", file=sys.stderr)
        sys.exit(1)
    out = []
    for line in m.group(1).split("\n"):
        mm = re.match(r"\s*([A-Za-z0-9_, ]+?)\s+\S+\s+`buf:", line)
        if mm:
            out.extend(n.strip() for n in mm.group(1).split(",") if n.strip())
    if not out:
        print(f"check-buffer-column-has-writer: MISCONFIGURED — parsed 0 `buf:` columns out of "
              f"{struct}; the struct format changed and this guard would check nothing", file=sys.stderr)
        sys.exit(1)
    return out

GENERATED_GO = "tools/topology-vscode/Buffer/buffer_layout_gen_singletons.go"
try:
    generated = open(GENERATED_GO, encoding="utf-8").read()
except OSError:
    print(f"check-buffer-column-has-writer: MISCONFIGURED — {GENERATED_GO} not found (renamed?); "
          f"cannot tell which blocks still have rows", file=sys.stderr)
    sys.exit(1)

def has_row(struct):
    block = struct.replace("bufLayout", "")
    return f"func Set{block}Row(" in generated

fail = False
for schema_path, struct, adapter_path in PAIRS:
    if not has_row(struct):
        continue
    try:
        adapter = {w.lower() for w in re.findall(r"[A-Za-z0-9_]+",
                                                 open(adapter_path, encoding="utf-8").read())}
    except OSError:
        print(f"check-buffer-column-has-writer: MISCONFIGURED — adapter {adapter_path} not found "
              f"(moved?); refusing vacuous pass", file=sys.stderr)
        sys.exit(1)

    for col in columns(schema_path, struct):
        if col.lower() in adapter:
            continue
        source = next((s for pat, s, _ in DERIVED if pat.match(col)), None)
        if source and source.lower() in adapter:
            continue
        fail = True
        print(f"BUFFER COLUMN WITH NO WRITER: {struct}.{col} is never named in {adapter_path}.")
        print(f"  That literal sets the row field by field, so the column streams a zero value while")
        print(f"  packing, decoding and every other check pass. Fix: set it in the {adapter_path}")
        print(f"  literal, or derive it there from a field that IS set (see DERIVED in this guard).")
        print(f"  check-no-dead-buffer-column stays GREEN through this: it guards the CONSUMER end")
        print(f"  (a TS reader must exist), and the column is unwritten at the PRODUCER end.")

if fail:
    sys.exit(1)
print("check-buffer-column-has-writer: clean (every Node/Overlay/Panel column is named by its adapter)")
PY
