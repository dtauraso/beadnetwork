#!/usr/bin/env bash

# PLACEMENT: none | this is a PreToolUse(Bash) hook gating shell commands, not a source-file guard
set -uo pipefail

input="$(cat)"
cmd="$(printf '%s' "$input" | jq -r '.tool_input.command // empty')"

emit() { # $1=allow|deny  $2=reason
  jq -nc --arg d "$1" --arg r "$2" \
    '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:$d,permissionDecisionReason:$r}}'
}

if [ -z "$cmd" ]; then emit allow "no command string"; exit 0; fi

CMD_HEAD='(^|[;&|(])[[:space:]]*'

INPLACE_RE="${CMD_HEAD}sed[[:space:]]+(-[[:alnum:]]*[[:space:]]+)*-i|${CMD_HEAD}(perl|ruby)[[:space:]]+-[[:alnum:]]*i|${CMD_HEAD}(python3?|perl|ruby|node)[[:space:]]+(-[[:alnum:]]+[[:space:]]+)*(-[ce]|-[[:alnum:]]*[ce][[:space:]])|<<[[:space:]]*['\"]?[A-Za-z_]"

REDIRECT_RE='>>?[[:space:]]*[^&[:space:]]|(^|[;&|(])[[:space:]]*tee[[:space:]]'

WRITE_CALL_RE="\.write[[:alnum:]_]*\(|open\([^)]*['\"][wa]|write_text\(|writeFile|\bfputs\b|\bprint\([^)]*file="

mechanism=""

if [ -z "$mechanism" ] \
  && printf '%s' "$cmd" | grep -Eq "$INPLACE_RE" \
  && printf '%s' "$cmd" | grep -Eq "$WRITE_CALL_RE"; then
  mechanism="an inline script that writes a file"
fi
targets="$(printf '%s' "$cmd" | grep -Eo '>>?[[:space:]]*[^&[:space:]]+|(^|[;&|(])[[:space:]]*tee[[:space:]]+(-a[[:space:]]+)?[^[:space:]]+' | sed -E 's/^.*(>|tee)[[:space:]]*//; s/^-a[[:space:]]+//')"
if [ -z "$mechanism" ] && [ -n "$targets" ]; then mechanism="a redirect"; fi
if [ -z "$mechanism" ]; then emit allow "no shell write mechanism"; exit 0; fi

scan="$cmd"
if [ "$mechanism" = "a redirect" ]; then scan="$targets"; fi

stripped="$(printf '%s' "$scan" \
  | sed -E 's#(/private)?/tmp/[^[:space:]"'"'"']*##g' \
  | sed -E 's#[^[:space:]"'"'"']*/(node_modules|\.probe|handoff-archive)/[^[:space:]"'"'"']*##g' \
  | sed -E 's#[^[:space:]"'"'"']*/?out/[^[:space:]"'"'"']*##g' \
  | sed -E 's#/dev/(null|stderr|stdout)##g')"

SOURCE_RE='[^[:space:]"'"'"'<>|;&()]+\.(go|ts|tsx|js|jsx|css|html|json|md|sh|py|mod|sum)([[:space:]"'"'"'<>|;&)]|$)'

if ! printf '%s' "$stripped" | grep -Eq "$SOURCE_RE"; then
  emit allow "write mechanism, but no source path targeted"; exit 0
fi

hit="$(printf '%s' "$stripped" | grep -Eo "$SOURCE_RE" | head -1 | tr -d '[:space:]')"

emit deny "Shell write to a source file blocked ($mechanism, targeting ${hit}). Edit source with the Edit/Write TOOLS — a shell write skips placement-brief-hook.sh, skips read-before-edit tracking, and is the exact route-around memory/feedback/process/guardrails/feedback_hook_block_means_stop.md names. The Edit tool handles several changes to one file natively; make several Edit calls rather than one heredoc."
exit 0
