#!/usr/bin/env bash

# PLACEMENT: src/**/buffer_block.go,src/schema/buffer-layout/layout_version.go,src/Node/Wiring/inputcodec/input_fingerprint.go,src/Node/*/SPEC.md | changing a generator source means running `go generate ./...` in the SAME commit
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)" || {
  echo "check-generated: MISCONFIGURED — cannot locate repo root." >&2
  exit 1
}
cd "$REPO_ROOT"

GEN_OUT=$(npm run --silent gen:node-defs 2>&1) || {
  echo "check-generated: generator failed" >&2
  echo "$GEN_OUT" >&2
  exit 1
}

FILES=$(printf '%s\n' "$GEN_OUT" \
  | sed -nE 's|^[A-Za-z][A-Za-z0-9/]*: wrote ([^ ]+).*$|\1|p' \
  | sed "s|^$REPO_ROOT/||" \
  | sort -u)

if [[ -z "$FILES" ]]; then
  echo "check-generated: MISCONFIGURED — parsed 0 generated files from the generator output." >&2
  echo "  Its 'wrote <path>' format likely changed; this guard would silently check nothing." >&2
  echo "  Generator said:" >&2
  printf '%s\n' "$GEN_OUT" | sed 's/^/    /' >&2
  exit 1
fi

UNTRACKED=0
while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  if ! git ls-files --error-unmatch "$f" >/dev/null 2>&1; then
    echo "check-generated: generated file is not tracked by git: $f"
    echo "  (an untracked generated file can never be reported stale — commit it)"
    UNTRACKED=1
  fi
done <<< "$FILES"
[[ $UNTRACKED -eq 0 ]] || exit 1

stale=$(git status --porcelain -- $FILES 2>/dev/null || true)

if [ -n "$stale" ]; then
  echo "check-generated: stale generated file(s) — commit the regenerated output:"
  echo "$stale"
  exit 1
fi

echo "check-generated: clean ($(printf '%s\n' "$FILES" | grep -c .) generated files checked)"
exit 0
