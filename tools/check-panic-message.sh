#!/usr/bin/env bash
# check-panic-message.sh — fail if a network panic does not name the invariant it broke.
# Run from repo root: bash tools/check-panic-message.sh
#
# WHY THIS EXISTS: a panic in nodes/, Buffer/, or Trace/ is an ASSERTION — it fires only via
# a code bug, never via ordinary traffic (MODEL.md "Assertions"). Its message is therefore
# read exactly once, by whoever is debugging, and it is the ONLY context they get. A message
# that names the site, the bound, and the mechanism that should have prevented it ends the
# investigation in one step; "limit exceeded" starts a file-reading expedition. That
# difference is the whole value, so it is the thing worth guarding.
#
# Until now the convention was a GREP INSTRUCTION: paced_wire.go told the next reader to
# "grep panic( in this package/nodes/Wiring for the convention". That makes every reader
# re-derive an unwritten rule from examples. MODEL.md "Assertions" now states it; this guard
# keeps the corpus matching the statement.
#
# THE RULE — every panic message in the network must:
#   1. open with a SITE TAG: an identifier followed by ": ", naming the function, method, or
#      subsystem that detected the violation (e.g. "paced_wire: ", "nodeMover(%s): ",
#      "BuildEdgeStreamFrame: "). The tag is what makes the message greppable back to here.
#   2. be SUBSTANTIVE — at least MIN_MSG_LEN characters. A tag alone ("wire: bad") passes (1)
#      while telling the reader nothing.
# and the network must contain no recover(): swallowing an assertion converts a loud, located
# failure into a silent wrong answer, which is strictly worse than the crash.
#
# Scope is non-test Go under nodes/, Buffer/, Trace/ — the network itself. Tests panic to
# ASSERT (nodes/Wiring/pending_bound_test.go drives a bound past its limit on purpose), and
# tools/ is build-time tooling whose failures are read at a terminal with a stack trace.
#
# Exit 0 clean (empty), exit 1 with a report — matches the guard-loop contract in
# scripts/stop-checks.sh (auto-discovered via tools/check-*.sh glob).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Shortest message that can plausibly name a site AND what broke. The current corpus's
# shortest is wire.Register's 39-char "wire.Register: kind already registered: ", so 30
# leaves headroom without admitting "wire: bad".
readonly MIN_MSG_LEN=30

ROOTS=()
for d in nodes Buffer Trace; do
  [ -d "$d" ] && ROOTS+=("$d")
done
# Nothing to enforce in a checkout without the network packages.
[ ${#ROOTS[@]} -eq 0 ] && exit 0

# One find for every non-test .go file under the network roots, reused by both passes below
# (per-file process spawns are what made the older guards slow — see commit e9896c4a).
FILES="$(find "${ROOTS[@]}" -name '*.go' ! -name '*_test.go' -type f)"
[ -z "$FILES" ] && exit 0

fail=0

# Pass 1 — panic messages. awk carries state across lines because the message frequently sits
# on the line AFTER `panic(fmt.Sprintf(`; we scan forward to the first double-quoted string.
report="$(echo "$FILES" | tr '\n' '\0' | xargs -0 awk -v minlen="$MIN_MSG_LEN" '
  FNR == 1 { seeking = 0 }

  # Enter seek mode at a panic( call site; remember where it was for the report.
  /panic\(/ { seeking = 1; panicline = FNR; msg = ""; }

  seeking {
    # First double-quoted string at or after the panic( line is the message.
    if (match($0, /"[^"]*"/)) {
      msg = substr($0, RSTART + 1, RLENGTH - 2)
      seeking = 0

      # A site tag is an identifier — optionally with dots, or a (%s)-style qualifier —
      # followed by a colon and a space. Anchored, so the colon must be at the FRONT of the
      # message and not merely somewhere inside it.
      hastag = (msg ~ /^[A-Za-z_][A-Za-z0-9_.]*(\([^)]*\))?: /)

      if (!hastag) {
        printf "%s:%d: panic message does not open with a site tag (want \"funcOrSubsystem: ...\"): %s\n", FILENAME, panicline, msg
        bad = 1
      } else if (length(msg) < minlen) {
        printf "%s:%d: panic message is %d chars, under the %d-char minimum — name what broke and what should have prevented it: %s\n", FILENAME, panicline, length(msg), minlen, msg
        bad = 1
      }
    }
  }
  END { exit 0 }
')"

if [ -n "$report" ]; then
  echo "panic-message: a network panic is an assertion; its message is the only context the"
  echo "next debugger gets. See MODEL.md \"Assertions\"."
  echo "$report"
  fail=1
fi

# Pass 2 — recover() in the network turns a located crash into a silent wrong answer.
recovered="$(echo "$FILES" | tr '\n' '\0' | xargs -0 grep -n 'recover()' || true)"
if [ -n "$recovered" ]; then
  echo "panic-message: recover() in the network swallows an assertion — the failure becomes a"
  echo "silent wrong answer instead of a located crash. Let it panic."
  echo "$recovered"
  fail=1
fi

exit $fail
