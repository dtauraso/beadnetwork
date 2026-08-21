#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# -prune, not -not -path: the scene's view/ tree is rewritten constantly through
# tmp+rename, so a walk that DESCENDS into it hits entries that vanish mid-walk
# and exits nonzero, which under set -e kills this script with no output.
FOUND=$(find . \
  -type d \( -name node_modules -o -name out -o -name .git -o -name view \) -prune -o \
  -name 'check-*.sh' -type f -print \
  | sed 's|^\./||' | sort)

if [ "$(printf '%s\n' "$FOUND" | grep -c .)" -lt 40 ]; then
  echo "guard-list: MISCONFIGURED — found $(printf '%s\n' "$FOUND" | grep -c .) guards, expected at least 40." >&2
  echo "  Guards live beside what they guard, so they are found by searching, not by a list." >&2
  echo "  A search that suddenly finds far fewer means the search is wrong, not that the" >&2
  echo "  guards are gone — and a short list here silently shrinks every run that uses it." >&2
  exit 1
fi

printf '%s\n' "$FOUND"
