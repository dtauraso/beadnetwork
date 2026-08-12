#!/usr/bin/env bash

# PLACEMENT: .claude/settings.json,.githooks/* | a new hook command must be classified and added to the ALLOWED list here

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

SETTINGS=".claude/settings.json"
if [[ ! -f "$SETTINGS" ]]; then
  echo "check-hooks-allowlist: MISCONFIGURED — $SETTINGS not found; refusing vacuous pass" >&2
  exit 1
fi

readonly ALLOWED=(
  "stop-checks.sh"
  "check-stray-screenshots.sh"
  "bash-approve-guard.sh"
  "check-no-foreground-sim.sh"
  "block-open-html-hook.py"
  "check-no-shell-source-edits.sh"

  "placement-brief-hook.sh"
)
is_allowed() { local s="$1"; for a in "${ALLOWED[@]}"; do [[ "$s" == "$a" ]] && return 0; done; return 1; }

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

HOOKS_DIR=".githooks"
readonly EXPECTED_GIT_HOOKS=(
  "pre-push"
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

configured="$(git config --local --get core.hooksPath 2>/dev/null || true)"
if [[ "$configured" != "$HOOKS_DIR" ]]; then
  echo "GIT HOOKS NOT INSTALLED: core.hooksPath is '${configured:-<unset>}', expected '$HOOKS_DIR'."
  echo "  The tracked hooks in $HOOKS_DIR/ are INERT until git is told to use them:"
  echo "      git config core.hooksPath $HOOKS_DIR"
  echo "  (per-clone local config, so every fresh clone needs it once)"
  fail=1
fi

exit $fail
