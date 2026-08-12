#!/usr/bin/env bash

# Sourced by tools/docs/check-doc-citations.sh. Strips comment markers from each candidate
# file in $TMP/files.txt and flattens it to one line per file in $TMP/norm.txt, with
# $TMP/paths.txt carrying the matching path per line — then checks the three stayed in
# lockstep, since a silent drop here would mean a file was never actually checked.

write_normalized_candidates() {
  awk -v PATHS="$TMP/paths.txt" '
  function emit() {
    if (path == "") return
    gsub(/"[ \t]*"/, "", text)
    gsub(/\\"/, "\"", text)
    print text
    print path > PATHS
  }
  FNR == 1 { emit(); path = FILENAME; text = ""
    if (FILENAME ~ /\.(go|ts|tsx)$/)  e = "c"
    else if (FILENAME ~ /\.(sh|py)$/) e = "h"
    else                              e = "n"
  }
  { line = $0
    if (e == "c")      sub(/^[ \t]*(\/\/|\*)[ \t]?/, "", line)
    else if (e == "h") sub(/^[ \t]*#[ \t]?/, "", line)
    text = text line " "
  }
  END { emit() }
  ' $(cat "$TMP/files.txt") > "$TMP/norm.txt"

  local nf nn np
  nf=$(wc -l < "$TMP/files.txt" | tr -d ' ')
  nn=$(wc -l < "$TMP/norm.txt" | tr -d ' ')
  np=$(wc -l < "$TMP/paths.txt" | tr -d ' ')
  if [[ "$nn" != "$nf" || "$np" != "$nf" ]]; then
    echo "doc-citations: MISCONFIGURED — normalizer emitted $nn lines / $np paths for $nf files." >&2
    echo "  (a mismatch means files were silently skipped; refusing that)" >&2
    exit 1
  fi
  if grep -qE '^[[:space:]]*$' "$TMP/norm.txt"; then
    echo "doc-citations: MISCONFIGURED — normalizer produced an empty line for some file." >&2
    exit 1
  fi
}
