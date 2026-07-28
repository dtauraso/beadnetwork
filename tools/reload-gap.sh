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
# WHY EXTHOST.LOG ALONE IS NOT ENOUGH
# -----------------------------------
# A short-lived host can exit before it flushes anything, leaving NO started and
# NO exiting line in exthost.log — it is missing from that file entirely. Pairing
# exthost's own exiting->started then spans the invisible host and reports two
# reloads as one doubled gap. Observed 07-28: pid 2739 lived 01:22:19->01:22:20
# and appears only in main.log; exthost pairing printed 3.5s for what were two
# ~1.7s reloads. So exits come from the session's main.log, which logs an
# "exited with code" line for EVERY pid, and starts come from the window's
# exthost.log. Each start is paired with the latest exit before it. A host with
# no started line is still unmeasurable — it anchors the NEXT reload's gap
# correctly, but contributes no row of its own, so the printed reload count can
# undercount. Better a missing row than a doubled number.
#
# Measured baseline on this machine (see git log for the session that found it):
#   07-23 .. 07-26  ->  1.8s, flat across 44 reloads
#   07-27           ->  4.0-4.9s
#   07-28 (reboot)  ->  1.6-1.7s, back to baseline
# The regression did NOT line up with any commit, and a full quit+relaunch of VS
# Code did NOT fix it — a window measured 12 minutes into a fresh VS Code launch
# was still at 4s, which rules out accumulated editor/session state. What fixed
# it was a full macOS restart: host memory pressure (44 days uptime, 2.83 GB of
# 4 GB swap in use, load average ~3) went to zero swap and the gap returned to
# baseline in the same reload session. So when this reads >4s, reboot the
# machine; there is nothing to fix in this repo.
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
    # Print the whole comment header, however long it grows — a hardcoded line
    # range silently truncates the next time something is added to it.
    -h|--help) sed -n '2,/^[^#]/p' "$0" | sed '$d; s/^# \{0,1\}//'; exit 0 ;;
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

# Pair each host start (window exthost.log) with the latest host exit before it
# (session main.log — see "WHY EXTHOST.LOG ALONE IS NOT ENOUGH" above), and print
# the gap. Timestamps are "YYYY-MM-DD HH:MM:SS.mmm"; convert to epoch seconds with
# days_from_civil so a reload spanning midnight still measures correctly. Plain
# awk (no gawk mktime).
#
# $1 = main.log (may be missing — then exthost's own "exiting" lines are the only
#      exits available and short-lived hosts stay invisible), $2 = exthost.log.
# A start with no exit before it is a window opening, not a reload, so it is
# skipped. So is a gap over MAXGAP seconds: that is a window that sat closed,
# not a respawn.
gaps_for() {
  awk -v MAXGAP=60 -v MAIN="$1" '
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
    # Pass 1 (main.log) and the exiting lines of pass 2 both contribute exits.
    /Extension host with pid .* (exiting|exited) with code/ {
      n++; ed[n] = $1; et[n] = $2; ee[n] = epoch($1, $2); next
    }
    FILENAME == MAIN { next }   # nothing else in main.log is of interest
    /Extension host with pid .* started/ {
      s = epoch($1, $2)
      # Latest exit strictly before this start. Both inputs are chronological,
      # but main.log covers every window of the session while exthost.log covers
      # one, so the arrays interleave — scan rather than assume a running index.
      best = -1
      for (i = 1; i <= n; i++)
        if (ee[i] < s && ee[i] > best) { best = ee[i]; bi = i }
      if (best < 0 || s - best > MAXGAP) next
      printf "  %s %s  ->  %5.1fs\n", ed[bi], substr(et[bi], 1, 8), s - best
    }
  ' "$1" "$2"
}

found=0
for root in "${roots[@]}"; do
  # Newest session dirs first, but print each window's history oldest->newest so
  # a trend reads top-to-bottom.
  while IFS= read -r log; do
    # <session>/<window>/exthost/exthost.log -> <session>/main.log
    sessdir=$(dirname "$(dirname "$(dirname "$log")")")
    main="$sessdir/main.log"
    [ -f "$main" ] || main=/dev/null
    out=$(gaps_for "$main" "$log")
    [ -z "$out" ] && continue
    found=1
    win=$(basename "$(dirname "$(dirname "$log")")")
    sess=$(basename "$sessdir")
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
  >4s    regressed — reboot the machine. A VS Code quit+relaunch was measured
         and does NOT fix it; host memory/swap pressure does (see header).

this gap is process respawn only; extension activation is ~200ms and is never
the cause. Confirm with: grep "Eager extensions activated" on the same log.
EOF
