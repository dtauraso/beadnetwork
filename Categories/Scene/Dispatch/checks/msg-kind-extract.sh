#!/usr/bin/env bash

kinds_from_go() {
  {
    grep -haoE 'msg\.Type[[:space:]]*[!=]=[[:space:]]*"[^"]+"' $GO_FILES \
      | grep -oE '"[^"]+"' \
      | tr -d '"'

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
