#!/usr/bin/env bash

# PLACEMENT: nodes/**/*.go | no time.Sleep/time.After/time.NewTicker outside nodes/clock/clock.go; block on the tick channel instead
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
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
      echo "found outside nodes/clock/clock.go. A goroutine parked here cannot service its other"
      echo "channels for the wait — route through clock.NewRealClock()'s SleepCycle or"
      echo "clock.NewTickChan() instead, both backed by the process's one TickBroadcaster:"
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
