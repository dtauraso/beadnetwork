#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <branch>" >&2
  exit 1
fi

BRANCH="$1"
DOCS_DIR="docs"

if [[ ! -d "$DOCS_DIR" ]]; then
  echo "Error: $DOCS_DIR not found. Run from repo root." >&2
  exit 1
fi

matched=()
while IFS= read -r -d '' file; do
  head10=$(head -10 "$file" 2>/dev/null || true)

  if grep -qE "^branch: ${BRANCH}$" <<< "$head10"; then
    matched+=("$file")
    continue
  fi

  if grep -qE "^[[:space:]]*<!--[[:space:]]*branch:[[:space:]]*${BRANCH}[[:space:]]*-->[[:space:]]*$" <<< "$head10"; then
    matched+=("$file")
  fi
done < <(find "$DOCS_DIR" -type f \( -name "*.md" -o -name "*.html" \) -print0)

if [[ ${#matched[@]} -eq 0 ]]; then
  echo "No docs tagged with branch: $BRANCH — nothing to remove."
  exit 0
fi

echo "Removing ${#matched[@]} doc(s) tagged with branch: $BRANCH"
for f in "${matched[@]}"; do
  git rm "$f"
  echo "  removed: $f"
done

for f in "${matched[@]}"; do

  hits=$(grep -rln --exclude-dir=node_modules --exclude-dir=.git --exclude-dir=.claude -- "$f" . 2>/dev/null | grep -v '^\./docs/' || true)
  if [[ -n "$hits" ]]; then
    echo "WARNING: $f is still cited from:"
    echo "$hits" | sed 's/^/  /'
  fi
done
echo "Done."
