#!/usr/bin/env bash

# PLACEMENT: tools/topology-vscode/Buffer/buffer-layout*.ts | every buffer column is a channel: one writer, one reader. A second reader means a consumer is re-deriving something Go should send it.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

source "$REPO_ROOT/tools/lib/ts-roots.sh"

LAYOUT_FILES=(
  "tools/topology-vscode/Buffer/buffer-layout.ts"
  "tools/topology-vscode/Buffer/buffer-layout-rows-gen.ts"
  "tools/topology-vscode/Buffer/buffer-layout-rows2-gen.ts"
  "tools/topology-vscode/Buffer/buffer-layout-singletons-gen.ts"
)
for LAYOUT in "${LAYOUT_FILES[@]}"; do
  if [[ ! -f "$LAYOUT" ]]; then
    echo "check-one-reader-per-column: MISCONFIGURED — $LAYOUT not found (renamed?); refusing vacuous pass" >&2
    exit 1
  fi
done

export TS_ROOTS_JOINED="${TS_ROOTS[*]}"
export LAYOUT_JOINED="${LAYOUT_FILES[*]}"

python3 - <<'PY'
import os, re, sys, pathlib

roots = os.environ["TS_ROOTS_JOINED"].split()
layouts = [pathlib.Path(p) for p in os.environ["LAYOUT_JOINED"].split()]

OBSERVERS = {
    "tools/topology-vscode/src/webview/three/decode/decode-event-line.ts",
    "tools/topology-vscode/src/webview/three/decode/decode-event-node-geometry.ts",
    "tools/topology-vscode/src/webview/three/decode/decode-event-overlay.ts",
}

RATCHET = {
    # DRIFT
    "readNodeCX": 8,
    "readNodeCY": 8,
    "readNodeCZ": 8,
    "readNodeRadius": 4,
    "readNodeSelected": 3,
    "readNodeKindId": 2,
    "readNodeTopTiltVectorLen": 2,
    "readOverlayHoverRing": 2,
    "readOverlayNodeBody": 2,
    "readOverlayNodeRing": 2,
    "readOverlayRingPick": 2,
    "readOverlayRuleChannels": 2,
    "readOverlaySelectionRing": 2,
}

readers = set()
for p in layouts:
    readers |= set(re.findall(r"export function (read[A-Za-z0-9_]+)", p.read_text(encoding="utf-8")))
if not readers:
    print("check-one-reader-per-column: MISCONFIGURED — parsed 0 read* helpers; format changed, "
          "guard would check nothing", file=sys.stderr)
    sys.exit(1)

layout_set = {str(p) for p in layouts}
files = []
for root in roots:
    for f in pathlib.Path(root).rglob("*"):
        if f.suffix not in (".ts", ".tsx"):
            continue
        s = str(f)
        if "node_modules" in s or "/out/" in s or "/test/" in s or s.endswith(".test.ts"):
            continue
        if s in layout_set or s in OBSERVERS:
            continue
        files.append(f)

STRIP = re.compile(r"/\*.*?\*/|//[^\n]*", re.S)
word = {}
for f in files:
    text = STRIP.sub("", f.read_text(encoding="utf-8", errors="replace"))
    for tok in set(re.findall(r"[A-Za-z0-9_]+", text)):
        word.setdefault(tok, []).append(str(f))

fail = False
for fn in sorted(readers):
    who = word.get(fn, [])
    n = len(who)
    allowed = RATCHET.get(fn)
    if allowed is not None:
        if n == allowed:
            continue
        fail = True
        if n < allowed:
            print(f"RATCHET STALE: {fn} now has {n} reader(s), but the ratchet still says {allowed}.")
            print(f"  Lower it, or delete the entry entirely if {fn} is down to one reader.")
        else:
            print(f"RATCHET BROKEN: {fn} has {n} readers, up from the {allowed} recorded.")
            for w in sorted(who):
                print(f"    {w}")
            print(f"  A new reader means a consumer is deriving from this column instead of being")
            print(f"  sent what it draws. Send it the thing, do not raise the ratchet.")
        continue
    if n > 1:
        fail = True
        print(f"COLUMN WITH {n} READERS: {fn} is read by more than one consumer.")
        for w in sorted(who):
            print(f"    {w}")
        print(f"  A column is a channel: one writer, one reader. A second reader means that")
        print(f"  consumer is re-deriving geometry Go should send it directly.")

if fail:
    sys.exit(1)
print(f"check-one-reader-per-column: clean ({len(readers)} columns, {len(RATCHET)} on the ratchet)")
PY
