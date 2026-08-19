#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: nodes/Wiring/stdinreader/stdin_reader.go,src/runner/framing.ts | maxFrameBytes and MAX_FRAME_BYTES must be numerically equal

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

GO_FILE="$REPO_ROOT/nodes/Wiring/stdinreader/stdin_reader.go"
TS_FILE="$REPO_ROOT/src/runner/framing.ts"

for f in "$GO_FILE" "$TS_FILE"; do
  if [[ ! -f "$f" ]]; then
    echo "check-frame-bytes-parity: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

value_go() {
  grep -aE '^[[:space:]]*const[[:space:]]+maxFrameBytes[[:space:]]*=' "$GO_FILE" \
    | sed -E 's/^[[:space:]]*const[[:space:]]+maxFrameBytes[[:space:]]*=[[:space:]]*//'
}

value_ts() {
  grep -aE '^[[:space:]]*export const MAX_FRAME_BYTES[[:space:]]*=' "$TS_FILE" \
    | sed -E 's/^[[:space:]]*export const MAX_FRAME_BYTES[[:space:]]*=[[:space:]]*//' \
    | sed -E 's/;[[:space:]]*$//'
}

assert_nonempty() { # value label
  if [[ -z "$(printf '%s' "$1" | tr -d '[:space:]')" ]]; then
    echo "check-frame-bytes-parity: EMPTY value for '$2' — const declaration missing or renamed; refusing vacuous parity pass" >&2
    exit 1
  fi
}

RAW_GO=$(value_go)
RAW_TS=$(value_ts)

assert_nonempty "$RAW_GO" "nodes/Wiring/stdinreader/stdin_reader.go maxFrameBytes"
assert_nonempty "$RAW_TS" "runner/framing.ts MAX_FRAME_BYTES"

VAL_GO=$(( RAW_GO ))
VAL_TS=$(( RAW_TS ))

if [[ "$VAL_GO" != "$VAL_TS" ]]; then
  echo "check-frame-bytes-parity: Go and TS max-frame-length constants DIVERGE"
  echo ""
  echo "  Go  ($GO_FILE): maxFrameBytes = $RAW_GO  ($VAL_GO)"
  echo ""
  echo "  TS  ($TS_FILE): MAX_FRAME_BYTES = $RAW_TS  ($VAL_TS)"
  echo ""
  echo "These bound the SAME framed-binary protocol in both directions and must match."
  exit 1
fi

echo "check-frame-bytes-parity: clean (maxFrameBytes == MAX_FRAME_BYTES == $VAL_GO)"
exit 0
