#!/usr/bin/env bash
set -euo pipefail

# check-stream-fd-mismatch-reported.sh — a per-owner stream that Go wires CONDITIONALLY
# must say so when the condition fails.
#
# THE BUG THIS EXISTS FOR. main.go wires each per-owner stream behind
# `if base, ok := streamFDs[B.StreamKind<X>]; ok`. When the entry is absent the block is
# simply skipped: every mover of that kind keeps a nil streamOut and emits nothing, so an
# ENTIRE ENTITY CLASS vanishes from the editor while every file on disk is correct. That
# happened to edges on 2026-07-28 — a VS Code extension host left running across a change
# handed Go older fd plumbing (the host owns WIREFOLD_STREAM_FDS, see runCommand.ts), and
# the silence read as a code defect: the bundle was byte-identical to a fresh build and the
# Go headless edge test passed, because neither was where the fault was.
#
# It is the same class runCommand.ts's MAX_EDGE_STREAMS overflow was fixed for in 93d2e9b6
# — "the quietest possible failure for the loudest consequence." That fix and the Go-side
# one are both point fixes; this guard is what keeps the NEXT stream kind from landing with
# a silent skip, since the failure is invisible precisely when it matters.
#
# WHAT IT ASSERTS: every per-owner StreamKind constant declared in Buffer/stream_fds.go is
# named in a stream-fd mismatch report in main.go. Not that the report is correct — a guard
# cannot know that — only that the silent-skip path has a voice.
#
# DELIBERATELY EXCLUDED: StreamKindView. It is the singleton view stream, not a per-owner
# one: there is no population ("the graph has N of these") to compare an fd count against,
# and its absence is not silent in the same way — no camera/overlay state means an
# immediately dead scene rather than one intact class quietly missing. If view ever gains a
# per-entity fd range, delete it from EXCLUDED here and give it a report.
#
# WHAT NO GUARD CAN COVER: the operational cause itself. A stale extension host is a
# running process, not a fact about the tree, so this guard cannot detect one — it only
# ensures the runtime says which failure it is. The reload trigger stays in
# memory/feedback_two_process_editor_reload.md for that reason.
#
# Exit 0 clean, exit 1 with a named report.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

KINDS_FILE="Buffer/stream_fds.go"
MAIN_FILE="main.go"

# A missing scan root is a disarmed guard, not a clean one (check-guards-refuse-vacuous).
for f in "$KINDS_FILE" "$MAIN_FILE"; do
  if [ ! -f "$f" ]; then
    echo "check-stream-fd-mismatch-reported: $f not found — this guard cannot scan and is"
    echo "reporting that fact rather than passing vacuously. If the file moved, repoint it here."
    exit 1
  fi
done

EXCLUDED="StreamKindView"

# Discover the kinds instead of hardcoding them, so a NEW kind is covered the day it lands.
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
  # Only kinds main.go actually wires conditionally can have a silent-skip path.
  grep -qF "streamFDs[B.$k]" "$MAIN_FILE" || continue
  checked=$((checked + 1))
  # The report must name the kind constant and sit on a stream-fd mismatch line.
  if ! grep -F 'stream-fd mismatch' "$MAIN_FILE" | grep -qF "B.$k"; then
    if ! grep -A6 'stream-fd mismatch' "$MAIN_FILE" | grep -qF "B.$k"; then
      missing="$missing $k"
    fi
  fi
done

if [ "$checked" -eq 0 ]; then
  echo "check-stream-fd-mismatch-reported: found ${#kinds[@]} StreamKind constant(s) but none are"
  echo "wired via streamFDs[B.<kind>] in $MAIN_FILE. Either the wiring moved out of main.go or"
  echo "the pattern changed — - either way this guard is scanning nothing. Repoint it."
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
