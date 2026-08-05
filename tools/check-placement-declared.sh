#!/usr/bin/env bash
set -euo pipefail

# check-placement-declared.sh — every guard must say which files it constrains.
#
# PLACEMENT: none | this guard reads the guards themselves, not the tree
#
# WHAT THIS DEFENDS
# tools/placement-brief.sh answers "before I write this file, what will the guards demand of
# it?" by reading a declared line out of each guard:
#
#   # PLACEMENT: <glob>[,<glob>...] | <one-line rule>
#   # PLACEMENT: none | <why this guard is not path-scoped>
#
# The brief is only as complete as those declarations. A guard added WITHOUT one is invisible
# to it — and invisible in the worst way, because the brief still prints confidently, just
# missing a rule. Silence reads as "nothing applies here" whether that is true or not.
#
# So the declaration is mandatory, and "none" is an explicit, justified answer rather than an
# omission: a guard that constrains no particular path (a repo-wide hygiene check, a check of
# the guards themselves) says so out loud.
#
# GLOBS ARE FOR DECISIONS, NOT FOR LINT. A guard that applies uniformly to every file of a
# language — gofmt, staticcheck, eslint, the prose-hygiene checks — declares "none" even
# though it does constrain the file being written. It poses no CHOICE: nobody places code
# differently because gofmt exists. Declaring globs for those was tried and measured: one Go
# file drew 17 rules, of which 7 were unavoidable hygiene, and the 2 that mattered were
# buried among them. A brief nobody reads is worth less than no brief, so the bar for a glob
# is "someone could put this in the wrong place", not "this check touches the file".
#
# WHAT IT DOES NOT CHECK
# It does not verify that a glob matches anything, nor that the rule text is accurate. A stale
# rule is still possible — but it is stale IN the guard that enforces it, one screen from the
# code that would prove it wrong, which is the closest coupling available short of generating
# the guard from the rule.
#
# Exit 0 if clean; exit 1 with a report otherwise.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

shopt -s nullglob
guards=("$REPO_ROOT"/tools/check-*.sh)
shopt -u nullglob

if [ "${#guards[@]}" -eq 0 ]; then
  echo "check-placement-declared: MISCONFIGURED — no tools/check-*.sh found; refusing to report success." >&2
  exit 1
fi

missing=()
malformed=()
for guard in "${guards[@]}"; do
  name="$(basename "$guard")"
  lines="$(grep -h '^# PLACEMENT:' "$guard" 2>/dev/null || true)"
  if [ -z "$lines" ]; then
    missing+=("$name")
    continue
  fi
  # Every declared line must carry the " | rule" half; a glob with no rule is a line that
  # prints an empty brief, which is worse than no line at all (it looks answered).
  while IFS= read -r line; do
    body="${line#*PLACEMENT:}"
    case "$body" in
      *\|*) ;;
      *) malformed+=("$name: $line") ;;
    esac
    rule="${body#*|}"
    rule="$(printf '%s' "$rule" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
    [ -z "$rule" ] && malformed+=("$name: empty rule text in: $line")
  done <<< "$lines"
done

status=0
if [ "${#missing[@]}" -gt 0 ]; then
  status=1
  echo "check-placement-declared: FAIL — ${#missing[@]} guard(s) declare no PLACEMENT line:"
  for m in "${missing[@]}"; do echo "  $m"; done
  echo
  echo "Add one line near the top of each, so tools/placement-brief.sh can surface the rule"
  echo "BEFORE a file is written the wrong way rather than after:"
  echo
  echo "  # PLACEMENT: nodes/Wiring/*.go | <what this guard demands of such a file>"
  echo "  # PLACEMENT: none | <why this guard constrains no particular path>"
fi

if [ "${#malformed[@]}" -gt 0 ]; then
  status=1
  echo "check-placement-declared: FAIL — ${#malformed[@]} malformed PLACEMENT line(s):"
  for m in "${malformed[@]}"; do echo "  $m"; done
  echo
  echo "Format is: # PLACEMENT: <glob>[,<glob>...] | <one-line rule>"
fi

exit $status
