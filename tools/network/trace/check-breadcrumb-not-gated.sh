#!/usr/bin/env bash

# PLACEMENT: nodes/wire/*.go,nodes/Wiring/*.go | only stepAll's KindEdgeBead append may sit behind edgeBeadTraceEnabled; breadcrumbs always emit

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

IDENT="edgeBeadTraceEnabled"

files=()
while IFS= read -r f; do
  [ -n "$f" ] && files+=("$f")
done < <(grep -rl "$IDENT" --include="*.go" nodes Buffer Trace 2>/dev/null | grep -v '_test\.go$' || true)

if [ ${#files[@]} -eq 0 ]; then
  echo "check-breadcrumb-not-gated: FAIL — $IDENT not found anywhere; the gate itself" >&2
  echo "appears to have been deleted or renamed. If intentional, delete/update this guard" >&2
  echo "in the same commit; otherwise the trace-volume gate has silently vanished." >&2
  exit 1
fi

all_hits=()
while IFS= read -r line; do
  [ -n "$line" ] && all_hits+=("$line")
done < <(grep -Hn "$IDENT" "${files[@]}" 2>/dev/null || true)

total=${#all_hits[@]}

guard_hits=()
for h in "${all_hits[@]}"; do

  content="${h#*:}"
  content="${content#*:}"
  trimmed="$(echo "$content" | sed -e 's/^[[:space:]]*//')"
  case "$trimmed" in
    //*) continue ;;
  esac
  case "$content" in
    *"if "*"$IDENT"*|*"$IDENT"*"&&"*|*"&&"*"$IDENT"*|*"$IDENT"*"||"*|*"||"*"$IDENT"*)
      guard_hits+=("$h") ;;
  esac
done

fail=0
report=""

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

  h="${guard_hits[0]}"
  gfile="${h%%:*}"
  rest="${h#*:}"
  glineno="${rest%%:*}"
  window_end=$((glineno + 5))
  window="$(sed -n "${glineno},${window_end}p" "$gfile")"
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

IDENT2="StreamsActive"

files2=()
while IFS= read -r f; do
  [ -n "$f" ] && files2+=("$f")
done < <(grep -rl "$IDENT2" --include="*.go" nodes Buffer Trace 2>/dev/null | grep -v '_test\.go$' || true)

if [ ${#files2[@]} -eq 0 ]; then
  echo "check-breadcrumb-not-gated: FAIL — $IDENT2 not found anywhere; the gate itself" >&2
  echo "appears to have been deleted or renamed. If intentional, delete/update this guard" >&2
  echo "in the same commit; otherwise the pending-accumulation gate has silently vanished." >&2
  exit 1
fi

all_hits2=()
while IFS= read -r line; do
  [ -n "$line" ] && all_hits2+=("$line")
done < <(grep -Hn "$IDENT2" "${files2[@]}" 2>/dev/null || true)

guard_hits2=()
for h in "${all_hits2[@]}"; do
  content="${h#*:}"
  content="${content#*:}"
  trimmed="$(echo "$content" | sed -e 's/^[[:space:]]*//')"
  case "$trimmed" in
    //*) continue ;;
  esac
  case "$content" in
    *"if "*"$IDENT2"*|*"$IDENT2"*"&&"*|*"&&"*"$IDENT2"*|*"$IDENT2"*"||"*|*"||"*"$IDENT2"*)
      guard_hits2+=("$h") ;;
  esac
done

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
  for h in "${guard_hits2[@]}"; do
    content_lc="$(echo "$h" | tr '[:upper:]' '[:lower:]')"
    if echo "$content_lc" | grep -qi "breadcrumb"; then
      fail=1
      report+="A guarding use of $IDENT2 mentions breadcrumb — this flag must never gate"$'\n'
      report+="breadcrumb emission:"$'\n'
      report+="  $h"$'\n'
    fi
  done

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

if [ $fail -ne 0 ]; then
  echo "check-breadcrumb-not-gated: FAIL" >&2
  printf '%s' "$report" >&2
  exit 1
fi

exit 0
