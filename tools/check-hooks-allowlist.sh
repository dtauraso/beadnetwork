#!/usr/bin/env bash
# check-hooks-allowlist.sh — every .claude/settings.json hook command must be a KNOWN
# check/reminder script. Run from repo root: bash tools/check-hooks-allowlist.sh
#
# WHY THIS EXISTS (drift-checklist item #4 — "hidden repair loop"): the checklist asks "does
# a second pass silently rewrite the answer before delivery?" A repo guard can't observe the
# runtime, but every automated pass over a turn is a hook declared in settings.json. Today
# all of them are non-mutating checks/reminders (stop-checks, delegate reminders, screenshot/
# foreground-sim/bash-approval/open-html guards) — none transform output. This guard pins
# that set: a NEW hook command that isn't on the allowlist fails the build, forcing a human
# to look at it and confirm it is a check, not a silent output-rewriter, before allowlisting.
#
# When you legitimately add a hook: add its script basename below WITH a one-word note that
# it is check/reminder-only (never a rewriter). That review IS the point of the guard.
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

SETTINGS=".claude/settings.json"
if [[ ! -f "$SETTINGS" ]]; then
  echo "check-hooks-allowlist: MISCONFIGURED — $SETTINGS not found; refusing vacuous pass" >&2
  exit 1
fi

# Known hook scripts — each verified check/reminder-only (no output rewriting).
readonly ALLOWED=(
  "stop-checks.sh"              # Stop: runs the guard suite, blocks on failure
  "delegate-reminder-hook.py"  # UserPromptSubmit: prints a delegation nudge
  "force-delegate-hook.py"     # PreToolUse: delegation gate
  "check-stray-screenshots.sh" # PreToolUse(Bash): screenshot guard
  "bash-approve-guard.sh"      # PreToolUse(Bash): bash approval gate
  "check-no-foreground-sim.sh" # PreToolUse(Bash): blocks foreground sim runs
  "block-open-html-hook.py"    # PreToolUse(Bash): blocks opening html
)
is_allowed() { local s="$1"; for a in "${ALLOWED[@]}"; do [[ "$s" == "$a" ]] && return 0; done; return 1; }

# Extract every hook command's script basename via a proper JSON parse (python3 is present —
# it runs the hooks themselves).
cmds="$(python3 - "$SETTINGS" <<'PY'
import json, sys, os
d = json.load(open(sys.argv[1]))
for _grp, entries in d.get("hooks", {}).items():
    for e in entries:
        for h in e.get("hooks", []):
            c = h.get("command", "")
            # last token ending in .sh/.py is the script
            tok = [t for t in c.replace('"', ' ').split() if t.endswith(('.sh', '.py'))]
            if tok:
                print(os.path.basename(tok[-1]))
PY
)"

if [[ -z "$cmds" ]]; then
  echo "check-hooks-allowlist: MISCONFIGURED — parsed 0 hook commands from $SETTINGS (format changed?); refusing vacuous pass" >&2
  exit 1
fi

fail=0
while IFS= read -r s; do
  [[ -z "$s" ]] && continue
  if ! is_allowed "$s"; then
    echo "UNKNOWN HOOK: '$s' is declared in $SETTINGS but not on the allowlist."
    echo "  Confirm it is a CHECK/REMINDER (never rewrites output), then add it to ALLOWED in this guard."
    fail=1
  fi
done <<< "$cmds"

exit $fail
