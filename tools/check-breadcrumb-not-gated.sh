#!/usr/bin/env bash
#
# PLACEMENT: nodes/wire/*.go,nodes/Wiring/*.go | only stepAll's KindEdgeBead append may sit behind edgeBeadTraceEnabled; breadcrumbs always emit
# check-breadcrumb-not-gated.sh — forbid the WIREFOLD_EDGE_BEAD_TRACE gate
# (edgeBeadTraceEnabled, nodes/wire/paced_wire.go) from spreading beyond its one
# legitimate site: the T.KindEdgeBead append inside stepAll. Debug breadcrumbs
# (T.KindBreadcrumb, the .claude/rules/go-debugging.md channel, read via
# tools/probe-merge.sh --debug) MUST always emit — gating them, gating
# drainPendingEvents wholesale, or gating any future bulk kind's emit site behind
# this flag reproduces the exact regression this guard exists to prevent: a
# breadcrumb fires, is silently swallowed, and reads as "my breadcrumb is
# broken" instead of "the gate ate it" (see
# memory/feedback_make_bug_class_unrepresentable.md).
#
# WHY grep-based counting, not an AST/semantic check: the invariant we actually
# care about — "this flag has exactly one guarding use, and it is the
# KindEdgeBead append" — is small and lexical enough that a line-count + adjacency
# check on the identifier text has no realistic false-negative (a Go
# identifier's declaration and every read of it are syntactically
# `edgeBeadTraceEnabled` tokens; there is no aliasing/reflection path in this
# codebase that could reference it invisibly — check-no-network-locks.sh and
# check-comment-vocab.sh already trust exactly this kind of grep-on-identifier
# scan for the same reason). An AST walker would add a heavy new dependency to
# catch a case grep already catches deterministically. The one thing grep
# cannot see — "is this identifier really doing what its name says" — isn't
# what this guard checks; it checks *how many places* touch the switch, and
# whether the one guarding use sits next to KindEdgeBead, which is exactly the
# textual shape of the regression (a second `if edgeBeadTraceEnabled` wrapping
# something else, or a rename of the flag to widen its reach, both show up as
# an extra grep hit or a hit with no KindEdgeBead on/near the line).
#
# Scope: production (non-test) Go, excluding tools/ codegen. Comments are NOT
# stripped, so a comment that merely mentions the
# identifier (e.g. this file's own header, or paced_wire.go's doc comment)
# still counts as an "occurrence" toward the total-count budget — see the
# EXPECTED_TOTAL note below for why that is fine.
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh
# via the tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

IDENT="edgeBeadTraceEnabled"

# All non-test .go files under nodes/Buffer/Trace, referencing the identifier at all
# (declaration, doc comments mentioning it, and code uses alike). Portable read
# — macOS bash 3.2 has no mapfile (see check-no-dead-buffer-column.sh).
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

# Collect every matching line, file:lineno: text, across the scoped files.
all_hits=()
while IFS= read -r line; do
  [ -n "$line" ] && all_hits+=("$line")
done < <(grep -Hn "$IDENT" "${files[@]}" 2>/dev/null || true)

total=${#all_hits[@]}

# Code (non-comment, i.e. not a line whose only content up to the identifier is
# a "//" comment marker) uses of the identifier as an actual guard/condition —
# this is where a THIRD site (a new "if edgeBeadTraceEnabled" elsewhere) would
# show up. We identify "guarding use" lines as those containing the identifier
# in an `if`/`&&`/`||` boolean-use position, which is how both legitimate uses
# (the var declaration's RHS assignment is not a guarding use; the stepAll
# guard is) appear in source today.
guard_hits=()
for h in "${all_hits[@]}"; do
  # h looks like path:lineno:content
  content="${h#*:}"       # strip path
  content="${content#*:}" # strip lineno, leaves the code text
  trimmed="$(echo "$content" | sed -e 's/^[[:space:]]*//')"
  case "$trimmed" in
    //*) continue ;; # pure comment line, not a guard use
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
  # The single guarding use must be tied to KindEdgeBead — check the guard's own line and
  # a small window of following lines (the append block it guards) for the KindEdgeBead
  # token, since `if emit && edgeBeadTraceEnabled {` is immediately followed by the
  # `pw.pending = append(..., pendingWireEvent{kind: T.KindEdgeBead, ...})` line it guards.
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

# --- StreamsActive (nodes/wire/paced_wire.go's PacedWire field, added to gate
# pending-event accumulation on whether a real per-edge stream consumer is
# wired — see PacedWire.StreamsActive's doc comment and
# docs/planning/visual-editor/session-log.md) --- same shape as above:
# this flag exists ONLY to gate the KindEdgeBead append (alongside
# edgeBeadTraceEnabled) and the KindArrive append in emitArrive. It must NEVER
# spread to gate a breadcrumb append, same failure mode (silence) as above.
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

# Expect exactly TWO guarding uses today: the KindEdgeBead append in stepAll
# (alongside edgeBeadTraceEnabled) and the KindArrive append in emitArrive.
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
  # Each guarding use must sit next to the specific bulk-kind it guards
  # (KindEdgeBead or KindArrive), not something unrelated (most dangerously a
  # bare drainBreadcrumbEvents/breadcrumbCh call).
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
