#!/usr/bin/env bash

# PLACEMENT: scripts/verify.sh | verify.sh stays a thin delegator to stop-checks.sh; never reimplement a check in it
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
VERIFY="$REPO_ROOT/scripts/verify.sh"
STOP="$REPO_ROOT/scripts/stop-checks.sh"

if [ ! -f "$VERIFY" ] || [ ! -f "$STOP" ]; then
  echo "check-verify-thin-wrapper: scripts/verify.sh and scripts/stop-checks.sh must both exist"
  exit 1
fi

if ! grep -q 'stop-checks\.sh' "$VERIFY"; then
  echo "check-verify-thin-wrapper: scripts/verify.sh no longer calls stop-checks.sh — it must"
  echo "delegate, not reimplement. There must be exactly ONE copy of the checks."
  exit 1
fi

leak=$(grep -nE 'tools/check-|go (build|test|vet)\b|npm run|staticcheck|eslint|vitest|tsc ' "$VERIFY" || true)
if [ -n "$leak" ]; then
  echo "check-verify-thin-wrapper: scripts/verify.sh must not run checks itself — it is a thin"
  echo "wrapper on stop-checks.sh (one copy, no drift). Move these back to stop-checks.sh:"
  echo "$leak"
  exit 1
fi

exit 0
