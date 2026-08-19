#!/usr/bin/env bash

guard_slug() {
  local gp="$1"
  gp="${gp%.sh}"
  printf '%s' "${gp//\//__}"
}
export -f guard_slug 2>/dev/null || true

run_one_guard() {
  local gp="$1" gs go grc
  gs=$(guard_slug "$gp")
  go=$(bash "$gp" 2>&1); grc=$?
  printf '%s' "$go" > "$GDIR/$gs.out"
  printf '%s' "$grc" > "$GDIR/$gs.rc"
}
export -f run_one_guard 2>/dev/null || true

run_guards() {
  #   check-staticcheck / check-eslint / check-vitest — expensive; invoked above under their
  local GUARD_EXCLUDE="check-staticcheck|check-eslint|check-vitest|check-no-foreground-sim|check-stray-screenshots|check-no-shell-source-edits"

  local guards=()
  while IFS= read -r g; do
    [ -n "$g" ] && guards+=("$g")
  done < <(bash scripts/guard-list.sh)

  if [ ${#guards[@]} -eq 0 ]; then
    echo "stop-checks: MISCONFIGURED — scripts/guard-list.sh named no guards; refusing to report success." >&2
    exit 1
  fi

  local GUARD_SERIAL="check-generated|check-added-comments"
  local JOBS
  JOBS=$(sysctl -n hw.ncpu 2>/dev/null || nproc 2>/dev/null || echo 4)

  local guard_selected=()
  local chk_path chk
  for chk_path in "${guards[@]}"; do
    chk=$(basename "$chk_path" .sh)
    if echo "$chk" | grep -qE "^($GUARD_EXCLUDE)$"; then continue; fi
    guard_selected+=("$chk_path")
  done

  GDIR=$(mktemp -d)
  trap 'rm -rf "$GDIR"' EXIT

  for chk_path in "${guard_selected[@]}"; do
    chk=$(basename "$chk_path" .sh)
    echo "$chk" | grep -qE "^($GUARD_SERIAL)$" || continue
    run_one_guard "$chk_path"
  done

  printf '%s\n' "${guard_selected[@]}" \
    | grep -vE "/($GUARD_SERIAL)\.sh$" \
    | GDIR="$GDIR" xargs -P "$JOBS" -I@ bash -c 'run_one_guard "@"'

  local grc chk_out gs
  for chk_path in "${guard_selected[@]}"; do
    gs=$(guard_slug "$chk_path")
    grc=$(cat "$GDIR/$gs.rc" 2>/dev/null || echo 1)
    if [ "$grc" != "0" ]; then
      chk_out=$(cat "$GDIR/$gs.out" 2>/dev/null || echo "(no output captured)")
      out+="$chk_path failed:\n$chk_out\n\n"
      fail=1
    fi
  done
}
