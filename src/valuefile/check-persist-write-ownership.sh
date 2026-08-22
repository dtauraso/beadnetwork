#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: src/**/*.go | writeJSONAtomic/entityReadModifyWrite calls must live in the owning file (node/edge/view mover), never elsewhere

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
WIRING_DIR="$REPO_ROOT/src"
PLUMBING="value_file.go"
PLUMBING_PATH="$WIRING_DIR/valuefile/$PLUMBING"

if [[ ! -d "$WIRING_DIR" ]]; then
  echo "check-persist-write-ownership: MISCONFIGURED — $WIRING_DIR not found (moved/renamed?)." >&2
  exit 1
fi
if [[ ! -f "$PLUMBING_PATH" ]]; then
  echo "check-persist-write-ownership: MISCONFIGURED — $PLUMBING_PATH not found." >&2
  echo "  This guard exempts src/valuefile as the shared write primitive; if it is" >&2
  echo "  gone, the invariant it enforces no longer has a home. Update the guard deliberately." >&2
  exit 1
fi

NODE_OWNERS=("node_mover.go" "new_node_files.go" "node_base_file.go" "quant_offset_persist.go" "scene_anchor_persist.go" "drag_file.go")

EDGE_OWNERS=("edge_file.go" "edge_delta_file.go" "edge_rule_active.go" "out_edges.go")

VIEW_OWNERS=("scene_camera_persist.go" "overlays_persist.go" "panels_persist.go" "scene_sphere_persist.go" "scene_selection_persist.go" "scene_speed_persist.go" "lattice.go" "scene_spawn_persist.go")

TREE_OWNERS=("counts.go")

in_list() {
  local needle="$1"; shift
  for x in "$@"; do [[ "$needle" == "$x" ]] && return 0; done
  return 1
}

GO_FILE_COUNT=$(find "$WIRING_DIR" -name "*.go" -not -name "*_test.go" | wc -l | tr -d ' ')
if [[ "$GO_FILE_COUNT" -eq 0 ]]; then
  echo "check-persist-write-ownership: MISCONFIGURED — no non-test .go files found under $WIRING_DIR." >&2
  exit 1
fi

eligible_files=()
while IFS= read -r file; do
  [[ "${file##*/}" == "$PLUMBING" ]] && continue
  eligible_files+=("$file")
done < <(find "$WIRING_DIR" -name "*.go" -not -name "*_test.go")

all_hits=""
if [[ ${#eligible_files[@]} -gt 0 ]]; then
  all_hits="$(grep -nE 'valuefile\.WriteAtomic\(|valuefile\.WriteAtomicIfChanged\(' "${eligible_files[@]}" 2>/dev/null || true)"
fi

HITS=0
report() { printf '%s\n' "$1"; HITS=$((HITS + 1)); }

TOTAL_CALLS=0
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  file="${hit%%:*}"; rest="${hit#*:}"; lineno="${rest%%:*}"; content="${rest#*:}"
  base="${file##*/}"

  if grep -qE '^[[:space:]]*func[[:space:]]+(\([^)]*\)[[:space:]]+)?(writeJSONAtomic|entityReadModifyWrite)[[:space:]]*\(' <<< "$content"; then
    continue
  fi
  TOTAL_CALLS=$((TOTAL_CALLS + 1))

  if in_list "$base" "${VIEW_OWNERS[@]}"; then
    continue
  fi
  if in_list "$base" "${TREE_OWNERS[@]}"; then
    continue
  fi
  if in_list "$base" "${EDGE_OWNERS[@]}"; then
    continue
  fi
  if in_list "$base" "${NODE_OWNERS[@]}"; then

    if [[ "$content" == *"edges"* ]]; then
      report "node-owner-wrote-edge-path: $file: $lineno:$content"
    fi
    continue
  fi
  report "unauthorized-write: $file: $lineno:$content"
done <<< "$all_hits"

if [[ "$TOTAL_CALLS" -eq 0 ]]; then
  echo "check-persist-write-ownership: MISCONFIGURED — zero writeJSONAtomic()/entityReadModifyWrite() call sites found outside $PLUMBING." >&2
  echo "  The scan must actually see real write call sites; refusing a vacuous pass." >&2
  exit 1
fi

if [[ $HITS -ne 0 ]]; then
  echo ""
  echo "check-persist-write-ownership: $HITS hit(s) — a per-node file may be written only by the node's own mover (${NODE_OWNERS[*]}), a per-edge outgoing-edge file only by its SOURCE NODE (out_edges.go), and a scene-level view/ file only by the view-owner goroutine's own files (${VIEW_OWNERS[*]})."
  exit 1
fi

echo "check-persist-write-ownership: clean ($TOTAL_CALLS write call site(s) scanned across $GO_FILE_COUNT files; every write stays with its owner)"
exit 0
