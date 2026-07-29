#!/usr/bin/env bash
set -euo pipefail

# check-scene-path-resolution.sh — guard: PER-OWNER path construction.
#
# The monolithic-vs-tree topologyPath form this guard originally defended against is gone
# (docs/planning/decentralized-persistence.md step 1: topologyPath is always the tree root
# directory now). The invariant this guard enforces has moved on to the CURRENT plan
# (same doc, "The model" / step 3-5): a path is constructed only by the goroutine that
# owns the file it names.
#
#   - view/*.json (camera, overlays, sphere, the legacy scene.json sidecar) — scene_paths.go
#     only. Enforced the original way: no hand-rolled os.Stat+IsDir or filepath.Join("view",
#     "scene.json") outside scene_paths.go (still real — scene_camera.go et al. must call
#     the shared sceneJSONPath/sceneCameraPath/sceneViewFilePath resolvers, not reimplement
#     them, so a future scene file split can't silently diverge again).
#   - nodes/<id>/... (position, local-polars, cascade-edges, inputs/outputs port files) —
#     node_mover.go only (plus loader_tree.go, which READS these paths to build the graph
#     at load time — an explicitly out-of-scope concern, see the plan's "Explicitly out of
#     scope: Loading").
#   - nodes/<id>/edges/<label>.json — reserved for edge_mover.go once a Go-side writer
#     exists (today there is none; loader_tree.go's read is the only occurrence, also
#     exempted as loading).
#
# UNLIKE the view-file rule, the node-path rule below matches on the LITERAL "nodes"
# path segment rather than requiring a shared resolver function — no such function
# exists (or is wanted) for node paths: node_mover.go's positionFilePath/
# localPolarsFilePath/cascadeEdgesFilePath/nodePortFilePath ARE the resolvers, called by
# quant_offset_persist.go / scene_anchor_persist.go, so this guard's job is just "no
# OTHER file hand-rolls a nodes/ path with its own filepath.Join".
#
# Exit 0 when clean.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WIRING_DIR="$REPO_ROOT/nodes/Wiring"
RESOLVER="$WIRING_DIR/scene_paths.go"

# POSITIVE ASSERTIONS — this guard had none, and was the only one in the suite that could
# pass vacuously. `find` on a missing dir writes to stderr and emits nothing; process-
# substitution failure does NOT trip `set -e`; the while loop then reads zero lines, HITS
# stays 0, and it prints "clean (all IsDir path-resolution lives in scene_paths.go)" — a
# confident all-clear having scanned nothing. Rename nodes/Wiring/ and it congratulates you.
#
# Its siblings already do this (check-ts-computes-no-geometry.sh has both a dir check and a
# "the scan must actually see source files" count); that hardening never propagated here.
if [[ ! -d "$WIRING_DIR" ]]; then
  echo "check-scene-path-resolution: MISCONFIGURED — $WIRING_DIR not found (moved/renamed?)." >&2
  echo "  Refusing to report clean without scanning anything; update WIRING_DIR in $(basename "$0")." >&2
  exit 1
fi

# The file whose authority is the entire point of the guard must exist. Exempting a resolver
# that has been deleted would make every other file trivially compliant.
if [[ ! -f "$RESOLVER" ]]; then
  echo "check-scene-path-resolution: MISCONFIGURED — $RESOLVER not found." >&2
  echo "  This guard exempts scene_paths.go as the authoritative resolver; if it is gone," >&2
  echo "  the invariant it enforces no longer has a home. Update the guard deliberately." >&2
  exit 1
fi

GO_FILE_COUNT=$(find "$WIRING_DIR" -maxdepth 1 -name "*.go" | wc -l | tr -d ' ')
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

# Collect the eligible file list ONCE (excluding test files and the resolver itself), and
# reuse it for all three passes below, instead of walking $WIRING_DIR three separate times
# plus spawning `basename`/`grep` once PER FILE PER PASS (121 files x 3 passes = ~370 greps
# for this alone). `${f##*/}` replaces `basename "$f"` inline — same result, no process.
eligible_files=()
while IFS= read -r file; do
  [[ "$file" == *"_test.go" ]] && continue
  [[ "${file##*/}" == "scene_paths.go" ]] && continue
  eligible_files+=("$file")
done < <(find "$WIRING_DIR" -maxdepth 1 -name "*.go" -not -path "*/node_modules/*")

# One `grep -n` across the WHOLE eligible file list for all three patterns at once (IsDir(),
# the three resolver-call names, and filepath.Join() ) — grep prefixes each hit with
# "file:line:" when given multiple files, so each pass below re-parses that prefix and
# classifies the hit by which pattern matched, applying that pass's own exemptions exactly
# as the original per-pass loop did.
all_hits=""
if [[ ${#eligible_files[@]} -gt 0 ]]; then
  all_hits="$(grep -nE 'IsDir\(\)|sceneTreeRoot\(|sceneJSONPath\(|sceneCameraPath\(|filepath\.Join\(' \
    "${eligible_files[@]}" 2>/dev/null || true)"
fi

# Pass 1 — hand-rolled IsDir(), excluding lines carrying the exemption marker.
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  file="${hit%%:*}"; rest="${hit#*:}"; lineno="${rest%%:*}"; content="${rest#*:}"
  [[ "$content" == *"IsDir()"* ]] || continue
  [[ "$content" == *"// path-resolution-ok:"* ]] && continue
  report "hand-rolled-IsDir: $file: $lineno:$content"
done <<< "$all_hits"

if [[ $HITS -ne 0 ]]; then
  echo ""
  echo "check-scene-path-resolution: $HITS hit(s) — resolve topologyPath via sceneTreeRoot/sceneJSONPath in scene_paths.go, not hand-rolled IsDir. Mark unrelated uses with '// path-resolution-ok:'"
  exit 1
fi

# POSITIVE ASSERTION #2 — the IsDir scan above only proves nobody hand-rolls os.Stat+IsDir.
# It says nothing about whether the resolver functions themselves are actually CALLED
# anywhere in the package outside their own definition file: a persister could resolve a
# scene path with a hand-rolled filepath.Join("view", "scene.json") that never touches
# os.Stat/IsDir at all and this guard would still report clean. Require at least one real
# call site of sceneTreeRoot(/sceneJSONPath(/sceneCameraPath( outside scene_paths.go and
# outside tests, proving the resolver is load-bearing, not dead code the IsDir scan
# vacuously credits.
CALL_SITES=0
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  content="${hit#*:*:}"
  case "$content" in
    *sceneTreeRoot\(*|*sceneJSONPath\(*|*sceneCameraPath\(*) CALL_SITES=$((CALL_SITES + 1)) ;;
  esac
done <<< "$all_hits"

if [[ "$CALL_SITES" -eq 0 ]]; then
  echo "check-scene-path-resolution: MISCONFIGURED — zero call sites of sceneTreeRoot()/sceneJSONPath()/sceneCameraPath() found outside scene_paths.go." >&2
  echo "  The resolver exists but nothing calls it; the IsDir-only scan above would pass vacuously." >&2
  exit 1
fi

# POSITIVE ASSERTION #3 — reject a persister that resolves scene.json by hand-rolling
# filepath.Join with the literal path segments ("view", "scene.json") instead of calling
# the shared resolver. This is the exact bug shape the resolver exists to make
# unrepresentable, just spelled without IsDir: no os.Stat, no IsDir, straight Join — passes
# the scan above clean while silently breaking the file-form topologyPath case.
JOIN_HITS=0
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  file="${hit%%:*}"; rest="${hit#*:}"; lineno="${rest%%:*}"; content="${rest#*:}"
  [[ "$content" == *"filepath.Join("* ]] || continue
  if [[ "$content" == *'"view"'* && "$content" == *'"scene.json"'* ]]; then
    printf 'hand-rolled-join: %s: %s:%s\n' "$file" "$lineno" "$content"
    JOIN_HITS=$((JOIN_HITS + 1))
  fi
done <<< "$all_hits"

if [[ "$JOIN_HITS" -ne 0 ]]; then
  echo ""
  echo "check-scene-path-resolution: $JOIN_HITS hand-rolled filepath.Join(\"view\", \"scene.json\") hit(s) outside scene_paths.go — call sceneJSONPath/sceneCameraPath instead."
  exit 1
fi

# POSITIVE ASSERTION #4 — node-path ownership. filepath.Join(...) with a literal "nodes"
# path segment is allowed only in node_mover.go (the node-path resolvers), edge_mover.go
# (reserved for a future edges/<label>.json writer), and loader_tree.go (the tree reader —
# loading is explicitly out of scope for the owner-writes-its-path invariant). Any other
# file hand-rolling a nodes/ path is exactly the bug class this pass exists to make
# unrepresentable — a persister reaching around node_mover.go's resolvers to build its own
# path, which is precisely how positionFilePath/localPolarsFilePath/cascadeEdgesFilePath/
# nodePortFilePath used to be scattered before step 3.
NODE_PATH_OWNERS=("node_mover.go" "edge_mover.go" "loader_tree.go")
is_node_path_owner() {
  local f="$1"
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
  echo "check-scene-path-resolution: $NODE_JOIN_HITS hand-rolled nodes/ filepath.Join(...) hit(s) outside node_mover.go/edge_mover.go/loader_tree.go — a node/port path belongs to its owning mover; call node_mover.go's resolvers instead of reconstructing the path."
  exit 1
fi

echo "check-scene-path-resolution: clean ($GO_FILE_COUNT files scanned; $CALL_SITES resolver call site(s); all IsDir/Join path-resolution lives in scene_paths.go; node-path construction lives in node_mover.go/edge_mover.go)"
exit 0
