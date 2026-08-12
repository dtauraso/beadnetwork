#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: none | repo-wide hygiene check that every *.go/*.ts/*.tsx/*.js/*.jsx/*.sh/*.py file is git-visible (tracked or intent-to-add)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

INCLUDE_RE='\.(go|ts|tsx|js|jsx|sh|py)$'

untracked=()
while IFS= read -r f; do
  [ -n "$f" ] && untracked+=("$f")
done < <(git ls-files --others --exclude-standard -z \
  | tr '\0' '\n' \
  | grep -E "$INCLUDE_RE" || true)

if [ ${#untracked[@]} -eq 0 ]; then
  echo "no-untracked-source: clean (every source file is visible to git ls-files)"
  exit 0
fi

echo "no-untracked-source: ${#untracked[@]} untracked source file(s) — these are INVISIBLE to the six git-ls-files-driven guards (doc-symbols, doc-citations, no-nul-bytes, send-rule-parity, generated, kind-imports), so they are currently unchecked:" >&2
for f in "${untracked[@]}"; do
  echo "  $f" >&2
done
echo >&2
echo "Make them visible without staging their content:" >&2
printf '  git add -N' >&2
for f in "${untracked[@]}"; do
  printf ' %q' "$f" >&2
done
echo >&2
exit 1
