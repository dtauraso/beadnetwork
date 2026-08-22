#!/usr/bin/env bash

# PLACEMENT: Categories/Node/**/*.go | a channel name must encode the two endpoints it connects (Categories/Node/audit-channel-names.sh)

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
bash "$REPO_ROOT/Categories/Node/audit-channel-names.sh"
