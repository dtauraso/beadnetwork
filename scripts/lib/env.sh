#!/usr/bin/env bash

# Sourced by scripts/stop-checks.sh. Shares its shell: no set -e/-u here, no subshell —
# a library option change would leak into the orchestrator that sources it.

emit_block() {
  if [ "$MODE" = "cli" ]; then
    printf 'stop-checks: %s\n' "$1" >&2
    exit 1
  fi
  python3 -c "import json,sys; print(json.dumps({'decision':'block','reason':sys.stdin.read()}))" <<< "$1"
  exit 0
}

resolve_repo_root_or_block() {
  ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)" || {
    echo "stop-checks: MISCONFIGURED — 'git -C \"$SCRIPT_DIR\" rev-parse --show-toplevel' failed; cannot locate repo root." >&2
    exit 1
  }
  if [ -z "$ROOT" ]; then
    echo "stop-checks: MISCONFIGURED — repo root resolved to empty; refusing to run checks against the wrong tree." >&2
    exit 1
  fi

  repo_common="$(git -C "$SCRIPT_DIR" rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)"
  caller_common="$(git -C "$CALLER_CWD" rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)"
  if [ -z "$caller_common" ] || [ -z "$repo_common" ] || [ "$caller_common" != "$repo_common" ]; then
    emit_block "did NOT run — shell cwd is outside the repo: '$CALLER_CWD' (looks like a scratchpad). cd back to the repo root ('$ROOT') and stop again."
  fi
}

cd_to_root_or_die() {
  cd "$ROOT" || {
    echo "stop-checks: MISCONFIGURED — cannot cd to repo root '$ROOT'." >&2
    exit 1
  }
}
