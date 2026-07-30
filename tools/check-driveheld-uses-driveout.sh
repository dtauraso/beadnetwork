#!/usr/bin/env bash
# check-driveheld-uses-driveout.sh — every node kind that spawns a gatecommon.DriveHeld
# drive goroutine must resolve ITS DRIVEN output(s) via BuildArgs.DriveOut, never via plain
# BuildArgs.Out. Run from repo root: bash tools/check-driveheld-uses-driveout.sh
#
# WHY THIS EXISTS: docs/interior-stream-framing.md documents the live-editor "bad frame
# length" desync — gatecommon.DriveHeld spawns its OWN goroutine, a SEPARATE goroutine from
# a node's Update loop, and that goroutine used to write the SAME shared *interiorStream
# the Update loop writes (because both got their eventSink from a.Out(...), which routes
# through the node's one shared getStream). The fix is BuildArgs.DriveOut
# (nodes/Wiring/build_args.go), which routes a driven output through its OWN dedicated
# per-(node,slot) fd instead (Buffer.StreamKindDrive). A reviewer or a future kind can
# reintroduce exactly the original bug by writing `a.Out("Out")` for a port that also gets
# handed to `gatecommon.DriveHeld(...)` two lines later — nothing about that compiles
# differently or fails any existing unit test in isolation (nodes/Wiring's own
# TestDriveStreamNeverSharesNodesInteriorStream only checks the GETTERS wire correctly, not
# that every kind actually CALLS the right one). This guard is the one check that reads the
# kind's own source and catches that specific mistake.
#
# WHAT IT CHECKS: for every nodes/<Kind>/node.go that calls `gatecommon.DriveHeld(`, the
# same file must also contain at least one `a.DriveOut(` call, and must NOT contain a bare
# `a.Out("Out")` or `a.Out("OutFanout")` — the two port names every current DriveHeld-using
# kind drives (Pulse/PulseLeft/PulseRight/holdflip; see Buffer.DriveSlotsPerNode's doc
# comment for why "Out"/"OutFanout" are the only names in play today). A node kind that
# calls DriveHeld on a differently-named port would need this list extended alongside it —
# see the PORT_NAMES array below.
#
# SCOPE: nodes/*/node.go only (production kind sources) — not _test.go files, which
# legitimately construct bare shared streams by hand to pin the mechanism
# (nodes/Wiring/interior_stream_concurrent_write_test.go's writeTwoCallFrame).
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# The port names every current DriveHeld-driving kind names its driven output(s). A new
# kind driving a differently-named port needs this extended in the SAME commit that adds
# it, or this guard silently misses it — same "small hand-kept list" shape as
# runCommand.ts's DRIVE_SLOTS_PER_NODE.
PORT_NAMES=("Out" "OutFanout")

fail=0
report=""

# Every node.go (production kind source, never a _test.go) that calls gatecommon.DriveHeld.
# while-read (not mapfile) so this runs under macOS's bash 3.2, which has no mapfile.
while IFS= read -r f; do
  [ -z "$f" ] && continue
  if ! grep -q 'a\.DriveOut(' "$f"; then
    report="${report}${f}: calls gatecommon.DriveHeld but never calls a.DriveOut — its driven output(s) are not on a dedicated fd (docs/interior-stream-framing.md)\n"
    fail=1
    continue
  fi
  for port in "${PORT_NAMES[@]}"; do
    if grep -qE "a\.Out\(\"${port}\"\)" "$f"; then
      report="${report}${f}: resolves \"${port}\" via a.Out(...) in a file that also calls gatecommon.DriveHeld — a DriveHeld-driven output must use a.DriveOut(...), or it shares the node's interior-stream getter (the original framing-desync bug)\n"
      fail=1
    fi
  done
done < <(grep -rl --include='node.go' 'gatecommon\.DriveHeld(' nodes 2>/dev/null || true)

if [ "$fail" -ne 0 ]; then
  echo -e "check-driveheld-uses-driveout: FAILED\n${report}" >&2
  exit 1
fi
exit 0
