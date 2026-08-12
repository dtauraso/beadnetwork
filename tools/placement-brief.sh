#!/usr/bin/env bash
set -uo pipefail


















#   # PLACEMENT: <glob>[,<glob>...] | <one-line rule>
#   # PLACEMENT: none | <why this guard is not path-scoped>











SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ "$#" -lt 1 ]; then
  echo "usage: placement-brief.sh <path> [<path>...]" >&2
  exit 2
fi



repo_relative() {
  local p="$1"
  case "$p" in
    "$REPO_ROOT"/*) printf '%s\n' "${p#"$REPO_ROOT"/}" ;;
    /*) printf '%s\n' "$p" ;;
    *) printf '%s\n' "$p" ;;
  esac
}

matched_any=0
brief_for() {
  local path; path="$(repo_relative "$1")"
  local hits=""

  local guard
  for guard in "$REPO_ROOT"/tools/*/check-*.sh "$REPO_ROOT"/tools/*/*/check-*.sh; do
    [ -f "$guard" ] || continue
    local line
    while IFS= read -r line; do

      local body="${line#*PLACEMENT:}"
      local globs="${body%%|*}"
      local rule="${body#*|}"

      globs="$(printf '%s' "$globs" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
      rule="$(printf '%s' "$rule" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
      [ "$globs" = "none" ] && continue

      local glob

      IFS=',' read -ra _globs <<< "$globs"
      for glob in "${_globs[@]}"; do
        glob="$(printf '%s' "$glob" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"

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
