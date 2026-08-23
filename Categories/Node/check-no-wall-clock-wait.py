#!/usr/bin/env python3
"""check-no-wall-clock-wait.py — worker for check-no-wall-clock-wait.sh.

See that script's header comment for the full rationale. This file just does the
source scan; the .sh wrapper formats the report and sets the exit code, matching the
shape of check-no-network-locks.sh.
"""
import os
import re

roots = ["Categories/Node", "Categories/NodeKinds", "Categories/Ring/Bead", "Categories/Clock"]
wait_pat = re.compile(r'\btime\.(Sleep|After|NewTicker)\s*\(')

EXEMPT_FILES = {
    os.path.join("Categories/Clock", "clock.go"),
    os.path.join("Categories/Clock", "pulse.go"),
}

ALLOWED = set()

hits = []
seen_allowed = set()

for root in roots:
    if not os.path.isdir(root):
        continue
    for dp, _dn, fns in os.walk(root):
        for fn in fns:
            if not fn.endswith(".go") or fn.endswith("_test.go"):
                continue
            p = os.path.join(dp, fn)
            rel = os.path.relpath(p, ".")
            if rel in EXEMPT_FILES:
                continue
            with open(p, encoding="utf-8", errors="replace") as fh:
                for i, line in enumerate(fh, 1):
                    code = line.split("//", 1)[0]
                    if wait_pat.search(code):
                        trimmed = code.strip()
                        key = (rel, trimmed)
                        if key in ALLOWED:
                            seen_allowed.add(key)
                        else:
                            hits.append(f"{p}:{i}: {line.strip()}")

out = []
if hits:
    out.append("WAIT")
    out += hits
stale = ALLOWED - seen_allowed
if stale:
    out.append("STALE_ALLOW")
    out += [f"{f}: {c}" for f, c in sorted(stale)]
print("\n".join(out))
