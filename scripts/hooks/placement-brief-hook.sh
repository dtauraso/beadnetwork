#!/usr/bin/env bash
set -uo pipefail

# PLACEMENT: none | this is the hook wrapper itself; it reads the guards, it does not constrain a source path

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

payload="$(cat)"

path="$(printf '%s' "$payload" | python3 -c '
import json, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)
print(d.get("tool_input", {}).get("file_path", "") or "")
' 2>/dev/null || true)"

[ -z "$path" ] && exit 0

[ -e "$path" ] && exit 0

brief="$(bash "$SCRIPT_DIR/placement-brief.sh" "$path" 2>/dev/null || true)"
[ -z "$brief" ] && exit 0

printf '%s' "$brief" | python3 -c '
import json, sys
brief = sys.stdin.read()
print(json.dumps({
    "hookSpecificOutput": {
        "hookEventName": "PreToolUse",
        "additionalContext": (
            "Guard rules that will apply to this new file (from scripts/placement-brief.sh, "
            "read out of the guards themselves). Satisfy them now rather than after "
            "stop-checks fails:\n\n" + brief
        ),
    }
}))
'
exit 0
