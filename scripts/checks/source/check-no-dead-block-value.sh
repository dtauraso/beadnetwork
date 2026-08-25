#!/usr/bin/env bash

# PLACEMENT: **/*_values.go | every declared block value needs a non-test production consumer; delete an unused one rather than allowlisting it. The buffer-column half of this guard went with the trace events: the buffer layout declares no columns any more, so there are no read* helpers left to find a consumer for.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

source "$REPO_ROOT/scripts/lib/ts-roots.sh"
SRC="."

readonly ALLOWED_DEAD=()

is_allowed() {
  local fn="$1"
  [[ ${#ALLOWED_DEAD[@]} -eq 0 ]] && return 1
  for a in "${ALLOWED_DEAD[@]}"; do [[ "$fn" == "$a" ]] && return 0; done
  return 1
}

prod_files=()
while IFS= read -r f; do prod_files+=("$f"); done < <(
  find "${TS_ROOTS[@]}" -type f \( -name '*.ts' -o -name '*.tsx' \) \
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

VALUES_GEN_FILES=()
while IFS= read -r f; do [[ -n "$f" ]] && VALUES_GEN_FILES+=("$f"); done < <(git ls-files '*-values-gen.ts')
if (( ${#VALUES_GEN_FILES[@]} < 5 )); then
  echo "check-no-dead-block-value: MISCONFIGURED — found only ${#VALUES_GEN_FILES[@]} *-values-gen.ts" >&2
  echo "  file(s); the generated value lists moved or are no longer generated, so this half of the" >&2
  echo "  guard checks almost nothing." >&2
  exit 1
fi

readonly DERIVED_FAMILIES=(
  '^ringM[0-9]+$:RING_NAMES'
  '^bodyM[0-9]+$:BODY_NAMES'
  '^tiltShaftM[0-9]+$:TILT_SHAFT_NAMES'
  '^tiltHeadM[0-9]+$:TILT_HEAD_NAMES'
  '^channelShaftM[0-9]+$:VECTOR_SHAFT_NAMES'
  '^channelHeadM[0-9]+$:VECTOR_HEAD_NAMES'
)

VALUE_CORPUS="$(mktemp)"
trap 'rm -f "$CODE_ONLY_CORPUS" "$VALUE_CORPUS"' EXIT
value_consumers=()
for f in "${prod_files[@]}"; do
  case "$(basename "$f")" in *-values-gen.ts) continue ;; esac
  value_consumers+=("$f")
done
if [[ ${#value_consumers[@]} -gt 0 ]]; then
  strip_ts_comments "${value_consumers[@]}" \
    | grep -ohE '"[A-Za-z0-9_]+"|'"'"'[A-Za-z0-9_]+'"'"'|[A-Z0-9_]*NAMES' \
    | tr -d "\"'" | sort -u > "$VALUE_CORPUS"
else
  : > "$VALUE_CORPUS"
fi

names=()
while IFS= read -r line; do [[ -n "$line" ]] && names+=("$line"); done < <(
  grep -ohE '^  "[A-Za-z0-9_]+",$' "${VALUES_GEN_FILES[@]}" | tr -d ' ",' | sort -u
)
if (( ${#names[@]} < 50 )); then
  echo "check-no-dead-block-value: MISCONFIGURED — parsed only ${#names[@]} value names from" >&2
  echo "  ${#VALUES_GEN_FILES[@]} *-values-gen.ts file(s); the generated form changed." >&2
  exit 1
fi

covered_by_family() {
  local name="$1" entry pattern const
  for entry in "${DERIVED_FAMILIES[@]}"; do
    pattern="${entry%%:*}"
    const="${entry##*:}"
    if [[ "$name" =~ $pattern ]] && grep -q "$const" "$VALUE_CORPUS"; then
      return 0
    fi
  done
  return 1
}

for n in "${names[@]}"; do
  if grep -qxF "$n" "$VALUE_CORPUS"; then
    continue
  fi
  if covered_by_family "$n"; then
    continue
  fi
  if is_allowed "$n"; then
    continue
  fi
  echo "DEAD BLOCK VALUE: \"$n\" is declared in a *-values-gen.ts list but no production file names"
  echo "  it. Go writes that section of the block file every tick and nothing reads it."
  echo "  Fix: read it, or drop it from the Go *ValueNames list and regenerate."
  fail=1
done

for a in "${ALLOWED_DEAD[@]+"${ALLOWED_DEAD[@]}"}"; do
  present=false
  corpus="$CODE_ONLY_CORPUS"
  for fn in "${readers[@]}"; do [[ "$fn" == "$a" ]] && present=true && break; done
  if ! $present; then
    for c in "${names[@]}"; do
      [[ "$c" == "$a" ]] && present=true && corpus="$VALUE_CORPUS" && break
    done
  fi
  if ! $present; then
    echo "STALE ALLOWLIST: '$a' is neither a generated read* helper nor a declared block value"
    echo "  — the thing is gone; remove it from ALLOWED_DEAD."
    fail=1
    continue
  fi
  if grep -qxF "$a" "$corpus"; then
    echo "STALE ALLOWLIST: '$a' now HAS a production consumer — remove it from ALLOWED_DEAD (no longer dead)."
    fail=1
  fi
done

if [[ $fail -eq 0 && ${#ALLOWED_DEAD[@]} -gt 0 ]]; then
  echo "check-no-dead-block-value: clean, but ${#ALLOWED_DEAD[@]} column(s) are packed every frame and read by nothing:"
  printf '  %s\n' "${ALLOWED_DEAD[@]}"
  echo "  Allowlisted, NOT resolved: consume it, or delete the field from its buffer_block.go."
fi

exit $fail
