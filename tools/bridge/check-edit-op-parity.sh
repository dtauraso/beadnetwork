#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: nodes/Wiring/stdin_dispatch.go,tools/topology-vscode/src/messages.ts,tools/topology-vscode/src/schema/input-layout-gen.ts,tools/topology-vscode/src/webview/three/controls/flags/overlay-flags.ts | edit ops/update-kinds/overlay flags must stay listed identically on both sides of the bridge
#
# Verifies the editor->Go geometry-CRUD "edit" bridge stays in parity across every
# axis below the top-level msg.Type (which check-message-kind-parity.sh covers).
# The bridge's sole op is "update" (create/delete were removed end-to-end — no live
# TS sender ever emitted them, and their only trigger tore down a live wire's in-flight
# beads via PacedWire.Restore()); op="update" sets an attribute on a typed ENTITY (kind:
# node/edge/camera/overlays/scene). Overlay visibility is one named-boolean FLAG
# attribute per overlay. A value added on one side and forgotten on another silently
# no-ops at runtime (CLAUDE.md "Bridge surface"). Three axes are checked:
#
#   1. ops          — messages.ts EditMsg  vs  nodes/Wiring's applyEdit op table
#                     (both now reduce to the single set {update}).
#   2. update kinds — messages.ts EditMsg  vs  nodes/Wiring's applyUpdate kind table
#                     vs  handle-message.ts update-dispatch switch (3-way).
#   3. overlay flags— messages.ts OVERLAY_FLAG_NAMES  vs  the HAND-AUTHORED overlay-flags.ts
#                     renderer (readOverlay* reads + OverlayFlagVals keys), by cardinality.
#
# (Axis 4 — stdinGuideVisPayload fields vs OverlayState/flags — was removed when the
# attr="set" full-visibility-install path was dropped: its only TS caller, the load-time
# main.tsx push, was deleted and the generated stdinGuideVisPayload struct with it.)
#
# Sentinel comments (X_START / X_END) bound each region so the greps cannot sweep in
# unrelated literals (viewpoint sub-kinds, attr labels, trace kinds).
# Exit 0 if clean; exit 1 with a report otherwise.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# The Go side is located by SCANNING the nodes/Wiring package for the sentinel fences, not
# by naming a file (memory/feedback_guards_hardcoding_single_file_break_on_split.md): the
# EDIT_OPS / EDIT_UPDATE_KINDS tables moved out of stdin_reader.go into stdin_dispatch.go
# when that file was split by job, and a path-hardcoded guard is exactly what such a split
# blinds. Tests are excluded so a fixture can never supply a fence. Finding no file with the
# fence is MISCONFIGURED below, not a pass.
GO_PKG_DIR="$REPO_ROOT/nodes/Wiring"
go_fence_files() { # start-marker
  grep -rl --include='*.go' -E "^[[:space:]]*//[[:space:]]*$1[[:space:]]*$" "$GO_PKG_DIR" \
    | grep -v '_test\.go$' || true
}
MESSAGES_TS="$REPO_ROOT/tools/topology-vscode/src/messages.ts"
# The 3rd update-kind parity source moved from handle-message.ts's dispatch switch (removed
# when the TS→Go bridge became a binary buffer) to the shared IN_UPDATE_KINDS schema, which
# is the single TS list of edit-update entity kinds the encoders key off. IN_UPDATE_KINDS
# itself is GENERATED (from Go's InputLayoutFingerprint) into input-layout-gen.ts, so that is
# the file carrying the sentinel-bound literal now, not the hand-authored input-encode.ts.
HANDLE_MSG="$REPO_ROOT/tools/topology-vscode/src/schema/input-layout-gen.ts"
# overlay-flags.ts is the HAND-AUTHORED overlay renderer: it reflects each Go-owned overlay
# column out of the binary content buffer. Its per-flag bit reads (readOverlay*) + its
# OverlayFlagVals object literal are the TS-side consumer that must stay in sync with the
# overlay flag list (axis 3). (The old JSON-trace consumer pump.ts was removed in the
# content-buffer erase; overlay state now round-trips through the buffer.)
OVERLAY_FLAGS_TS="$REPO_ROOT/tools/topology-vscode/src/webview/three/controls/flags/overlay-flags.ts"

for f in "$MESSAGES_TS" "$HANDLE_MSG" "$OVERLAY_FLAGS_TS"; do
  if [[ ! -f "$f" ]]; then
    echo "edit-op-parity: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done
if [[ ! -d "$GO_PKG_DIR" ]]; then
  echo "edit-op-parity: MISCONFIGURED — dir not found: $GO_PKG_DIR" >&2
  exit 1
fi

GO_OPS_FILES=$(go_fence_files EDIT_OPS_START)
GO_KINDS_FILES=$(go_fence_files EDIT_UPDATE_KINDS_START)
for pair in "GO_OPS_FILES:EDIT_OPS_START" "GO_KINDS_FILES:EDIT_UPDATE_KINDS_START"; do
  var="${pair%%:*}"; marker="${pair#*:}"
  if [[ -z "${!var}" ]]; then
    echo "edit-op-parity: MISCONFIGURED — no non-test .go file under $GO_PKG_DIR carries the" >&2
    echo "  $marker sentinel. The fenced table was renamed, deleted, or moved out of the" >&2
    echo "  package; repoint this guard rather than letting it scan nothing." >&2
    exit 1
  fi
done

# Extract the lines of FILE strictly between the sentinel comment lines START and END.
#
# Markers are matched ANCHORED: a comment line containing the marker and NOTHING else.
# The previous `index($0,s)` was an unanchored substring match, and that is a trap — the
# moment any prose in the scanned file names the sentinel (e.g. a header saying "the op
# switch is fenced by EDIT_OPS_START/END", which is exactly the style stdin_dispatch.go and
# CLAUDE.md already use), the fence opens on that prose line and the extracted set becomes
# silently WRONG. It was armed but unexploded here.
#
# assert_nonempty does NOT protect against this: an unanchored match yields a non-empty,
# wrong set rather than an empty one. Vacuous-pass refusal is orthogonal to fence
# correctness. Same fix as check-message-kind-parity.sh, which cites this file as its model.
#
# Takes ONE OR MORE files (the Go side is now a package scan, so a fence may live in any of
# several files); each file's fence state is reset at its first line so an unterminated
# fence in one file cannot leak into the next.
between() { # start end file...
  local s="$1" e="$2"; shift 2
  awk -v s="$s" -v e="$e" '
    FNR==1 { p=0 }
    $0 ~ "^[ \t]*(//|#)[ \t]*" s "[ \t]*$" { p=1; next }
    $0 ~ "^[ \t]*(//|#)[ \t]*" e "[ \t]*$" { p=0 }
    p
  ' "$@"
}

# Double-quoted literal values from a stream.
quoted() { grep -aoE '"[^"]+"' | tr -d '"' | sort -u; }

# Top-level Go `case "..."` labels OR top-level dispatch-table string keys (`"foo": ...,`):
# exactly one leading tab (nested cases/keys have two or more, so they are excluded).
# BSD grep lacks -P, so match with awk (\t = tab). The two alternatives let this extractor
# survive either a switch or a map[string]func(...) dispatch table at the fenced level —
# applyEdit/applyUpdate (stdin_dispatch.go) moved from switch to table form; this keeps the
# same fences discoverable either way.
toplevel_case() { awk '/^\tcase "/ || /^\t"[^"]+":/'; }

# Refuse a vacuous pass: if a sentinel-bounded extractor returns an EMPTY set, a
# sentinel pair was deleted/renamed on that side and comm would compare empty-to-empty
# and "pass". Every extracted axis set must be non-empty. (Positive-assertion pattern,
# per check-ts-shading-from-go.sh / check-no-await-on-bridge.sh.)
assert_nonempty() { # value label
  if [[ -z "$(printf '%s' "$1" | tr -d '[:space:]')" ]]; then
    echo "edit-op-parity: EMPTY extracted set for '$2' — sentinel block missing/renamed; refusing vacuous parity pass" >&2
    exit 1
  fi
}

HITS=0
report_diff() { # label missing_in_a a_name missing_in_b b_name
  local missing_a="$1" a_name="$2" missing_b="$3" b_name="$4"
  if [[ -n "$missing_a" ]]; then
    while IFS= read -r v; do [[ -z "$v" ]] && continue
      echo "  $v: present in $b_name but missing in $a_name"; HITS=$((HITS+1)); done <<< "$missing_a"
  fi
  if [[ -n "$missing_b" ]]; then
    while IFS= read -r v; do [[ -z "$v" ]] && continue
      echo "  $v: present in $a_name but missing in $b_name"; HITS=$((HITS+1)); done <<< "$missing_b"
  fi
}

# --- Axis 1: ops ------------------------------------------------------------
# NOTE `|| true` on every extractor assignment below. Without it, `set -euo pipefail` kills
# the script AT THE ASSIGNMENT whenever an extractor's grep legitimately matches nothing —
# so the assert_nonempty diagnostic underneath, which exists precisely to explain that case,
# could never print. The script still exited nonzero, so it failed SAFE but SILENTLY,
# defeating the message. Verified with a minimal repro.
TS_OPS=$(between EDIT_MSG_START EDIT_MSG_END "$MESSAGES_TS" | grep -aoE 'op: "[^"]+"' | quoted) || true
GO_OPS=$(between EDIT_OPS_START EDIT_OPS_END $GO_OPS_FILES | toplevel_case | quoted) || true
assert_nonempty "$TS_OPS" "axis1 messages.ts ops"
assert_nonempty "$GO_OPS" "axis1 nodes/Wiring ops"
report_diff "$(comm -13 <(echo "$GO_OPS") <(echo "$TS_OPS"))" "nodes/Wiring ops" \
            "$(comm -23 <(echo "$GO_OPS") <(echo "$TS_OPS"))" "messages.ts ops"

# --- Axis 2: update entity kinds (3-way) ------------------------------------
TS_KINDS=$(between EDIT_MSG_START EDIT_MSG_END "$MESSAGES_TS" | grep -aoE 'kind: "[^"]+"' | quoted) || true
GO_KINDS=$(between EDIT_UPDATE_KINDS_START EDIT_UPDATE_KINDS_END $GO_KINDS_FILES | toplevel_case | quoted) || true
HM_KINDS=$(between EDIT_UPDATE_KINDS_START EDIT_UPDATE_KINDS_END "$HANDLE_MSG" | quoted) || true
assert_nonempty "$TS_KINDS" "axis2 messages.ts update kinds"
assert_nonempty "$GO_KINDS" "axis2 nodes/Wiring update kinds"
assert_nonempty "$HM_KINDS" "axis2 handle-message.ts update kinds"
report_diff "$(comm -13 <(echo "$GO_KINDS") <(echo "$TS_KINDS"))" "nodes/Wiring kinds" \
            "$(comm -23 <(echo "$GO_KINDS") <(echo "$TS_KINDS"))" "messages.ts kinds"
report_diff "$(comm -13 <(echo "$HM_KINDS") <(echo "$TS_KINDS"))" "handle-message.ts kinds" \
            "$(comm -23 <(echo "$HM_KINDS") <(echo "$TS_KINDS"))" "messages.ts kinds"

# --- Axis 3: overlay flags → hand-authored renderer -------------------------
# Repointed (was messages.ts OVERLAY_FLAG_NAMES vs the GENERATED OverlayToggles map in
# nodes/Wiring/viewstate/overlay_state.go — circular, since the latter is generated from
# the former; flag→Go
# parity is already covered by check-generated.sh regenerate+diff and the overlay
# behavior test). The value axis 3 adds is flag→RENDERER parity: a flag added to
# OVERLAY_FLAG_NAMES but never wired into the hand-authored overlay-flags.ts renderer
# (per-flag readOverlay* bit reads + OverlayFlagVals object literal) would silently never
# reflect out of the buffer. Nothing else forces those two lists to track the flag list —
# that is this axis.
#
# CARDINALITY, not normalized-name, correspondence: the flag→buffer-column mapping is
# non-mechanical (tori→readOverlaySceneTori, overlays→readOverlayOverlaysVis) so a
# camelCase↔read-name compare would false-diverge. Counts are robust and catch the dominant
# failure (flag added/removed on one side only). The three independent hand-authored lists
# (flags, readOverlay* reads, OverlayFlagVals object keys) must have equal cardinality.
TS_FLAGS=$(between OVERLAY_FLAGS_START OVERLAY_FLAGS_END "$MESSAGES_TS" | quoted) || true
assert_nonempty "$TS_FLAGS" "axis3 messages.ts overlay flags"
# Per-flag buffer reads: the distinct readOverlay* function names used in overlay-flags.ts.
RENDER_READS=$(grep -aoE 'readOverlay[A-Za-z]+\(v\)' "$OVERLAY_FLAGS_TS" | sort -u)
# OverlayFlagVals object keys: property lines inside the `… : OverlayFlagVals = { … };`
# literal. Keyed off the OverlayFlagVals TYPE annotation (stable), not the assigned var name
# (which changed from `cachedVals` to a `next` local when the bit-packing cache became a
# value-equality cache) — so a variable rename can't blind this extraction.
RENDER_KEYS=$(awk '/OverlayFlagVals = \{/{p=1;next} p&&/^[[:space:]]*};/{p=0} p&&/^[[:space:]]*[a-zA-Z_]+:/{print}' "$OVERLAY_FLAGS_TS" | grep -aoE '^[[:space:]]*[a-zA-Z_]+:' | sort -u)
assert_nonempty "$RENDER_READS" "axis3 overlay-flags.ts readOverlay* reads"
assert_nonempty "$RENDER_KEYS" "axis3 overlay-flags.ts OverlayFlagVals keys"
N_FLAGS=$(printf '%s\n' "$TS_FLAGS" | grep -c .)
N_READS=$(printf '%s\n' "$RENDER_READS" | grep -c .)
N_KEYS=$(printf '%s\n' "$RENDER_KEYS" | grep -c .)
if [[ "$N_FLAGS" -ne "$N_READS" || "$N_FLAGS" -ne "$N_KEYS" ]]; then
  echo "  overlay flag/renderer cardinality mismatch: OVERLAY_FLAG_NAMES=$N_FLAGS, overlay-flags reads=$N_READS, OverlayFlagVals keys=$N_KEYS"
  echo "    (a flag was added/removed in messages.ts but not wired into overlay-flags.ts's renderer, or vice versa)"
  HITS=$((HITS+1))
fi

# (Camera viewpoint sub-kinds axis removed: camera edits are produced in-process by the
# gesture FSM from raw-input and no longer cross the editor→Go seam, so there is no vp.Kind
# TS↔Go vocabulary left to keep in parity.)
# (Axis 4 — stdinGuideVisPayload fields — removed: the attr="set" path was dropped; see
# header comment above.)

if [[ $HITS -eq 0 ]]; then
  echo "edit-op-parity: clean (ops + update kinds + overlay flags in parity)"
  exit 0
fi
echo ""
echo "edit-op-parity: $HITS divergence(s) found"
exit 1
