#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: Input/Stdin/stdin_reader.go,extension/messages.ts,extension/handle-message.ts,extension/webview/** | a new message kind needs matching entries in Go's MSG_TYPES fence+doc, messages.ts, and a live sender

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

GO_PKG_DIR="$REPO_ROOT"
MESSAGES_TS="$REPO_ROOT/extension/messages.ts"
HANDLE_MESSAGE_TS="$REPO_ROOT/extension/handle-message.ts"
WEBVIEW_SRC_DIR="$REPO_ROOT/extension/webview"

source "$SCRIPT_DIR/msg-kind-extract.sh"
source "$SCRIPT_DIR/msg-kind-checks.sh"

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

GO_KINDS=$(kinds_from_go) || true
GO_DOC_KINDS=$(kinds_from_go_doc) || true
TS_KINDS=$(kinds_from_ts) || true

for pair in "GO_KINDS:stdin_reader.go MSG_TYPES fenced switch" \
            "GO_DOC_KINDS:stdin_reader.go MSG_TYPES_DOC header list" \
            "TS_KINDS:WEBVIEW_TO_HOST_TYPES kinds"; do
  var="${pair%%:*}"; label="${pair#*:}"
  require_nonempty_set_or_die "$var" "$label"
done

HITS=0
check_go_vs_ts_parity
check_go_doc_parity

LIVE_CASE_KINDS=$(kinds_from_live_cases) || true
require_nonempty_set_or_die LIVE_CASE_KINDS "handle-message.ts LIVE_CASES fence"
check_live_cases_have_senders

DECLARED_NOT_SENT_KINDS=$(kinds_from_declared_not_sent) || true
require_nonempty_set_or_die DECLARED_NOT_SENT_KINDS "handle-message.ts DECLARED_NOT_SENT fence"
check_declared_not_sent_has_go_parity

if [[ $HITS -eq 0 ]]; then
  echo "message-kind-parity: clean"
  exit 0
fi

echo ""
echo "message-kind-parity: $HITS divergence(s) found"
exit 1
