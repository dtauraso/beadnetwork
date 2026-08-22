#!/usr/bin/env bash

# PLACEMENT: Categories/Node/**/*.go | a panic must NAME the invariant it broke, not just the symptom

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

readonly MIN_MSG_LEN=30

ROOTS=()
for d in Node NodeKinds Categories/Ring/Bead; do
  [ -d "$d" ] && ROOTS+=("$d")
done

[ ${#ROOTS[@]} -eq 0 ] && exit 0

FILES="$(find "${ROOTS[@]}" -name '*.go' ! -name '*_test.go' -type f)"
[ -z "$FILES" ] && exit 0

fail=0

report="$(echo "$FILES" | tr '\n' '\0' | xargs -0 awk -v minlen="$MIN_MSG_LEN" '
  FNR == 1 { seeking = 0 }

  /panic\(/ { seeking = 1; panicline = FNR; msg = ""; }

  seeking {
    if (match($0, /"[^"]*"/)) {
      msg = substr($0, RSTART + 1, RLENGTH - 2)
      seeking = 0

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

recovered="$(echo "$FILES" | tr '\n' '\0' | xargs -0 grep -n 'recover()' || true)"
if [ -n "$recovered" ]; then
  echo "panic-message: recover() in the network swallows an assertion — the failure becomes a"
  echo "silent wrong answer instead of a located crash. Let it panic."
  echo "$recovered"
  fail=1
fi

exit $fail
