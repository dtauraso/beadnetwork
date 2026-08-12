#!/usr/bin/env bash

# PLACEMENT: docs/pair-node/**/*.html | a data-src naming a definition must name one that still exists
set -euo pipefail


















REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

fail=0
note() { echo "check-docs-symbols: $*" >&2; fail=1; }


refs=$(grep -ho 'data-src="[^"]*"' docs/pair-node/*.html docs/pair-node/*/*.html 2>/dev/null \
  | sed -E 's/^data-src="//; s/"$//' | sort -u)

if [ -z "$refs" ]; then
  echo "check-docs-symbols: MISCONFIGURED — no data-src attributes found; refusing to report success." >&2
  exit 1
fi

while IFS= read -r ref; do
  [ -n "$ref" ] || continue
  file="${ref%%#*}"
  symbol=""
  case "$ref" in *#*) symbol="${ref#*#}" ;; esac

  if [ ! -f "$file" ]; then
    note "$file does not exist (referenced as \"$ref\")"
    continue
  fi
  [ -n "$symbol" ] || continue



  if ! grep -qE \
    "^func[[:space:]]+(\([^)]*\)[[:space:]]*)?${symbol}\b|\
^(type|const|var)[[:space:]]+${symbol}\b|\
^	${symbol}[[:space:]=]|\
^[[:space:]]*(export[[:space:]]+)?(async[[:space:]]+)?function[[:space:]]+${symbol}\b|\
^[[:space:]]*(export[[:space:]]+)?(abstract[[:space:]]+)?(class|interface|type|enum)[[:space:]]+${symbol}\b|\
^[[:space:]]*(export[[:space:]]+)?(const|let|var)[[:space:]]+${symbol}\b" \
    "$file"; then
    note "$file has no definition of \"$symbol\" — the docs page names something that moved or is gone"
  fi
done <<< "$refs"

exit "$fail"
