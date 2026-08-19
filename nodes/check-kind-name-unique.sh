#!/usr/bin/env bash

# PLACEMENT: nodes/*/node.go | a kind's wire.Register("<Kind>") name must be unique across every node package
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

shopt -s nullglob
node_files=("$REPO_ROOT"/nodes/*/node.go)
shopt -u nullglob

if [ ${#node_files[@]} -eq 0 ]; then

  exit 0
fi

dups=$(grep -hoE '(wire|Wiring)\.Register\("[^"]+"' "${node_files[@]}" \
  | sed -E 's/.*Register\("([^"]+)"/\1/' \
  | sort | uniq -d || true)

if [ -n "$dups" ]; then
  echo "check-kind-name-unique: a kind name is registered by more than one production"
  echo "node package (registry.go would panic on this at startup). Duplicate kind(s):"
  echo "$dups" | sed 's/^/  /'
  exit 1
fi

exit 0
