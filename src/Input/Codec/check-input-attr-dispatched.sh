#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: src/Input/Codec/edit_update_decode.go,src/Input/Codec/edit_update_decode_scene.go,src/Input/Dispatch/dispatch_edit.go | a new addressed-edit attribute must reach a handler, not just decode off the wire

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

TABLE_FILE=$(grep -rl 'updateKindHandlers = map' --include='*.go' . \
  --exclude-dir=node_modules --exclude-dir=out --exclude-dir=.git 2>/dev/null | head -1 || true)
if [ -z "$TABLE_FILE" ]; then
  echo "check-input-attr-dispatched: MISCONFIGURED — no file declares 'updateKindHandlers = map'."
  echo "The edit-update dispatch table was renamed or removed; this guard is scanning nothing"
  echo "rather than passing. Repoint its pattern."
  exit 1
fi
PKG_DIR="$(dirname "$TABLE_FILE")"

CODEC_FILES=$(grep -rlE 'Kind: "[a-zA-Z]+", Attr: "[a-zA-Z]+"' --include='*.go' . \
  --exclude-dir=node_modules --exclude-dir=out --exclude-dir=.git 2>/dev/null || true)
if [ -z "$CODEC_FILES" ]; then
  echo "check-input-attr-dispatched: MISCONFIGURED — nothing in the repo produces a"
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

TABLE_BODY=$(awk '/updateKindHandlers = map/{p=1} p{print} p&&/^\}/{exit}' "$TABLE_FILE" || true)
if [ -z "$TABLE_BODY" ]; then
  echo "check-input-attr-dispatched: MISCONFIGURED — could not read the updateKindHandlers"
  echo "table body out of $TABLE_FILE."
  exit 1
fi

func_body() {
  awk -v fn="$1" '
    $0 ~ "^func " fn "\\(" { p=1 }
    p { print }
    p && /^\}/ { exit }
  ' $PKG_GO_FILES
}

var_block() {
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

  for tbl in $(printf '%s\n' "$BODY" | grep -oE '[A-Za-z0-9_]+AttrHandlers' | sort -u || true); do
    BODY="$BODY
$(var_block "$tbl" || true)"
  done

  if ! grep -qF "\"$attr\"" <<< "$BODY"; then
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
