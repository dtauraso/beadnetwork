#!/usr/bin/env bash








# PLACEMENT: none | thin wrapper invoking scripts/audit-doc-drift.mjs so it runs in the guard loop
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
node "$REPO_ROOT/scripts/audit-doc-drift.mjs"
