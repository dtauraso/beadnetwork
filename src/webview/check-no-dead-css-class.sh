#!/usr/bin/env bash

# PLACEMENT: src/webview/*.css | every CSS class needs a renderer; delete an unused one rather than allowlisting it

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

source "$REPO_ROOT/scripts/lib/ts-roots.sh"
SRC="src"
CSS_DIR="$SRC/webview"

if [[ ! -d "$CSS_DIR" ]]; then
  echo "check-no-dead-css-class: MISCONFIGURED — $CSS_DIR not found (moved?); refusing vacuous pass" >&2
  exit 1
fi

shopt -s nullglob
CSS_FILES=("$CSS_DIR"/*.css)
shopt -u nullglob

if [[ ${#CSS_FILES[@]} -eq 0 ]]; then
  STYLED=$(grep -rlE "(className|class)=" --include="*.ts" --include="*.tsx" "${TS_ROOTS[@]}" 2>/dev/null || true)
  if [[ -n "$STYLED" ]]; then
    echo "STYLED WITH NO STYLESHEET: no CSS under $CSS_DIR, but these still carry class attributes," >&2
    echo "so their styles come from somewhere this guard cannot read:" >&2
    echo "$STYLED" | sed 's/^/  /' >&2
    exit 1
  fi
  exit 0
fi

readonly UNDECIDABLE=(clean toolbar app)

is_undecidable() {
  local c="$1"
  for u in "${UNDECIDABLE[@]}"; do [[ "$c" == "$u" ]] && return 0; done
  return 1
}

class_is_rendered() {
  local c="$1"
  grep -rqE "(className|class)=[^>]*\b${c}\b" --include="*.ts" --include="*.tsx" "${TS_ROOTS[@]}" 2>/dev/null && return 0
  grep -rqE "[\"'\`][^\"'\`]*\b${c}\b[^\"'\`]*[\"'\`]" --include="*.ts" --include="*.tsx" "${TS_ROOTS[@]}" 2>/dev/null && return 0
  return 1
}

dead_found=0
scanned=0

for css in "${CSS_FILES[@]}"; do
  classes=$(grep -oE "\.[a-zA-Z][a-zA-Z0-9_-]*" "$css" | sed 's/^\.//' | sort -u)
  file_dead=()
  for c in $classes; do
    is_undecidable "$c" && continue
    scanned=$((scanned + 1))
    class_is_rendered "$c" || file_dead+=("$c")
  done
  if [[ ${#file_dead[@]} -gt 0 ]]; then
    dead_found=1
    echo "DEAD CSS CLASS: no TS/TSX renders these, so the rule styles nothing:"
    for c in "${file_dead[@]}"; do echo "  ${css#"$REPO_ROOT"/}: .$c"; done
  fi
done

if [[ "$dead_found" -eq 1 ]]; then
  echo "Delete the rule. If the class is applied in a way this cannot see, say how in the"
  echo "commit and add it to UNDECIDABLE with that reason — do not leave it looking live."
  exit 1
fi

echo "check-no-dead-css-class: clean ($scanned class(es) scanned across ${#CSS_FILES[@]} stylesheet(s); ${#UNDECIDABLE[@]} undecidable name(s) skipped)"
