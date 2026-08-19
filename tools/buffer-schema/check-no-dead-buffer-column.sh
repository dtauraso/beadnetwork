#!/usr/bin/env bash

# PLACEMENT: tools/topology-vscode/**/buffer_block.go,tools/topology-vscode/src/Buffer/bufschema/layout*.go,tools/topology-vscode/src/Buffer/buffer-layout*.ts | every buffer column needs a non-test production consumer; delete an unused one rather than allowlisting it

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

LAYOUT_FILES=()
for LAYOUT in "tools/topology-vscode/src/Buffer/buffer-layout.ts" \
              tools/topology-vscode/src/Buffer/buffer-layout-rows*-gen.ts \
              "tools/topology-vscode/src/Buffer/buffer-layout-singletons-gen.ts"; do
  [[ -f "$LAYOUT" ]] && LAYOUT_FILES+=("$LAYOUT")
done
source "$REPO_ROOT/tools/lib/ts-roots.sh"
SRC="tools/topology-vscode/src"

if [[ ! -f "tools/topology-vscode/src/Buffer/buffer-layout.ts" ]] || (( ${#LAYOUT_FILES[@]} < 2 )); then
  echo "check-no-dead-buffer-column: MISCONFIGURED — found ${#LAYOUT_FILES[@]} layout file(s) under Buffer/ (renamed?); refusing vacuous pass" >&2
  exit 1
fi

readonly ALLOWED_DEAD=()

is_allowed() {
  local fn="$1"
  [[ ${#ALLOWED_DEAD[@]} -eq 0 ]] && return 1
  for a in "${ALLOWED_DEAD[@]}"; do [[ "$fn" == "$a" ]] && return 0; done
  return 1
}

LAYOUT_EXCLUDES=()
for LAYOUT in "${LAYOUT_FILES[@]}"; do
  LAYOUT_EXCLUDES+=(-not -path "*/${LAYOUT#tools/topology-vscode/}")
done

readers=()
while IFS= read -r line; do
  [[ -n "$line" ]] && readers+=("$line")
done < <(grep -ohE 'export function (read[A-Za-z0-9_]+)' "${LAYOUT_FILES[@]}" | awk '{print $3}' | sort -u)

if [[ ${#readers[@]} -eq 0 ]]; then
  echo "check-no-dead-buffer-column: MISCONFIGURED — parsed 0 read* helpers from ${LAYOUT_FILES[*]}; format changed, guard would check nothing" >&2
  exit 1
fi

prod_files=()
while IFS= read -r f; do prod_files+=("$f"); done < <(
  find "${TS_ROOTS[@]}" -type f \( -name '*.ts' -o -name '*.tsx' \) \
    "${LAYOUT_EXCLUDES[@]}" \
    -not -path '*/test/*' \
    -not -name '*.test.ts' 2>/dev/null
)

strip_ts_comments() {
  perl -0777pe 's{/\*.*?\*/}{}gs; s{//[^\n]*}{}g' "$@" 2>/dev/null
}

CODE_ONLY_CORPUS="$(mktemp)"
trap 'rm -f "$CODE_ONLY_CORPUS"' EXIT
if [[ ${#prod_files[@]} -gt 0 ]]; then
  strip_ts_comments "${prod_files[@]}" | grep -ohE '[A-Za-z0-9_]+' | sort -u > "$CODE_ONLY_CORPUS"
else
  : > "$CODE_ONLY_CORPUS"
fi

fail=0
for fn in "${readers[@]}"; do
  if grep -qxF "$fn" "$CODE_ONLY_CORPUS"; then
    continue
  fi
  if is_allowed "$fn"; then
    continue
  fi
  echo "DEAD BUFFER COLUMN: $fn has no production consumer — the column is packed + decoded but used by nothing."
  echo "  Fix: consume it, remove the column from its block (Buffer/bufschema/ for a model block, the concern's own buffer_block.go otherwise) and regenerate, or (if intentionally staged) add it to ALLOWED_DEAD with a reason."
  fail=1
done

for a in "${ALLOWED_DEAD[@]+"${ALLOWED_DEAD[@]}"}"; do
  present=false
  for fn in "${readers[@]}"; do [[ "$fn" == "$a" ]] && present=true && break; done
  if ! $present; then
    echo "STALE ALLOWLIST: '$a' is no longer a generated read* helper — remove it from ALLOWED_DEAD."
    fail=1
    continue
  fi
  if grep -qxF "$a" "$CODE_ONLY_CORPUS"; then
    echo "STALE ALLOWLIST: '$a' now HAS a production consumer — remove it from ALLOWED_DEAD (no longer dead)."
    fail=1
  fi
done

exit $fail
