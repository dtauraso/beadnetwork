#!/usr/bin/env bash

# PLACEMENT: src/webview/scene/interaction-controls.ts | the nav handler does no Cartesian math; angle/sphere math lives in the polar helpers

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
NAV_DIR="$REPO_ROOT/src/webview/scene"

shopt -s nullglob
NAV_FILES=( "$NAV_DIR"/interaction-*.ts )
shopt -u nullglob

if [ ${#NAV_FILES[@]} -eq 0 ]; then
  echo "✗ polar-only nav: MISCONFIGURED — no interaction-*.ts files under $NAV_DIR" >&2
  echo "  (nav handlers moved/renamed? update NAV_DIR in $(basename "$0"))" >&2
  exit 1
fi

PATTERN='setFromUnitVectors|\.cross\(|new THREE\.Raycaster|\.unproject\(|setFromAxisAngle|setFromMatrixColumn|new THREE\.Spherical|Math\.atan2|Math\.acos'

hits=$(grep -anE "$PATTERN" "${NAV_FILES[@]}" 2>/dev/null | grep -v 'polar-nav-ok' || true)

if [ -n "$hits" ]; then
  echo "✗ polar-nav violation(s) found — all rotation/axis math must live in polar.ts:"
  echo "$hits"
  echo "  (banned: setFromUnitVectors, .cross(, new THREE.Raycaster, .unproject(, setFromAxisAngle, setFromMatrixColumn, new THREE.Spherical, Math.atan2, Math.acos)"
  exit 1
fi

echo "✓ polar-only nav: no banned Cartesian rotation math in ${NAV_FILES[*]}."
