#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: Categories/Scene/Drag/encode.ts,Categories/Scene/Drag/raw_input_decode.go | the raw-input record is positional, so the writer's field order and width must match the reader's exactly

REPO_ROOT="$(git rev-parse --show-toplevel)"

ENCODE_TS="$REPO_ROOT/Categories/Scene/Drag/encode.ts"
DECODE_GO="$REPO_ROOT/Categories/Scene/Drag/raw_input_decode.go"

for f in "$ENCODE_TS" "$DECODE_GO"; do
  if [[ ! -f "$f" ]]; then
    echo "✗ raw-input-parity: MISCONFIGURED — file not found: $f" >&2
    exit 1
  fi
done

python3 - "$ENCODE_TS" "$DECODE_GO" <<'PY'
import re, sys

encode_path, decode_path = sys.argv[1], sys.argv[2]

body = open(encode_path).read().split("export function encodeRawInput", 1)[-1]
writes = re.findall(r"\bw\.(u8|i32|f64|bool|str)\(", body)
if writes and writes[0] == "u8":
    writes = writes[1:]

HELPER = {"f": "f64", "i": "i32", "b": "bool", "u": "u8"}
reads = []
for line in open(decode_path):
    line = line.strip()
    if re.match(r"ev\.[A-Za-z.]+\s*=\s*enumAt\(", line):
        reads.append("u8")
    elif re.match(r"ev\.[A-Za-z.]+\s*=\s*[fibu]\(\)$", line):
        reads.append(HELPER[line[line.rindex("=") + 1:].strip()[0]])
    elif re.match(r"ev\.[A-Za-z.]+\s*=\s*\w+\{", line):
        reads.extend(HELPER[c] for c in re.findall(r"\b([fibu])\(\)", line))
    elif "r.Str()" in line:
        reads.append("str")

if writes == reads:
    sys.exit(0)

out = [
    "raw-input-parity: encode.ts and raw_input_decode.go disagree.",
    "  The record carries no field names, so every field is found by offset:",
    "  the first mismatch misreads everything after it, and both sides still",
    "  compile and both guards stay green.",
]
first = next((i for i in range(max(len(writes), len(reads)))
              if (writes[i:i + 1] or [None]) != (reads[i:i + 1] or [None])), None)
for i in range(max(len(writes), len(reads))):
    w = writes[i] if i < len(writes) else "-"
    r = reads[i] if i < len(reads) else "-"
    mark = "   <-- first divergence" if i == first else ""
    out.append(f"    [{i:2}] write {w:<5} read {r:<5}{mark}")
print("\n".join(out), file=sys.stderr)
sys.exit(1)
PY
