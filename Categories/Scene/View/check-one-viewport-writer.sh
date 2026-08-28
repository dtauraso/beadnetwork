#!/usr/bin/env bash

# PLACEMENT: Categories/Scene/View/edit_apply.go | SetViewport is the only writer of the viewport

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
OWNER="Categories/Scene/View/edit_apply.go"

hits=$(git -C "$REPO_ROOT" grep -n -E '\.(ViewW|ViewH)[[:space:]]*=[^=]' -- '*.go' \
  | grep -v "^${OWNER}:" \
  | grep -v '^Categories/Scene/View/ui_state.go:' || true)

if [ -n "$hits" ]; then
  echo "✗ one viewport writer: something other than SetViewport assigns ViewW/ViewH" >&2
  echo "$hits" >&2
  echo >&2
  echo "  The viewport is reported by the frame that measures the canvas, in CSS pixels," >&2
  echo "  through $OWNER's SetViewport. A pointer event's bounding rect is a DIFFERENT" >&2
  echo "  quantity — unrounded, and right for hit-testing — and assigning it here gives the" >&2
  echo "  chrome two centres to alternate between." >&2
  exit 1
fi

echo "✓ one viewport writer"
