#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: src/schema/messages.ts,src/Overlay/overlay_tables_gen.go | OVERLAY_FLAG_NAMES (TS) and OverlayToggles keys (Go) must be the exact same name set

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
TS="$REPO_ROOT/src/schema/messages.ts"
GO="$REPO_ROOT/src/Overlay/overlay_tables_gen.go"

if [ ! -f "$TS" ] || [ ! -f "$GO" ]; then
  echo "check-overlay-flag-name-parity: MISCONFIGURED — one or both of these are missing:" >&2
  echo "  $TS" >&2
  echo "  $GO" >&2
  echo "  Both are the checked-in generated/source pair this guard exists to compare; a" >&2
  echo "  missing file is not 'nothing to compare', it means the guard has been silently" >&2
  echo "  disarmed. Refusing a vacuous pass." >&2
  exit 1
fi

ts_names=$(awk '/OVERLAY_FLAGS_START/{on=1;next} /OVERLAY_FLAGS_END/{on=0} on' "$TS" \
  | grep -oE '"[^"]+"' | tr -d '"' | sort)

# overlay_tables_gen.go: the map KEYS between OVERLAY_TOGGLES_START / OVERLAY_TOGGLES_END

go_names=$(awk '/OVERLAY_TOGGLES_START/{on=1;next} /OVERLAY_TOGGLES_END/{on=0} on' "$GO" \
  | grep -oE '"[^"]+"[[:space:]]*:' | grep -oE '"[^"]+"' | tr -d '"' | sort)

if [ "$ts_names" != "$go_names" ]; then
  echo "check-overlay-flag-name-parity: OVERLAY_FLAG_NAMES (messages.ts) and the"
  echo "OverlayToggles keys (overlay_tables_gen.go) diverge. Diff (< messages.ts, > overlay_tables_gen.go):"
  diff <(printf '%s\n' "$ts_names") <(printf '%s\n' "$go_names") || true
  echo "If you changed the overlay vocabulary, edit OVERLAY_FLAG_NAMES in messages.ts and"
  echo "regenerate (go run ./cmd/gen-node-defs) so overlay_tables_gen.go matches."
  exit 1
fi

exit 0
