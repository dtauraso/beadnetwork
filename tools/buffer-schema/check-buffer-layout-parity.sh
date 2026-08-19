#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: tools/topology-vscode/src/Buffer/buffer_layout_gen.go,tools/topology-vscode/src/Buffer/buffer-layout.ts | BUF_LAYOUT_FINGERPRINT must match between the two generated layout files

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

GO_FILE="$REPO_ROOT/tools/topology-vscode/src/Buffer/buffer_layout_gen.go"
TS_FILE="$REPO_ROOT/tools/topology-vscode/src/Buffer/buffer-layout.ts"

for f in "$GO_FILE" "$TS_FILE"; do
  if [[ ! -f "$f" ]]; then
    echo "check-buffer-layout-parity: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

fingerprint_go() {
  grep -a 'BUF_LAYOUT_FINGERPRINT:' "$GO_FILE" \
    | sed 's/.*BUF_LAYOUT_FINGERPRINT: //'
}

fingerprint_ts() {
  grep -a 'BUF_LAYOUT_FINGERPRINT:' "$TS_FILE" \
    | sed 's/.*BUF_LAYOUT_FINGERPRINT: //'
}

assert_nonempty() {
  if [[ -z "$(printf '%s' "$1" | tr -d '[:space:]')" ]]; then
    echo "check-buffer-layout-parity: EMPTY fingerprint for '$2' — BUF_LAYOUT_FINGERPRINT comment missing or file renamed; refusing vacuous parity pass" >&2
    exit 1
  fi
}

FP_GO=$(fingerprint_go)
FP_TS=$(fingerprint_ts)

assert_nonempty "$FP_GO" "tools/topology-vscode/src/Buffer/buffer_layout_gen.go"
assert_nonempty "$FP_TS" "buffer-layout.ts"

if [[ "$FP_GO" != "$FP_TS" ]]; then
  echo "check-buffer-layout-parity: Go and TS buffer layout fingerprints DIVERGE"
  echo ""
  echo "  Go  ($GO_FILE):"
  echo "    $FP_GO"
  echo ""
  echo "  TS  ($TS_FILE):"
  echo "    $FP_TS"
  echo ""
  echo "Regenerate with: cd tools/topology-vscode && npm run gen:node-defs"
  exit 1
fi

echo "check-buffer-layout-parity: clean (fingerprint matches)"
exit 0
