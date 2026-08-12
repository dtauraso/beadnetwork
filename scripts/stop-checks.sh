#!/usr/bin/env bash




set -u










MODE="hook"
if [ "${1:-}" = "--cli" ]; then
  MODE="cli"
fi


CALLER_CWD="$PWD"




emit_block() {
  if [ "$MODE" = "cli" ]; then
    printf 'stop-checks: %s\n' "$1" >&2
    exit 1
  fi
  python3 -c "import json,sys; print(json.dumps({'decision':'block','reason':sys.stdin.read()}))" <<< "$1"
  exit 0
}






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





repo_common="$(git -C "$SCRIPT_DIR" rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)"
caller_common="$(git -C "$CALLER_CWD" rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)"
if [ -z "$caller_common" ] || [ -z "$repo_common" ] || [ "$caller_common" != "$repo_common" ]; then
  emit_block "did NOT run — shell cwd is outside the repo: '$CALLER_CWD' (looks like a scratchpad). cd back to the repo root ('$ROOT') and stop again."
fi

cd "$ROOT" || {
  echo "stop-checks: MISCONFIGURED — cannot cd to repo root '$ROOT'." >&2
  exit 1
}







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








css_changed=$(echo "$changed" | grep -E 'tools/topology-vscode/.*\.css$' || true)

fail=0
out=""

if [ -n "$go_changed" ]; then
  if ! go_out=$(go build ./... 2>&1); then
    out+="go build failed:\n$go_out\n\n"
    fail=1
  fi

















  if ! gotest_out=$(go test ./... 2>&1); then
    out+="go test failed:\n$gotest_out\n\n"
    fail=1
  fi
  # go vet + staticcheck. staticcheck COMPILES the whole module, so it is


  if ! sc_out=$(bash tools/lang/check-staticcheck.sh 2>&1); then
    out+="check-staticcheck failed:\n$sc_out\n\n"
    fail=1
  fi
fi

if [ -n "$ts_changed" ] || [ -n "$css_changed" ]; then
  if [ -n "$ts_changed" ] && ! tsc_out=$(cd tools/topology-vscode && npx --no-install tsc --noEmit 2>&1); then
    out+="tsc --noEmit failed:\n$tsc_out\n\n"
    fail=1
  fi




  webview_out="tools/topology-vscode/out/webview.js"


  bundle_ts_changed=$(printf '%s\n%s' "$(echo "$ts_changed" | grep -v 'tools/topology-vscode/test/' || true)" "$css_changed" | grep -v '^$' || true)
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


  if ! eslint_out=$(bash tools/lang/check-eslint.sh 2>&1); then
    out+="check-eslint failed:\n$eslint_out\n\n"
    fail=1
  fi




fi















#   check-staticcheck / check-eslint / check-vitest — expensive; invoked above under their



GUARD_EXCLUDE="check-staticcheck|check-eslint|check-vitest|check-no-foreground-sim|check-stray-screenshots|check-no-shell-source-edits"

shopt -s nullglob
guards=(tools/*/check-*.sh tools/*/*/check-*.sh)
shopt -u nullglob

if [ ${#guards[@]} -eq 0 ]; then
  echo "stop-checks: MISCONFIGURED — no tools/*/check-*.sh found; refusing to report success." >&2
  exit 1
fi













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

  local gp="$1" gn go grc
  gn=$(basename "$gp" .sh)
  go=$(bash "$gp" 2>&1); grc=$?
  printf '%s' "$go" > "$GDIR/$gn.out"
  printf '%s' "$grc" > "$GDIR/$gn.rc"
}
export -f run_one_guard 2>/dev/null || true


for chk_path in "${guard_selected[@]}"; do
  chk=$(basename "$chk_path" .sh)
  [ "$chk" = "$GUARD_SERIAL" ] || continue
  run_one_guard "$chk_path"
done






printf '%s\n' "${guard_selected[@]}" \
  | grep -v "/${GUARD_SERIAL}\.sh$" \
  | GDIR="$GDIR" xargs -P "$JOBS" -I@ bash -c '
      gp="@"; gn=$(basename "$gp" .sh)
      go=$(bash "$gp" 2>&1); grc=$?
      printf "%s" "$go" > "$GDIR/$gn.out"
      printf "%s" "$grc" > "$GDIR/$gn.rc"
    ' 


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
