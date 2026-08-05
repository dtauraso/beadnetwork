#!/usr/bin/env bash
# check-breadcrumb-label-registered.sh — every literal label passed to a
# tr.Breadcrumb("label", ...)/In.Breadcrumb("label", ...) call site must be a member of
# Trace.BreadcrumbLabels, or nodes/wire/ports.go's breadcrumbLabelFor DROPS it SILENTLY:
# its `default:` returns (0, false), the row never reaches the buffer, and the breadcrumb
# PLACEMENT: Trace/Trace.go,nodes/wire/ports.go,nodes/**/*.go,Buffer/**/*.go | a .Breadcrumb("label") literal must be added to Trace.BreadcrumbLabels
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
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

TRACE_GO="Trace/Trace.go"
PORTS_GO="nodes/wire/ports.go"

for f in "$TRACE_GO" "$PORTS_GO"; do
  if [[ ! -f "$f" ]]; then
    echo "check-breadcrumb-label-registered: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

# Registered labels: Trace.BreadcrumbLabels's string-literal elements, in order. Same
# "slurp from the var name to the closing brace" shape as gen-node-defs' own
# parseBreadcrumbLabels (tools/gen-node-defs/trace_kinds.go), just in awk instead of Go's
# ast package — both read the identical literal.
registered() {
  awk '/var BreadcrumbLabels = \[\]string\{/,/^\}/' "$TRACE_GO" \
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
  echo "and nodes/wire/ports.go's breadcrumbLabelFor switch case — follow how drag.commit" >&2
  echo "is wired) or it is silently dropped before it ever reaches the buffer/logs." >&2
  exit 1
fi

echo "check-breadcrumb-label-registered: clean"
exit 0
