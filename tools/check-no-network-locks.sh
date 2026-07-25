#!/usr/bin/env bash
# check-no-network-locks.sh — forbid shared-synchronization primitives (sync.Mutex/RWMutex
# AND sync/atomic) in the concurrent node network. Run from repo root:
# bash tools/check-no-network-locks.sh
#
# WHY THIS EXISTS (audit shared-state cluster): the network's model is ownership +
# message-passing with ZERO shared memory — each piece of mutable state is owned by exactly
# one goroutine, and anything another goroutine needs to know is SENT to it as a value, not
# read from a shared location. Under that model a synchronization primitive is not a tool,
# it is a DEFECT MARKER: if code reaches for a mutex OR an atomic, it shared something it
# should have owned. A mutex is forbidden outright. An atomic is forbidden too — needing one
# means a shared cross-goroutine read that should instead be a push-to-owned-copy.
#
# GRANDFATHERED: two atomic sites predate this rule and are KNOWN defects being removed
# (see the ALLOWED_ATOMIC list). The list may only SHRINK — a NEW atomic fails the build,
# and an allowlisted one that is deleted must be dropped from the list (rot-checked below).
#
# Scope: production (non-test) Go under nodes/, Buffer/, Trace/ — the runtime network.
# tools/ codegen and other non-network code are NOT scanned. sync.WaitGroup / sync.Once are
# allowed (they coordinate goroutine lifetime, they do not share mutable state). Comments are
# stripped before matching, so prose about a removed primitive does not trip it.
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

report="$(python3 - <<'PY'
import os, re

roots = ["nodes", "Buffer", "Trace"]
mutex_pat  = re.compile(r'sync\.(Mutex|RWMutex)\b')
atomic_pat = re.compile(r'\batomic\.')

# Known atomic sites that predate the no-atomic rule — DEFECTS being removed, matched by the
# trimmed code text of the line. This list may only shrink. Each is pending a redesign to
# push-to-owned-copies / single-threaded ownership; see the branch descriptions.
#
# Empty as of the atomic #3 removal (node_mover.go's snap atomic.Pointer[centerSnap] —
# redesigned to push-to-owned-copies: bucket-3/quantize reads now use nm.partnerCenters,
# bucket-1/camera reads now use the dispatch goroutine's owned centerMirror). The network
# is atomic-free.
ALLOWED_ATOMIC = set()

mutex_hits = []
atomic_hits = []            # non-allowlisted atomic usage — forbidden
seen_allowed = set()        # allowlisted atomic lines actually found (for rot-check)

for root in roots:
    if not os.path.isdir(root):
        continue
    for dp, _dn, fns in os.walk(root):
        for fn in fns:
            if not fn.endswith(".go") or fn.endswith("_test.go"):
                continue
            p = os.path.join(dp, fn)
            with open(p, encoding="utf-8", errors="replace") as fh:
                for i, line in enumerate(fh, 1):
                    code = line.split("//", 1)[0]          # strip line comment (prose exempt)
                    if mutex_pat.search(code):
                        mutex_hits.append(f"{p}:{i}: {line.strip()}")
                    if atomic_pat.search(code):
                        trimmed = line.strip()
                        if trimmed in ALLOWED_ATOMIC:
                            seen_allowed.add(trimmed)
                        else:
                            atomic_hits.append(f"{p}:{i}: {trimmed}")

out = []
if mutex_hits:
    out.append("MUTEX")
    out += mutex_hits
if atomic_hits:
    out.append("ATOMIC")
    out += atomic_hits
stale = ALLOWED_ATOMIC - seen_allowed
if stale:
    out.append("STALE_ALLOW")
    out += sorted(stale)
print("\n".join(out))
PY
)"

if [[ -z "$report" ]]; then
  exit 0
fi

section=""
fail=0
while IFS= read -r line; do
  case "$line" in
    MUTEX)
      echo "NETWORK MUTEX FORBIDDEN: sync.Mutex/RWMutex in the node network — the model is"
      echo "ownership + message-passing, zero shared memory. Give the state one owning goroutine:"
      fail=1; section=body; continue ;;
    ATOMIC)
      echo "NETWORK ATOMIC FORBIDDEN: a non-allowlisted atomic in the node network. An atomic here"
      echo "is a shared-state DEFECT — a cross-goroutine read that should be a push-to-owned-copy."
      echo "Do not add one; send the value to the goroutine that needs it and let it own its copy:"
      fail=1; section=body; continue ;;
    STALE_ALLOW)
      echo "STALE ALLOWLIST: an ALLOWED_ATOMIC entry no longer exists — remove it (the list must"
      echo "only shrink as these defects are eliminated):"
      fail=1; section=body; continue ;;
    *)
      printf '  %s\n' "$line" ;;
  esac
done <<< "$report"

exit $fail
