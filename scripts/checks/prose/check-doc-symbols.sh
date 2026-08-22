#!/usr/bin/env bash

# PLACEMENT: none | universal prose hygiene: every file's comments are checked, no placement decision
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

ALLOWLIST="$SCRIPT_DIR/../doc-symbols-allow.txt"

IDENT_RE='([A-Z][a-z0-9]+([A-Z][a-z0-9]*)+|[a-z][a-z0-9]*([A-Z][a-z0-9]*)+)'

HISTORY_RE='(gone|removed|retired|deleted|erased|obsolete|legacy|superseded|replaced|no longer|used to|formerly|gutted|gone away|gone now|dead|gone\.|was |were |gone,|old |gone;|pre-|reverted|rejected|abandoned|do not re-|don.t re-|never re-|does not exist|doesn.t exist|never existed|unbuilt|no such|there is no|do not cite)'

DOC_HISTORICAL_RE='<meta[[:space:]]+name="doc-status"[[:space:]]+content="historical"'

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

git ls-files -z '*.go' '*.ts' '*.tsx' '*.js' '*.sh' 2>/dev/null \
  | tr '\0' '\n' | grep -v '/check-\|^check-' | tr '\n' '\0' \
  | xargs -0 -r sed -e 's|//.*||' 2>/dev/null \
  | grep -oE "$IDENT_RE" \
  | sort -u > "$TMP/code.txt" || true

git ls-files -z '*.go' '*.ts' '*.tsx' 2>/dev/null \
  | xargs -0 -r grep -hE '^[[:space:]]*(//|\*)' 2>/dev/null \
  | grep -viE "$HISTORY_RE" \
  | grep -oE '`[^`]+`' \
  | grep -oE "$IDENT_RE" \
  | sort -u > "$TMP/from_go.txt" || true

DOCS=(CLAUDE.md MODEL.md)
while IFS= read -r spec; do DOCS+=("$spec"); done < <(git ls-files 'Categories/NodeKinds/*/SPEC.md' 2>/dev/null || true)

: > "$TMP/from_md.txt"
for d in "${DOCS[@]}"; do
  [[ -f "$d" ]] || continue

  grep -viE "$HISTORY_RE" "$d" 2>/dev/null \
    | grep -oE '`[^`]+`' \
    | grep -oE "$IDENT_RE" >> "$TMP/from_md.txt" || true
done
sort -u -o "$TMP/from_md.txt" "$TMP/from_md.txt"

: > "$TMP/from_html.txt"
while IFS= read -r page; do
  [[ -f "$page" ]] || continue
  case "$page" in docs/planning/*) continue ;; esac
  if grep -qiE "$DOC_HISTORICAL_RE" "$page" 2>/dev/null; then continue; fi
  grep -viE "$HISTORY_RE" "$page" 2>/dev/null \
    | grep -oE '<code[^>]*>[^<]*</code>' \
    | sed -e 's/<[^>]*>//g' \
    | grep -oE "$IDENT_RE" >> "$TMP/from_html.txt" || true
done < <(git ls-files 'docs/**/*.html' 'docs/**/*.html' 2>/dev/null || true)
sort -u -o "$TMP/from_html.txt" "$TMP/from_html.txt"

cat "$TMP/from_go.txt" "$TMP/from_md.txt" "$TMP/from_html.txt" | sort -u > "$TMP/candidates.txt"

for pair in "code.txt:code-identifier corpus" "candidates.txt:comment/doc candidates"; do
  f="${pair%%:*}"; label="${pair#*:}"
  if [[ ! -s "$TMP/$f" ]]; then
    echo "doc-symbols: EMPTY extracted set for '$label' — extractor broken; refusing vacuous pass" >&2
    exit 1
  fi
done

: > "$TMP/allow.txt"
if [[ -f "$ALLOWLIST" ]]; then
  sed -e 's/#.*//' -e 's/[[:space:]]//g' "$ALLOWLIST" \
    | grep -vE '^$' \
    | sort -u > "$TMP/allow.txt" || true
fi

comm -23 "$TMP/candidates.txt" "$TMP/code.txt" > "$TMP/ghosts_raw.txt"
comm -23 "$TMP/ghosts_raw.txt" "$TMP/allow.txt" > "$TMP/ghosts.txt"

if [[ ! -s "$TMP/ghosts.txt" ]]; then
  echo "doc-symbols: clean"
  exit 0
fi

echo "doc-symbols: a BACKTICKED symbol exists nowhere in code:"
echo ""
while IFS= read -r sym; do
  echo "  ghost: \`$sym\`"

  git grep -nE "(\`[^\`]*${sym}[^\`]*\`|<code[^>]*>[^<]*${sym}[^<]*</code>)" \
    -- '*.go' '*.ts' '*.tsx' CLAUDE.md MODEL.md 'Categories/NodeKinds/*/SPEC.md' 'docs/**/*.html' 'docs/**/*.html' 2>/dev/null \
    | head -3 | sed 's/^/      /' || true
done < "$TMP/ghosts.txt"

COUNT=$(grep -c . "$TMP/ghosts.txt")
echo ""
echo "doc-symbols: $COUNT ghost symbol(s) found"
echo ""
echo "  Each is a comment pointing a reader at something that does not exist."
echo "  Fix by DELETING the stale claim, not by refreshing it — an unenforceable claim"
echo "  will just drift again. If the token is prose and not a symbol (a library name, a"
echo "  product), add it to scripts/checks/prose/doc-symbols-allow.txt."
exit 1
