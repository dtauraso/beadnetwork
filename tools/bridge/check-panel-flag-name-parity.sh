#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: tools/topology-vscode/src/messages.ts,tools/topology-vscode/OverlaysDropdown/panel_state.go | PANEL_FLAG_NAMES (TS) and PanelToggles keys (Go) must be the exact same name set

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TS="$REPO_ROOT/tools/topology-vscode/src/messages.ts"
GO="$REPO_ROOT/tools/topology-vscode/OverlaysDropdown/panel_state.go"

if [ ! -f "$TS" ] || [ ! -f "$GO" ]; then
  echo "check-panel-flag-name-parity: MISCONFIGURED — one or both of these are missing:" >&2
  echo "  $TS" >&2
  echo "  $GO" >&2
  echo "  Both are the checked-in source pair this guard exists to compare; a" >&2
  echo "  missing file is not 'nothing to compare', it means the guard has been silently" >&2
  echo "  disarmed. Refusing a vacuous pass." >&2
  exit 1
fi

ts_names=$(awk '/PANEL_FLAGS_START/{on=1;next} /PANEL_FLAGS_END/{on=0} on' "$TS" \
  | grep -oE '"[^"]+"' | tr -d '"' | sort)

# panel_state.go: the map KEYS between PANEL_TOGGLES_START / PANEL_TOGGLES_END

go_names=$(awk '/PANEL_TOGGLES_START/{on=1;next} /PANEL_TOGGLES_END/{on=0} on' "$GO" \
  | grep -oE '"[^"]+"[[:space:]]*:' | grep -oE '"[^"]+"' | tr -d '"' | sort)

if [ "$ts_names" != "$go_names" ]; then
  echo "check-panel-flag-name-parity: PANEL_FLAG_NAMES (messages.ts) and the"
  echo "PanelToggles keys (panel_state.go) diverge. Diff (< messages.ts, > panel_state.go):"
  diff <(printf '%s\n' "$ts_names") <(printf '%s\n' "$go_names") || true
  echo "If you changed the panel vocabulary, edit PANEL_FLAG_NAMES in messages.ts and"
  echo "PanelToggles in panel_state.go so they match (panel state is hand-written, not generated)."
  exit 1
fi

exit 0
