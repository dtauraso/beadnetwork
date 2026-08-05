#!/usr/bin/env bash
set -uo pipefail

# placement-brief-hook.sh — PreToolUse(Write|Edit) hook: before a file is written, say what
# the guards will demand of it.
#
# PLACEMENT: none | this is the hook wrapper itself; it reads the guards, it does not constrain a source path
#
# It is ADVISORY, never a block. Exit is always 0 and stdin's tool call is never modified —
# the hook adds CONTEXT (a PreToolUse `additionalContext` payload) and nothing else. The
# guards themselves remain the enforcement; this only moves their content earlier, from
# "stop-checks failed, restructure the file" to "before you write it".
#
# Silent when no rule matches, which is the common case. A hook that speaks on every edit is
# a hook that gets skimmed past, and then the one time it mattered it was noise like all the
# rest (see check-placement-declared.sh's "globs are for decisions, not for lint").

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

payload="$(cat)"

# The path being written. Prefer python3 for a real JSON parse; the repo already depends on
# it for another hook (scripts/block-open-html-hook.py).
path="$(printf '%s' "$payload" | python3 -c '
import json, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(0)
print(d.get("tool_input", {}).get("file_path", "") or "")
' 2>/dev/null || true)"

[ -z "$path" ] && exit 0

# Only for files that do not exist yet. An EDIT to an existing file is a change within a
# placement decision already made; the brief is about where a NEW thing goes, and repeating
# it on every edit of a long-lived file is exactly the noise this stays quiet to avoid.
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
            "Guard rules that will apply to this new file (from tools/placement-brief.sh, "
            "read out of the guards themselves). Satisfy them now rather than after "
            "stop-checks fails:\n\n" + brief
        ),
    }
}))
'
exit 0
