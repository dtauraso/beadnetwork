#!/usr/bin/env bash
set -euo pipefail

# check-persist-write-ownership.sh — guard: PER-OWNER file WRITES (not just path
# construction — that half is check-scene-path-resolution.sh's job).
#
# docs/planning/decentralized-persistence.md "The model": the goroutine that owns a piece
# of state writes its own file. Steps 4a-4c moved every write call site onto its owning
# goroutine (a node's own nodeMover for its own files; the view-owner goroutine,
# RunStdinReader, for the three scene-level files). This guard is what stops a FUTURE
# refactor from quietly routing a write back through a shared coordinator — the exact
# regression step 3's path-construction guard alone cannot catch, because it only checks
# who builds a path string, not who actually calls the OS write.
#
# Every real on-disk write in this package funnels through exactly two primitives
# (scene_persist.go): writeJSONAtomic (whole-file marshal+write) and entityReadModifyWrite
# (read-modify-write, used only by the port-anchor writer). scene_persist.go itself is
# exempted as shared plumbing every owner calls THROUGH — same exemption shape as
# scene_paths.go in check-scene-path-resolution.sh.
#
# Ownership is matched by PATH PATTERN, not directory (docs/planning/decentralized-
# persistence.md "The model": after step 2, nodes/<id>/ legitimately holds files owned by
# TWO different kinds of mover — the node's own, and each outgoing edge's):
#
#   - nodes/<id>/position.json, local-polars.json, {inputs,outputs}/<port>.json
#     (i.e. everything under nodes/<id>/ EXCEPT edges/) — written only by the node's own
#     mover. Today that write call lives in quant_offset_persist.go (position.json,
#     local-polars.json) and scene_anchor_persist.go (port-anchor files) — both define
#     nodeMover methods (persistQuantOffset/persistLocalPolars/persistPortAnchor) called
#     exclusively from that node's own goroutine. node_mover.go itself holds the path
#     resolvers (positionFilePath/localPolarsFilePath/nodePortFilePath) but no direct
#     writeJSONAtomic/entityReadModifyWrite call of its own today; it is listed as an
#     allowed owner so a future write added directly there does not need this guard
#     touched again.
#   - nodes/<id>/edges/<label>.json — reserved for edge_mover.go. NO writer exists in this
#     codebase today (edges are editor-authored; loader_tree.go only reads them — see the
#     plan's "Explicitly out of scope: Loading"), so the correct state is ZERO write call
#     sites naming an edges/ path anywhere. This guard does not invent a fake writer to
#     give itself something to match; instead it asserts two things that stay true whether
#     or not edge_mover.go exists yet: (1) no NODE-OWNER file's write call resolves an
#     edges/ path (that would be a node mover reaching into an edge's file — the exact
#     coupling this model forbids), and (2) if a write call ever appears in edge_mover.go,
#     it is accepted without editing this guard.
#   - view/camera.json, view/overlays.json, view/sphere.json — written only by the
#     view-owner goroutine's own files: scene_camera_persist.go, scene_overlays_persist.go,
#     scene_sphere_persist.go.
#
# Exit 0 when clean.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WIRING_DIR="$REPO_ROOT/nodes/Wiring"
PLUMBING="scene_persist.go"

if [[ ! -d "$WIRING_DIR" ]]; then
  echo "check-persist-write-ownership: MISCONFIGURED — $WIRING_DIR not found (moved/renamed?)." >&2
  exit 1
fi
if [[ ! -f "$WIRING_DIR/$PLUMBING" ]]; then
  echo "check-persist-write-ownership: MISCONFIGURED — $WIRING_DIR/$PLUMBING not found." >&2
  echo "  This guard exempts scene_persist.go as the shared write primitive; if it is gone," >&2
  echo "  the invariant it enforces no longer has a home. Update the guard deliberately." >&2
  exit 1
fi

# Node-owner files: files whose writeJSONAtomic/entityReadModifyWrite call acts on behalf
# of a node's own mover (a nodeMover method, called only from that node's own goroutine).
NODE_OWNERS=("node_mover.go" "quant_offset_persist.go" "scene_anchor_persist.go")
# Edge-owner files: reserved for a future Go-side edges/<label>.json writer.
EDGE_OWNERS=("edge_mover.go")
# View-owner files: the view-owner goroutine's (RunStdinReader) own scene-level writers.
VIEW_OWNERS=("scene_camera_persist.go" "scene_overlays_persist.go" "scene_sphere_persist.go")

in_list() {
  local needle="$1"; shift
  for x in "$@"; do [[ "$needle" == "$x" ]] && return 0; done
  return 1
}

GO_FILE_COUNT=$(find "$WIRING_DIR" -maxdepth 1 -name "*.go" -not -name "*_test.go" | wc -l | tr -d ' ')
if [[ "$GO_FILE_COUNT" -eq 0 ]]; then
  echo "check-persist-write-ownership: MISCONFIGURED — no non-test .go files found under $WIRING_DIR." >&2
  exit 1
fi

eligible_files=()
while IFS= read -r file; do
  [[ "${file##*/}" == "$PLUMBING" ]] && continue
  eligible_files+=("$file")
done < <(find "$WIRING_DIR" -maxdepth 1 -name "*.go" -not -name "*_test.go")

all_hits=""
if [[ ${#eligible_files[@]} -gt 0 ]]; then
  all_hits="$(grep -nE 'writeJSONAtomic\(|entityReadModifyWrite\(' "${eligible_files[@]}" 2>/dev/null || true)"
fi

HITS=0
report() { printf '%s\n' "$1"; HITS=$((HITS + 1)); }

TOTAL_CALLS=0
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  file="${hit%%:*}"; rest="${hit#*:}"; lineno="${rest%%:*}"; content="${rest#*:}"
  base="${file##*/}"
  # Skip the function DEFINITION lines (func writeJSONAtomic(... / func (...)
  # entityReadModifyWrite(...) — matched only in scene_persist.go, already excluded above,
  # but a defensive skip here in case a definition is ever duplicated elsewhere).
  [[ "$content" == *"func "* ]] && continue
  TOTAL_CALLS=$((TOTAL_CALLS + 1))

  if in_list "$base" "${VIEW_OWNERS[@]}"; then
    continue
  fi
  if in_list "$base" "${EDGE_OWNERS[@]}"; then
    continue
  fi
  if in_list "$base" "${NODE_OWNERS[@]}"; then
    # A node-owner file may never resolve an edges/ path — that would be a node's own
    # mover reaching into an edge's file, the exact per-owner violation this model
    # forbids. edge_mover.go is the only file allowed to write there (once it exists).
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
  echo "check-persist-write-ownership: $HITS hit(s) — a per-node file may be written only by the node's own mover (${NODE_OWNERS[*]}), a per-edge outgoing-edge file only by edge_mover.go, and a scene-level view/ file only by the view-owner goroutine's own files (${VIEW_OWNERS[*]})."
  exit 1
fi

echo "check-persist-write-ownership: clean ($TOTAL_CALLS write call site(s) scanned across $GO_FILE_COUNT files; every write stays with its owner)"
exit 0
