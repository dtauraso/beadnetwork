#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: nodes/Wiring/stdin_reader.go,tools/topology-vscode/src/runCommand.ts | maxFrameBytes and MAX_FRAME_BYTES must be numerically equal
#
# Verifies that the two sides of the SAME framed-binary protocol ([len:u32-LE][payload])
# agree on the maximum frame length they will accept:
#   nodes/Wiring/stdin_reader.go        `const maxFrameBytes = ...`   (TS -> Go direction)
#   tools/topology-vscode/src/runCommand.ts  `export const MAX_FRAME_BYTES = ...`  (Go -> TS direction)
#
# This protocol is used in BOTH directions on the same wire shape, and each side's reader
# is the one that must bound ITS OWN allocation — a bound that exists on only one side is
# the defect (a corrupt/hostile length on the unguarded side grows the carry-over buffer
# without limit). This guard is generated-vs-generated in spirit but hand-authored-vs-
# hand-authored in practice: it does not regenerate anything, it only asserts the two
# CONSTANTS are numerically equal. It does NOT verify either side's logic actually rejects
# an over-limit frame (see the reader code itself / the guard's own file comments for that);
# it only catches the constants drifting apart.
#
# Exit 0 if clean; exit 1 with a report if the values diverge.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

GO_FILE="$REPO_ROOT/nodes/Wiring/stdin_reader.go"
TS_FILE="$REPO_ROOT/tools/topology-vscode/src/runCommand.ts"

for f in "$GO_FILE" "$TS_FILE"; do
  if [[ ! -f "$f" ]]; then
    echo "check-frame-bytes-parity: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

# Extract the literal expression after the `=` on the const declaration line, e.g.
# "1 << 20". Evaluated with bash arithmetic so "1 << 20" and "1048576" compare equal.
value_go() {
  grep -aE '^[[:space:]]*const[[:space:]]+maxFrameBytes[[:space:]]*=' "$GO_FILE" \
    | sed -E 's/^[[:space:]]*const[[:space:]]+maxFrameBytes[[:space:]]*=[[:space:]]*//'
}

value_ts() {
  grep -aE '^[[:space:]]*export const MAX_FRAME_BYTES[[:space:]]*=' "$TS_FILE" \
    | sed -E 's/^[[:space:]]*export const MAX_FRAME_BYTES[[:space:]]*=[[:space:]]*//' \
    | sed -E 's/;[[:space:]]*$//'
}

# Refuse a vacuous pass: an empty extraction means the declaration is missing (renamed,
# moved, comment-only, etc.). Assert non-empty.
assert_nonempty() { # value label
  if [[ -z "$(printf '%s' "$1" | tr -d '[:space:]')" ]]; then
    echo "check-frame-bytes-parity: EMPTY value for '$2' — const declaration missing or renamed; refusing vacuous parity pass" >&2
    exit 1
  fi
}

RAW_GO=$(value_go)
RAW_TS=$(value_ts)

assert_nonempty "$RAW_GO" "nodes/Wiring/stdin_reader.go maxFrameBytes"
assert_nonempty "$RAW_TS" "runCommand.ts MAX_FRAME_BYTES"

# Evaluate as bash arithmetic so "1 << 20" (Go/TS shift syntax happens to match bash's)
# compares equal to a plain decimal literal on either side.
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
