#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: main.go,runtopology/edge_stream.go,runtopology/node_stream.go,tools/topology-vscode/src/Buffer/streamframe/stream_fds.go | every conditionally-wired per-owner StreamKind must have a named stream-fd mismatch report

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

KINDS_FILE="tools/topology-vscode/src/Buffer/streamframe/stream_fds.go"

MAIN_FILE="main.go runtopology/edge_stream.go runtopology/node_stream.go"

for f in "$KINDS_FILE" $MAIN_FILE; do
  if [ ! -f "$f" ]; then
    echo "check-stream-fd-mismatch-reported: $f not found — this guard cannot scan and is"
    echo "reporting that fact rather than passing vacuously. If the file moved, repoint it here."
    exit 1
  fi
done

EXCLUDED="StreamKindView"

kinds=()
while IFS= read -r k; do
  [ -n "$k" ] && kinds+=("$k")
done < <(grep -oE '^const (StreamKind[A-Za-z]+) =' "$KINDS_FILE" | awk '{print $2}' || true)

if [ ${#kinds[@]} -eq 0 ]; then
  echo "check-stream-fd-mismatch-reported: no StreamKind constants found in $KINDS_FILE."
  echo "The declaration form probably changed; this guard is now blind. Fix its pattern."
  exit 1
fi

missing=""
checked=0
for k in "${kinds[@]}"; do
  case " $EXCLUDED " in
    *" $k "*) continue ;;
  esac

  grep -qE "streamFDs\[[A-Za-z_][A-Za-z0-9_]*\.$k\]" $MAIN_FILE || continue
  checked=$((checked + 1))

  if ! grep -F 'stream-fd mismatch' $MAIN_FILE | grep -qE "[A-Za-z_][A-Za-z0-9_]*\.$k\b"; then
    if ! grep -A6 'stream-fd mismatch' $MAIN_FILE | grep -qE "[A-Za-z_][A-Za-z0-9_]*\.$k\b"; then
      missing="$missing $k"
    fi
  fi
done

if [ "$checked" -eq 0 ]; then
  echo "check-stream-fd-mismatch-reported: found ${#kinds[@]} StreamKind constant(s) but none are"
  echo "wired via streamFDs[B.<kind>] in $MAIN_FILE. Either the wiring moved or"
  echo "the pattern changed — either way this guard is scanning nothing. Repoint it."
  exit 1
fi

if [ -n "${missing// /}" ]; then
  echo "check-stream-fd-mismatch-reported: per-owner stream kind(s) wired behind a conditional"
  echo "in $MAIN_FILE with NO stream-fd mismatch report when the fds are absent:"
  for k in $missing; do echo "    B.$k"; done
  echo
  echo "Skipping the block silently leaves every mover of that kind with a nil stream, so the"
  echo "whole class disappears from the editor while the tree looks correct. Add a report on"
  echo "stderr naming the kind (see the StreamKindEdge case in $MAIN_FILE)."
  exit 1
fi

exit 0
