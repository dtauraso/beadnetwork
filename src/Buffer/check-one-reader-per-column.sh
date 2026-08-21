#!/usr/bin/env bash

# PLACEMENT: src/Buffer/buffer-layout*.ts | every generated buffer read* helper is a channel: one writer, one reader. A second reader means a consumer is re-deriving something Go should send it.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

source "$REPO_ROOT/scripts/lib/ts-roots.sh"

LAYOUT_FILES=()
for LAYOUT in "src/Buffer/buffer-layout.ts" \
              src/Buffer/buffer-layout-rows*-gen.ts \
              "src/Buffer/buffer-layout-singletons-gen.ts"; do
  [[ -f "$LAYOUT" ]] && LAYOUT_FILES+=("$LAYOUT")
done
if [[ ! -f "src/Buffer/buffer-layout.ts" ]] || (( ${#LAYOUT_FILES[@]} < 2 )); then
  echo "check-one-reader-per-column: MISCONFIGURED — found ${#LAYOUT_FILES[@]} layout file(s) under Buffer/ (renamed?); refusing vacuous pass" >&2
  exit 1
fi

export TS_ROOTS_JOINED="${TS_ROOTS[*]}"
export LAYOUT_JOINED="${LAYOUT_FILES[*]}"

python3 - <<'PY'
import os, re, sys, pathlib

roots = os.environ["TS_ROOTS_JOINED"].split()
layouts = [pathlib.Path(p) for p in os.environ["LAYOUT_JOINED"].split()]

WRITERS = {
    "src/extension/runner/stream-demux.ts",
}

OBSERVERS = {
    "src/webview/decode/decode-event-line.ts",
    "src/Node/Edge/check-edge-lands-on-node.ts",

    "src/webview/main.tsx",
}

RATCHET = {}

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
        if s in layout_set or s in OBSERVERS or s in WRITERS:
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
print(f"check-one-reader-per-column: clean ({len(readers)} read* helpers, "
      f"{len(RATCHET)} on the ratchet)")
PY
