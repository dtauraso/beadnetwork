#!/usr/bin/env bash

# Sourced by tools/network/trace/check-breadcrumb-not-gated.sh. Audits StreamsActive:
# exactly two guarding uses (stepAll's KindEdgeBead append, emitArrive's KindArrive
# append), neither mentioning breadcrumb, both tied to KindEdgeBead or KindArrive.
# Reads/writes the caller's globals ($fail, $report) directly.

check_streams_active_gate() {
  local IDENT2="StreamsActive"

  local files2=()
  while IFS= read -r f; do
    [ -n "$f" ] && files2+=("$f")
  done < <(find_ident_files "$IDENT2")

  if [ ${#files2[@]} -eq 0 ]; then
    echo "check-breadcrumb-not-gated: FAIL — $IDENT2 not found anywhere; the gate itself" >&2
    echo "appears to have been deleted or renamed. If intentional, delete/update this guard" >&2
    echo "in the same commit; otherwise the pending-accumulation gate has silently vanished." >&2
    exit 1
  fi

  local guard_hits2=()
  while IFS= read -r h; do
    [ -n "$h" ] && guard_hits2+=("$h")
  done < <(find_guarding_hits "$IDENT2" "${files2[@]}")

  if [ ${#guard_hits2[@]} -ne 2 ]; then
    fail=1
    report+="Expected exactly TWO guarding uses of $IDENT2 (stepAll's KindEdgeBead append and"$'\n'
    report+="emitArrive's KindArrive append); found ${#guard_hits2[@]}:"$'\n'
    for h in "${guard_hits2[@]:-}"; do
      [ -n "$h" ] && report+="  $h"$'\n'
    done
    report+=$'\n'
    report+="$IDENT2 exists ONLY to suppress pending-event accumulation when no stream"$'\n'
    report+="consumer is wired (see PacedWire.StreamsActive's doc comment). Debug breadcrumbs"$'\n'
    report+="(T.KindBreadcrumb, breadcrumbCh, drainBreadcrumbEvents) must NEVER be gated by it —"$'\n'
    report+="reproduces the same silent-swallow regression class as edgeBeadTraceEnabled would."$'\n'
  else
    local h content_lc
    for h in "${guard_hits2[@]}"; do
      content_lc="$(echo "$h" | tr '[:upper:]' '[:lower:]')"
      if echo "$content_lc" | grep -qi "breadcrumb"; then
        fail=1
        report+="A guarding use of $IDENT2 mentions breadcrumb — this flag must never gate"$'\n'
        report+="breadcrumb emission:"$'\n'
        report+="  $h"$'\n'
      fi
    done

    local gfile rest glineno window_end window
    for h in "${guard_hits2[@]}"; do
      gfile="${h%%:*}"
      rest="${h#*:}"
      glineno="${rest%%:*}"
      window_end=$((glineno + 3))
      window="$(sed -n "${glineno},${window_end}p" "$gfile")"
      if ! echo "$window" | grep -qE "KindEdgeBead|KindArrive"; then
        fail=1
        report+="A guarding use of $IDENT2 is not tied to KindEdgeBead or KindArrive:"$'\n'
        report+="  $h"$'\n'
        report+=$'\n'
        report+="This flag exists only to gate those two bulk-event appends — a guard use that"$'\n'
        report+="does not sit next to either means it has spread to a different emit site,"$'\n'
        report+="most dangerously a breadcrumb one (see"$'\n'
        report+="memory/feedback_make_bug_class_unrepresentable.md)."$'\n'
      fi
    done
  fi
}
