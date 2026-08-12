#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: tools/topology-vscode/src/**/*.ts,tools/topology-vscode/src/**/*.tsx | TS→Go sends (postGoRecord/sendRawInput/writeStdin/postMessage) must be fire-and-forget, no await/.then

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

SRC_DIR="$REPO_ROOT/tools/topology-vscode/src"

WRITESTDIN_FILES=$(grep -rlE '\bwriteStdin\(' --include="*.ts" "$SRC_DIR" 2>/dev/null \
  | xargs grep -lE '\bwriteStdin\([^)]*\)[[:space:]]*:' 2>/dev/null || true)

SEND_FNS='postGoRecord|sendRawInput|writeStdin|postMessage'

HITS=0
report() {
  printf '%s\n' "$1"
  HITS=$((HITS + 1))
}

while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  report "await-on-send: $line"
done < <(grep -arnE "await[[:space:]]+([A-Za-z0-9_]+\.)*($SEND_FNS)\(" \
  --include="*.ts" --include="*.tsx" "$SRC_DIR" 2>/dev/null || true)

while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  report "promise-chain-on-send: $line"
done < <(grep -arnE "($SEND_FNS)\(.*\)[[:space:]]*\.(then|catch|finally)\b" \
  --include="*.ts" --include="*.tsx" "$SRC_DIR" 2>/dev/null || true)

[[ -n "$WRITESTDIN_FILES" ]] || {
  echo "no-await-on-bridge: MISCONFIGURED — nothing under $SRC_DIR declares writeStdin(...): <type>." >&2
  echo "The runner's send method moved or was renamed; this guard is scanning nothing rather than" >&2
  echo "passing vacuously. Repoint it." >&2
  exit 1
}

while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  report "writeStdin-returns-promise: $line"
done < <(grep -nE 'writeStdin\([^)]*\)[[:space:]]*:[[:space:]]*(Promise|Thenable)' $WRITESTDIN_FILES 2>/dev/null || true)

if ! grep -qE 'writeStdin\([^)]*\)[[:space:]]*:[[:space:]]*void' $WRITESTDIN_FILES; then
  report "writeStdin-not-void: $WRITESTDIN_FILES does not declare writeStdin(...): void — the TS→Go send must be fire-and-forget"
fi

SEND_CALL_COUNT=$(grep -arlE "($SEND_FNS)\(" --include="*.ts" --include="*.tsx" "$SRC_DIR" 2>/dev/null | wc -l | tr -d '[:space:]')
if [[ "$SEND_CALL_COUNT" -eq 0 ]]; then
  report "no-send-calls-found: none of ($SEND_FNS) appear anywhere under $SRC_DIR — the send surface names likely changed; update SEND_FNS or this guard is scanning nothing"
fi

if [[ $HITS -eq 0 ]]; then
  echo "no-await-on-bridge: clean (TS→Go send is fire-and-forget; no await/Promise/request-response on the bridge)"
  exit 0
fi

echo ""
echo "no-await-on-bridge: $HITS hit(s) — the TS→Go send path must be fire-and-forget (no await, no Promise chain, no request/response)"
exit 1
