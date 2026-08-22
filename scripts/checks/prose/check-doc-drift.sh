#!/usr/bin/env bash

# PLACEMENT: none | thin wrapper invoking scripts/checks/prose/audit-doc-drift.mjs so it runs in the guard loop
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
node "$REPO_ROOT/scripts/checks/prose/audit-doc-drift.mjs"
