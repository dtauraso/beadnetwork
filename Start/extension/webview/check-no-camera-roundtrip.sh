#!/usr/bin/env bash

# PLACEMENT: Start/extension/webview/**,Categories/Camera/** | camera/nav code must not reconstruct angles from a Cartesian position (setFromVector3/makeSafe/THREE.Spherical banned outside polar.ts, PanPolarOverlay.tsx)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
DIRS=("$REPO_ROOT/Start/extension/webview" "$REPO_ROOT/Categories/Camera")
for d in "${DIRS[@]}"; do if [ ! -d "$d" ]; then
  echo "✗ no-camera-roundtrip: MISCONFIGURED — scan dir not found: $d" >&2
  exit 1
fi; done

EXCLUDE='polar\.ts|PanPolarOverlay\.tsx'

PATTERN='setFromVector3|setFromCartesianCoords|\.makeSafe\(|new THREE\.Spherical'

hits=$(grep -arnE "$PATTERN" "${DIRS[@]}" --include='*.ts' --include='*.tsx' 2>/dev/null \
       | grep -vE "$EXCLUDE" || true)

if [ -n "$hits" ]; then
  echo "✗ camera round-trip fingerprint(s) found — the camera must stay angle-of-record:"
  echo "$hits"
  echo "  (reconstructing angles from a position reintroduces the pole singularity / makeSafe.)"
  exit 1
fi

echo "✓ no camera round-trip: camera state is not reconstructed from a Cartesian position."
