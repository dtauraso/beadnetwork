#!/usr/bin/env bash
set -euo pipefail


































# PLACEMENT: none | universal prose hygiene: any file quoting CLAUDE.md/MODEL.md must quote it verbatim



SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

for doc in CLAUDE.md MODEL.md; do
  if [[ ! -f "$doc" ]]; then
    echo "doc-citations: MISCONFIGURED — $doc not found (renamed? a missing doc would vacuously pass)" >&2
    exit 1
  fi
done




HISTORY_RE='(gone|removed|retired|deleted|erased|obsolete|legacy|superseded|replaced|no longer|used to|formerly|dead|was |were |old |reverted|rejected|abandoned|does not exist|doesn.t exist|never existed|unbuilt|no such|there is no|do not cite|do not re-add|until )'

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT



normalize() {
  tr '\n' ' ' < "$1" | tr -s '[:space:]' ' ' | tr '[:upper:]' '[:lower:]' | sed 's/[*`]//g'
}
normalize CLAUDE.md > "$TMP/claude.txt"
normalize MODEL.md  > "$TMP/model.txt"

for f in claude model; do
  if [[ ! -s "$TMP/$f.txt" ]]; then
    echo "doc-citations: EMPTY normalized text for $f — extractor broken; refusing vacuous pass" >&2
    exit 1
  fi
done
















: > "$TMP/cites.txt"





: > "$TMP/mentions.txt"
git grep -nE '(CLAUDE|MODEL)\.md' -- '*.go' '*.ts' '*.tsx' '*.sh' '*.py' '*.md' '*.html' \
  > "$TMP/mentions.txt" 2>/dev/null || true
















: > "$TMP/files.txt"
: > "$TMP/filelines.txt"
while IFS=$'\t' read -r path _line; do
  case "$path" in
    docs/planning/*) continue ;;
    tools/docs/check-doc-citations.sh) continue ;;
  esac
  if [[ "$path" == *.html ]] && grep -qiE '<meta[[:space:]]+name="doc-status"[[:space:]]+content="historical"' "$path" 2>/dev/null; then
    continue
  fi
  printf '%s\n' "$path" >> "$TMP/files.txt"
  printf '%s\n' "$_line" >> "$TMP/filelines.txt"
done < <(awk -F: '!seen[$1]++{print $1 "\t" $2}' "$TMP/mentions.txt" | sort || true)

if [[ ! -s "$TMP/files.txt" ]]; then
  echo "doc-citations: MISCONFIGURED — no candidate files survived filtering." >&2
  exit 1
fi



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



FIRSTLINE=(0)
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





HITS=0
while IFS=$'\t' read -r path line doc quoted; do
  case "$doc" in
    CLAUDE) blob="$TMP/claude.txt" ;;
    MODEL)  blob="$TMP/model.txt" ;;
    *) continue ;;
  esac



  needle=${quoted//[*\`]/}
  needle=${needle//$'\t'/ }
  needle=${needle//$'\n'/ }
  while [[ $needle == *"  "* ]]; do needle=${needle//  / }; done



  if ! grep -qiF -- "$needle" "$blob"; then
    if [[ $HITS -eq 0 ]]; then
      echo "doc-citations: a citation quotes text that is NOT in the doc it cites:"
      echo ""
    fi
    echo "  $path:$line"
    echo "      cites $doc.md \"$quoted\""
    echo "      but that text does not appear in $doc.md"
    HITS=$((HITS + 1))
  fi
done < "$TMP/cites.txt"

if [[ $HITS -eq 0 ]]; then
  echo "doc-citations: clean ($(wc -l < "$TMP/cites.txt" | tr -d ' ') citations checked)"
  exit 0
fi

echo ""
echo "doc-citations: $HITS bad citation(s)"
echo ""
echo "  Either quote the doc verbatim, or drop the citation and state the point directly."
echo "  A paraphrase presented as a citation is how a retired rule gets re-imposed as live"
echo "  doctrine (see this script's header). If the doc changed, the citer must change too."
exit 1
