#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: tools/topology-vscode/Buffer/bufschema/layout_event.go | bufLayoutEvent may declare at most one `<Name>Off uint32` free-form string section

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LAYOUT_FILE="$REPO_ROOT/tools/topology-vscode/Buffer/bufschema/layout_event.go"

if [[ ! -f "$LAYOUT_FILE" ]]; then
  echo "check-event-string-section-singular: MISCONFIGURED — file not found: $LAYOUT_FILE" >&2
  exit 1
fi

BODY="$(awk '
  /^type bufLayoutEvent struct \{/ { grab=1; next }
  grab && /^\}/                    { grab=0 }
  grab                             { print }
' "$LAYOUT_FILE")"

if [[ -z "$BODY" ]]; then
  echo "check-event-string-section-singular: MISCONFIGURED — bufLayoutEvent struct not found"
  echo "  in $LAYOUT_FILE. If the struct was renamed, update this guard to match."
  exit 1
fi

OFF_FIELDS="$(printf '%s\n' "$BODY" | grep -oE '^[[:space:]]*[A-Za-z0-9_]+Off[[:space:]]+uint32' || true)"
OFF_COUNT="$(printf '%s' "$OFF_FIELDS" | grep -c . || true)"

if [[ "$OFF_COUNT" -gt 1 ]]; then
  echo "check-event-string-section-singular: bufLayoutEvent declares $OFF_COUNT string sections"
  echo "  (fields matching '<Name>Off uint32'):"
  printf '%s\n' "$OFF_FIELDS" | sed 's/^[[:space:]]*/    /'
  echo
  echo "  The event row may carry AT MOST ONE free-form string section. Multiple string"
  echo "  sections reintroduce the per-payload opaque-string sprawl the binary breadcrumb"
  echo "  conversion removed. New payload data must either REUSE an existing typed column"
  echo "  (Value/X/Y/Z/NodeRow/...) or ride the single sanctioned free-form text section —"
  echo "  never a new *Off/*Len string blob. See tools/buffer-schema/check-event-string-section-singular.sh."
  exit 1
fi

echo "check-event-string-section-singular: clean ($OFF_COUNT string section)"
exit 0
