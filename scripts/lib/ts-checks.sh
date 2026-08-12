#!/usr/bin/env bash

# Sourced by scripts/stop-checks.sh. Reads/writes the orchestrator's globals
# ($ts_changed, $css_changed, $out, $fail) directly.

run_ts_checks() {
  if [ -z "$ts_changed" ] && [ -z "$css_changed" ]; then
    return
  fi

  if [ -n "$ts_changed" ] && ! tsc_out=$(cd tools/topology-vscode && npx --no-install tsc --noEmit 2>&1); then
    out+="tsc --noEmit failed:\n$tsc_out\n\n"
    fail=1
  fi

  local webview_out="tools/topology-vscode/out/webview.js"

  local bundle_ts_changed need_build
  bundle_ts_changed=$(printf '%s\n%s' "$(echo "$ts_changed" | grep -v 'tools/topology-vscode/test/' || true)" "$css_changed" | grep -v '^$' || true)
  need_build=0
  if [ -n "$bundle_ts_changed" ]; then
    need_build=1
    if [ -f "$webview_out" ]; then
      local out_mtime newer f f_mtime
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
}
