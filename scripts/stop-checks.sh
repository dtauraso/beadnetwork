#!/usr/bin/env bash

set -u

MODE="hook"
if [ "${1:-}" = "--cli" ]; then
  MODE="cli"
fi

CALLER_CWD="$PWD"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)" || {
  echo "stop-checks: MISCONFIGURED — cannot resolve script directory." >&2
  exit 1
}
LIB_DIR="$SCRIPT_DIR/lib"

# Each lib below is sourced (not exec'd) into THIS shell on purpose: the phases share
# $MODE/$ROOT/$out/$fail as plain globals, and the always-exit-0 JSON-block protocol below
# has to run in the one process that decides pass/fail.
source "$LIB_DIR/env.sh"
source "$LIB_DIR/changed-files.sh"
source "$LIB_DIR/go-checks.sh"
source "$LIB_DIR/ts-checks.sh"
source "$LIB_DIR/guards.sh"

resolve_repo_root_or_block
cd_to_root_or_die

changed="$(collect_changed_files)"
go_changed=$(echo "$changed" | grep -E '\.go$' || true)
ts_changed=$(echo "$changed" | grep -E 'tools/topology-vscode/.*\.(ts|tsx)$' || true)
css_changed=$(echo "$changed" | grep -E 'tools/topology-vscode/.*\.css$' || true)

fail=0
out=""

run_go_checks
run_ts_checks
run_guards

if [ $fail -ne 0 ]; then
  if [ "$MODE" = "cli" ]; then

    printf 'stop-checks: FAILED\n\n%b\n' "$out" >&2
    exit 1
  fi

  python3 -c "
import json, sys
reason = 'Pre-stop checks failed. Fix before stopping:\n\n' + sys.stdin.read()
print(json.dumps({'decision': 'block', 'reason': reason}))
" <<< "$(printf '%b' "$out")"
  exit 0
fi

if [ "$MODE" = "cli" ]; then
  echo "stop-checks: clean" >&2
fi
exit 0
