#!/usr/bin/env bash

# PLACEMENT: nodes/SPEC-FORMAT.md,cmd/gen-node-defs/kindscan/spec_md.go | the `## View` field table must name exactly the view.* fields parseSpecMD reads
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

GEN_DIR="cmd/gen-node-defs"
DOC="nodes/SPEC-FORMAT.md"

if [[ ! -d "$GEN_DIR" ]]; then
  echo "check-spec-format-view-fields: MISSING $GEN_DIR — cannot verify parity, refusing vacuous pass" >&2
  exit 1
fi
if [[ ! -f "$DOC" ]]; then
  echo "check-spec-format-view-fields: MISSING $DOC — cannot verify parity, refusing vacuous pass" >&2
  exit 1
fi

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

grep -rhoE 'vmap\["[A-Za-z0-9_]+"\]' "$GEN_DIR" --include="*.go" \
  | sed -E 's/vmap\["([A-Za-z0-9_]+)"\]/\1/' \
  | sort -u > "$TMP/code_fields.txt"

if [[ ! -s "$TMP/code_fields.txt" ]]; then
  echo "check-spec-format-view-fields: EMPTY vmap[...] extraction from $GEN_DIR/*.go — extractor broken, refusing vacuous pass" >&2
  exit 1
fi

awk '
  /^\| *Field *\| *Value *\|/ { intable=1; next }
  intable && /^\|[-: ]+\|[-: ]+\|/ { next }
  intable && /^\|/ {
    line=$0
    sub(/^\|/, "", line)
    split(line, parts, "|")
    field=parts[1]
    gsub(/^[ \t]+|[ \t]+$/, "", field)
    if (field != "") print field
    next
  }
  intable && !/^\|/ { intable=0 }
' "$DOC" | sort -u > "$TMP/doc_fields.txt"

if [[ ! -s "$TMP/doc_fields.txt" ]]; then
  echo "check-spec-format-view-fields: EMPTY field extraction from $DOC — extractor broken, refusing vacuous pass" >&2
  exit 1
fi

DOC_ONLY=$(comm -23 "$TMP/doc_fields.txt" "$TMP/code_fields.txt" || true)
CODE_ONLY=$(comm -13 "$TMP/doc_fields.txt" "$TMP/code_fields.txt" || true)

if [[ -z "$DOC_ONLY" && -z "$CODE_ONLY" ]]; then
  echo "check-spec-format-view-fields: clean"
  exit 0
fi

echo "check-spec-format-view-fields: $DOC's ## View field table is out of parity with"
echo "$GEN_DIR's parseSpecMD vmap[...] reads:"
echo ""
if [[ -n "$DOC_ONLY" ]]; then
  echo "  documented in $DOC but NOT read by the generator (dead/ghost field — remove from the doc, or if it should do something, wire it into main.go):"
  echo "$DOC_ONLY" | sed 's/^/    - /'
fi
if [[ -n "$CODE_ONLY" ]]; then
  echo "  read by the generator (vmap[\"...\"]) but NOT documented in $DOC (undiscoverable field — document it, or if vestigial, delete the vmap read):"
  echo "$CODE_ONLY" | sed 's/^/    - /'
fi
echo ""
exit 1
