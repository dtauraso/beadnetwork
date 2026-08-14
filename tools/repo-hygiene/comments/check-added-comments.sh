#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: none | repo-wide: no comment line added since the base ref survives in a hand-edited file

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

report="$(python3 "$SCRIPT_DIR/strip_added_comments.py")"

summary="$(head -n1 <<< "$report")"
rest="$(tail -n +2 <<< "$report")"
echo "$summary"
[[ -z "$rest" ]] && exit 0

fail=0
while IFS= read -r line; do
  case "$line" in
    REMOVED)
      echo "ADDED COMMENTS REMOVED: these comment lines were added since the base ref and have"
      echo "been deleted from the working tree. Re-read the diff before committing — if one of"
      echo "them was load-bearing, put it back deliberately and it will not be touched again"
      echo "once it is part of the base:"
      fail=1; continue ;;
    *) printf '  %s\n' "$line" ;;
  esac
done <<< "$rest"

exit $fail
