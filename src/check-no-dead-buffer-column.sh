#!/usr/bin/env bash

# PLACEMENT: src/**/buffer_block.go,src/Buffer/layout_version.go,src/Buffer/buffer-layout*.ts | every buffer column needs a non-test production consumer; delete an unused one rather than allowlisting it

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

LAYOUT_FILES=()
for LAYOUT in "src/Buffer/buffer-layout.ts" \
              src/Buffer/buffer-layout-rows*-gen.ts \
              "src/Buffer/buffer-layout-singletons-gen.ts"; do
  [[ -f "$LAYOUT" ]] && LAYOUT_FILES+=("$LAYOUT")
done
source "$REPO_ROOT/scripts/lib/ts-roots.sh"
SRC="src"

if [[ ! -f "src/Buffer/buffer-layout.ts" ]] || (( ${#LAYOUT_FILES[@]} < 2 )); then
  echo "check-no-dead-buffer-column: MISCONFIGURED — found ${#LAYOUT_FILES[@]} layout file(s) under Buffer/ (renamed?); refusing vacuous pass" >&2
  exit 1
fi

readonly ALLOWED_DEAD=(
  COL_STREAM_NODE_NODE_ID
  COL_STREAM_NODE_HAS_KIND_RULE
  COL_STREAM_NODE_LATTICE_POINTS
  COL_STREAM_NODE_TOP_TILT_VECTOR_LEN
  COL_STREAM_EDGE_BEAD_EDGE_ROW
  COL_STREAM_RULES_PANEL_TOGGLE_W
  COL_STREAM_RULES_PANEL_ROW_DEPTH
  COL_STREAM_RULES_PANEL_ROW_VALUE_W
  COL_STREAM_RULES_PANEL_ROW_VALUE_H
  COL_STREAM_RULES_PANEL_MENU_CHECK_W
  COL_STREAM_RULES_PANEL_MENU_CHECK_H
)

is_allowed() {
  local fn="$1"
  [[ ${#ALLOWED_DEAD[@]} -eq 0 ]] && return 1
  for a in "${ALLOWED_DEAD[@]}"; do [[ "$fn" == "$a" ]] && return 0; done
  return 1
}

LAYOUT_EXCLUDES=()
for LAYOUT in "${LAYOUT_FILES[@]}"; do
  LAYOUT_EXCLUDES+=(-not -path "*/${LAYOUT}")
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

COLUMNS_GEN_FILES=()
while IFS= read -r f; do [[ -n "$f" ]] && COLUMNS_GEN_FILES+=("$f"); done < <(git ls-files '*/columns-gen.ts')
if [[ ${#COLUMNS_GEN_FILES[@]} -eq 0 ]]; then
  echo "check-no-dead-buffer-column: MISCONFIGURED — no */columns-gen.ts is tracked; the per-column" >&2
  echo "  constants moved or are no longer generated, so this half of the guard checks nothing." >&2
  exit 1
fi

constants=()
while IFS= read -r line; do [[ -n "$line" ]] && constants+=("$line"); done < <(
  grep -ohE 'export const COL_STREAM_[A-Z0-9_]+' "${COLUMNS_GEN_FILES[@]}" \
    | awk '{print $3}' | grep -v '^COL_STREAM_BASE_' | sort -u
)
if [[ ${#constants[@]} -eq 0 ]]; then
  echo "check-no-dead-buffer-column: MISCONFIGURED — parsed 0 COL_STREAM_* constants from" >&2
  echo "  ${#COLUMNS_GEN_FILES[@]} columns-gen.ts file(s); the generated form changed." >&2
  exit 1
fi

CONST_CORPUS="$(mktemp)"
trap 'rm -f "$CODE_ONLY_CORPUS" "$CONST_CORPUS"' EXIT
const_consumers=()
for f in "${prod_files[@]}"; do
  [[ "$(basename "$f")" == "columns-gen.ts" ]] && continue
  const_consumers+=("$f")
done
if [[ ${#const_consumers[@]} -gt 0 ]]; then
  strip_ts_comments "${const_consumers[@]}" | grep -ohE '[A-Za-z0-9_]+' | sort -u > "$CONST_CORPUS"
else
  : > "$CONST_CORPUS"
fi

for c in "${constants[@]}"; do
  if grep -qxF "$c" "$CONST_CORPUS"; then
    continue
  fi
  if is_allowed "$c"; then
    continue
  fi
  echo "DEAD BUFFER COLUMN: $c is generated but named by no production file — Go packs the column"
  echo "  every frame and nothing reads it. A singleton column has no generated read* helper, so the"
  echo "  half of this guard above cannot see it; this half is what covers it."
  echo "  Fix: consume it, or remove the field from its block's buffer_block.go and regenerate."
  fail=1
done

for a in "${ALLOWED_DEAD[@]+"${ALLOWED_DEAD[@]}"}"; do
  present=false
  corpus="$CODE_ONLY_CORPUS"
  for fn in "${readers[@]}"; do [[ "$fn" == "$a" ]] && present=true && break; done
  if ! $present; then
    for c in "${constants[@]}"; do
      [[ "$c" == "$a" ]] && present=true && corpus="$CONST_CORPUS" && break
    done
  fi
  if ! $present; then
    echo "STALE ALLOWLIST: '$a' is neither a generated read* helper nor a generated COL_STREAM_*"
    echo "  constant — the column is gone; remove it from ALLOWED_DEAD."
    fail=1
    continue
  fi
  if grep -qxF "$a" "$corpus"; then
    echo "STALE ALLOWLIST: '$a' now HAS a production consumer — remove it from ALLOWED_DEAD (no longer dead)."
    fail=1
  fi
done

if [[ $fail -eq 0 && ${#ALLOWED_DEAD[@]} -gt 0 ]]; then
  echo "check-no-dead-buffer-column: clean, but ${#ALLOWED_DEAD[@]} column(s) are packed every frame and read by nothing:"
  printf '  %s\n' "${ALLOWED_DEAD[@]}"
  echo "  They are allowlisted, NOT resolved. Each is plausibly a half-wired feature rather than"
  echo "  dead weight, so deleting would cement the renderer's omission - the RULES_PANEL entries"
  echo "  are the W/H of Rect groups whose X/Y ARE read, so the renderer sizes those rects some"
  echo "  other way. NODE_NODE_ID is the id-vs-row check .claude/rules/bridge-surface.md describes"
  echo "  and decodeNodeStreamFrame never performs. Decide per entry: consume it, or delete the"
  echo "  field from its buffer_block.go and regenerate."
fi

exit $fail
