#!/usr/bin/env python3
"""check-no-wall-clock-wait.py — worker for check-no-wall-clock-wait.sh.

See that script's header comment for the full rationale. This file just does the
source scan; the .sh wrapper formats the report and sets the exit code, matching the
shape of check-no-network-locks.sh.
"""
import os
import re

roots = ["nodes"]
wait_pat = re.compile(r'\btime\.(Sleep|After|NewTicker)\s*\(')

EXEMPT_FILES = {os.path.join("nodes", "wire", "clock.go")}

# Known pre-existing sites that are NOT mover/node tick-pacing waits — each is a distinct,
# short-lived, already-documented poll unrelated to the clock/SleepCycle mechanism this rule
# governs. The list may only SHRINK.
ALLOWED = {
    # waitForCenterSettle polls another goroutine's already-written position after a
    # synchronous dispatch call, bounded by a 200ms deadline -- not a per-cycle pacing loop,
    # and already flagged in its own doc comment for deletion once dispatch stops measuring
    # across goroutines (see the comment above ApplyDistanceGroupTarget's call site).
    ("nodes/Wiring/distance_groups.go", "time.Sleep(time.Millisecond)"),
}

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
                    code = line.split("//", 1)[0]  # strip line comment (prose exempt)
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
