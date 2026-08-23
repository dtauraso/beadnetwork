#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: Start/extension/webview/**,Categories/Node/**,Categories/Scene/**,Categories/Ring/**,Categories/Chrome/** | Go-owned shading props (roughness, ior, etc.) must reference shading-params.ts, not a numeric literal

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

SCAN_DIR=(
  "$REPO_ROOT/Start/extension/webview"
  "$REPO_ROOT/Categories/Node"
  "$REPO_ROOT/Categories/Scene"
  "$REPO_ROOT/Categories/Ring"
  "$REPO_ROOT/Categories/Chrome"
)

if [[ ! -d "$SCAN_DIR" ]]; then
  echo "ts-shading-from-go: render dir not found at $SCAN_DIR" >&2
  exit 1
fi

PROPS=(
  transmission
  thickness
  roughness
  ior
  metalness
  clearcoat
  clearcoatRoughness
  sheen
  sheenRoughness
  iridescence
  iridescenceIOR
  envMapIntensity
  attenuationDistance
  specularIntensity
  reflectivity
  emissiveIntensity
)

EXCLUDED_FILES=(SelectionHighlight.tsx PolarFrame.tsx PolarHandholds.tsx)

NUM='-?(0x[0-9a-fA-F]+|[0-9]+\.?[0-9]*([eE][-+]?[0-9]+)?)'

HITS=0
for prop in "${PROPS[@]}"; do
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    file="$(basename "${line%%:*}")"
    skip=0
    for ex in "${EXCLUDED_FILES[@]}"; do
      [[ "$file" == "$ex" ]] && { skip=1; break; }
    done
    [[ $skip -eq 1 ]] && continue

    if [[ "$prop" == "metalness" || "$prop" == "emissiveIntensity" ]] \
      && grep -qE "${prop}=\{0\}|${prop}:[[:space:]]*0([^0-9.]|\$)" <<< "$line"; then
      continue
    fi
    printf '%s  (forbidden shading literal: prop "%s" assigned a numeric literal — must reference a shading-params.ts import)\n' "$line" "$prop"
    HITS=$((HITS + 1))
  done < <(grep -arnE "\b${prop}[[:space:]]*=\{[[:space:]]*${NUM}[[:space:]]*\}|\b${prop}:[[:space:]]*${NUM}\b" \
    --include='*.ts' --include='*.tsx' "${SCAN_DIR[@]}" 2>/dev/null || true)
done

SHADING_SOURCES=(
  "Categories/Node/Edge/shading_params.go"
  "Categories/Node/Interior/shading_params.go"
  "Categories/Node/ChannelVectors/shading_params.go"
  "Categories/Node/Poles/shading_params.go"
  "Categories/Node/shading_params.go"
  "Categories/Ring/Bead/shading_params.go"
  "Categories/Ring/NodeShape/shading_params.go"
  "Categories/Scene/Environment/shading_params.go"
)
for src in "${SHADING_SOURCES[@]}"; do
  if [[ ! -f "$REPO_ROOT/$src" ]]; then
    echo "ts-shading-from-go: MISCONFIGURED — $src not found; a shading source moved and this guard" >&2
    echo "  would stop checking that category's params. Point it at where they live now." >&2
    exit 1
  fi
  ts="$REPO_ROOT/${src%/shading_params.go}/shading-params.ts"
  if [[ ! -f "$ts" ]]; then
    echo "ts-shading-from-go: $src has no generated shading-params.ts beside it — run go generate ./..."
    HITS=$((HITS + 1))
  fi
done

if ! grep -arq --include='*.ts' --include='*.tsx' 'shading-params"' "${SCAN_DIR[@]}"; then
  echo 'ts-shading-from-go: nothing imports a shading-params.ts — shading params must come from Go'
  HITS=$((HITS + 1))
fi
if ! grep -arq --include='*.ts' --include='*.tsx' 'SHADING_PARAM_NODE_TRANSMISSION' "${SCAN_DIR[@]}"; then
  echo 'ts-shading-from-go: SHADING_PARAM_NODE_TRANSMISSION not used — node glass material is not reading Go params'
  HITS=$((HITS + 1))
fi

if [[ $HITS -eq 0 ]]; then
  echo "ts-shading-from-go: clean (TS binds Go-supplied shading params; no relocated shading literals in three/)"
  exit 0
fi

echo ""
echo "ts-shading-from-go: $HITS hit(s) — shading parameter VALUES must live in Go, in the shading_params.go of the category they describe, not TS"
exit 1
