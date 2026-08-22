#!/usr/bin/env bash

# PLACEMENT: src/extension/webview/**/*.ts,src/extension/webview/**/*.tsx | TS plots what Go streams: no bead positions, no edge curves, no traversal timing computed here
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

SRC_DIR="$REPO_ROOT/src"

if [[ ! -d "$SRC_DIR" ]]; then
  echo "ts-computes-no-geometry: MISCONFIGURED — scan dir not found: $SRC_DIR" >&2
  exit 1
fi

ts_file_count=$(find "$SRC_DIR" \( -name "*.ts" -o -name "*.tsx" \) | head -1 | wc -l | tr -d ' ')
if [[ "$ts_file_count" -eq 0 ]]; then
  echo "ts-computes-no-geometry: MISCONFIGURED — no .ts/.tsx files under $SRC_DIR" >&2
  exit 1
fi

FORBIDDEN=(
  "getPointAt"
  "rfArcLength"
  "arcLengthToSimLatencyMs"
  "patchPulse"
  "buildPortCurve"
  "buildEdgeCurve"
)

HITS=0
for token in "${FORBIDDEN[@]}"; do
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    printf '%s  (forbidden: %s)\n' "$line" "$token"
    HITS=$((HITS + 1))
  done < <(grep -arn --include="*.ts" --include="*.tsx" "$token" "$SRC_DIR" 2>/dev/null || true)
done

if [[ $HITS -eq 0 ]]; then
  echo "ts-computes-no-geometry: clean (TS plots Go's position stream; computes no bead geometry)"
  exit 0
fi

echo ""
echo "ts-computes-no-geometry: $HITS hit(s) — bead position/geometry math must live in Go, not TS"
exit 1
