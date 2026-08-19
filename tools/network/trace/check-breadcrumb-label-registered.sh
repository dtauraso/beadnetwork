#!/usr/bin/env bash

# PLACEMENT: tools/topology-vscode/src/Trace/Trace.go,nodes/bead/inport/in_port.go,nodes/**/*.go,tools/topology-vscode/src/Buffer/**/*.go | a .Breadcrumb("label") literal must be added to Trace.BreadcrumbLabels

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

TRACE_DIR="tools/topology-vscode/src/Trace"

if [[ ! -d "$TRACE_DIR" ]]; then
  echo "check-breadcrumb-label-registered: MISCONFIGURED — dir not found: $TRACE_DIR" >&2
  exit 1
fi

LABEL_FN_FILE=$(grep -rl '^func breadcrumbLabelFor(' --include="*.go" nodes/bead 2>/dev/null | head -1 || true)
if [[ -z "$LABEL_FN_FILE" ]]; then
  echo "check-breadcrumb-label-registered: MISCONFIGURED — breadcrumbLabelFor not found in" >&2
  echo "nodes/bead; it was renamed or deleted. The switch this guard exists to keep in sync" >&2
  echo "with Trace.BreadcrumbLabels no longer has a home — update this guard in the same commit." >&2
  exit 1
fi

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

call_site_hits() {
  grep -rnoE '\.Breadcrumb\("[^"]*"' --include="*.go" nodes Buffer Trace 2>/dev/null \
    | grep -v '_test\.go:' || true
}

CALL_HITS=$(call_site_hits)
if [[ -z "$(printf '%s' "$CALL_HITS" | tr -d '[:space:]')" ]]; then
  echo "check-breadcrumb-label-registered: MISCONFIGURED — no .Breadcrumb(\"...\") call" >&2
  echo "sites found under nodes/ or tools/topology-vscode/src/Trace; the extraction pattern itself is likely stale." >&2
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
