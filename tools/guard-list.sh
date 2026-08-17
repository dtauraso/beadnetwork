#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

find tools -name 'check-*.sh' -type f \
  -not -path '*/node_modules/*' \
  -not -path '*/out/*' \
  | sort
