#!/usr/bin/env bash
set -euo pipefail

# check-doc-citations.sh — a quoted citation of CLAUDE.md / MODEL.md must be QUOTING it.
#
# WHY THIS EXISTS
# ---------------
# The delegate-reminder hook told the model, on every audit-shaped prompt:
#
#     Per CLAUDE.md "Model routing", delegate multi-step lookups ... to a general-purpose
#     subagent with model "sonnet", rather than grinding inline on Opus.
#
# There is no "Model routing" section. There was: 66268d97 (2026-05-05) added the Delegation
# doctrine, 24de543c (2026-05-13) wrote the hook citing it — correct at the time — and
# c123b83e (2026-06-16) REMOVED the doctrine and softened the hook threshold 1->8. The hook
# kept citing the deleted section for a month, so a retired rule was being re-imposed as
# live doctrine, in the imperative, with a citation that made it look authoritative. It also
# contradicted memory/feedback_no_nested_agents (implementer, not general-purpose).
#
# That is the WaitTick bug class aimed at the agent's instructions instead of its code. A
# citation is a claim about another file, and it was checkable all along.
#
# THE RULE
# --------
# If you write CLAUDE.md "X" or MODEL.md "X" (or CLAUDE.md's "X"), then X must appear as
# literal text in that file. Matching is case-insensitive and whitespace-normalized, so
# "no blow-up, by construction" legitimately cites "**No blow-up, by construction.**".
#
# This also enforces memory/feedback_dont_invent_doctrine — "don't paraphrase a one-off note
# into a rule and cite the paraphrase as project doctrine; grep for the literal phrasing
# first". A paraphrase cannot pass: quote it or don't cite it.
#
# SCOPE: tracked *.go, *.ts, *.tsx, *.sh, *.py, *.md — excluding docs/planning/** and HTML
# docs marked <meta name="doc-status" content="historical">, which are dated snapshots whose
# citations are pinned to their moment (same exemption as check-doc-symbols.sh).
#
# PLACEMENT: none | universal prose hygiene: any file quoting CLAUDE.md/MODEL.md must quote it verbatim
#
# Exit 0 if clean; exit 1 with a report otherwise.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

for doc in CLAUDE.md MODEL.md; do
  if [[ ! -f "$doc" ]]; then
    echo "doc-citations: MISCONFIGURED — $doc not found (renamed? a missing doc would vacuously pass)" >&2
    exit 1
  fi
done

# A citation discussing its own history ("this hook CITED X UNTIL c123b83e REMOVED it") is
# exempt. Mirrors check-doc-symbols.sh: the repo documents what it retired on purpose, and a
# guard that punished that would delete the most useful prose in the file.
HISTORY_RE='(gone|removed|retired|deleted|erased|obsolete|legacy|superseded|replaced|no longer|used to|formerly|dead|was |were |old |reverted|rejected|abandoned|does not exist|doesn.t exist|never existed|unbuilt|no such|there is no|do not cite|do not re-add|until )'

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

# Normalize a doc to one lowercase whitespace-collapsed blob so a citation spanning a line
# break, or sitting inside **bold**/`code` markup, still matches its source.
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

# Collect citations: <file>\t<line>\t<DOC>\t<quoted text>
#
# Files are NORMALIZED before extraction, because a raw per-line grep misses the exact case
# this guard exists for. The hook's citation looked like this in the source:
#
#     "Heads up: ... Per CLAUDE.md "
#     "\"Model routing\", delegate multi-step lookups ... "
#
# Split across concatenated string literals AND backslash-escaped, so the text
# `CLAUDE.md "Model routing"` never appears on any single line. A line-oriented regex reports
# CLEAN on the very lie that motivated the check. Verified: it did.
#
# So: join lines, drop Python/JS implicit-concat seams (`" "` between adjacent literals),
# then unescape `\"`. Line numbers come from locating the doc mention afterwards — an
# approximate line beats a precise miss.
: > "$TMP/cites.txt"

# Pre-filter once for the whole tree instead of spawning `grep -qE`/`grep -nE|head|cut`
# per candidate file: `git grep -nE` already gives us, in one process, both "does this file
# mention the docs at all" AND "what's the first matching line" for every file that does.
# The per-file exclusions below are still applied to this pre-computed list, not dropped.
: > "$TMP/mentions.txt"
git grep -nE '(CLAUDE|MODEL)\.md' -- '*.go' '*.ts' '*.tsx' '*.sh' '*.py' '*.md' '*.html' \
  > "$TMP/mentions.txt" 2>/dev/null || true

# ONE pass, not five spawns per file. This loop used to run, per candidate file:
#   awk (line lookup) + sed (strip) + tr (join) + sed (normalize) + grep -oE (extract)
# = ~5 processes x 95 files ~= 2.0s, which was 44% of the whole guard suite. The work is
# identical for every file, so it is done once for all of them instead:
#
#   1. one awk normalizes EVERY candidate file to a single line (strip leading comment
#      markers, join, merge adjacent quotes, unescape) -> norm.txt, paths.txt in parallel
#   2. one `grep -noE` extracts every citation window from norm.txt; because each file is
#      exactly one line, the match's LINE NUMBER is its file (paths.txt line N)
#
# Only the per-WINDOW work stays per-item, and there are ~26 windows, not 95 files.
#
# Verified equivalent, not assumed: the citation set extracted here was diffed against the
# per-file implementation's and is byte-identical (24/24). Measured 2.04s -> 0.07s for the
# normalize step.
: > "$TMP/files.txt"
: > "$TMP/filelines.txt"
while IFS=$'\t' read -r path _line; do
  case "$path" in
    docs/planning/*) continue ;;
    tools/check-doc-citations.sh) continue ;;  # quotes the example that motivated it
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

# Per-extension comment stripping happens INSIDE awk (same expressions as the old per-file
# `sed -E`), keyed off FILENAME.
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

# The old code asserted per file that the normalizer produced text, because a broken strip
# expression otherwise degrades to "this file has no citations" (that exact bug shipped
# once). Same assertion, now on the batch: one normalized line per input file, none empty.
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

# FIRSTLINE[N] = first-mention line of the file on paths.txt line N. Built once; the loop
# below indexes it instead of running `awk` over mentions.txt per window.
FIRSTLINE=(0)
while IFS= read -r _l; do FIRSTLINE+=("$_l"); done < "$TMP/filelines.txt"

# Extraction is awk, NOT `grep -oE`, and the context prefix is a substr rather than a
# regex. MEASURED on this repo's normalized corpus (87 lines, one 47k-char line):
#
#   /usr/bin/grep -noE '.{0,110}<citation>'   1.656s   <- 83% of the whole guard
#   awk, same greedy .{0,110} prefix          0.854s
#   awk, citation match + substr context      0.019s   <- 87x faster than grep
#
# The cost was the GREEDY BOUNDED PREFIX: for every position on a 47k-char line, BSD grep
# tries up to 110 leading characters. Matching the citation first (anchored on the literal
# "CLAUDE.md"/"MODEL.md") and then taking the 110 preceding characters by offset does the
# same job with no backtracking at all.
#
# Beware benchmarking this from an interactive shell: `grep` there may be a shell FUNCTION
# rather than /usr/bin/grep, which is why an earlier round of profiling kept measuring
# 0.08s for a call that costs 1.65s inside the script. Compare /usr/bin/grep explicitly.
#
# Output format is unchanged: "<line-number>:<window>", line number = file (paths.txt).
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
  # A citation discussing its own history is exempt — this hook's docstring explains
  # that it CITED "Model routing" until that section was removed. Same exemption as
  # check-doc-symbols.sh: deliberate history is the valuable prose, not the bug.
  # MEASURED, not assumed: doing these two matches with bash's own `=~` under `nocasematch`
  # cost ~60ms per window (~1.6s over 26) — bash's regex engine is far slower here than
  # grep's, which does the same work in ~10ms total for the price of a spawn. The spawn is
  # the cheap half. Only the first-mention lookup stays in-process (FIRSTLINE), because
  # that one was an `awk` re-reading a file per window.
  if printf '%s' "$window" | grep -qiE "$HISTORY_RE"; then continue; fi
  hit=$(printf '%s' "$window" | grep -oE '(CLAUDE|MODEL)\.md'"'"'?s?[[:space:]]+"[^"]{3,}"' | tail -1)
  [[ -n "$hit" ]] || continue
  approx_line=${FIRSTLINE[$n]:-1}
  doc="${hit%%.md*}"
  quoted="${hit#*\"}"; quoted="${quoted%\"}"
  printf '%s\t%s\t%s\t%s\n' "$path" "$approx_line" "$doc" "$quoted" >> "$TMP/cites.txt"
done < "$TMP/windows.txt"

if [[ ! -s "$TMP/cites.txt" ]]; then
  # Zero citations repo-wide is implausible given CLAUDE.md/MODEL.md are the doctrine docs;
  # treat it as a broken extractor rather than a clean tree.
  echo "doc-citations: EMPTY citation set — the extractor found no CLAUDE.md/MODEL.md citations at all." >&2
  echo "  That is almost certainly a broken regex, not a clean repo; refusing vacuous pass." >&2
  exit 1
fi

# The blobs are read ONCE into memory and matched with bash's own pattern matching, rather
# than `tr | tr | sed | grep -qF` per citation (4 processes x 24 citations). `nocasematch`
# replaces the lowercasing `tr` entirely — the comparison is case-insensitive instead of
# the operands being lowered.
HITS=0
while IFS=$'\t' read -r path line doc quoted; do
  case "$doc" in
    CLAUDE) blob="$TMP/claude.txt" ;;
    MODEL)  blob="$TMP/model.txt" ;;
    *) continue ;;
  esac
  # Same normalization the pipeline did: squeeze whitespace runs, drop * and ` (markdown
  # emphasis in a citation should not change what it matches). Case is handled by
  # nocasematch above.
  needle=${quoted//[*\`]/}
  needle=${needle//$'\t'/ }
  needle=${needle//$'\n'/ }
  while [[ $needle == *"  "* ]]; do needle=${needle//  / }; done
  # grep -qiF, not a bash `==` against the blob in memory: with `nocasematch` on, bash's
  # own substring match against a ~50KB string costs ~60ms per citation (~1.5s over 24),
  # while grep does it in C for the price of one spawn. -i replaces the lowercasing `tr`.
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
