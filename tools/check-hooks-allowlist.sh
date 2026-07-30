#!/usr/bin/env bash
# check-hooks-allowlist.sh — every .claude/settings.json hook command must be a KNOWN
# check/reminder script. Run from repo root: bash tools/check-hooks-allowlist.sh
#
# WHY THIS EXISTS (drift-checklist item #4 — "hidden repair loop"): the checklist asks "does
# a second pass silently rewrite the answer before delivery?" A repo guard can't observe the
# runtime, but every automated pass over a turn is a hook declared in settings.json. This
# guard pins the set: a NEW hook command that isn't on the allowlist fails the build, forcing
# a human to look at it and classify it before allowlisting.
#
# WHAT IS ACTUALLY BANNED is rewriting OUTPUT — a second pass that repairs the answer on its
# way to the reader, so a problem looks solved instead of being surfaced. The reader cannot
# tell the difference between "went right" and "went wrong and got patched", which is what
# makes it drift rather than automation.
#
# Rewriting INPUT is a different act and is ALLOWED, under one condition: the hook must
# DISCLOSE the change (PreToolUse `additionalContext` naming what it altered and why), so the
# rewrite lands in the transcript instead of behind it. A disclosed input-rewrite adds
# information; a silent output-rewrite removes it. This distinction was added deliberately
# when git-runs-in-the-worktree.sh landed — the alternative was a deny-only hook that spent a
# round trip telling the caller to re-issue a command the hook could already fix.
#
# When you legitimately add a hook: add its script basename below WITH a one-word note
# classifying it — check/reminder, or disclosed input-rewrite. Never an output-rewriter.
# That classification IS the point of the guard.
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

# Known hook scripts — each verified to rewrite no OUTPUT. An input-rewriter is marked as
# such and must disclose its change (see the doctrine note above).
readonly ALLOWED=(
  "stop-checks.sh"              # Stop: runs the guard suite, blocks on failure
  "delegate-reminder-hook.py"  # UserPromptSubmit: prints a delegation nudge
  "force-delegate-hook.py"     # PreToolUse: delegation gate
  "check-stray-screenshots.sh" # PreToolUse(Bash): screenshot guard
  "bash-approve-guard.sh"      # PreToolUse(Bash): bash approval gate
  "check-no-foreground-sim.sh" # PreToolUse(Bash): blocks foreground sim runs
  "block-open-html-hook.py"    # PreToolUse(Bash): blocks opening html
  "git-runs-in-the-worktree.sh" # PreToolUse(Bash): DISCLOSED INPUT-REWRITE — prefixes a
                                # tree-less mutating git command with a cd to the task
                                # worktree, and says so; denies when >1 worktree makes the
                                # target ambiguous. Rewrites no output.
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
    echo "  Classify it first — check/reminder, or DISCLOSED input-rewrite (it states what it"
    echo "  changed and why). An OUTPUT-rewriter is never allowed. Then add it to ALLOWED here."
    fail=1
  fi
done <<< "$cmds"

# ---------------------------------------------------------------------------
# GIT hooks — the other place an automated pass can run
# ---------------------------------------------------------------------------
# settings.json hooks are Claude's; .githooks/ holds git's. Both are automated passes over
# work in flight, so both belong to this guard — otherwise adding a git hook would create an
# enforcement mechanism that the repo's own hook audit cannot see.
#
# The failure this catches is SILENCE, not misconfiguration: .git/hooks is not
# version-controlled and core.hooksPath is per-clone local config, so a tracked hook that
# git was never pointed at is inert while looking installed. That is worse than no hook —
# it reads as coverage that does not exist.
HOOKS_DIR=".githooks"
readonly EXPECTED_GIT_HOOKS=(
  "pre-push"   # runs scripts/verify.sh; blocks the push on failure
)

for h in "${EXPECTED_GIT_HOOKS[@]}"; do
  if [[ ! -f "$HOOKS_DIR/$h" ]]; then
    echo "MISSING GIT HOOK: $HOOKS_DIR/$h is expected by this guard but does not exist."
    echo "  Either restore it or drop it from EXPECTED_GIT_HOOKS in this guard."
    fail=1
  elif [[ ! -x "$HOOKS_DIR/$h" ]]; then
    echo "GIT HOOK NOT EXECUTABLE: $HOOKS_DIR/$h — git silently ignores a non-executable hook."
    echo "  Fix: chmod +x $HOOKS_DIR/$h"
    fail=1
  fi
done

# The hook only runs if git has been pointed at the tracked directory. Checked with
# --local so a stray global/system setting cannot satisfy this for a clone that lacks it.
configured="$(git config --local --get core.hooksPath 2>/dev/null || true)"
if [[ "$configured" != "$HOOKS_DIR" ]]; then
  echo "GIT HOOKS NOT INSTALLED: core.hooksPath is '${configured:-<unset>}', expected '$HOOKS_DIR'."
  echo "  The tracked hooks in $HOOKS_DIR/ are INERT until git is told to use them:"
  echo "      git config core.hooksPath $HOOKS_DIR"
  echo "  (per-clone local config, so every fresh clone needs it once)"
  fail=1
fi

exit $fail
