#!/usr/bin/env bash

# PLACEMENT: none | universal prose hygiene: retired vocabulary is banned everywhere, not in one place
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

readonly DEAD_COMMENT_TOKENS=(
  "fan-in is safe"
  "fan-in safe"

  "SnapshotState"
  "handleFd3"
  "fd-3 fallback"
  "fd3 fallback"

  "flood-to-all"
  "plain flood"
  "full flood"
  "flood fallback"

  "fd3 side channel"
  "fd-3 side channel"
  "fd-3 content buffer"
  "fd3 content buffer"
  "fd-3 SCENE"
  "fd3 SCENE"
  "fd-3 node frame"
  "fd-3 Node block"
  "fd3 binary side"
  "fd3 snapshot"

  "atomically-published"
  "atomic-snapshot-backed"
  "the atomic held"

  "alphabetically-sorted"

  "Wiring.Register("

  "useCameraStore"
  "CameraFromStore"
  "pump.ts"

  "Go mirror of the port-to-port segment geometry"

  "discovered by reflectPorts"
  "derived by reflectPorts"
  "filled in by reflectBuild"
  "injected by reflectBuild"

  "window-and-inhibit gate loop"

  "reads topology.json"
  "topology/edges/"

  "falls back to the legacy scene.json"
  "falls back to legacyScenePath"
  "tries camera.json first and falls back"
  "tries sphere.json first and falls back"
  "tries overlays.json first and falls back"
  "Legacy fallback: pre-split topology only has scene.json"
)

fail=0

all_hits="$(git ls-files -z --cached --others --exclude-standard '*.go' '*.ts' '*.tsx' \
    | xargs -0 grep -nIF -f <(printf '%s\n' "${DEAD_COMMENT_TOKENS[@]}") -- 2>/dev/null \
    | grep -vF "tools/docs/check-comment-vocab.sh" \
    | grep -E ':[[:space:]]*(//|\*|#)' || true)"

for token in "${DEAD_COMMENT_TOKENS[@]}"; do

  hits="$(printf '%s\n' "$all_hits" | awk -F: -v t="$token" '
    { content = $0; sub(/^[^:]*:[0-9]*:/, "", content); if (index(content, t) > 0) print }
  ' || true)"
  if [ -n "$hits" ]; then
    echo "RETIRED COMMENT VOCAB: '$token' — remove or reword; it contradicts the current model:"
    printf '%s\n' "$hits"
    fail=1
  fi
done

exit $fail
