#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROBE_DIR="$REPO_ROOT/.probe"

GO_FILE="$PROBE_DIR/go.jsonl"
GO_NODE_FILE="$PROBE_DIR/go-node.jsonl"
GO_EDGE_FILE="$PROBE_DIR/go-edge.jsonl"
GO_INTERIOR_FILE="$PROBE_DIR/go-interior.jsonl"
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
      read_file "$GO_NODE_FILE"
      read_file "$GO_EDGE_FILE"
      read_file "$GO_INTERIOR_FILE"
      read_file "$GO_ERR_FILE"
      read_file "$TS_FILE"
      read_file "$TS_ERR_FILE"
    } | jq -s --argjson step "$STEP" '[.[] | select(.step == $step)] | sort_by(.ts_ms) | .[]' -c
    ;;
  --go)
    merge_and_sort "$GO_FILE" "$GO_NODE_FILE" "$GO_EDGE_FILE" "$GO_INTERIOR_FILE" "$GO_ERR_FILE"
    ;;
  --debug)

    {
      read_file "$GO_FILE"
      read_file "$GO_NODE_FILE"
      read_file "$GO_EDGE_FILE"
      read_file "$GO_INTERIOR_FILE"
    } | jq -s '[.[] | select(.kind == "breadcrumb" and .debug == true)] | sort_by(.ts_ms) | .[]' -c
    ;;
  --ts)
    merge_and_sort "$TS_FILE" "$TS_ERR_FILE"
    ;;
  "")
    merge_and_sort "$GO_FILE" "$GO_NODE_FILE" "$GO_EDGE_FILE" "$GO_INTERIOR_FILE" "$GO_ERR_FILE" "$TS_FILE" "$TS_ERR_FILE"
    ;;
  *)
    echo "Unknown option: $MODE" >&2
    echo "Usage: probe-merge.sh [--errors | --step N | --go | --debug | --ts]" >&2
    exit 1
    ;;
esac
