#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
PROBE_DIR="$REPO_ROOT/.probe"

GO_FILE="$PROBE_DIR/go.jsonl"

OWNER_FILES=()
for owner in node edge interior bead; do
  for f in "$PROBE_DIR/$owner"/*.jsonl; do
    [[ -f "$f" ]] && OWNER_FILES+=("$f")
  done
done

GO_ERR_FILE="$PROBE_DIR/go-errors.jsonl"
TS_FILE="$PROBE_DIR/ts.jsonl"
TS_ERR_FILE="$PROBE_DIR/ts-errors.jsonl"

read_file() {
  local f="$1"
  if [[ -f "$f" ]]; then
    cat "$f"
  fi
}

merge_and_sort() {

  local files=("$@")
  {
    for f in "${files[@]}"; do
      read_file "$f"
    done
  } | jq -s 'sort_by(.ts_ms) | .[]' -c
}

MODE="${1:-}"

case "$MODE" in
  --errors)
    merge_and_sort "$GO_ERR_FILE" "$TS_ERR_FILE"
    ;;
  --step)
    STEP="${2:?Usage: probe-merge.sh --step N}"
    {
      read_file "$GO_FILE"
      for f in "${OWNER_FILES[@]:-}"; do [[ -n "$f" ]] && read_file "$f"; done
      read_file "$GO_ERR_FILE"
      read_file "$TS_FILE"
      read_file "$TS_ERR_FILE"
    } | jq -s --argjson step "$STEP" '[.[] | select(.step == $step)] | sort_by(.ts_ms) | .[]' -c
    ;;
  --go)
    merge_and_sort "$GO_FILE" "${OWNER_FILES[@]:-}" "$GO_ERR_FILE"
    ;;
  --debug)

    {
      read_file "$GO_FILE"
      for f in "${OWNER_FILES[@]:-}"; do [[ -n "$f" ]] && read_file "$f"; done
    } | jq -s '[.[] | select(.kind == "breadcrumb" and .debug == true)] | sort_by(.ts_ms) | .[]' -c
    ;;
  --ts)
    merge_and_sort "$TS_FILE" "$TS_ERR_FILE"
    ;;
  "")
    merge_and_sort "$GO_FILE" "${OWNER_FILES[@]:-}" "$GO_ERR_FILE" "$TS_FILE" "$TS_ERR_FILE"
    ;;
  *)
    echo "Unknown option: $MODE" >&2
    echo "Usage: probe-merge.sh [--errors | --step N | --go | --debug | --ts]" >&2
    exit 1
    ;;
esac
