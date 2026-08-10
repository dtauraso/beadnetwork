#!/usr/bin/env bash
# check-no-shell-source-edits.sh — PreToolUse(Bash) guard.
#
# Source files are edited with the Edit/Write TOOLS, never by a shell command. A shell
# write — `python3 - <<'PY'`, `sed -i`, `perl -pi`, `cat > file.ts` — reaches the same bytes
# but bypasses everything the tool path provides:
#
#   * placement-brief-hook.sh (PreToolUse Write|Edit) never fires, so a NEW file is created
#     without the brief naming the guard rules that will apply to it;
#   * the harness's read-before-edit and file-state tracking are skipped, so an edit can
#     land on content that has changed underneath it;
#   * memory/feedback_hook_block_means_stop.md names python3/sed -i/shell redirect
#     EXPLICITLY as the route-around that must not be taken when a write hook says no.
#     A guard that only fires on Write|Edit is not a guard if Bash can write the file.
#
# This was written after a session in which several multi-part edits went through
# `python3 - <<'PY'` heredocs purely because they touched one file in several places —
# which is what the Edit tool does natively, so the shell bought nothing and lost the above.
#
# SCOPE — narrow on purpose. It fires only when BOTH are true:
#   1. the command uses a WRITE MECHANISM (in-place editor, redirect, or tee), and
#   2. the target is a SOURCE PATH — a repo-relative path with a source extension.
# Reading, grepping, listing, `go run` of a generator (a program writing its own output is
# not a shell write), and anything under the exempt trees below are allowed silently.
#
# NO ESCAPE HATCH, deliberately. Every previous "just this once" category in this repo
# became a standing exemption. If a shell write is genuinely the only way, the user runs it
# — that is the review step. See CLAUDE.md's landing rule for the same reasoning applied to
# merges.
#
# Exit 0 always; the decision is carried in the emitted JSON, same protocol as
# check-no-foreground-sim.sh.
#
# PLACEMENT: none | this is a PreToolUse(Bash) hook gating shell commands, not a source-file guard
set -uo pipefail

input="$(cat)"
cmd="$(printf '%s' "$input" | jq -r '.tool_input.command // empty')"

emit() { # $1=allow|deny  $2=reason
  jq -nc --arg d "$1" --arg r "$2" \
    '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:$d,permissionDecisionReason:$r}}'
}

if [ -z "$cmd" ]; then emit allow "no command string"; exit 0; fi

# --- 1. Does the command carry a write mechanism? -----------------------------------
#
# CMD_HEAD anchors an invocation to COMMAND POSITION (line start, or after a shell
# separator) for the same reason check-no-foreground-sim.sh does: matching a bare word
# anywhere in the string turns every MENTION of a tool — in prose, a path, a grep pattern —
# into a match. Redirect and tee are matched anywhere, since a redirect is by nature not at
# the head of a command.
CMD_HEAD='(^|[;&|(])[[:space:]]*'

# In-place editors and interpreters fed a script inline. `python3`/`perl`/`ruby`/`node` only
# count with -c/-e/-i or a heredoc, so `python3 script.py` (running a checked-in program)
# stays allowed.
INPLACE_RE="${CMD_HEAD}sed[[:space:]]+(-[[:alnum:]]*[[:space:]]+)*-i|${CMD_HEAD}(perl|ruby)[[:space:]]+-[[:alnum:]]*i|${CMD_HEAD}(python3?|perl|ruby|node)[[:space:]]+(-[[:alnum:]]+[[:space:]]+)*(-[ce]|-[[:alnum:]]*[ce][[:space:]])|<<[[:space:]]*['\"]?[A-Za-z_]"

# A redirect or tee onto a path. `>&`, `2>&1`, and `>/dev/null` are not file writes.
REDIRECT_RE='>>?[[:space:]]*[^&[:space:]]|(^|[;&|(])[[:space:]]*tee[[:space:]]'

# An inline script or heredoc is not a write by itself — it is just as often reading,
# parsing, or carrying test data that happens to MENTION a path. It counts as a write only
# when the script text also contains a write CALL. Without this, every heredoc naming a
# source path was denied, including the read-only ones (the first version of this guard
# blocked its own test harness that way). `sed -i`/`perl -i` need no such evidence: the
# in-place flag IS the write.
WRITE_CALL_RE="\.write[[:alnum:]_]*\(|open\([^)]*['\"][wa]|write_text\(|writeFile|\bfputs\b|\bprint\([^)]*file="

mechanism=""
if printf '%s' "$cmd" | grep -Eq "${CMD_HEAD}sed[[:space:]]+(-[[:alnum:]]*[[:space:]]+)*-i|${CMD_HEAD}(perl|ruby)[[:space:]]+-[[:alnum:]]*i"; then
  mechanism="an in-place edit"
fi
if [ -z "$mechanism" ] \
  && printf '%s' "$cmd" | grep -Eq "$INPLACE_RE" \
  && printf '%s' "$cmd" | grep -Eq "$WRITE_CALL_RE"; then
  mechanism="an inline script that writes a file"
fi
# A redirect is judged on its TARGET ALONE, never on the whole command. Scanning the whole
# string made `bash scripts/stop-checks.sh > /dev/null` read as a write to a .sh file — the
# path matched was the script being RUN, on the far side of the `>`. Only what the arrow
# points at can be written.
targets="$(printf '%s' "$cmd" | grep -Eo '>>?[[:space:]]*[^&[:space:]]+|(^|[;&|(])[[:space:]]*tee[[:space:]]+(-a[[:space:]]+)?[^[:space:]]+' | sed -E 's/^.*(>|tee)[[:space:]]*//; s/^-a[[:space:]]+//')"
if [ -z "$mechanism" ] && [ -n "$targets" ]; then mechanism="a redirect"; fi
if [ -z "$mechanism" ]; then emit allow "no shell write mechanism"; exit 0; fi

# The text the source-path match runs against: for a redirect that is the targets only; for
# an in-place edit or inline script it is the whole command, since the path being written
# is an argument rather than a redirect target.
scan="$cmd"
if [ "$mechanism" = "a redirect" ]; then scan="$targets"; fi

# --- 2. Does it name a source path? --------------------------------------------------
#
# Exempt trees FIRST: scratch space, build output, logs and vendored code are not source,
# and writing them from the shell is ordinary. Stripping them before the source match means
# a command that only touches exempt paths cannot match on them.
stripped="$(printf '%s' "$scan" \
  | sed -E 's#(/private)?/tmp/[^[:space:]"'"'"']*##g' \
  | sed -E 's#[^[:space:]"'"'"']*/(node_modules|\.probe|handoff-archive)/[^[:space:]"'"'"']*##g' \
  | sed -E 's#[^[:space:]"'"'"']*/?out/[^[:space:]"'"'"']*##g' \
  | sed -E 's#/dev/(null|stderr|stdout)##g')"

# Source extensions this repo actually holds. A path is only a source path if it carries
# one — a bare name with no extension is a binary, a directory, or a log.
SOURCE_RE='[^[:space:]"'"'"'<>|;&()]+\.(go|ts|tsx|js|jsx|css|html|json|md|sh|py|mod|sum)([[:space:]"'"'"'<>|;&)]|$)'

if ! printf '%s' "$stripped" | grep -Eq "$SOURCE_RE"; then
  emit allow "write mechanism, but no source path targeted"; exit 0
fi

hit="$(printf '%s' "$stripped" | grep -Eo "$SOURCE_RE" | head -1 | tr -d '[:space:]')"

emit deny "Shell write to a source file blocked ($mechanism, targeting ${hit}). Edit source with the Edit/Write TOOLS — a shell write skips placement-brief-hook.sh, skips read-before-edit tracking, and is the exact route-around memory/feedback_hook_block_means_stop.md names. The Edit tool handles several changes to one file natively; make several Edit calls rather than one heredoc."
exit 0
