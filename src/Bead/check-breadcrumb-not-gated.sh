#!/usr/bin/env bash

# PLACEMENT: src/Bead/*.go,src/Node/Wiring/**/*.go | only emitArrive's KindArrive append may sit behind StreamsActive; breadcrumbs always emit

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

source "$REPO_ROOT/scripts/lib/trace/gate-audit-common.sh"
source "$REPO_ROOT/scripts/lib/trace/gate-streams-active.sh"

fail=0
report=""

check_streams_active_gate

if [ $fail -ne 0 ]; then
  echo "check-breadcrumb-not-gated: FAIL" >&2
  printf '%s' "$report" >&2
  exit 1
fi

exit 0
