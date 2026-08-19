#!/usr/bin/env bash

# PLACEMENT: nodes/bead/*.go,nodes/Wiring/**/*.go | only emitArrive's KindArrive append may sit behind StreamsActive; breadcrumbs always emit

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

source "$SCRIPT_DIR/lib/gate-audit-common.sh"
source "$SCRIPT_DIR/lib/gate-streams-active.sh"

fail=0
report=""

check_streams_active_gate

if [ $fail -ne 0 ]; then
  echo "check-breadcrumb-not-gated: FAIL" >&2
  printf '%s' "$report" >&2
  exit 1
fi

exit 0
