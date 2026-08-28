#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

INTERVAL="${1:-10}"
ANCHOR="topology"
LOG=".probe/file-growth.log"

die() { printf 'watch-file-growth: %s\n' "$1" >&2; exit 1; }

[ -d "$ANCHOR" ] || die "no $ANCHOR/ directory here — run from the repo root"

selected="$(cat "$ANCHOR/view/scene/selected.bin" 2>/dev/null || true)"
SCENE="${selected:-$ANCHOR}"
[ -d "$SCENE" ] || die "selected scene '$SCENE' is not a directory"

size_of() { stat -f%z "$1" 2>/dev/null || echo 0; }

mkdir -p .probe
printf '=== watch-file-growth  scene=%s  interval=%ss  started=%s\n' \
  "$SCENE" "$INTERVAL" "$(date '+%H:%M:%S')" >>"$LOG"

PREV="$(mktemp)"
trap 'rm -f "$PREV"' EXIT
: >"$PREV"

while :; do
  now="$(date '+%H:%M:%S')"
  total=0
  grown=""
  CUR="$(mktemp)"

  while IFS= read -r f; do
    sz="$(size_of "$f")"
    total=$((total + sz))
    printf '%s %s\n' "$sz" "$f" >>"$CUR"
    was="$(awk -v p="$f" '$2 == p {print $1; exit}' "$PREV")"
    [ -z "$was" ] && continue
    [ "$sz" -eq "$was" ] && continue
    delta=$((sz - was))
    rate=$((delta / INTERVAL))
    grown="$grown$(printf '    %+10d B  %+8d B/s  %10d B  %s' "$delta" "$rate" "$sz" "$f")
"
  done < <(find "$SCENE/view" .probe -type f ! -name 'file-growth.log' 2>/dev/null)

  mv "$CUR" "$PREV"

  go_rss="$(ps -axo rss,command | awk '/beadnetwork -topology/ && !/awk/ && !v {v=int($1/1024)} END {print v+0}')"
  code_rss="$(ps -axo rss,command | awk '/Visual Studio Code/ && !/awk/ {s+=$1} END {print int(s/1024)}')"

  printf '%s  tree=%dK  go=%sM  vscode=%sM\n' \
    "$now" "$((total / 1024))" "$go_rss" "$code_rss" >>"$LOG"
  if [ -n "$grown" ]; then printf '%s' "$grown" >>"$LOG"; fi

  sleep "$INTERVAL"
done
