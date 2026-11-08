#!/usr/bin/env bash

# PLACEMENT: Categories/Node/Links/*.go, Categories/Chrome/Pills/Builder/*.go | Δ and increment are whole numbers (Ints) and s only multiplies: no float Δ/increment, no division by s

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

report="$(python3 - <<'PY'
import os, re

hits = []
link = "Categories/Node/Links/link.go"
src = open(link, encoding="utf-8").read()
if not re.search(r'^type Ints struct\{ Phi, Theta, R, S int \}$', src, re.M):
    hits.append(f"{link}: Ints is not `struct{{ Phi, Theta, R, S int }}` — Δ and increment numbers are whole")
for field in ("Delta", "Increment"):
    if not re.search(rf'^\s+{field}\s+Ints$', src, re.M):
        hits.append(f"{link}: Link.{field} is not Ints")

divide = re.compile(r'/\s*(float64\()?\s*(s|S|[\w.]*\.S|ShownS\([^)]*\)|ScaleOf\([^)]*\)|SOf\([^)]*\))\b\)?')
for root in ("Categories/Node/Links", "Categories/Chrome/Pills/Builder", "Categories/Scene/Dispatch"):
    for fn in sorted(os.listdir(root)):
        if not fn.endswith(".go"):
            continue
        p = os.path.join(root, fn)
        for i, line in enumerate(open(p, encoding="utf-8"), 1):
            code = line.split("//", 1)[0]
            if divide.search(code):
                hits.append(f"{p}:{i}: divides by s: {line.strip()}")
print("\n".join(hits))
PY
)"

if [ -n "$report" ]; then
  echo "check-no-divide-by-s: s only multiplies (shown = number × s); a number is never divided by s and Δ/increment are never floats:" >&2
  echo "$report" >&2
  exit 1
fi
