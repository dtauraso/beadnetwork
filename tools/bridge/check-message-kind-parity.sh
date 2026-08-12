#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: nodes/Wiring/stdinreader/stdin_reader.go,tools/topology-vscode/src/messages.ts,tools/topology-vscode/src/extension/handle-message.ts,tools/topology-vscode/src/webview/** | a new message kind needs matching entries in Go's MSG_TYPES fence+doc, messages.ts, and a live sender





#      MSG_TYPES_DOC block, and vice versa — so the header cannot undercount its switch.













SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"


# (memory/feedback_guards_hardcoding_single_file_break_on_split.md): the MSG_TYPES fence and
# its MSG_TYPES_DOC header sit in stdin_reader.go today, but that file has already been split



GO_PKG_DIR="$REPO_ROOT/nodes/Wiring"
MESSAGES_TS="$REPO_ROOT/tools/topology-vscode/src/messages.ts"
HANDLE_MESSAGE_TS="$REPO_ROOT/tools/topology-vscode/src/extension/handle-message.ts"
WEBVIEW_SRC_DIR="$REPO_ROOT/tools/topology-vscode/src/webview"

if [[ ! -d "$GO_PKG_DIR" ]]; then
  echo "message-kind-parity: MISCONFIGURED — dir not found: $GO_PKG_DIR" >&2
  exit 1
fi

GO_FILES=$(find "$GO_PKG_DIR" -name '*.go' ! -name '*_test.go' | sort)
if [[ -z "$GO_FILES" ]]; then
  echo "message-kind-parity: MISCONFIGURED — no non-test .go files under $GO_PKG_DIR" >&2
  exit 1
fi

for f in "$MESSAGES_TS" "$HANDLE_MESSAGE_TS"; do
  if [[ ! -f "$f" ]]; then
    echo "message-kind-parity: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done
if [[ ! -d "$WEBVIEW_SRC_DIR" ]]; then
  echo "message-kind-parity: MISCONFIGURED — dir not found: $WEBVIEW_SRC_DIR" >&2
  exit 1
fi




#   case "...":  inside the MSG_TYPES_START/END fence
kinds_from_go() {
  {
    grep -haoE 'msg\.Type[[:space:]]*[!=]=[[:space:]]*"[^"]+"' $GO_FILES \
      | grep -oE '"[^"]+"' \
      | tr -d '"'


    # for top-level types. Same pattern as EDIT_OPS_START/END in applyEdit.







    awk '
      FNR==1 { inblk=0 }
      /^[[:space:]]*\/\/[[:space:]]*MSG_TYPES_START[[:space:]]*$/ { inblk=1; next }
      /^[[:space:]]*\/\/[[:space:]]*MSG_TYPES_END[[:space:]]*$/   { inblk=0 }
      inblk
    ' $GO_FILES \
      | grep -aoE 'case[[:space:]]+"[^"]+"' \
      | grep -oE '"[^"]+"' \
      | tr -d '"'
  } | sort -u
}

# Extract the types DECLARED in the package's MSG_TYPES_DOC block (stdin_reader.go's header). Only numbered


kinds_from_go_doc() {
  awk '
    FNR==1 { inblk=0 }
    /^[[:space:]]*\/\/[[:space:]]*MSG_TYPES_DOC_START[[:space:]]*$/ { inblk=1; next }
    /^[[:space:]]*\/\/[[:space:]]*MSG_TYPES_DOC_END[[:space:]]*$/   { inblk=0 }
    inblk && /^\/\/[[:space:]]+[0-9]+\.[[:space:]]+"/
  ' $GO_FILES \
    | grep -oE '"[^"]+"' \
    | tr -d '"' \
    | sort -u
}




kinds_from_ts() {
  awk '/WEBVIEW_TO_HOST_TYPES/,/\]\)/' "$MESSAGES_TS" \
    | grep -avE 'flatMap|WEBVIEW_TO_HOST_TYPES|\]\)' \
    | grep -o '"[^"]*"' \
    | tr -d '"' \
    | sort -u
}






GO_KINDS=$(kinds_from_go) || true
GO_DOC_KINDS=$(kinds_from_go_doc) || true
TS_KINDS=$(kinds_from_ts) || true




for pair in "GO_KINDS:stdin_reader.go MSG_TYPES fenced switch" \
            "GO_DOC_KINDS:stdin_reader.go MSG_TYPES_DOC header list" \
            "TS_KINDS:WEBVIEW_TO_HOST_TYPES kinds"; do
  var="${pair%%:*}"; label="${pair#*:}"
  if [[ -z "$(printf '%s' "${!var}" | tr -d '[:space:]')" ]]; then
    echo "message-kind-parity: EMPTY extracted set for '$label' — switch/const missing or renamed; refusing vacuous parity pass" >&2
    exit 1
  fi
done

MISSING=$(comm -23 <(echo "$GO_KINDS") <(echo "$TS_KINDS"))

HITS=0
if [[ -n "$MISSING" ]]; then
  echo "message-kind-parity: kinds in stdin_reader.go but missing from WEBVIEW_TO_HOST_TYPES:"
  while IFS= read -r k; do
    echo "  missing: \"$k\""
    HITS=$((HITS + 1))
  done <<< "$MISSING"
fi


UNDOCUMENTED=$(comm -23 <(echo "$GO_KINDS") <(echo "$GO_DOC_KINDS"))
PHANTOM=$(comm -13 <(echo "$GO_KINDS") <(echo "$GO_DOC_KINDS"))

if [[ -n "$UNDOCUMENTED" ]]; then
  echo "message-kind-parity: types dispatched in the MSG_TYPES fence but NOT documented in MSG_TYPES_DOC:"
  while IFS= read -r k; do
    echo "  undocumented: \"$k\"  (add a numbered entry to the header)"
    HITS=$((HITS + 1))
  done <<< "$UNDOCUMENTED"
fi

if [[ -n "$PHANTOM" ]]; then
  echo "message-kind-parity: types documented in MSG_TYPES_DOC that the switch does NOT dispatch:"
  while IFS= read -r k; do
    echo "  phantom: \"$k\"  (the header describes a type that no longer exists)"
    HITS=$((HITS + 1))
  done <<< "$PHANTOM"
fi










kinds_from_live_cases() {
  awk '
    /^[[:space:]]*\/\/[[:space:]]*LIVE_CASES_START[[:space:]]*$/ { inblk=1; next }
    /^[[:space:]]*\/\/[[:space:]]*LIVE_CASES_END[[:space:]]*$/   { inblk=0 }
    inblk
  ' "$HANDLE_MESSAGE_TS" \
    | grep -aoE 'case[[:space:]]+"[^"]+"' \
    | grep -oE '"[^"]+"' \
    | tr -d '"' \
    | sort -u
}

LIVE_CASE_KINDS=$(kinds_from_live_cases) || true
if [[ -z "$(printf '%s' "$LIVE_CASE_KINDS" | tr -d '[:space:]')" ]]; then
  echo "message-kind-parity: EMPTY extracted set for 'handle-message.ts LIVE_CASES fence' — fence missing or renamed; refusing vacuous parity pass" >&2
  exit 1
fi

while IFS= read -r k; do
  [[ -z "$k" ]] && continue
  if ! grep -arE "type:[[:space:]]*\"$k\"" "$WEBVIEW_SRC_DIR" >/dev/null 2>&1; then
    echo "  no sender: \"$k\" is a LIVE_CASES handler in handle-message.ts but nothing under" \
         "$WEBVIEW_SRC_DIR posts { type: \"$k\" } (move it to DECLARED_NOT_SENT ONLY if Go" \
         "still dispatches it — that hatch is not a free pass, see check (4) below)"
    HITS=$((HITS + 1))
  fi
done <<< "$LIVE_CASE_KINDS"



# MSG_TYPES switch does NOT recognize is dead on both sides: no sender, no real handler


kinds_from_declared_not_sent() {
  awk '
    /^[[:space:]]*\/\/[[:space:]]*DECLARED_NOT_SENT_START[[:space:]]*$/ { inblk=1; next }
    /^[[:space:]]*\/\/[[:space:]]*DECLARED_NOT_SENT_END[[:space:]]*$/   { inblk=0 }
    inblk
  ' "$HANDLE_MESSAGE_TS" \
    | grep -aoE 'case[[:space:]]+"[^"]+"' \
    | grep -oE '"[^"]+"' \
    | tr -d '"' \
    | sort -u
}

DECLARED_NOT_SENT_KINDS=$(kinds_from_declared_not_sent) || true
if [[ -z "$(printf '%s' "$DECLARED_NOT_SENT_KINDS" | tr -d '[:space:]')" ]]; then
  echo "message-kind-parity: EMPTY extracted set for 'handle-message.ts DECLARED_NOT_SENT fence' — fence missing or renamed; refusing vacuous parity pass" >&2
  exit 1
fi

NOT_GO_PARITY=$(comm -23 <(echo "$DECLARED_NOT_SENT_KINDS") <(echo "$GO_KINDS"))
if [[ -n "$NOT_GO_PARITY" ]]; then
  echo "message-kind-parity: kinds in DECLARED_NOT_SENT that Go's msg.Type switch does NOT dispatch:"
  while IFS= read -r k; do
    echo "  dead-on-both-sides: \"$k\" has no live TS sender AND stdin_reader.go does not" \
         "recognize it — DECLARED_NOT_SENT is only for Go-parity kinds; remove this kind" \
         "or wire a real sender/handler"
    HITS=$((HITS + 1))
  done <<< "$NOT_GO_PARITY"
fi

if [[ $HITS -eq 0 ]]; then
  echo "message-kind-parity: clean"
  exit 0
fi

echo ""
echo "message-kind-parity: $HITS divergence(s) found"
exit 1
