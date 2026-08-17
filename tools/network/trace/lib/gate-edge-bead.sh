#!/usr/bin/env bash

check_edge_bead_trace_gate() {
  local IDENT="edgeBeadTraceEnabled"

  local files=()
  while IFS= read -r f; do
    [ -n "$f" ] && files+=("$f")
  done < <(find_ident_files "$IDENT")

  if [ ${#files[@]} -eq 0 ]; then
    echo "check-breadcrumb-not-gated: FAIL — $IDENT not found anywhere; the gate itself" >&2
    echo "appears to have been deleted or renamed. If intentional, delete/update this guard" >&2
    echo "in the same commit; otherwise the trace-volume gate has silently vanished." >&2
    exit 1
  fi

  local guard_hits=()
  while IFS= read -r h; do
    [ -n "$h" ] && guard_hits+=("$h")
  done < <(find_guarding_hits "$IDENT" "${files[@]}")

  if [ ${#guard_hits[@]} -ne 1 ]; then
    fail=1
    report+="Expected exactly ONE guarding use of $IDENT (the stepAll KindEdgeBead append);"$'\n'
    report+="found ${#guard_hits[@]}:"$'\n'
    for h in "${guard_hits[@]:-}"; do
      [ -n "$h" ] && report+="  $h"$'\n'
    done
    report+=$'\n'
    report+="Debug breadcrumbs (T.KindBreadcrumb) and any other bulk-event kind must NEVER be"$'\n'
    report+="gated by this flag — it exists ONLY to bound KindEdgeBead volume. Widening it"$'\n'
    report+="(wrapping drainPendingEvents, a breadcrumb append, or a future kind's emit site)"$'\n'
    report+="reproduces the exact silent-swallow regression this guard exists to prevent — see"$'\n'
    report+=".claude/rules/go-debugging.md's 'Debugging the Go layer (probe breadcrumbs)' section and"$'\n'
    report+="memory/feedback_make_bug_class_unrepresentable.md."$'\n'
  else
    local h="${guard_hits[0]}"
    local gfile="${h%%:*}"
    local rest="${h#*:}"
    local glineno="${rest%%:*}"
    local window_end=$((glineno + 5))
    local window="$(sed -n "${glineno},${window_end}p" "$gfile")"
    if ! echo "$window" | grep -q "KindEdgeBead"; then
      fail=1
      report+="The single guarding use of $IDENT is not tied to KindEdgeBead:"$'\n'
      report+="  $h"$'\n'
      report+=$'\n'
      report+="This flag exists ONLY to gate T.KindEdgeBead volume (see"$'\n'
      report+="nodes/wire/paced_wire.go's edgeBeadTraceEnabled doc comment). A guard use that"$'\n'
      report+="does not guard a KindEdgeBead emit site means the gate has spread to a different"$'\n'
      report+="event kind — most dangerously KindBreadcrumb, whose failure mode is SILENCE (see"$'\n'
      report+="memory/feedback_make_bug_class_unrepresentable.md)."$'\n'
    fi
  fi
}
