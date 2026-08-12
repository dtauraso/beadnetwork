#!/usr/bin/env bash

# Sourced by tools/docs/check-doc-citations.sh. Turns CLAUDE.md/MODEL.md into flat,
# lowercased, markup-stripped blobs ($TMP/claude.txt, $TMP/model.txt) that a citation's
# quoted text can be substring-matched against.

require_docs_present() {
  for doc in CLAUDE.md MODEL.md; do
    if [[ ! -f "$doc" ]]; then
      echo "doc-citations: MISCONFIGURED — $doc not found (renamed? a missing doc would vacuously pass)" >&2
      exit 1
    fi
  done
}

normalize() {
  tr '\n' ' ' < "$1" | tr -s '[:space:]' ' ' | tr '[:upper:]' '[:lower:]' | sed 's/[*`]//g'
}

write_normalized_docs() {
  normalize CLAUDE.md > "$TMP/claude.txt"
  normalize MODEL.md  > "$TMP/model.txt"

  for f in claude model; do
    if [[ ! -s "$TMP/$f.txt" ]]; then
      echo "doc-citations: EMPTY normalized text for $f — extractor broken; refusing vacuous pass" >&2
      exit 1
    fi
  done
}
