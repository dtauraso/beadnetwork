#!/usr/bin/env bash

# PLACEMENT: none | this is a PreToolUse(Bash) hook gating shell commands, not a source-file guard
set -uo pipefail

input="$(cat)"
cmd="$(printf '%s' "$input" | jq -r '.tool_input.command // empty')"
bg="$(printf '%s' "$input" | jq -r '.tool_input.run_in_background // false')"

emit() {
  jq -nc --arg d "$1" --arg r "$2" \
    '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:$d,permissionDecisionReason:$r}}'
}

if [ -z "$cmd" ]; then emit allow "no command string"; exit 0; fi

CMD_HEAD='(^|[;&|(])[[:space:]]*'
SIM_RE="${CMD_HEAD}(\./)?wirefold([[:space:]]|\$)|${CMD_HEAD}go[[:space:]]+run[[:space:]]+(\./?|github\.com/dtauraso/wirefold)([[:space:]]|\$)"
if ! printf '%s' "$cmd" | grep -Eq "$SIM_RE"; then
  emit allow "not a sim run"; exit 0
fi

if [ "$bg" = "true" ]; then emit allow "sim run is harness-backgrounded"; exit 0; fi
if printf '%s' "$cmd" | grep -Eq 'run-bounded\.sh'; then emit allow "sim run is bounded"; exit 0; fi
if printf '%s' "$cmd" | grep -Eq '&[[:space:]]*$'; then emit allow "sim run is shell-backgrounded"; exit 0; fi

emit deny "Foreground sim run blocked (memory/feedback/process/guardrails/feedback_no_foreground_sim_runs.md): the sim can fail to exit and will hang the Bash call. Re-run backgrounded (run_in_background=true), or wrap it: scripts/run-bounded.sh <seconds> <cmd…>."
exit 0
