#!/usr/bin/env bash

# Sourced by tools/docs/check-doc-citations.sh. Scans $TMP/norm.txt for a doc name
# immediately followed by a quoted string (the citation shape this whole guard enforces),
# drops ones that read as history (HISTORY_RE), and writes the survivors to
# $TMP/cites.txt as path\tline\tdoc\tquoted.

HISTORY_RE='(gone|removed|retired|deleted|erased|obsolete|legacy|superseded|replaced|no longer|used to|formerly|dead|was |were |old |reverted|rejected|abandoned|does not exist|doesn.t exist|never existed|unbuilt|no such|there is no|do not cite|do not re-add|until )'

extract_citations() {
  : > "$TMP/cites.txt"

  local FIRSTLINE=(0)
  while IFS= read -r _l; do FIRSTLINE+=("$_l"); done < "$TMP/filelines.txt"

  awk '{
    s = $0
    while (match(s, /(CLAUDE|MODEL)\.md'"'"'?s?[ \t]+"[^"]{3,}"/)) {
      st = RSTART; ln = RLENGTH
      pstart = st - 110; if (pstart < 1) pstart = 1
      print FNR ":" substr(s, pstart, st - pstart + ln)
      s = substr(s, st + ln)
    }
  }' "$TMP/norm.txt" > "$TMP/windows.txt" 2>/dev/null || true

  local rec n window path hit approx_line doc quoted
  while IFS= read -r rec; do
    n="${rec%%:*}"; window="${rec#*:}"
    path=$(sed -n "${n}p" "$TMP/paths.txt")

    if printf '%s' "$window" | grep -qiE "$HISTORY_RE"; then continue; fi
    hit=$(printf '%s' "$window" | grep -oE '(CLAUDE|MODEL)\.md'"'"'?s?[[:space:]]+"[^"]{3,}"' | tail -1)
    [[ -n "$hit" ]] || continue
    approx_line=${FIRSTLINE[$n]:-1}
    doc="${hit%%.md*}"
    quoted="${hit#*\"}"; quoted="${quoted%\"}"
    printf '%s\t%s\t%s\t%s\n' "$path" "$approx_line" "$doc" "$quoted" >> "$TMP/cites.txt"
  done < "$TMP/windows.txt"

  if [[ ! -s "$TMP/cites.txt" ]]; then
    echo "doc-citations: EMPTY citation set — the extractor found no CLAUDE.md/MODEL.md citations at all." >&2
    echo "  That is almost certainly a broken regex, not a clean repo; refusing vacuous pass." >&2
    exit 1
  fi
}
