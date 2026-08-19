#!/usr/bin/env bash

set -euo pipefail

limit=20
show_all=0
while [ $# -gt 0 ]; do
  case "$1" in
    -n) limit="${2:?-n needs a count}"; shift 2 ;;
    -a) show_all=1; shift ;;

    -h|--help) sed -n '2,/^[^#]/p' "$0" | sed '$d; s/^# \{0,1\}//'; exit 0 ;;
    *) echo "unknown arg: $1 (try -h)" >&2; exit 2 ;;
  esac
done

roots=()
for app in Code VSCodium Cursor "Code - Insiders"; do
  d="$HOME/Library/Application Support/$app/logs"
  [ -d "$d" ] && roots+=("$d")
done
if [ ${#roots[@]} -eq 0 ]; then
  echo "no VS Code log directory found under ~/Library/Application Support/" >&2
  exit 1
fi

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
    /Extension host with pid .* (exiting|exited) with code/ {
      n++; ed[n] = $1; et[n] = $2; ee[n] = epoch($1, $2); next
    }
    FILENAME == MAIN { next }   # nothing else in main.log is of interest
    /Extension host with pid .* started/ {
      s = epoch($1, $2)
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

  while IFS= read -r log; do

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
         and does NOT fix it; host memory or swap pressure does (see header).

this gap is process respawn only; extension activation is ~200ms and is never
the cause. Confirm with: grep "Eager extensions activated" on the same log.
EOF
