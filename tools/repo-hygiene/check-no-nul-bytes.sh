#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: **/*.{go,ts,tsx,js,jsx,json,md,sh,css} | must not contain a literal 0x00 byte

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

INCLUDE_EXT_RE='\.(go|ts|tsx|js|jsx|json|md|sh|css)$'

HITS=0
report() {
  printf '%s\n' "$1"
  HITS=$((HITS + 1))
}

while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  f="${hit%%:*}"
  rest="${hit#*:}"
  byte_off="${rest%%:*}"
  line_no="${rest##*:}"
  report "nul-byte: $f: byte offset $byte_off (line $line_no)"
done < <(git ls-files | grep -E "$INCLUDE_EXT_RE" | python3 -c "
import sys, os

for path in sys.stdin.read().splitlines():
    if not path:
        continue
    if not os.path.isfile(path):
        continue
    try:
        data = open(path, 'rb').read()
    except Exception:
        continue
    idx = data.find(b'\x00')
    if idx == -1:
        continue
    line_no = data.count(b'\n', 0, idx) + 1
    print(f'{path}:{idx}:{line_no}')
" || true)

if [[ $HITS -eq 0 ]]; then
  echo "no-nul-bytes: clean (no literal NUL bytes in tracked source files)"
  exit 0
fi

echo ""
echo "no-nul-bytes: $HITS hit(s) — literal 0x00 byte(s) found in tracked source; this silently turns the file BINARY to git (diffs show Bin X -> Y, grep goes silent, all grep-based guards are blinded). Almost always a stray '\\0' escape that landed as a real byte instead of two characters. Fix by replacing the literal NUL with the intended escape sequence."
exit 1
