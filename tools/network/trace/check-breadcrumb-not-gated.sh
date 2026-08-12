#!/usr/bin/env bash

# PLACEMENT: nodes/wire/*.go,nodes/Wiring/*.go | only stepAll's KindEdgeBead append may sit behind edgeBeadTraceEnabled; breadcrumbs always emit

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

source "$SCRIPT_DIR/lib/gate-audit-common.sh"
source "$SCRIPT_DIR/lib/check-edge-bead-gate.sh"
source "$SCRIPT_DIR/lib/check-streams-active-gate.sh"

fail=0
report=""

check_edge_bead_trace_gate
check_streams_active_gate

if [ $fail -ne 0 ]; then
  echo "check-breadcrumb-not-gated: FAIL" >&2
  printf '%s' "$report" >&2
  exit 1
fi

exit 0
