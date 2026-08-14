#!/usr/bin/env bash

# Sourced by scripts/stop-checks.sh. Reads/writes the orchestrator's globals
# ($ts_changed, $css_changed, $out, $fail) directly.

webview_bundle_stale() {
  local webview_out="tools/topology-vscode/out/webview.js"
  local src_dir="tools/topology-vscode/src"

  if [ ! -d "$src_dir" ]; then
    out+="ts-checks: MISCONFIGURED — webview source dir not found: $src_dir\n\n"
    fail=1
    return 1
  fi

  [ -f "$webview_out" ] || return 0

  local newer
  newer=$(find "$src_dir" \( -name '*.ts' -o -name '*.tsx' -o -name '*.css' \) \
    -newer "$webview_out" -print -quit 2>/dev/null)
  [ -n "$newer" ]
}

run_ts_checks() {
  if webview_bundle_stale; then
    if ! build_out=$(cd tools/topology-vscode && npm run --silent build 2>&1); then
      out+="webview build failed:\n$build_out\n\n"
      fail=1
    fi
  fi

  if [ -z "$ts_changed" ] && [ -z "$css_changed" ]; then
    return
  fi

  if [ -n "$ts_changed" ] && ! tsc_out=$(cd tools/topology-vscode && npx --no-install tsc --noEmit 2>&1); then
    out+="tsc --noEmit failed:\n$tsc_out\n\n"
    fail=1
  fi

  if ! eslint_out=$(bash tools/lang/check-eslint.sh 2>&1); then
    out+="check-eslint failed:\n$eslint_out\n\n"
    fail=1
  fi
}
