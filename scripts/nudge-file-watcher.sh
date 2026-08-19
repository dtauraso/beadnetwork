#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

INCLUDE_RE='\.(go|ts|tsx)$'

{
  git diff --name-only HEAD -- . 2>/dev/null || true
  git ls-files --others --exclude-standard 2>/dev/null || true
} | grep -E "$INCLUDE_RE" | sort -u > /tmp/.nudge-list.$$ || true

count=0
while IFS= read -r f; do
  [ -f "$f" ] || continue
  touch "$f"
  count=$((count + 1))
done < /tmp/.nudge-list.$$
rm -f /tmp/.nudge-list.$$

if [ "$count" -gt 0 ]; then
  echo "nudge-file-watcher: re-emitted events for $count changed source file(s)"
fi
exit 0
