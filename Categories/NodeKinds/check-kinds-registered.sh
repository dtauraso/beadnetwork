#!/usr/bin/env bash

# PLACEMENT: Categories/NodeKinds/kinds.go,Categories/NodeKinds/node-defs.ts | every kind directory must appear in BOTH registries — these two files are hand-edited now that each kind generates only itself, and a kind missing from either is a scene type that silently cannot be built or drawn

set -euo pipefail
cd "$(dirname "$0")"

SWITCH=kinds.go
DEFS=node-defs.ts
fail=0

for d in */; do
  dir=${d%/}
  [ -f "$dir/node.go" ] || continue
  grep -rq 'BuilderFor("' "$dir" --include="*.go" || continue

  kind=$(grep -rho 'BuilderFor("[A-Za-z_][A-Za-z0-9_]*"' "$dir" --include="*.go" | head -1 | sed 's/BuilderFor("//; s/"//')
  if [ -z "$kind" ]; then
    echo "UNREGISTERED: $dir has a BuilderFor call with no readable kind name"
    fail=1
    continue
  fi

  if ! grep -q "case \"$kind\":" "$SWITCH"; then
    echo "UNREGISTERED: kind $kind ($dir) has no 'case \"$kind\":' in $SWITCH — loading a scene that names it would fail with nothing building it"
    fail=1
  fi
  if ! grep -q "\"./$dir/node-def-gen\"" "$DEFS"; then
    echo "UNREGISTERED: kind $kind ($dir) is not imported by $DEFS — the editor would not draw it or offer it in the palette"
    fail=1
  fi

  frag="$dir/node-def-gen.ts"
  if [ ! -f "$frag" ]; then
    echo "MISSING FRAGMENT: $frag — run go generate ./..."
    fail=1
  fi
done

while read -r kind; do
  found=0
  for d in */; do
    dir=${d%/}
    [ -f "$dir/node.go" ] || continue
    grep -rq "BuilderFor(\"$kind\"" "$dir" --include="*.go" && found=1 && break
  done
  if [ "$found" = 0 ]; then
    echo "STALE REGISTRATION: $SWITCH has a case for $kind but no directory here declares BuilderFor(\"$kind\")"
    fail=1
  fi
done < <(grep -o 'case "[A-Za-z_][A-Za-z0-9_]*":' "$SWITCH" | sed 's/case "//; s/"://')

exit $fail
