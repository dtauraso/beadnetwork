#!/usr/bin/env bash

# PLACEMENT: Node/**/*.go | a channel name must encode the two endpoints it connects (Node/audit-channel-names.sh)

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
bash "$REPO_ROOT/Node/audit-channel-names.sh"
