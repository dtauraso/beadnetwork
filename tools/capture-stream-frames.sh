#!/usr/bin/env bash

# PLACEMENT: none | repo-wide: capture the real per-goroutine stream frames Go emits, for reading back with decode-stream-frames.py

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

OUT="${1:-}"
SCENE="${2:-topology}"
SECONDS_TO_RUN="${3:-10}"

if [[ -z "$OUT" ]]; then
  echo "usage: capture-stream-frames.sh <out-dir> [scene] [seconds]" >&2
  echo "  reads what Go ACTUALLY emits, rather than recomputing it. A check that" >&2
  echo "  reimplements the same arithmetic on both sides cannot disagree with itself." >&2
  exit 2
fi

mkdir -p "$OUT"

exec 3>"$OUT/view.bin"
for i in $(seq 0 9); do eval "exec $((4 + i))>\"$OUT/edge$i.bin\""; done
for i in $(seq 0 8); do eval "exec $((14 + i))>\"$OUT/node$i.bin\""; done
for i in $(seq 0 8); do eval "exec $((23 + i))>\"$OUT/interior$i.bin\""; done
for i in $(seq 0 8); do eval "exec $((32 + i))>\"$OUT/bead$i.bin\""; done

WIREFOLD_STREAM_FDS="view:3,edge:4,node:14,interior:23,bead:32" \
  bash "$SCRIPT_DIR/run-bounded.sh" "$SECONDS_TO_RUN" go run . -topology "$SCENE" \
  >"$OUT/stdout.txt" 2>"$OUT/stderr.txt"

exec 3>&-
for i in $(seq 0 9); do eval "exec $((4 + i))>&-"; done
for i in $(seq 0 8); do eval "exec $((14 + i))>&-"; done
for i in $(seq 0 8); do eval "exec $((23 + i))>&-"; done
for i in $(seq 0 8); do eval "exec $((32 + i))>&-"; done

if grep -q "stream-fd mismatch" "$OUT/stderr.txt" 2>/dev/null; then
  echo "capture-stream-frames: Go reported a stream-fd mismatch — the capture is PARTIAL:" >&2
  head -c 400 "$OUT/stderr.txt" >&2
  echo >&2
  exit 1
fi

echo "capture-stream-frames: $SCENE for ${SECONDS_TO_RUN}s -> $OUT"
echo "  read it with: python3 $SCRIPT_DIR/decode-stream-frames.py $OUT"
