#!/usr/bin/env bash

# PLACEMENT: tools/topology-vscode/src/webview/**/*.ts,tools/topology-vscode/src/webview/**/*.tsx | no zustand and no useSyncExternalStore outside the named buffer-reflect resources
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

WEBVIEW_DIR="$REPO_ROOT/tools/topology-vscode/src/webview"

if [[ ! -d "$WEBVIEW_DIR" ]]; then
  echo "no-webview-state: MISCONFIGURED — webview dir not found at $WEBVIEW_DIR" >&2
  exit 1
fi

HITS=0
report() {
  printf '%s\n' "$1"
  HITS=$((HITS + 1))
}

while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  report "zustand-import: $line  (Zustand store in the webview — domain state must live in Go)"
done < <(grep -arnE 'from[[:space:]]+"zustand"' \
  --include="*.ts" --include="*.tsx" "$WEBVIEW_DIR" 2>/dev/null || true)

while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  report "zustand-create: $line  (store constructor in the webview — domain state must live in Go)"
done < <(grep -arnE '\bcreate[<(]' \
  --include="*.ts" --include="*.tsx" "$WEBVIEW_DIR" 2>/dev/null || true)

while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  f="${line%%:*}"
  base="$(basename "$f")"
  case "$base" in
    snapshot-buffer.ts|overlay-flags.ts|buffer-nav.ts|scene-tabs.ts) continue ;;
    overlay-flags-drag.ts|overlay-flags-edit-refused.ts|overlay-flags-scene.ts) continue ;;
    overlay-flags-selection.ts|overlay-flags-distance-groups.ts) continue ;;
    overlay-flags-speed.ts|overlay-flags-tilt-vectors.ts) continue ;;
  esac
  report "domain-hook: $line  (useSyncExternalStore outside the allowed buffer-reflect resources)"
done < <(grep -arnE '\buseSyncExternalStore\b' \
  --include="*.ts" --include="*.tsx" "$WEBVIEW_DIR" 2>/dev/null || true)

while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  report "reducer: $line  (useReducer in the webview — a domain state machine must live in Go)"
done < <(grep -arnE '\buseReducer\b' \
  --include="*.ts" --include="*.tsx" "$WEBVIEW_DIR" 2>/dev/null || true)

while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  f="${line%%:*}"
  base="$(basename "$f")"
  case "$base" in
    scene-env.tsx) continue ;;
  esac
  report "context: $line  (createContext outside the allowed render-infra files — domain state must live in Go)"
done < <(grep -arnE '\bcreateContext\b' \
  --include="*.ts" --include="*.tsx" "$WEBVIEW_DIR" 2>/dev/null || true)

if [[ $HITS -eq 0 ]]; then
  echo "no-webview-state: clean (webview holds no domain state; render + forward only, Go owns the model)"
  exit 0
fi

echo ""
echo "no-webview-state: $HITS hit(s) — the webview must hold no domain state (no Zustand store, no stateful domain hook); the model lives in Go and streams as the binary content buffer"
exit 1
