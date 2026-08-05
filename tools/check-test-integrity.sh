#!/usr/bin/env bash
set -euo pipefail

# check-test-integrity.sh — makes tests-edited-to-go-green VISIBLE and DELIBERATE.
#
# The verify suite (go test, vitest, staticcheck, eslint, 30+ guards) all asks the same
# question: "does the suite pass?" None asks "did the suite change IN ORDER to pass?" An
# agent that deletes an assertion, loosens a comparison, adds t.Skip / .only, or drops a
# table case produces the same green as one that fixed the bug — and the Stop hook accepts
# it. This is the last big false-green escape route in the repo; every other rule here is
# guarded, and the verify gate itself was the exception.
#
# DETECTION, NOT PROHIBITION. Tests must stay editable — new tests, renames, refactors, and
# genuinely-wrong assertions being corrected are all legitimate. This guard fails only when a
# test change SHEDS strength (a fix should generally add or hold assertions, not lose them),
# and it names the file and what was lost so the change is a deliberate decision, not a
# reflex.
#
# Signals (diffed against the merge-base with origin/main, matching stop-checks' branch-ahead
# gating — so it is a no-op on main and only looks at this branch's own changes, committed or
# working-tree):
#   1. NET assertion loss across all changed test files (t.Fatal/Error[f], require./assert.,
#      expect(). Counted across ALL files so moving a test between files is net-zero.
#   2. Newly ADDED skips/onlys: t.Skip(Now), .skip(, .only( (worst — silently disables every
#      OTHER test in a vitest file), xit/xdescribe.
#   3. Newly ADDED os.Exit / recover() inside test files (fake-pass / swallow-failure tricks).
#
# ESCAPE HATCH WITH A COST, AND IT CLOSES BEHIND ITSELF: put the marker — the bracketed
# form `allow-test-weakening:` followed by one or more paths — AT THE START OF A LINE
# in a commit message on the branch to state a deliberate removal (mirrors how
# strip-branch-local-docs keys off a marker). It is a commit-message marker ON PURPOSE — not a
# CLI flag the agent can pass itself, and not silent. Uncommitted weakening stays flagged until
# you commit it with the marker.
#
# The marker EXEMPTS THE FILES IT NAMES, nothing else. It used to be a bare marker that exited
# the whole guard 0 for the rest of the branch: one commit stating one deliberate removal
# disabled test-integrity checking for every later commit, in every file. That is what happened
# in the explicit-upper-bounds work — two legitimate `recover()`s asserting that a ceiling
# panics (which WEAKEN_RE cannot tell apart from a `recover()` swallowing a failure) switched
# the guard off for the remainder of the branch. A hatch whose scope outlives its justification
# is a disarmed guard that still reads as a passing one.
#
# A marker naming NO path is a FAILURE, not a pass. Silently ignoring it would restore exactly
# the false-clean this narrowing exists to remove: the author believes they stated an
# exemption, the guard believes nothing was claimed, and the disagreement is invisible.
# Exempted files are dropped from the net-assertion counters entirely, not merely from the
# report — otherwise their removals still drag the branch-wide net negative, and their
# additions could mask a real loss in some other file.
#
# The line-start anchor is what lets this file, and a commit message, WRITE the syntax down:
# an indented example or an inline mention is prose, only a line that BEGINS with the marker
# is a claim. Without it there is no way to document the hatch without invoking it.
#
# Exit 0 clean, exit 1 with a named report.
#
# PLACEMENT: **/*_test.go,**/*.test.ts,**/*.test.tsx,**/*.spec.ts,**/*.spec.tsx | a test edit must not net-remove assertions or add skip/only/exit/recover without an [allow-test-weakening: <paths>] commit marker

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Base ref, same precedence stop-checks.sh uses. No base (no origin/main and no main, e.g. a
# shallow CI checkout) → nothing to diff against, no-op.
if git rev-parse --verify -q origin/main >/dev/null 2>&1; then
  base_ref="origin/main"
elif git rev-parse --verify -q main >/dev/null 2>&1; then
  base_ref="main"
else
  exit 0
fi
base=$(git merge-base "$base_ref" HEAD 2>/dev/null || true)
[ -z "$base" ] && exit 0

# Escape hatch: collect the files named by [allow-test-weakening: ...] markers in this
# branch's commit messages. Each marker exempts ONLY the paths it lists.
#
# NB: read the markers with grep -oE (NOT grep -q). Under `set -o pipefail`, grep -q closes
# the pipe on its first match — and the marker rides the newest commit, emitted FIRST — so
# git log, still streaming the rest of the branch's messages, dies with SIGPIPE (141) and
# pipefail turns the whole condition non-zero, silently skipping the hatch on any branch
# whose log exceeds the pipe buffer. grep -o drains all input, so no early close, no SIGPIPE.
exempt_list=""
bare_marker=0
if [ "$base" != "$(git rev-parse HEAD)" ]; then
  msgs=$(git log --format='%B' "$base"..HEAD 2>/dev/null || true)
  # Well-formed: marker, colon, at least one path, closing bracket.
  while IFS= read -r m; do
    [ -z "$m" ] && continue
    # Strip the marker prefix and the trailing ']', leaving whitespace-separated paths.
    paths=${m#*allow-test-weakening:}
    paths=${paths%]}
    for p in $paths; do
      exempt_list="$exempt_list $p"
    done
  done < <(printf '%s\n' "$msgs" | grep -oE '^\[allow-test-weakening:[^]]*\]' || true)
  # Bare (path-less) markers: [allow-test-weakening] or [allow-test-weakening:<blank>].
  # grep -oE, never -q, for the same SIGPIPE reason as above.
  bare=$(printf '%s\n' "$msgs" \
    | grep -oE '^\[allow-test-weakening(\]|:[[:space:]]*\])' || true)
  [ -n "$bare" ] && bare_marker=1
fi

if [ "$bare_marker" -ne 0 ]; then
  echo "check-test-integrity: a commit on this branch carries a PATH-LESS [allow-test-weakening]"
  echo "marker. The marker exempts the files it names and nothing else — a bare one exempts"
  echo "nothing, so it cannot be honoured, and honouring it silently would disarm this guard for"
  echo "the whole branch (the failure mode this form exists to prevent)."
  echo "Reword the commit message to name the files, e.g.:"
  echo "  [allow-test-weakening: nodes/Wiring/foo_test.go]"
  exit 1
fi

# Test files changed on this branch (committed-ahead + working tree) vs the base.
# while-read (not mapfile) so this runs under macOS's bash 3.2.
files=()
while IFS= read -r line; do
  [ -n "$line" ] && files+=("$line")
done < <(git diff --name-only "$base" -- '*_test.go' '*.test.ts' '*.test.tsx' '*.spec.ts' '*.spec.tsx' 2>/dev/null || true)
[ ${#files[@]} -eq 0 ] && exit 0

ASSERT_RE='t\.(Fatal|Fatalf|Error|Errorf)\b|\b(require|assert)\.|expect\('
WEAKEN_RE='\bt\.Skip(Now)?\b|\.(skip|only)\(|\bxit\b|\bxdescribe\b|\bos\.Exit\b|\brecover\(\)'

total_removed=0
total_added=0
report=""
exempted_seen=""

is_exempt() {
  local want="$1" p
  for p in $exempt_list; do
    [ "$p" = "$want" ] && return 0
  done
  return 1
}

for f in "${files[@]}"; do
  # Exempt files are skipped ENTIRELY — not counted, not reported. Counting them would let a
  # marked file's removals push the branch-wide net negative anyway, and let its additions
  # offset an unmarked file's real loss.
  if is_exempt "$f"; then
    exempted_seen="$exempted_seen $f"
    continue
  fi
  [ -f "$f" ] || { report+="  $f: test file DELETED (all its coverage removed)\n"; continue; }
  fdiff=$(git diff "$base" -- "$f" 2>/dev/null || true)
  [ -z "$fdiff" ] && continue

  # Added/removed lines only (skip the +++/--- headers).
  added_lines=$(printf '%s\n' "$fdiff" | grep -E '^\+' | grep -vE '^\+\+\+' || true)
  removed_lines=$(printf '%s\n' "$fdiff" | grep -E '^-' | grep -vE '^---' || true)

  r=$(printf '%s\n' "$removed_lines" | grep -cE "$ASSERT_RE" || true)
  a=$(printf '%s\n' "$added_lines" | grep -cE "$ASSERT_RE" || true)
  total_removed=$((total_removed + r))
  total_added=$((total_added + a))
  if [ "$r" -gt "$a" ]; then
    report+="  $f: assertions ${a} added / ${r} removed (net -$((r - a)))\n"
  fi

  weak=$(printf '%s\n' "$added_lines" | grep -nE "$WEAKEN_RE" | sed 's/^/      /' || true)
  if [ -n "$weak" ]; then
    report+="  $f: added a skip/only/exit/recover:\n${weak}\n"
  fi
done

fail=0
if [ "$total_removed" -gt "$total_added" ]; then
  fail=1
fi
[ -n "$report" ] && fail=1

# A marker naming a path this branch never changed exempts nothing. Same reasoning as the
# path-less form and as check-guards-refuse-vacuous: the author thinks a claim was recorded,
# the guard sees none, and a typo'd path reads identically to a real exemption.
stale=""
for p in $exempt_list; do
  case " $exempted_seen " in
    *" $p "*) ;;
    *) stale="$stale $p" ;;
  esac
done
if [ -n "${stale// /}" ]; then
  echo "check-test-integrity: [allow-test-weakening] names path(s) this branch does not change:"
  for p in $stale; do echo "    $p"; done
  echo "An exemption for an unchanged file exempts nothing. Fix the path (or drop it) in the"
  echo "commit message — a typo here reads exactly like a granted exemption."
  exit 1
fi

if [ "$fail" -ne 0 ]; then
  echo "check-test-integrity: this branch's test changes SHED strength (assertions removed,"
  echo "or a skip/only/exit added). A fix should add or hold assertions, not lose them."
  echo "Net assertions across changed test files: ${total_added} added / ${total_removed} removed."
  [ -n "$report" ] && printf '%b' "$report"
  echo "If the removal is deliberate (a genuinely-wrong assertion, a retired test), state it in a"
  echo "commit message on this branch, NAMING THE FILES it covers and why:"
  echo "  [allow-test-weakening: path/to/foo_test.go path/to/bar_test.go]"
  echo "It exempts exactly those paths — every other test file on the branch stays guarded."
  exit 1
fi

exit 0
