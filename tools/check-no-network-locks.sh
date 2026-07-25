#!/usr/bin/env bash
# check-no-network-locks.sh — forbid sync.Mutex/RWMutex in the concurrent node network.
# Run from repo root: bash tools/check-no-network-locks.sh
#
# WHY THIS EXISTS (audit shared-state cluster): the network's core doctrine is "ownership
# replaces locking" (MODEL.md) — each piece of mutable state is owned by exactly one
# goroutine; cross-goroutine reads use single-writer atomic.Pointer snapshots, never a lock.
# The concurrency audit CONFIRMED zero production mutexes; this guard keeps it that way for
# free, so the invariant holds without a re-audit. An LLM's prior is to add a mutex "for
# safety" on shared-looking state — this fails that at commit time and points at the model.
#
# Scope: production (non-test) Go under nodes/, Buffer/, Trace/ — the runtime network.
# tools/ codegen and other non-network code may use locks legitimately and are NOT scanned.
# sync.WaitGroup / sync.Once / sync/atomic are fine — only Mutex/RWMutex are forbidden.
# Comments are stripped before matching, so prose ABOUT a (deleted) mutex does not trip it.
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
pat = re.compile(r'sync\.(Mutex|RWMutex)\b')
hits = []
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
                    code = line.split("//", 1)[0]  # strip line comment (prose is exempt)
                    if pat.search(code):
                        hits.append(f"{p}:{i}: {line.strip()}")
for h in hits:
    print(h)
PY
)"

if [[ -n "$report" ]]; then
  echo "NETWORK LOCK FORBIDDEN: sync.Mutex/RWMutex found in the node network — the model is"
  echo "'ownership replaces locking' (MODEL.md). Give the state a single owning goroutine and"
  echo "publish cross-goroutine reads via an atomic.Pointer snapshot, not a lock:"
  printf '%s\n' "$report" | sed 's/^/  /'
  exit 1
fi

exit 0
