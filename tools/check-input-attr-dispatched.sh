#!/usr/bin/env bash
set -euo pipefail

# check-input-attr-dispatched.sh — an attribute Go can DECODE must be DISPATCHED.
#
# PLACEMENT: nodes/Wiring/edit_update_decode.go,nodes/Wiring/stdin_dispatch.go,nodes/Wiring/stdin_apply.go | a new addressed-edit attribute must reach a handler, not just decode off the wire
#
# THE BUG THIS EXISTS FOR. A new addressed-edit ATTRIBUTE (the only way new addressed
# capability lands — .claude/rules/bridge-surface.md: new entity kind or attribute, never a
# new op) is spread across four files: the TS const + encoder (src/schema/input-attrs.ts,
# src/schema/input-encode.ts),
# Go's decode (edit_update_decode.go), the dispatch TABLE (stdin_dispatch.go's updateKindHandlers /
# the per-kind *AttrHandlers tables), and the HANDLER that says what it means
# (stdin_apply.go). Miss the last two and the attribute round-trips the wire PERFECTLY —
# encodes, frames, decodes into a stdinMsg with the right Kind/Attr — and is then dropped on
# the floor, because every dispatch level is deliberately forward-compat ("unknown
# kinds/attrs are ignored"). The click does nothing and nothing anywhere says why.
#
# WHAT THE EXISTING GUARDS ALREADY COVER (this one does not repeat them):
#   - check-edit-op-parity.sh: the OP set, the ENTITY-KIND set 3 ways (messages.ts vs
#     updateKindHandlers vs the generated IN_UPDATE_KINDS), and overlay-flag cardinality. It
#     is kind-level; it never looks at an attr.
#   - check-message-kind-parity.sh: top-level msg.Type kinds and their senders. Above this
#     seam entirely.
#   - INPUT_LAYOUT_FINGERPRINT (input_codec.go vs input-layout-gen.ts): pins the attr NAME LIST
#     and its index order across the two languages, so an encoder and a decoder cannot
#     disagree about byte 5. It says nothing about whether anything RUNS afterwards.
#   The gap this guard closes: decoded-but-never-dispatched.
#
# WHAT IT ASSERTS. For every (entity kind, attr) pair Go's decoder actually PRODUCES —
# extracted from the `Kind: "<k>", Attr: "<a>"` stdinMsg literals in the codec — both:
#   1. <k> has an entry in the top-level updateKindHandlers dispatch table, and
#   2. the attr literal "<a>" appears in that entry's handler function body, or in a
#      per-kind `*AttrHandlers` table that body routes through (the two live dispatch
#      shapes: applyUpdateClock/applyUpdateOverlays defer to an attr table;
#      applyUpdateScene/applyUpdateTiltVector/applyUpdateDistanceGroup switch or compare
#      on msg.Attr inline).
#
# WHAT IT DELIBERATELY DOES NOT ASSERT. That the handler does the RIGHT thing with the
# attribute, or anything at all — a `case "x":` with an empty body passes. It is a
# reachability check on the routing surface, not a semantics check. It also does not check
# the reverse direction (a handler for an attr nothing decodes); that is dead code, not a
# silently-ignored user action, and it is cheap to spot.
#
# Files are LOCATED BY SCANNING (memory/feedback_guards_hardcoding_single_file_break_on_split.md
# — the handlers moved out of stdin_reader.go into stdin_apply.go on exactly this branch, and
# a path-hardcoded guard is what that split breaks). Finding nothing to scan reports
# MISCONFIGURED and exits 1 rather than passing vacuously.
#
# Exit 0 clean, exit 1 with a named report.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# `|| true` on every extractor assignment: under `set -e` a non-matching grep inside a
# VAR=$(...) assignment kills the script with exit 1 and NO output, which would make the
# MISCONFIGURED branches below unreachable (a sibling guard hit exactly this).

# --- Locate the dispatch TABLE, and with it the package dir -------------------
TABLE_FILE=$(grep -rl 'updateKindHandlers = map' --include='*.go' . \
  --exclude-dir=node_modules --exclude-dir=out --exclude-dir=.git 2>/dev/null | head -1 || true)
if [ -z "$TABLE_FILE" ]; then
  echo "check-input-attr-dispatched: MISCONFIGURED — no file declares 'updateKindHandlers = map'."
  echo "The edit-update dispatch table was renamed or removed; this guard is scanning nothing"
  echo "rather than passing. Repoint its pattern."
  exit 1
fi
PKG_DIR="$(dirname "$TABLE_FILE")"

# --- Locate the DECODER: whatever file builds stdinMsg{... Kind: "x", Attr: "y" } ---
CODEC_FILES=$(grep -rlE 'Kind: "[a-zA-Z]+", Attr: "[a-zA-Z]+"' --include='*.go' "$PKG_DIR" 2>/dev/null || true)
if [ -z "$CODEC_FILES" ]; then
  echo "check-input-attr-dispatched: MISCONFIGURED — nothing in $PKG_DIR produces a"
  echo "'Kind: \"…\", Attr: \"…\"' decoded message literal. The decoder's shape changed;"
  echo "repoint this guard rather than trusting its silence."
  exit 1
fi

PAIRS=$(grep -hoE 'Kind: "[a-zA-Z]+", Attr: "[a-zA-Z]+"' $CODEC_FILES \
  | sed -E 's/Kind: "([a-zA-Z]+)", Attr: "([a-zA-Z]+)"/\1 \2/' | sort -u || true)
if [ -z "$PAIRS" ]; then
  echo "check-input-attr-dispatched: MISCONFIGURED — found $CODEC_FILES but extracted no"
  echo "(kind, attr) pairs from it."
  exit 1
fi

# Extract the top-level dispatch table's "kind": handlerFunc entries.
TABLE_BODY=$(awk '/updateKindHandlers = map/{p=1} p{print} p&&/^\}/{exit}' "$TABLE_FILE" || true)
if [ -z "$TABLE_BODY" ]; then
  echo "check-input-attr-dispatched: MISCONFIGURED — could not read the updateKindHandlers"
  echo "table body out of $TABLE_FILE."
  exit 1
fi

# Print the body of a top-level Go func by name, from any .go in the package dir.
func_body() { # name
  awk -v fn="$1" '
    $0 ~ "^func " fn "\\(" { p=1 }
    p { print }
    p && /^\}/ { exit }
  ' $PKG_GO_FILES
}

# Print a top-level `var <name> = map[...]{ … }` block by name.
var_block() { # name
  awk -v vn="$1" '
    $0 ~ "^var " vn " = " { p=1 }
    p { print }
    p && /^\}/ { exit }
  ' $PKG_GO_FILES
}

PKG_GO_FILES=$(ls "$PKG_DIR"/*.go 2>/dev/null || true)
if [ -z "$PKG_GO_FILES" ]; then
  echo "check-input-attr-dispatched: MISCONFIGURED — no .go files in $PKG_DIR."
  exit 1
fi

HITS=0
while read -r kind attr; do
  [ -z "$kind" ] && continue

  handler=$(printf '%s\n' "$TABLE_BODY" | grep -E "^[[:space:]]*\"$kind\":" \
    | sed -E 's/.*:[[:space:]]*([A-Za-z0-9_]+).*/\1/' | head -1 || true)
  if [ -z "$handler" ]; then
    echo "check-input-attr-dispatched: entity kind \"$kind\" is DECODED (attr \"$attr\") but has"
    echo "  no entry in updateKindHandlers ($TABLE_FILE) — every edit for it decodes cleanly and"
    echo "  is then dropped by applyUpdate's forward-compat fallthrough."
    HITS=$((HITS + 1))
    continue
  fi

  BODY=$(func_body "$handler" || true)
  if [ -z "$BODY" ]; then
    echo "check-input-attr-dispatched: MISCONFIGURED — updateKindHandlers routes \"$kind\" to"
    echo "  $handler, but no 'func $handler(' exists in $PKG_DIR. Handler moved out of the"
    echo "  package; repoint this guard."
    exit 1
  fi

  # Follow one level of indirection: a per-kind attr TABLE the handler routes through.
  for tbl in $(printf '%s\n' "$BODY" | grep -oE '[A-Za-z0-9_]+AttrHandlers' | sort -u || true); do
    BODY="$BODY
$(var_block "$tbl" || true)"
  done

  if ! printf '%s\n' "$BODY" | grep -qF "\"$attr\""; then
    echo "check-input-attr-dispatched: attr \"$attr\" on entity \"$kind\" is DECODED but never"
    echo "  DISPATCHED: $handler (and any *AttrHandlers table it routes through) never mentions"
    echo "  \"$attr\". The edit crosses the wire intact and is then silently ignored — the click"
    echo "  does nothing and nothing reports why. Add the case/table entry and its handler."
    HITS=$((HITS + 1))
  fi
done <<< "$PAIRS"

if [ "$HITS" -ne 0 ]; then
  echo ""
  echo "check-input-attr-dispatched: $HITS undispatched attribute(s) found"
  exit 1
fi

exit 0
