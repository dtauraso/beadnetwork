#!/usr/bin/env bash

# PLACEMENT: none | PreToolUse(Bash) hook on git commit; it inspects the repo root for stray screenshots, not a source path
set -uo pipefail

input="$(cat)"
cmd="$(printf '%s' "$input" | jq -r '.tool_input.command // empty')"

if [ -z "$cmd" ]; then exit 0; fi
if ! printf '%s' "$cmd" | grep -Eq '(^|[;&|[:space:]])git[[:space:]]+commit([[:space:]]|$)'; then
  exit 0
fi

files=$(find . -maxdepth 1 \( -name 'Screenshot*.png' -o -name 'Screen Shot*.png' \) 2>/dev/null | sed 's|^\./||' | sort)
if [ -n "$files" ]; then
  python3 -c "
import json, sys
files = sys.stdin.read().strip().splitlines()
msg = 'Stray screenshot(s) in repo root: ' + ', '.join(files) + '. Per the visual-editor convention, move them under docs/planning/visual-editor/screenshots/ with a date-prefixed kebab name (e.g. 2026-05-05-<topic>-N.png) and reference them from docs/planning/visual-editor/session-log.md in the same commit as the work they motivate.'
print(json.dumps({'hookSpecificOutput': {'hookEventName': 'PreToolUse', 'permissionDecision': 'deny', 'permissionDecisionReason': msg}}))
" <<< "$files"
fi
exit 0
