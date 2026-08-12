#!/usr/bin/env bash
# check-breadcrumb-label-registered.sh — every literal label passed to a
# tr.Breadcrumb("label", ...)/In.Breadcrumb("label", ...) call site must be a member of
# Trace.BreadcrumbLabels, or nodes/wire/in_port.go's breadcrumbLabelFor DROPS it SILENTLY:
# its `default:` returns (0, false), the row never reaches the buffer, and the breadcrumb
# PLACEMENT: Trace/Trace.go,nodes/wire/in_port.go,nodes/**/*.go,Buffer/**/*.go | a .Breadcrumb("label") literal must be added to Trace.BreadcrumbLabels
#
# reads exactly like a passing (i.e. never-hit) probe — you cannot tell "this code path
# didn't run" from "this code path ran but its label was never registered" just by staring
# at an empty .probe log. Three temporary probe breadcrumbs (drag.jump, probe.commitLocal,
# probe.enterCommit) cost a real debugging round trip this way: none of them appeared in
# the logs, and their absence was read as evidence the code wasn't running, when the code
# ran fine and the labels were simply unregistered
# (memory/feedback_check_the_signal_the_check_emits.md).
#
# WHY grep-based, not a Go AST walk: BreadcrumbLabels is a flat []string literal and every
# call site passes a string LITERAL (never a variable) for label — this codebase's existing
# cross-file parity guards (check-message-kind-parity.sh, check-comment-vocab.sh) already
# trust exactly this shape of literal-extraction diff, and a real Go compile can't catch
# this bug at all: breadcrumbLabelFor's `default: return 0, false` is valid, type-checked
# Go — the missing registration is a DATA problem (two lists out of sync), not a syntax one.
#
# Exit 0 if every call-site label is registered; exit 1 with a report otherwise.
# Auto-discovered by scripts/stop-checks.sh via the tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

TRACE_DIR="Trace"

if [[ ! -d "$TRACE_DIR" ]]; then
  echo "check-breadcrumb-label-registered: MISCONFIGURED — dir not found: $TRACE_DIR" >&2
  exit 1
fi

# breadcrumbLabelFor is located BY NAME, not by a hardcoded filename: it used to live in
# nodes/wire/ports.go and moved to nodes/wire/in_port.go when that god file was split by
# job. A path constant would have kept pointing at a file that still exists (ports.go was
# not deleted), so the existence check would have gone on passing while pointing at the
# wrong file — a guard blind in exactly the way memory/
# feedback_guards_hardcoding_single_file_break_on_split.md describes. Scan the package.
LABEL_FN_FILE=$(grep -rl '^func breadcrumbLabelFor(' --include="*.go" nodes/wire 2>/dev/null | head -1 || true)
if [[ -z "$LABEL_FN_FILE" ]]; then
  echo "check-breadcrumb-label-registered: MISCONFIGURED — breadcrumbLabelFor not found in" >&2
  echo "nodes/wire; it was renamed or deleted. The switch this guard exists to keep in sync" >&2
  echo "with Trace.BreadcrumbLabels no longer has a home — update this guard in the same commit." >&2
  exit 1
fi

# Registered labels: Trace.BreadcrumbLabels's string-literal elements, in order. Same
# "slurp from the var name to the closing brace" shape as gen-node-defs' own
# parseBreadcrumbLabels (tools/gen-node-defs/trace_kinds.go), just in awk instead of Go's
# ast package — both read the identical literal. Scanned across EVERY non-test *.go file
# under Trace/ (not one hardcoded filename), matching parseBreadcrumbLabels' own
# whole-dir scan, so a future file split (like this one moving BreadcrumbLabels out of
# Trace.go into breadcrumb_labels.go) cannot go blind the way
# memory/feedback_guards_hardcoding_single_file_break_on_split.md describes.
registered() {
  awk '/var BreadcrumbLabels = \[\]string\{/,/^\}/' "$TRACE_DIR"/*.go \
    | grep -oE '"[^"]*"' \
    | tr -d '"' \
    | sort -u
}

REGISTERED=$(registered) || true
if [[ -z "$(printf '%s' "$REGISTERED" | tr -d '[:space:]')" ]]; then
  echo "check-breadcrumb-label-registered: EMPTY extracted set for Trace.BreadcrumbLabels —" >&2
  echo "the var was renamed or its shape changed; refusing a vacuous pass." >&2
  exit 1
fi

# Call-site labels: every `.Breadcrumb("literal"` across production (non-test, non-tools)
# Go source. Scoped the same way check-breadcrumb-not-gated.sh scopes its identifier scan
# (nodes/Buffer/Trace, excluding _test.go) — call sites live
# under nodes/ today, but Buffer/Trace are included in case one lands there later.
call_site_hits() {
  grep -rnoE '\.Breadcrumb\("[^"]*"' --include="*.go" nodes Buffer Trace 2>/dev/null \
    | grep -v '_test\.go:' || true
}

CALL_HITS=$(call_site_hits)
if [[ -z "$(printf '%s' "$CALL_HITS" | tr -d '[:space:]')" ]]; then
  echo "check-breadcrumb-label-registered: MISCONFIGURED — no .Breadcrumb(\"...\") call" >&2
  echo "sites found under nodes/Buffer/Trace; the extraction pattern itself is likely stale." >&2
  exit 1
fi

MISSING=0
report=""
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  loc="${hit%%:.Breadcrumb*}"
  label="$(echo "$hit" | grep -oE '"[^"]*"' | tr -d '"')"
  if ! printf '%s\n' "$REGISTERED" | grep -qxF "$label"; then
    report+="  $loc: label \"$label\" is not in Trace.BreadcrumbLabels"$'\n'
    MISSING=$((MISSING + 1))
  fi
done <<< "$CALL_HITS"

if [[ $MISSING -ne 0 ]]; then
  echo "check-breadcrumb-label-registered: FAIL — $MISSING unregistered breadcrumb label(s):" >&2
  printf '%s' "$report" >&2
  echo "" >&2
  echo "Add each label to Trace.BreadcrumbLabels (and its matching BreadcrumbLabel* const" >&2
  echo "and $LABEL_FN_FILE's breadcrumbLabelFor switch case — follow how drag.commit" >&2
  echo "is wired) or it is silently dropped before it ever reaches the buffer/logs." >&2
  exit 1
fi

echo "check-breadcrumb-label-registered: clean"
exit 0
