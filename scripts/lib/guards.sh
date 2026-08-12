#!/usr/bin/env bash

# Sourced by scripts/stop-checks.sh. Discovers and runs every tools/*/check-*.sh (and
# tools/*/*/check-*.sh) guard except the ones the orchestrator already ran above as part
# of the language phases. Reads/writes the orchestrator's globals ($out, $fail) directly.

run_one_guard() {
  local gp="$1" gn go grc
  gn=$(basename "$gp" .sh)
  go=$(bash "$gp" 2>&1); grc=$?
  printf '%s' "$go" > "$GDIR/$gn.out"
  printf '%s' "$grc" > "$GDIR/$gn.rc"
}
export -f run_one_guard 2>/dev/null || true

run_guards() {
  #   check-staticcheck / check-eslint / check-vitest — expensive; invoked above under their
  local GUARD_EXCLUDE="check-staticcheck|check-eslint|check-vitest|check-no-foreground-sim|check-stray-screenshots|check-no-shell-source-edits"

  shopt -s nullglob
  local guards=(tools/*/check-*.sh tools/*/*/check-*.sh)
  shopt -u nullglob

  if [ ${#guards[@]} -eq 0 ]; then
    echo "stop-checks: MISCONFIGURED — no tools/*/check-*.sh found; refusing to report success." >&2
    exit 1
  fi

  local GUARD_SERIAL="check-generated"
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
    [ "$chk" = "$GUARD_SERIAL" ] || continue
    run_one_guard "$chk_path"
  done

  printf '%s\n' "${guard_selected[@]}" \
    | grep -v "/${GUARD_SERIAL}\.sh$" \
    | GDIR="$GDIR" xargs -P "$JOBS" -I@ bash -c '
        gp="@"; gn=$(basename "$gp" .sh)
        go=$(bash "$gp" 2>&1); grc=$?
        printf "%s" "$go" > "$GDIR/$gn.out"
        printf "%s" "$grc" > "$GDIR/$gn.rc"
      '

  local grc chk_out
  for chk_path in "${guard_selected[@]}"; do
    chk=$(basename "$chk_path" .sh)
    grc=$(cat "$GDIR/$chk.rc" 2>/dev/null || echo 1)
    if [ "$grc" != "0" ]; then
      chk_out=$(cat "$GDIR/$chk.out" 2>/dev/null || echo "(no output captured)")
      out+="$chk failed:\n$chk_out\n\n"
      fail=1
    fi
  done
}
