#!/usr/bin/env bash

collect_citing_files() {
  : > "$TMP/mentions.txt"
  git grep -nE '(CLAUDE|MODEL)\.md' -- '*.go' '*.ts' '*.tsx' '*.sh' '*.py' '*.md' '*.html' \
    > "$TMP/mentions.txt" 2>/dev/null || true

  : > "$TMP/files.txt"
  : > "$TMP/filelines.txt"
  while IFS=$'\t' read -r path _line; do
    case "$path" in
      docs/planning/*) continue ;;
      scripts/checks/prose/check-doc-citations.sh) continue ;;
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
