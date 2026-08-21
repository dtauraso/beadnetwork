#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: src/Input/dispatch/dispatch_edit.go,src/Input/messages.ts,src/Input/input-layout-gen.ts,src/Overlay/paths/,src/Chrome/Panels/Panel/paths/ | edit ops/update-kinds/overlay flags must stay listed identically on both sides of the bridge

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

GO_PKG_DIR="$REPO_ROOT/src/Input"
go_fence_files() {
  grep -rl --include='*.go' -E "^[[:space:]]*//[[:space:]]*$1[[:space:]]*$" "$GO_PKG_DIR" \
    | grep -v '_test\.go$' || true
}
MESSAGES_TS="$REPO_ROOT/src/Input/messages.ts"

HANDLE_MSG="$REPO_ROOT/src/Input/input-layout-gen.ts"


PANEL_STATE_GO="$REPO_ROOT/src/Chrome/Panels/Panel/panel_state.go"

for f in "$MESSAGES_TS" "$HANDLE_MSG" "$PANEL_STATE_GO"; do
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

between() {
  local s="$1" e="$2"; shift 2
  awk -v s="$s" -v e="$e" '
    FNR==1 { p=0 }
    $0 ~ "^[ \t]*(//|#)[ \t]*" s "[ \t]*$" { p=1; next }
    $0 ~ "^[ \t]*(//|#)[ \t]*" e "[ \t]*$" { p=0 }
    p
  ' "$@"
}

quoted() { grep -aoE '"[^"]+"' | tr -d '"' | sort -u; }

toplevel_case() { awk '/^\tcase "/ || /^\t"[^"]+":/'; }

assert_nonempty() {
  if [[ -z "$(printf '%s' "$1" | tr -d '[:space:]')" ]]; then
    echo "edit-op-parity: EMPTY extracted set for '$2' — sentinel block missing/renamed; refusing vacuous parity pass" >&2
    exit 1
  fi
}

HITS=0
report_diff() {
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

TS_OPS=$(between EDIT_MSG_START EDIT_MSG_END "$MESSAGES_TS" | grep -aoE 'op: "[^"]+"' | quoted) || true
GO_OPS=$(between EDIT_OPS_START EDIT_OPS_END $GO_OPS_FILES | toplevel_case | quoted) || true
assert_nonempty "$TS_OPS" "axis1 messages.ts ops"
assert_nonempty "$GO_OPS" "axis1 src/Input ops"
report_diff "$(comm -13 <(echo "$GO_OPS") <(echo "$TS_OPS"))" "src/Input ops" \
            "$(comm -23 <(echo "$GO_OPS") <(echo "$TS_OPS"))" "messages.ts ops"

TS_KINDS=$(between EDIT_MSG_START EDIT_MSG_END "$MESSAGES_TS" | grep -aoE 'kind: "[^"]+"' | quoted) || true
GO_KINDS=$(between EDIT_UPDATE_KINDS_START EDIT_UPDATE_KINDS_END $GO_KINDS_FILES | toplevel_case | quoted) || true
HM_KINDS=$(between EDIT_UPDATE_KINDS_START EDIT_UPDATE_KINDS_END "$HANDLE_MSG" | quoted) || true
assert_nonempty "$TS_KINDS" "axis2 messages.ts update kinds"
assert_nonempty "$GO_KINDS" "axis2 src/Input update kinds"
assert_nonempty "$HM_KINDS" "axis2 handle-message.ts update kinds"
report_diff "$(comm -13 <(echo "$GO_KINDS") <(echo "$TS_KINDS"))" "src/Input kinds" \
            "$(comm -23 <(echo "$GO_KINDS") <(echo "$TS_KINDS"))" "messages.ts kinds"
report_diff "$(comm -13 <(echo "$HM_KINDS") <(echo "$TS_KINDS"))" "handle-message.ts kinds" \
            "$(comm -23 <(echo "$HM_KINDS") <(echo "$TS_KINDS"))" "messages.ts kinds"

TS_FLAGS=$(between OVERLAY_FLAGS_START OVERLAY_FLAGS_END "$MESSAGES_TS" | quoted) || true
assert_nonempty "$TS_FLAGS" "axis3 messages.ts overlay flags"

RENDER_READS=$(awk '/^var FlagNames = \[\]string\{/{p=1;next} p&&/^\}/{p=0} p' \
  "$REPO_ROOT/src/Overlay/flag_paths_gen.go" | grep -aoE '"[a-zA-Z]+"' | tr -d '"' | sort -u)
assert_nonempty "$RENDER_READS" "axis3 Overlay FlagNames block layout"
N_FLAGS=$(printf '%s\n' "$TS_FLAGS" | grep -c .)
N_READS=$(printf '%s\n' "$RENDER_READS" | grep -c .)
if [[ "$N_FLAGS" -ne "$N_READS" ]]; then
  echo "  overlay flag/renderer cardinality mismatch: OVERLAY_FLAG_NAMES=$N_FLAGS, Overlay FlagNames=$N_READS"
  echo "    (a flag was added/removed in messages.ts without regenerating flag_paths_gen.go, so it has no slot in the block the renderer reads)"
  HITS=$((HITS+1))
fi

TS_PANEL_FLAGS=$(between PANEL_FLAGS_START PANEL_FLAGS_END "$MESSAGES_TS" | quoted) || true
assert_nonempty "$TS_PANEL_FLAGS" "axis4 messages.ts panel flags"

PANEL_RENDER_READS=$(awk '/PanelOpen = map\[string\]/{p=1;next} p&&/^}/{p=0} p' "$PANEL_STATE_GO" \
  | grep -aoE '^[[:space:]]*"[a-zA-Z]+":' | sort -u)

PANEL_RENDER_KEYS=$(awk '/PanelToggles = map\[string\]/{p=1;next} p&&/^}/{p=0} p' "$PANEL_STATE_GO" \
  | grep -aoE '^[[:space:]]*"[a-zA-Z]+":' | sort -u)
assert_nonempty "$PANEL_RENDER_READS" "axis4 panel_state.go PanelOpen keys"
assert_nonempty "$PANEL_RENDER_KEYS" "axis4 panel_state.go PanelToggles keys"
N_PANEL_FLAGS=$(printf '%s\n' "$TS_PANEL_FLAGS" | grep -c .)
N_PANEL_READS=$(printf '%s\n' "$PANEL_RENDER_READS" | grep -c .)
N_PANEL_KEYS=$(printf '%s\n' "$PANEL_RENDER_KEYS" | grep -c .)
if [[ "$N_PANEL_FLAGS" -ne "$N_PANEL_READS" || "$N_PANEL_FLAGS" -ne "$N_PANEL_KEYS" ]]; then
  echo "  panel flag/renderer cardinality mismatch: PANEL_FLAG_NAMES=$N_PANEL_FLAGS, PanelOpen keys=$N_PANEL_READS, PanelToggles keys=$N_PANEL_KEYS"
  echo "    (a flag was added/removed in messages.ts but not wired into panel_state.go's tables, or vice versa)"
  HITS=$((HITS+1))
fi

PANEL_PATH_FILES=$(awk '/^var FlagNames = \[\]string\{/{p=1;next} p&&/^\}/{p=0} p' \
  "$REPO_ROOT/src/Chrome/Panels/Panel/flag_paths_gen.go" | grep -aoE '"[a-zA-Z]+"' | tr -d '"' | sort -u)
assert_nonempty "$PANEL_PATH_FILES" "axis4 Panel FlagNames block layout"
N_PANEL_PATHS=$(printf '%s\n' "$PANEL_PATH_FILES" | grep -c .)
if [[ "$N_PANEL_FLAGS" -ne "$N_PANEL_PATHS" ]]; then
  echo "  panel flag/renderer cardinality mismatch: PANEL_FLAG_NAMES=$N_PANEL_FLAGS, Panel FlagNames=$N_PANEL_PATHS"
  echo "    (a flag was added/removed in messages.ts without regenerating flag_paths_gen.go, so it has no slot in the block the renderer reads)"
  HITS=$((HITS+1))
fi

if [[ $HITS -eq 0 ]]; then
  echo "edit-op-parity: clean (ops + update kinds + overlay flags + panel flags in parity)"
  exit 0
fi
echo ""
echo "edit-op-parity: $HITS divergence(s) found"
exit 1
