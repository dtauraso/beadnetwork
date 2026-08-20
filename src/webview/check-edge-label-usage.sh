#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: src/webview/**/*.ts,src/webview/**/*.tsx | only buffer-decode-edge.ts/buffer-layout.ts may read EdgeLabelOff/Len; the renderer must not

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
WEBVIEW_DIR="$REPO_ROOT/src/webview"

if [ ! -d "$WEBVIEW_DIR" ]; then
  echo "check-edge-label-usage: MISCONFIGURED — $WEBVIEW_DIR not found; refusing a vacuous pass." >&2
  echo "  This is the render tree the guard exists to police (EdgeTube.tsx et al.); if it" >&2
  echo "  moved or was deleted, the invariant it enforces no longer has a home. Update the" >&2
  echo "  guard deliberately." >&2
  exit 1
fi

PATTERN='readEdgeEdgeLabelOff|readEdgeEdgeLabelLen|EdgeLabelOff|EdgeLabelLen'

hits=$(grep -rnE "$PATTERN" "$WEBVIEW_DIR" --include="*.ts" --include="*.tsx" \
  | grep -v '/buffer-decode-edge.ts:' \
  | grep -v '/buffer-layout.ts:' \
  || true)

if [ -n "$hits" ]; then
  echo "check-edge-label-usage: the edge label columns must be read ONLY by buffer-decode-edge.ts"
  echo "(the .probe decoder), never by the render tree (.claude/rules/wire-props.md: label never"
  echo "feeds the render path). Offending references:"
  echo "$hits"
  exit 1
fi

exit 0
