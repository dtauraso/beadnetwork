#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: **/*.go,Node/*.go,Node/BeadAnimation/send-rule.ts | a new SendRule const must also appear in the SEND_RULES array in send-rule.ts, which sits beside send_rule.go

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

WIRING_DIR="$REPO_ROOT"
WIRE_DIR="$REPO_ROOT/Node"
TYPES_TS="$REPO_ROOT/Node/BeadAnimation/send-rule.ts"

if [[ ! -d "$WIRING_DIR" ]]; then
  echo "send-rule-parity: MISCONFIGURED — dir not found: $WIRING_DIR" >&2
  exit 1
fi
if [[ ! -d "$WIRE_DIR" ]]; then
  echo "send-rule-parity: MISCONFIGURED — dir not found: $WIRE_DIR" >&2
  exit 1
fi
if [[ ! -f "$TYPES_TS" ]]; then
  echo "send-rule-parity: MISCONFIGURED — file not found: $TYPES_TS" >&2
  exit 1
fi

rules_from_go() {
  local go_files
  go_files=$(cd "$REPO_ROOT" && git ls-files '*.go')
  [[ -z "$go_files" ]] && return
  ( cd "$REPO_ROOT" && grep -haE '[A-Za-z0-9_]+[[:space:]]+SendRule[[:space:]]*=[[:space:]]*"[^"]+"' $go_files ) \
    | grep -o '"[^"]*"' \
    | tr -d '"' \
    | sort
}

rules_from_ts() {
  grep -a 'SEND_RULES' "$TYPES_TS" \
    | grep -o '"[^"]*"' \
    | tr -d '"' \
    | sort
}

GO_RULES=$(rules_from_go) || true
TS_RULES=$(rules_from_ts) || true

assert_nonempty() {
  if [[ -z "$(printf '%s' "$1" | tr -d '[:space:]')" ]]; then
    echo "send-rule-parity: EMPTY extracted set for '$2' — const/array missing or renamed; refusing vacuous parity pass" >&2
    exit 1
  fi
}
assert_nonempty "$GO_RULES" "SendRule consts (Node/BeadAnimation/send_rule.go)"
assert_nonempty "$TS_RULES" "SEND_RULES array (send-rule.ts)"

MISSING=$(comm -23 <(echo "$GO_RULES") <(echo "$TS_RULES"))
EXTRA=$(comm -13 <(echo "$GO_RULES") <(echo "$TS_RULES"))

HITS=0
if [[ -n "$MISSING" ]]; then
  while IFS= read -r k; do
    [[ -z "$k" ]] && continue
    echo "  SendRule in send_rule.go but missing from SEND_RULES in send-rule.ts: \"$k\""
    HITS=$((HITS + 1))
  done <<< "$MISSING"
fi

if [[ -n "$EXTRA" ]]; then
  while IFS= read -r k; do
    [[ -z "$k" ]] && continue
    echo "  SendRule in SEND_RULES (send-rule.ts) but not defined in send_rule.go: \"$k\""
    HITS=$((HITS + 1))
  done <<< "$EXTRA"
fi

if [[ $HITS -eq 0 ]]; then
  echo "send-rule-parity: clean"
  exit 0
fi

echo ""
echo "send-rule-parity: $HITS divergence(s) found"
exit 1
