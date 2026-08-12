#!/usr/bin/env bash

# Sourced by tools/docs/check-doc-citations.sh. Finds every file that mentions
# CLAUDE.md/MODEL.md and filters it to the candidate set worth checking, writing
# $TMP/files.txt (one path per line) and $TMP/filelines.txt (that file's first matching
# line number, same order).

collect_citing_files() {
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
}
