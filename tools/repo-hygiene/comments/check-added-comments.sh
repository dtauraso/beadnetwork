#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: none | repo-wide: no comment line added since the base ref survives in a hand-edited file

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

python3 "$SCRIPT_DIR/strip_added_comments.py" >/dev/null
