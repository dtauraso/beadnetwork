#!/usr/bin/env bash

# PLACEMENT: nodes/**/*.go | a channel name must encode the two endpoints it connects (scripts/audit-channel-names.sh)




set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
bash "$REPO_ROOT/scripts/audit-channel-names.sh"
