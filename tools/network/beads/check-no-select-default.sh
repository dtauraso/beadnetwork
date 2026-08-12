#!/usr/bin/env bash

# PLACEMENT: nodes/wire/beadchain/bead_actor.go,nodes/wire/*.go | a bead goroutine's select must have NO default: case — it parks, it never spins
set -euo pipefail

# nodes/wire/beadchain/bead_actor.go's Bead.run, fenced by BEAD-SELECT-START/END, NOT a repo-wide "no

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
FILE="$REPO_ROOT/nodes/wire/beadchain/bead_actor.go"

if [ ! -f "$FILE" ]; then
  echo "✗ no-select-default: MISCONFIGURED — file not found: $FILE" >&2
  exit 1
fi

body=$(awk '/^\/\/ BEAD-SELECT-START$/,/^\/\/ BEAD-SELECT-END$/' "$FILE")

if [ -z "$body" ]; then
  echo "✗ no-select-default: MISCONFIGURED — BEAD-SELECT-START/END fence not found in $FILE" >&2
  exit 1
fi

if echo "$body" | grep -qE '^\s*default:'; then
  echo "✗ Bead.run's select carries a default: case — that makes it non-blocking, so a loop"
  echo "  around it spins a core instead of parking at zero CPU. Remove the default: case;"
  echo "  every event this loop reacts to must arrive as a channel message, never be polled."
  exit 1
fi

echo "✓ no default: in Bead.run's select (nodes/wire/beadchain/bead_actor.go) — the loop parks, it does not spin."
