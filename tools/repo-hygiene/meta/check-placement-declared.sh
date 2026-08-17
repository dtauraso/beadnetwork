#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: none | this guard reads the guards themselves, not the tree

#   # PLACEMENT: <glob>[,<glob>...] | <one-line rule>
#   # PLACEMENT: none | <why this guard is not path-scoped>

# language — gofmt, staticcheck, eslint, the prose-hygiene checks — declares "none" even

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

guards=()
while IFS= read -r g; do
  [ -n "$g" ] && guards+=("$g")
done < <(bash tools/guard-list.sh)

if [ "${#guards[@]}" -eq 0 ]; then
  echo "check-placement-declared: MISCONFIGURED — tools/guard-list.sh named no guards; refusing to report success." >&2
  exit 1
fi

dupes="$(printf '%s\n' "${guards[@]}" | xargs -n1 basename | sort | uniq -d)"
if [ -n "$dupes" ]; then
  echo "check-placement-declared: FAIL — guard basename(s) used more than once:"
  printf '%s\n' "$dupes" | sed 's/^/  /'
  echo
  echo "Guard results are reported by name; two guards sharing a basename are"
  echo "indistinguishable in the failure report. Rename one."
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
