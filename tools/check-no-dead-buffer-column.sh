#!/usr/bin/env bash
#
# PLACEMENT: Buffer/layout.go | every buffer column needs a non-test production consumer; delete an unused one rather than allowlisting it
# check-no-dead-buffer-column.sh — fail if a generated buffer-column reader has NO
# production consumer. Run from repo root: bash tools/check-no-dead-buffer-column.sh
#
# WHY THIS EXISTS (audit gap A, lockstep-codec cluster): a column added to Buffer/layout.go
# gets a generated read<Block><Name>() helper in schema/buffer-layout.ts automatically, and
# check-generated.sh + check-buffer-layout-parity.sh both stay GREEN for it — a generated
# reader existing says nothing about whether any production src/ code CONSUMES the column.
# So a column can be packed on the Go side and decoded on the TS side yet used by nothing:
# dead wire surface that every other guard passes. This guard closes that: every generated
# read* helper must be referenced from non-test production src/.
#
# Ported pattern: the locked-allowlist / grep|exit1 shape of the Uncle-Bob raid exemplar
# check-ai-map-access.sh — known-dead columns are listed explicitly and must not grow.
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

LAYOUT="tools/topology-vscode/src/schema/buffer-layout.ts"
SRC="tools/topology-vscode/src"

if [[ ! -f "$LAYOUT" ]]; then
  echo "check-no-dead-buffer-column: MISCONFIGURED — $LAYOUT not found (renamed?); refusing vacuous pass" >&2
  exit 1
fi

# Known-dead columns permitted TODAY. Each is dead surface pending a decision to remove the
# column from Buffer/layout.go (a MODEL.md-gated codec-spine edit, not a guard-commit change).
# This list must never GROW without that decision — a new dead column is a bug, not an entry.
readonly ALLOWED_DEAD=(
)

is_allowed() {
  local fn="$1"
  [[ ${#ALLOWED_DEAD[@]} -eq 0 ]] && return 1   # empty-array-safe under set -u (bash 3.2)
  for a in "${ALLOWED_DEAD[@]}"; do [[ "$fn" == "$a" ]] && return 0; done
  return 1
}

# Every generated column reader. (Portable read — macOS bash 3.2 has no mapfile.)
readers=()
while IFS= read -r line; do
  [[ -n "$line" ]] && readers+=("$line")
done < <(grep -oE 'export function (read[A-Za-z0-9_]+)' "$LAYOUT" | awk '{print $3}' | sort -u)

if [[ ${#readers[@]} -eq 0 ]]; then
  echo "check-no-dead-buffer-column: MISCONFIGURED — parsed 0 read* helpers from $LAYOUT; format changed, guard would check nothing" >&2
  exit 1
fi

# Build the production-consumer reference corpus ONCE instead of one full recursive `grep
# -r` of src/ PER reader (was 97 readers x a whole-tree walk = quadratic). File exclusions
# (the generated definition file, /test/, *.test.ts) are applied here, at corpus-build time,
# not filtered out of results afterward — same effect, same files excluded.
#
# The corpus is a set of WHOLE-WORD tokens ([A-Za-z0-9_]+ runs), not raw file text: `\b$fn\b`
# on file contents only matches $fn as a complete identifier, never as a substring of a
# longer one. Extracting maximal identifier-character runs and doing exact-set membership on
# THAT reproduces exactly that word-boundary semantics — a naive `grep -F` of the corpus text
# would let a reader whose name is a substring of a longer identifier false-match and hide a
# genuinely dead column, which is the bug this guard exists to catch.
prod_files=()
while IFS= read -r f; do prod_files+=("$f"); done < <(
  find "$SRC" -type f \( -name '*.ts' -o -name '*.tsx' \) \
    -not -path '*/schema/buffer-layout.ts' \
    -not -path '*/test/*' \
    -not -name '*.test.ts' 2>/dev/null
)

CORPUS="$(mktemp)"
trap 'rm -f "$CORPUS"' EXIT
if [[ ${#prod_files[@]} -gt 0 ]]; then
  grep -ohE '[A-Za-z0-9_]+' "${prod_files[@]}" 2>/dev/null | sort -u > "$CORPUS"
else
  : > "$CORPUS"
fi

fail=0
for fn in "${readers[@]}"; do
  if grep -qxF "$fn" "$CORPUS"; then
    continue
  fi
  if is_allowed "$fn"; then
    continue
  fi
  echo "DEAD BUFFER COLUMN: $fn has no production consumer — the column is packed + decoded but used by nothing."
  echo "  Fix: consume it, remove the column from Buffer/layout.go (regenerate), or (if intentionally staged) add it to ALLOWED_DEAD with a reason."
  fail=1
done

# Guard the allowlist against rot: an entry that is no longer dead (now consumed, or the
# reader was renamed away) should be removed so the list stays honest.
for a in "${ALLOWED_DEAD[@]+"${ALLOWED_DEAD[@]}"}"; do   # empty-array-safe under set -u
  present=false
  for fn in "${readers[@]}"; do [[ "$fn" == "$a" ]] && present=true && break; done
  if ! $present; then
    echo "STALE ALLOWLIST: '$a' is no longer a generated read* helper — remove it from ALLOWED_DEAD."
    fail=1
    continue
  fi
  if grep -qxF "$a" "$CORPUS"; then
    echo "STALE ALLOWLIST: '$a' now HAS a production consumer — remove it from ALLOWED_DEAD (no longer dead)."
    fail=1
  fi
done

exit $fail
