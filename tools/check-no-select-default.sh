#!/usr/bin/env bash
set -euo pipefail

# check-no-select-default.sh — guard that the bead goroutine's own select carries no
# `default:` case.
#
# PLAN.md "two clocks per bead" / MODEL.md: a `select` WITH `default:` is non-blocking, so a
# loop wrapped around it SPINS and burns a core; WITHOUT one, the runtime parks the
# goroutine on every case's wait queue at zero CPU until something is ready. This is
# invisible to any behavioural test (both versions pass identically under a real drag) —
# only a source guard can catch a `default:` creeping back in. Scoped to the exact select in
# nodes/wire/bead_actor.go's Bead.run, fenced by BEAD-SELECT-START/END, NOT a repo-wide "no
# default in any select" rule — sendStepsNonBlocking and friends elsewhere in this package
# correctly rely on `default:` for a non-blocking latest-wins send, and banning it there
# would break that idiom.
#
# Exit 0 clean, exit 1 with a report.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FILE="$REPO_ROOT/nodes/wire/bead_actor.go"

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

echo "✓ no default: in Bead.run's select (nodes/wire/bead_actor.go) — the loop parks, it does not spin."
