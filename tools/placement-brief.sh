#!/usr/bin/env bash
set -uo pipefail

# placement-brief.sh — "before you write this file, here is what the guards will say about
# where it goes."
#
# THE PROBLEM THIS SOLVES
# The guards in tools/check-*.sh encode a placement discipline: a write to view/* belongs in
# a view-owner file; a filepath.Join("view", …) belongs in scene_paths.go; useSyncExternalStore
# belongs in a named buffer-reflect resource; a breadcrumb label must be registered. Each rule
# is documented — in the guard's own header, which nobody reads until that guard fails. The
# feedback therefore arrives AFTER the file is written the wrong way, and the fix is a
# restructure rather than a line.
#
# This flips the order: given a path, print the rules that will apply to it. It is the same
# information the guards already carry, surfaced before the work instead of after.
#
# DERIVED, NOT DUPLICATED
# The rules are read out of the guards themselves, from a declared line:
#
#   # PLACEMENT: <glob>[,<glob>...] | <one-line rule>
#   # PLACEMENT: none | <why this guard is not path-scoped>
#
# A guard may declare several. Because the text lives in the guard, a rule cannot drift from
# the check that enforces it — the failure mode of every hand-maintained "here are our
# conventions" document. check-placement-declared.sh fails the build if a guard declares none.
#
# Usage:
#   tools/placement-brief.sh <path> [<path>...]
#
# Prints nothing and exits 0 when no rule matches — silence is the common case, and a brief
# that speaks on every edit is one that gets ignored.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ "$#" -lt 1 ]; then
  echo "usage: placement-brief.sh <path> [<path>...]" >&2
  exit 2
fi

# repo_relative strips the repo root from an absolute path so the globs below can be written
# against repo-relative paths (which is how every guard talks about files).
repo_relative() {
  local p="$1"
  case "$p" in
    "$REPO_ROOT"/*) printf '%s\n' "${p#"$REPO_ROOT"/}" ;;
    /*) printf '%s\n' "$p" ;;  # outside the repo: leave alone, it will simply match nothing
    *) printf '%s\n' "$p" ;;
  esac
}

matched_any=0
brief_for() {
  local path; path="$(repo_relative "$1")"
  local hits=""

  local guard
  for guard in "$REPO_ROOT"/tools/check-*.sh; do
    [ -f "$guard" ] || continue
    local line
    while IFS= read -r line; do
      # Strip the marker, then split on the first " | " into globs and rule.
      local body="${line#*PLACEMENT:}"
      local globs="${body%%|*}"
      local rule="${body#*|}"
      # Trim surrounding whitespace from both halves.
      globs="$(printf '%s' "$globs" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
      rule="$(printf '%s' "$rule" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
      [ "$globs" = "none" ] && continue

      local glob
      # shellcheck disable=SC2086  # deliberate word split on commas
      IFS=',' read -ra _globs <<< "$globs"
      for glob in "${_globs[@]}"; do
        glob="$(printf '%s' "$glob" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
        # shellcheck disable=SC2053  # glob on the RIGHT is the point
        if [[ "$path" == $glob ]]; then
          hits+="  $(basename "$guard" .sh): $rule"$'\n'
          break
        fi
      done
    done < <(grep -h '^# PLACEMENT:' "$guard" 2>/dev/null || true)
  done

  if [ -n "$hits" ]; then
    matched_any=1
    printf 'placement rules for %s:\n%s' "$path" "$hits"
  fi
}

for p in "$@"; do
  brief_for "$p"
done

exit 0
