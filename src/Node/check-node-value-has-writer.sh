#!/usr/bin/env bash

# PLACEMENT: src/Node/node_values.go | a declared node value must be written by WriteNodeValues, or it crosses as an empty section and the renderer silently reads its fallback

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

python3 - <<'PY'
import re, sys

PATH = "src/Node/node_values.go"

try:
    src = open(PATH, encoding="utf-8").read()
except OSError:
    print(f"check-node-value-has-writer: MISCONFIGURED — {PATH} not found (moved?); "
          f"refusing vacuous pass", file=sys.stderr)
    sys.exit(1)

m = re.search(r"func buildNodeValueNames\(\) \[\]string \{(.*?)\n\}", src, re.S)
if not m:
    print("check-node-value-has-writer: MISCONFIGURED — buildNodeValueNames not found in "
          f"{PATH} (renamed?); refusing vacuous pass", file=sys.stderr)
    sys.exit(1)

declared = set(re.findall(r'"([a-zA-Z0-9]+)"', m.group(1)))

for i in range(16):
    declared.add(f"ringM{i}")

if len(declared) < 20:
    print(f"check-node-value-has-writer: MISCONFIGURED — parsed only {len(declared)} declared "
          f"values; the declaration format changed and this guard would check almost nothing",
          file=sys.stderr)
    sys.exit(1)

w = re.search(r"func WriteNodeValues\(.*?\n\}", src, re.S)
if not w:
    print("check-node-value-has-writer: MISCONFIGURED — WriteNodeValues not found in "
          f"{PATH} (renamed?); refusing vacuous pass", file=sys.stderr)
    sys.exit(1)

body = w.group(0)
written = set(re.findall(r'w\.(?:F32|I32|U8|U32|F64|I64|Bool|Text|Bytes)\("([a-zA-Z0-9]+)"', body))
if "RingName(m)" in body:
    written |= {f"ringM{i}" for i in range(16)}

missing = sorted(declared - written)
if missing:
    for name in missing:
        print(f"NODE VALUE WITH NO WRITER: {name} is declared in NodeValueNames but WriteNodeValues")
        print(f"  never sets it. It crosses as a ZERO-LENGTH section, so the renderer silently reads")
        print(f"  its fallback while packing, decoding and every other check pass. Fix: write it in")
        print(f"  WriteNodeValues, or drop it from NodeValueNames.")
    sys.exit(1)

print(f"check-node-value-has-writer: clean (all {len(declared)} node values are written)")
PY
