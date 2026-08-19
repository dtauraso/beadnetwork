#!/usr/bin/env bash

require_nonempty_set_or_die() {
  local var="$1" label="$2"
  if [[ -z "$(printf '%s' "${!var}" | tr -d '[:space:]')" ]]; then
    echo "message-kind-parity: EMPTY extracted set for '$label' — switch/const missing or renamed; refusing vacuous parity pass" >&2
    exit 1
  fi
}

check_go_vs_ts_parity() {
  local missing
  missing=$(comm -23 <(echo "$GO_KINDS") <(echo "$TS_KINDS"))
  if [[ -n "$missing" ]]; then
    echo "message-kind-parity: kinds in stdin_reader.go but missing from WEBVIEW_TO_HOST_TYPES:"
    while IFS= read -r k; do
      echo "  missing: \"$k\""
      HITS=$((HITS + 1))
    done <<< "$missing"
  fi
}

check_go_doc_parity() {
  local undocumented phantom
  undocumented=$(comm -23 <(echo "$GO_KINDS") <(echo "$GO_DOC_KINDS"))
  phantom=$(comm -13 <(echo "$GO_KINDS") <(echo "$GO_DOC_KINDS"))

  if [[ -n "$undocumented" ]]; then
    echo "message-kind-parity: types dispatched in the MSG_TYPES fence but NOT documented in MSG_TYPES_DOC:"
    while IFS= read -r k; do
      echo "  undocumented: \"$k\"  (add a numbered entry to the header)"
      HITS=$((HITS + 1))
    done <<< "$undocumented"
  fi

  if [[ -n "$phantom" ]]; then
    echo "message-kind-parity: types documented in MSG_TYPES_DOC that the switch does NOT dispatch:"
    while IFS= read -r k; do
      echo "  phantom: \"$k\"  (the header describes a type that no longer exists)"
      HITS=$((HITS + 1))
    done <<< "$phantom"
  fi
}

check_live_cases_have_senders() {
  while IFS= read -r k; do
    [[ -z "$k" ]] && continue
    if ! grep -arE "type:[[:space:]]*\"$k\"" "$WEBVIEW_SRC_DIR" >/dev/null 2>&1; then
      echo "  no sender: \"$k\" is a LIVE_CASES handler in handle-message.ts but nothing under" \
           "$WEBVIEW_SRC_DIR posts { type: \"$k\" } (move it to DECLARED_NOT_SENT ONLY if Go" \
           "still dispatches it — that hatch is not a free pass, see check (4) below)"
      HITS=$((HITS + 1))
    fi
  done <<< "$LIVE_CASE_KINDS"
}

check_declared_not_sent_has_go_parity() {
  local not_go_parity
  not_go_parity=$(comm -23 <(echo "$DECLARED_NOT_SENT_KINDS") <(echo "$GO_KINDS"))
  if [[ -n "$not_go_parity" ]]; then
    echo "message-kind-parity: kinds in DECLARED_NOT_SENT that Go's msg.Type switch does NOT dispatch:"
    while IFS= read -r k; do
      echo "  dead-on-both-sides: \"$k\" has no live TS sender AND stdin_reader.go does not" \
           "recognize it — DECLARED_NOT_SENT is only for Go-parity kinds; remove this kind" \
           "or wire a real sender/handler"
      HITS=$((HITS + 1))
    done <<< "$not_go_parity"
  fi
}
