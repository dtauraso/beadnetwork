#!/usr/bin/env bash

# PLACEMENT: tools/topology-vscode/src/webview/three/** | camera/nav code must not reconstruct angles from a Cartesian position (setFromVector3/makeSafe/THREE.Spherical banned outside polar.ts, PanPolarOverlay.tsx)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DIR="$REPO_ROOT/tools/topology-vscode/src/webview/three"
if [ ! -d "$DIR" ]; then
  echo "✗ no-camera-roundtrip: MISCONFIGURED — scan dir not found: $DIR" >&2
  exit 1
fi

EXCLUDE='polar\.ts|PanPolarOverlay\.tsx'

PATTERN='setFromVector3|setFromCartesianCoords|\.makeSafe\(|new THREE\.Spherical'

hits=$(grep -arnE "$PATTERN" "$DIR" --include='*.ts' --include='*.tsx' 2>/dev/null \
       | grep -vE "$EXCLUDE" || true)

if [ -n "$hits" ]; then
  echo "✗ camera round-trip fingerprint(s) found — the camera must stay angle-of-record:"
  echo "$hits"
  echo "  (reconstructing angles from a position reintroduces the pole singularity / makeSafe.)"
  exit 1
fi

echo "✓ no camera round-trip: camera state is not reconstructed from a Cartesian position."
