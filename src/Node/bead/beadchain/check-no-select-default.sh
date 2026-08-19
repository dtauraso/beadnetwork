#!/usr/bin/env bash

# PLACEMENT: src/Node/bead/beadchain/bead_actor.go,src/Node/bead/*.go | a bead goroutine's select must have NO default: case — it parks, it never spins
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
FILE="$REPO_ROOT/src/Node/bead/beadchain/bead_actor.go"

if [ ! -f "$FILE" ]; then
  echo "✗ no-select-default: MISCONFIGURED — file not found: $FILE" >&2
  exit 1
fi

body=$(awk '/^func \(b \*Bead\) run\(\) \{$/,/^\}$/' "$FILE")

if [ -z "$body" ]; then
  echo "✗ no-select-default: MISCONFIGURED — func (b *Bead) run() not found in $FILE" >&2
  exit 1
fi

if echo "$body" | grep -qE '^\s*default:'; then
  echo "✗ Bead.run's select carries a default: case — that makes it non-blocking, so a loop"
  echo "  around it spins a core instead of parking at zero CPU. Remove the default: case;"
  echo "  every event this loop reacts to must arrive as a channel message, never be polled."
  exit 1
fi

echo "✓ no default: in Bead.run's select (src/Node/bead/beadchain/bead_actor.go) — the loop parks, it does not spin."
