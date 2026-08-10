#!/usr/bin/env bash
#
# PLACEMENT: docs/pair-node/*.html | a data-src naming a definition must name one that still exists
set -euo pipefail

# check-docs-symbols.sh — every source name on the pair-node docs pages points at
# something real.
#
# The pages carry `data-src="nodes/PairNode/node.go#handleVectorCycle"`; clicking the name
# opens that file AT THAT DEFINITION (docs-open.ts's findDefinitionLine, which resolves
# the name at click time — the pages hold no line numbers, precisely so an edit above a
# definition cannot silently move a link).
#
# What a name CAN do is go stale by rename or deletion, and the failure is quiet: the
# file still opens, just at the top, and the page keeps claiming a function that is gone.
# This turns that into a check failure. It is the same reasoning as the docs pages'
# formulas table — a doc claim about code is only worth keeping if something verifies it
# (memory/feedback_check_the_signal_the_check_emits.md).
#
# The definition patterns MIRROR findDefinitionLine's. If one gains a form, so does the
# other, or this guard passes a name the extension cannot find.

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

fail=0
note() { echo "check-docs-symbols: $*" >&2; fail=1; }

# Every data-src value on every docs page, one per line: `path` or `path#symbol`.
refs=$(grep -ho 'data-src="[^"]*"' docs/*/*.html 2>/dev/null \
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

  # Mirrors findDefinitionLine: Go func/method, Go type/const/var, a name inside a
  # tab-indented const/var/type block, and the TS function/class/const forms.
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
