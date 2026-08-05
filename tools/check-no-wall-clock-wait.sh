#!/usr/bin/env bash
# check-no-wall-clock-wait.sh — forbid time.Sleep/time.After/time.NewTicker anywhere in
# the node network EXCEPT the single clock goroutine's own implementation. Run from repo
# root: bash tools/check-no-wall-clock-wait.sh
#
# WHY THIS EXISTS (PLAN.md "No sleeping" / two-clock-beads Phase A): a goroutine parked in
# time.After (or time.Sleep, or blocked on a time.Ticker) cannot service any of its other
# channels for the duration of the wait — that stall is exactly what the two-clock model
# must not have. The fix is ONE clock goroutine (nodes/wire/clock.go's TickBroadcaster) that
# is the only thing in the process that ever waits on wall time; every pacing loop instead
# blocks on RECEIVE from a dedicated channel that goroutine pushes to
# (RealClock.SleepCycle / wire.NewTickChan). A NEW time.Sleep/After/NewTicker anywhere else
# in the network re-adds a wall-time wait outside that one goroutine, silently reintroducing
# the stall this file guards against.
#
# Scope: production (non-test) Go under nodes/ — the runtime network. tools/ codegen and
# other non-network code are NOT scanned. Comments are stripped before matching, so prose
# about the removed pattern does not trip it.
#
# EXEMPT: nodes/wire/clock.go — this is the single clock goroutine's own implementation
# (TickBroadcaster.run's time.NewTicker); it is what everything else routes through.
#
# ALLOWLIST (may only shrink): sites that predate this rule and are NOT on the tick-pacing
# path — see the comment for each. A new entry must justify why it is not a mover/node
# pacing wait; anything that paces a network goroutine belongs on the tick channel instead.
#
# PLACEMENT: nodes/**/*.go | no time.Sleep/time.After/time.NewTicker outside nodes/wire/clock.go; block on the tick channel instead
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

report="$(python3 "$SCRIPT_DIR/check-no-wall-clock-wait.py")"

if [[ -z "$report" ]]; then
  exit 0
fi

section=""
fail=0
while IFS= read -r line; do
  case "$line" in
    WAIT)
      echo "WALL-CLOCK WAIT OUTSIDE THE CLOCK GOROUTINE: time.Sleep/time.After/time.NewTicker"
      echo "found outside nodes/wire/clock.go. A goroutine parked here cannot service its other"
      echo "channels for the wait — route through wire.NewRealClock()'s SleepCycle or"
      echo "wire.NewTickChan() instead, both backed by the process's one TickBroadcaster:"
      fail=1; section=body; continue ;;
    STALE_ALLOW)
      echo "STALE ALLOWLIST: an entry in check-no-wall-clock-wait.sh's ALLOWED set no longer"
      echo "exists — remove it (the list must only shrink):"
      fail=1; section=body; continue ;;
    *)
      printf '  %s\n' "$line" ;;
  esac
done <<< "$report"

exit $fail
