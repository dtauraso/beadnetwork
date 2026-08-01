#!/usr/bin/env bash
# Fast deterministic checks. Skips slow test suites. Two output modes (see MODE below):
#   default (hook)  — prints {"decision":"block",...} JSON on stdout and exits 0 (Stop hook).
#   --cli           — prints the failure reason to stderr and exits NONZERO (terminal use;
#                     scripts/verify.sh is the front door). Run `bash scripts/verify.sh`.
set -u

# Output MODE — determined FIRST so the scratchpad guard below can signal per-mode. Two
# callers, two failure-signalling contracts — ONE set of checks below so they can never
# drift (the reason this is a flag, not a second script):
#   - hook (default): the Stop hook wires this in .claude/settings.json. The hook protocol
#     signals failure by printing {"decision":"block",...} JSON on stdout and MUST exit 0;
#     a nonzero exit would be read as a hook error, not a blocked stop.
#   - cli (--cli): for a human/agent at the terminal. Prints the failure reason plainly and
#     EXITS NONZERO, so the reflexive `&& echo ok` / `$?` / `if ...; then` habit is correct
#     here. scripts/verify.sh is the obvious front door to this mode.
MODE="hook"
if [ "${1:-}" = "--cli" ]; then
  MODE="cli"
fi

# Caller's cwd, captured BEFORE any cd — the scratchpad guard compares it to the repo root.
CALLER_CWD="$PWD"

# emit_block REASON — signal a blocked stop per MODE, then exit. hook mode prints the
# Stop-hook JSON on stdout and exits 0 (a nonzero exit would read as a hook error, not a
# blocked stop); cli mode prints the reason to stderr and exits NONZERO.
emit_block() {
  if [ "$MODE" = "cli" ]; then
    printf 'stop-checks: %s\n' "$1" >&2
    exit 1
  fi
  python3 -c "import json,sys; print(json.dumps({'decision':'block','reason':sys.stdin.read()}))" <<< "$1"
  exit 0
}

# Resolve the repo root from the SCRIPT'S OWN location, not the caller's cwd. This script
# always lives at <repo>/scripts/stop-checks.sh, so `git -C "$SCRIPT_DIR"` finds the repo
# regardless of where the shell was when the hook fired. (Assignment, not `cd "$(...)"`:
# `ROOT=$(cmd)` propagates cmd's exit status, so the `|| { ... }` actually fires — the old
# `cd "$(...)" || exit` degraded to `cd ""`, a no-op that never tripped the guard.)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)" || {
  echo "stop-checks: MISCONFIGURED — cannot resolve script directory." >&2
  exit 1
}
ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)" || {
  echo "stop-checks: MISCONFIGURED — 'git -C \"$SCRIPT_DIR\" rev-parse --show-toplevel' failed; cannot locate repo root." >&2
  exit 1
}
if [ -z "$ROOT" ]; then
  echo "stop-checks: MISCONFIGURED — repo root resolved to empty; refusing to run checks against the wrong tree." >&2
  exit 1
fi

# Scratchpad guard (chosen behavior: BLOCK, do NOT run). If the caller's cwd is not inside
# THIS repo — a drifted shell parked in a scratchpad / tmp after background work — refuse to
# run and say so plainly, rather than silently compensating. A subdir of the repo is fine
# (git resolves the same root from anywhere inside); only a cwd OUTSIDE the repo trips this.
repo_common="$(git -C "$SCRIPT_DIR" rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)"
caller_common="$(git -C "$CALLER_CWD" rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)"
if [ -z "$caller_common" ] || [ -z "$repo_common" ] || [ "$caller_common" != "$repo_common" ]; then
  emit_block "did NOT run — shell cwd is outside the repo: '$CALLER_CWD' (looks like a scratchpad). cd back to the repo root ('$ROOT') and stop again."
fi

cd "$ROOT" || {
  echo "stop-checks: MISCONFIGURED — cannot cd to repo root '$ROOT'." >&2
  exit 1
}

# Files to consider for the EXPENSIVE language builds. This is the union of:
#   - the working tree (uncommitted changes), and
#   - committed-but-unmerged work (branch ahead of origin/main, or main if no origin).
# The guard LOOP below runs unconditionally regardless of this set — it is fast
# (sub-second greps) and is the real enforcement; gating it on a dirty tree made
# the whole suite decorative for the normal commit-then-stop workflow.
worktree_changed=$(git status --porcelain 2>/dev/null | awk '{print $NF}')

base=""
if git rev-parse --verify -q origin/main >/dev/null 2>&1; then
  base="origin/main"
elif git rev-parse --verify -q main >/dev/null 2>&1; then
  base="main"
fi
committed_changed=""
if [ -n "$base" ]; then
  committed_changed=$(git diff --name-only "$base"...HEAD 2>/dev/null || true)
fi

changed=$(printf '%s\n%s\n' "$worktree_changed" "$committed_changed")

go_changed=$(echo "$changed" | grep -E '\.go$' || true)
ts_changed=$(echo "$changed" | grep -E 'tools/topology-vscode/.*\.(ts|tsx)$' || true)

fail=0
out=""

if [ -n "$go_changed" ]; then
  if ! go_out=$(go build ./... 2>&1); then
    out+="go build failed:\n$go_out\n\n"
    fail=1
  fi
  # go test — runs on the same gate as go build (Go changed / branch ahead of
  # origin/main). This was the one verify step living outside the suite.
  #
  # -race was REMOVED (2026-07-28). Its original justification named "atomic center
  # snapshots" as the shared memory worth detecting races on. That code is gone:
  # check-no-network-locks.sh now enforces ZERO atomics and ZERO mutexes across
  # nodes/ Buffer/ Trace/ with an empty ALLOWED_ATOMIC set and a rot-check, so the
  # shared-memory bug class -race detects is structurally unrepresentable here —
  # per-mover ownership, dedicated per-pair channels, no shared state.
  #
  # The cost had also drifted 13x from what the original decision was priced on:
  # the comment claimed "4.9s -> 6.3s uncached"; measured 2026-07-28 it was 82s of
  # a ~112s suite, i.e. 73% of every Stop-hook run to check a bug class a guard
  # already makes impossible.
  #
  # If shared state is ever deliberately reintroduced, check-no-network-locks.sh
  # fails first — restore -race here in the same commit that allowlists the atomic.
  if ! gotest_out=$(go test ./... 2>&1); then
    out+="go test failed:\n$gotest_out\n\n"
    fail=1
  fi
  # go vet + staticcheck. staticcheck COMPILES the whole module, so it is
  # expensive — it lives here in the go-gated block (Go changed / branch ahead of
  # origin/main), never in the fast unconditional guard loop below.
  if ! sc_out=$(bash tools/check-staticcheck.sh 2>&1); then
    out+="check-staticcheck failed:\n$sc_out\n\n"
    fail=1
  fi
fi

if [ -n "$ts_changed" ]; then
  if ! tsc_out=$(cd tools/topology-vscode && npx --no-install tsc --noEmit 2>&1); then
    out+="tsc --noEmit failed:\n$tsc_out\n\n"
    fail=1
  fi
  # Rebuild the webview/extension bundle so Cmd-R in the host picks up
  # the latest TS changes without a manual `npm run build`. Skip when
  # out/webview.js is already newer than every changed TS file — avoids
  # paying full esbuild cost on no-op or pure-test edits.
  webview_out="tools/topology-vscode/out/webview.js"
  # Test files don't enter the bundle; skip build when only test/ changed.
  bundle_ts_changed=$(echo "$ts_changed" | grep -v 'tools/topology-vscode/test/' || true)
  need_build=0
  if [ -n "$bundle_ts_changed" ]; then
    need_build=1
    if [ -f "$webview_out" ]; then
      out_mtime=$(stat -f %m "$webview_out" 2>/dev/null || stat -c %Y "$webview_out" 2>/dev/null || echo 0)
      newer=0
      while IFS= read -r f; do
        [ -z "$f" ] && continue
        [ ! -f "$f" ] && continue
        f_mtime=$(stat -f %m "$f" 2>/dev/null || stat -c %Y "$f" 2>/dev/null || echo 0)
        if [ "$f_mtime" -gt "$out_mtime" ]; then newer=1; break; fi
      done <<< "$bundle_ts_changed"
      [ "$newer" -eq 0 ] && need_build=0
    fi
  fi
  if [ "$need_build" -eq 1 ]; then
    if ! build_out=$(cd tools/topology-vscode && npm run --silent build 2>&1); then
      out+="webview build failed:\n$build_out\n\n"
      fail=1
    fi
  fi
  # ESLint (react-hooks correctness guard) — TS-only, needs node_modules, so it
  # lives in the ts_changed block alongside tsc rather than the generic loop.
  if ! eslint_out=$(bash tools/check-eslint.sh 2>&1); then
    out+="check-eslint failed:\n$eslint_out\n\n"
    fail=1
  fi
  # Vitest unit suite (trace-event field contracts, parseSpec, round-trips) —
  # compiles + runs the tests, so EXPENSIVE; lives in the ts_changed block, not
  # the fast unconditional guard loop.
  if ! vitest_out=$(bash tools/check-vitest.sh 2>&1); then
    out+="check-vitest failed:\n$vitest_out\n\n"
    fail=1
  fi
fi

# DISCOVER the guard suite; do not hand-list it. The list used to be hardcoded here, which
# meant a new tools/check-*.sh was simply not run until someone remembered to edit this line
# — a guard nobody invokes is a guard that cannot fail. Globbing makes "written" and "run"
# the same fact.
#
# EXCLUDED, deliberately:
#   check-staticcheck / check-eslint / check-vitest — expensive; invoked above under their
#     language gate, not in this fast unconditional loop.
#   check-no-foreground-sim / check-stray-screenshots — PreToolUse hooks, not checks;
#     they read tool_input JSON from stdin and must always exit 0.
GUARD_EXCLUDE="check-staticcheck|check-eslint|check-vitest|check-no-foreground-sim|check-stray-screenshots"

shopt -s nullglob
guards=(tools/check-*.sh)
shopt -u nullglob

if [ ${#guards[@]} -eq 0 ]; then
  echo "stop-checks: MISCONFIGURED — no tools/check-*.sh found; refusing to report success." >&2
  exit 1
fi

# The guards are ~96% of this script's wall time (5.86s of 6.14s measured), they are 48
# independent read-only checks, and they used to run one at a time. Now they run
# CONCURRENTLY, which bounds the suite at its slowest single guard instead of the sum.
#
# ONE guard must not run concurrently with the others: check-generated.sh RUNS the
# generator (`npm run gen:node-defs`, its line 27), which WRITES the generated files into
# the working tree. Any guard reading those files while it rewrites them could see a
# half-written file and fail at random — a flaky guard suite is worse than a slow one. So
# it runs FIRST, alone, and the rest fan out only after it has finished.
#
# Output is collected per guard and replayed in sorted order, so a failure report reads
# identically to the serial version regardless of which guard finished first.
GUARD_SERIAL="check-generated"
JOBS=$(sysctl -n hw.ncpu 2>/dev/null || nproc 2>/dev/null || echo 4)

guard_selected=()
for chk_path in "${guards[@]}"; do
  chk=$(basename "$chk_path" .sh)
  if echo "$chk" | grep -qE "^($GUARD_EXCLUDE)$"; then continue; fi
  guard_selected+=("$chk_path")
done

GDIR=$(mktemp -d)
trap 'rm -rf "$GDIR"' EXIT

run_one_guard() {
  # $1 = path. Writes <name>.out and <name>.rc so the collector can report deterministically.
  local gp="$1" gn go grc
  gn=$(basename "$gp" .sh)
  go=$(bash "$gp" 2>&1); grc=$?
  printf '%s' "$go" > "$GDIR/$gn.out"
  printf '%s' "$grc" > "$GDIR/$gn.rc"
}
export -f run_one_guard 2>/dev/null || true

# 1. the tree-mutating guard, alone
for chk_path in "${guard_selected[@]}"; do
  chk=$(basename "$chk_path" .sh)
  [ "$chk" = "$GUARD_SERIAL" ] || continue
  run_one_guard "$chk_path"
done

# 2. everything else, concurrently. `xargs -P` rather than backgrounding in batches of
# $JOBS: a batch only advances when its SLOWEST member finishes, so with one 2s guard and
# 47 fast ones the batch containing it idles 7 cores for 2s. xargs keeps every slot busy —
# measured 3.98s (batched) vs 2.66s (streaming) on the same tree.
# bash 3.2 (macOS) has no `wait -n`, which is the other way to do this.
printf '%s\n' "${guard_selected[@]}" \
  | grep -v "/${GUARD_SERIAL}\.sh$" \
  | GDIR="$GDIR" xargs -P "$JOBS" -I@ bash -c '
      gp="@"; gn=$(basename "$gp" .sh)
      go=$(bash "$gp" 2>&1); grc=$?
      printf "%s" "$go" > "$GDIR/$gn.out"
      printf "%s" "$grc" > "$GDIR/$gn.rc"
    ' 

# 3. collect in sorted order — same report the serial loop produced
for chk_path in "${guard_selected[@]}"; do
  chk=$(basename "$chk_path" .sh)
  grc=$(cat "$GDIR/$chk.rc" 2>/dev/null || echo 1)
  if [ "$grc" != "0" ]; then
    chk_out=$(cat "$GDIR/$chk.out" 2>/dev/null || echo "(no output captured)")
    out+="$chk failed:\n$chk_out\n\n"
    fail=1
  fi
done

if [ $fail -ne 0 ]; then
  if [ "$MODE" = "cli" ]; then
    # CLI mode: human-readable reason to stderr, NONZERO exit — the obvious-correct signal.
    printf 'stop-checks: FAILED\n\n%b\n' "$out" >&2
    exit 1
  fi
  # Hook mode: the Stop-hook JSON-on-stdout contract, exit 0 (a nonzero exit here would be
  # read as a hook error, not a blocked stop).
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
