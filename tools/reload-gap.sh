#!/usr/bin/env bash
# reload-gap.sh — measure how long "Developer: Reload Window" actually takes.
#
# WHY THIS EXISTS
# ---------------
# Editor reload slowness is easy to misattribute. When it was first noticed
# (10s, up from ~1s) the standing theory was "the .probe logs are huge and VS
# Code is watching them" — plausible, already partly true, and WRONG for this
# symptom: files.watcherExclude and per-run log rotation had both landed and the
# reload stayed slow.
#
# The extension host log settles it, because it timestamps the two events that
# bracket the dead time:
#
#   Extension host with pid N exiting with code 0     <- old host gone
#   Extension host with pid M started                 <- new host up
#
# That gap is pure process respawn: nothing of ours runs in it. Separately,
# "Eager extensions activated" lands ~200ms after start, so extension ACTIVATION
# is not the cost and never was. This script prints the gap so the question
# "is reload slow, and did my change help?" is a number instead of a feeling.
#
# Measured baseline on this machine (see git log for the session that found it):
#   07-23 .. 07-26  ->  1.8s, flat across 44 reloads
#   07-27 onward    ->  4.0-7.5s
# The regression did NOT line up with any commit; the leading suspect is
# accumulated editor/session state (VS Code main process up for days, several
# windows, swap pressure). Hence the first thing to try: fully quit and relaunch
# VS Code, then re-run this. If it drops back to ~1.8s, that was it.
#
# USAGE
#   tools/reload-gap.sh              # all windows, last 20 reloads each
#   tools/reload-gap.sh -n 5         # last 5 reloads per window
#   tools/reload-gap.sh -a           # every reload on record, no tail limit
#
# Reads only VS Code's own logs under ~/Library/Application Support/*/logs/.
# Touches nothing in the repo and needs no editor state.

set -euo pipefail

limit=20
show_all=0
while [ $# -gt 0 ]; do
  case "$1" in
    -n) limit="${2:?-n needs a count}"; shift 2 ;;
    -a) show_all=1; shift ;;
    -h|--help) sed -n '2,40p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "unknown arg: $1 (try -h)" >&2; exit 2 ;;
  esac
done

# VS Code, VSCodium and Cursor all use the same log layout; take whichever exist.
roots=()
for app in Code VSCodium Cursor "Code - Insiders"; do
  d="$HOME/Library/Application Support/$app/logs"
  [ -d "$d" ] && roots+=("$d")
done
if [ ${#roots[@]} -eq 0 ]; then
  echo "no VS Code log directory found under ~/Library/Application Support/" >&2
  exit 1
fi

# Pair each "exiting" with the next "started" and print the gap. Timestamps are
# "YYYY-MM-DD HH:MM:SS.mmm"; convert to epoch seconds with days_from_civil so a
# reload spanning midnight still measures correctly. Plain awk (no gawk mktime).
gaps_for() {
  awk '
    function days(y, m, d,   era, yoe, doy, doe) {
      y -= (m <= 2)
      era = int((y >= 0 ? y : y - 399) / 400)
      yoe = y - era * 400
      doy = int((153 * (m + (m > 2 ? -3 : 9)) + 2) / 5) + d - 1
      doe = yoe * 365 + int(yoe / 4) - int(yoe / 100) + doy
      return era * 146097 + doe - 719468
    }
    function epoch(date, time,   D, T) {
      split(date, D, "-"); split(time, T, ":")
      return days(D[1] + 0, D[2] + 0, D[3] + 0) * 86400 + T[1] * 3600 + T[2] * 60 + T[3]
    }
    /Extension host with pid .* exiting/ { xd = $1; xt = $2; have = 1; next }
    /Extension host with pid .* started/ {
      if (have) {
        printf "  %s %s  ->  %5.1fs\n", xd, substr(xt, 1, 8), epoch($1, $2) - epoch(xd, xt)
        have = 0
      }
    }
  ' "$1"
}

found=0
for root in "${roots[@]}"; do
  # Newest session dirs first, but print each window's history oldest->newest so
  # a trend reads top-to-bottom.
  while IFS= read -r log; do
    out=$(gaps_for "$log")
    [ -z "$out" ] && continue
    found=1
    win=$(basename "$(dirname "$(dirname "$log")")")
    sess=$(basename "$(dirname "$(dirname "$(dirname "$log")")")")
    n=$(printf '%s\n' "$out" | wc -l | tr -d ' ')
    printf '\033[1m%s/%s\033[0m  (%s reloads)\n' "$sess" "$win" "$n"
    if [ "$show_all" = 1 ]; then
      printf '%s\n' "$out"
    else
      printf '%s\n' "$out" | tail -n "$limit"
    fi
    echo
  done < <(find "$root" -name exthost.log -type f 2>/dev/null | sort)
done

if [ "$found" = 0 ]; then
  echo "no extension-host reloads recorded yet (reload the window once, then re-run)"
  exit 0
fi

cat <<'EOF'
reading it:
  ~1.8s  healthy baseline for this repo
  >4s    regressed — first try a full quit+relaunch of VS Code, then re-measure

this gap is process respawn only; extension activation is ~200ms and is never
the cause. Confirm with: grep "Eager extensions activated" on the same log.
EOF
