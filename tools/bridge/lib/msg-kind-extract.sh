#!/usr/bin/env bash

# Sourced by tools/bridge/check-message-kind-parity.sh. Extracts the four message-kind
# sets the guard compares: what Go dispatches, what Go documents, what TS is allowed to
# send, and what TS actually handles/declares.

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
