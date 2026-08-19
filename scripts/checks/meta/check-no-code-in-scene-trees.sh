#!/usr/bin/env bash

# PLACEMENT: none | a scene tree is data the loader enumerates; one stray script in it and the loader refuses the whole scene, which presents as a blank editor

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

python3 - <<'PY'
import json
import pathlib
import sys

scenes = sorted(p for p in pathlib.Path(".").iterdir()
                if p.is_dir() and (p / "counts").is_dir() and (p / "view").is_dir())

if not scenes:
    print("check-no-code-in-scene-trees: MISCONFIGURED — found no scene tree (a directory "
          "with counts/ and view/); this would check nothing", file=sys.stderr)
    sys.exit(1)

CODE = {".sh", ".py", ".go", ".ts", ".tsx", ".mjs", ".js"}

offenders = []
for scene in scenes:
    for f in scene.rglob("*"):
        if f.is_file() and f.suffix in CODE:
            offenders.append(f)

bad_ids = []
for scene in scenes:
    nodes = scene / "nodes"
    if not nodes.is_dir():
        continue
    for entry in nodes.iterdir():
        if not entry.name.isdigit():
            bad_ids.append(entry)

if offenders or bad_ids:
    print("CODE INSIDE A SCENE TREE — the loader reads these directories as data:")
    for f in sorted(set(offenders) | set(bad_ids)):
        print(f"  {f}")
    print()
    print("  loadTree parses every entry under <scene>/nodes/ as a node id and refuses the")
    print("  whole scene when one does not, so a single script dropped in here takes the")
    print("  editor down to a blank window with the reason only in .probe/go-errors.jsonl.")
    print("  A guard belongs next to the CODE that depends on the data, not inside the data.")
    sys.exit(1)

counted = sum(1 for s in scenes for _ in s.rglob("*"))
print(f"check-no-code-in-scene-trees: clean ({len(scenes)} scene(s), {counted} entries, no code)")
PY
