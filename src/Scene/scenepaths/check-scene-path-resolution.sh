#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: src/**/*.go | view/*.json path resolution lives only in scene_paths.go; nodes/ path Join lives only in node_mover.go/new_node_files.go/edge_mover.go/loader_tree.go/tree_shape.go

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
WIRING_DIR="$REPO_ROOT/src"
RESOLVER="$REPO_ROOT/src/Scene/scenepaths/scene_paths.go"

if [[ ! -d "$WIRING_DIR" ]]; then
  echo "check-scene-path-resolution: MISCONFIGURED — $WIRING_DIR not found (moved/renamed?)." >&2
  echo "  Refusing to report clean without scanning anything; update WIRING_DIR in $(basename "$0")." >&2
  exit 1
fi

if [[ ! -f "$RESOLVER" ]]; then
  echo "check-scene-path-resolution: MISCONFIGURED — $RESOLVER not found." >&2
  echo "  This guard exempts scene_paths.go as the authoritative resolver; if it is gone," >&2
  echo "  the invariant it enforces no longer has a home. Update the guard deliberately." >&2
  exit 1
fi

GO_FILE_COUNT=$(find "$WIRING_DIR" -name "*.go" | wc -l | tr -d ' ')
if [[ "$GO_FILE_COUNT" -eq 0 ]]; then
  echo "check-scene-path-resolution: MISCONFIGURED — no .go files found under $WIRING_DIR." >&2
  echo "  The scan must actually see source files; refusing a vacuous pass." >&2
  exit 1
fi

HITS=0
report() {
  printf '%s\n' "$1"
  HITS=$((HITS + 1))
}

eligible_files=()
while IFS= read -r file; do
  [[ "$file" == *"_test.go" ]] && continue
  [[ "${file##*/}" == "scene_paths.go" ]] && continue
  eligible_files+=("$file")
done < <(find "$WIRING_DIR" -name "*.go" -not -path "*/node_modules/*")

all_hits=""
if [[ ${#eligible_files[@]} -gt 0 ]]; then
  all_hits="$(grep -nE 'IsDir\(\)|scenepaths\.InputFilePath\(|scenepaths\.SelectionFilePath\(|scenepaths\.SpeedFilePath\(|scenepaths\.LatticeFilePath\(|filepath\.Join\(' \
    "${eligible_files[@]}" 2>/dev/null || true)"
fi

while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  file="${hit%%:*}"; rest="${hit#*:}"; lineno="${rest%%:*}"; content="${rest#*:}"
  [[ "$content" == *"IsDir()"* ]] || continue
  [[ "$content" == *"// path-resolution-ok:"* ]] && continue
  report "hand-rolled-IsDir: $file: $lineno:$content"
done <<< "$all_hits"

if [[ $HITS -ne 0 ]]; then
  echo ""
  echo "check-scene-path-resolution: $HITS hit(s) — resolve topologyPath via the scene_paths.go resolvers, not hand-rolled IsDir. Mark unrelated uses with '// path-resolution-ok:'"
  exit 1
fi

CALL_SITES=0
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  content="${hit#*:*:}"
  case "$content" in
    *scenepaths.InputFilePath\(*|*scenepaths.SelectionFilePath\(*|*scenepaths.SpeedFilePath\(*|*scenepaths.LatticeFilePath\(*) CALL_SITES=$((CALL_SITES + 1)) ;;
  esac
done <<< "$all_hits"

if [[ "$CALL_SITES" -eq 0 ]]; then
  echo "check-scene-path-resolution: MISCONFIGURED — zero call sites of scenepaths.InputFilePath()/scenepaths.SelectionFilePath()/scenepaths.SpeedFilePath()/scenepaths.LatticeFilePath() found outside scenepaths/scene_paths.go." >&2
  echo "  The resolvers exist but nothing calls them; the IsDir-only scan above would pass vacuously." >&2
  exit 1
fi

JOIN_HITS=0
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  file="${hit%%:*}"; rest="${hit#*:}"; lineno="${rest%%:*}"; content="${rest#*:}"
  [[ "$content" == *"filepath.Join("* ]] || continue
  if [[ "$content" == *'"view"'* ]]; then
    printf 'hand-rolled-join: %s: %s:%s\n' "$file" "$lineno" "$content"
    JOIN_HITS=$((JOIN_HITS + 1))
  fi
done <<< "$all_hits"

if [[ "$JOIN_HITS" -ne 0 ]]; then
  echo ""
  echo "check-scene-path-resolution: $JOIN_HITS hand-rolled filepath.Join(\"view\", ...) hit(s) outside scene_paths.go — call the shared resolver in scene_paths.go instead."
  exit 1
fi

NODE_PATH_OWNERS=("new_node_files.go" "edge_file.go" "edge_delta_file.go" "loader_tree.go" "tree_shape.go" "drag_file.go")
is_node_path_owner() {
  local f="$1"
  [[ "$f" == */gen/* ]] && return 0
  for owner in "${NODE_PATH_OWNERS[@]}"; do
    [[ "${f##*/}" == "$owner" ]] && return 0
  done
  return 1
}
NODE_JOIN_HITS=0
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  file="${hit%%:*}"; rest="${hit#*:}"; lineno="${rest%%:*}"; content="${rest#*:}"
  [[ "$content" == *"filepath.Join("* ]] || continue
  [[ "$content" == *'"nodes"'* ]] || continue
  is_node_path_owner "$file" && continue
  printf 'hand-rolled-node-path: %s: %s:%s\n' "$file" "$lineno" "$content"
  NODE_JOIN_HITS=$((NODE_JOIN_HITS + 1))
done <<< "$all_hits"

if [[ "$NODE_JOIN_HITS" -ne 0 ]]; then
  echo ""
  echo "check-scene-path-resolution: $NODE_JOIN_HITS hand-rolled nodes/ filepath.Join(...) hit(s) outside node_mover.go/edge_mover.go/edge_file.go/loader_tree.go/tree_shape.go/drag_file.go — a node/port path belongs to its owning mover; call node_mover.go's or dragfile's resolvers instead of reconstructing the path."
  exit 1
fi

echo "check-scene-path-resolution: clean ($GO_FILE_COUNT files scanned; $CALL_SITES resolver call site(s); all IsDir/Join path-resolution lives in scene_paths.go; node-path construction lives in node_mover.go/edge_mover.go/edge_file.go/dragfile)"
exit 0
