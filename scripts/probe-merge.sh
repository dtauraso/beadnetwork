#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
PROBE_DIR="$REPO_ROOT/.probe"

GO_FILE="$PROBE_DIR/go.log"

OWNER_FILES=()
TRACE_BINS=()
while IFS= read -r f; do
  [[ -n "$f" ]] && TRACE_BINS+=("$f")
done < <(find "$REPO_ROOT"/topology*/view -name 'trace.bin' -o -name '*-trace.bin' 2>/dev/null | sort)

if (( ${#TRACE_BINS[@]} > 0 )); then
  DECODED_DIR="$(mktemp -d)"
  trap 'rm -rf "$DECODED_DIR"' EXIT
  for f in "${TRACE_BINS[@]}"; do
    rel="${f#"$REPO_ROOT"/}"
    out="$DECODED_DIR/$(echo "${rel%.bin}" | tr '/' '-').log"
    case "$f" in
      */interior-trace.bin) reader="Categories/Node/Interior/readtrace" ;;
      */beads-trace.bin)    reader="Categories/Node/BeadAnimation/readtrace" ;;
      */view/trace.bin)     reader="Categories/Scene/viewstate/readtrace" ;;
      *)                    reader="Categories/Node/nodeactor/owners/readtrace" ;;
    esac
    if go run "$REPO_ROOT/$reader" "$f" > "$out" 2>/dev/null && [[ -s "$out" ]]; then
      OWNER_FILES+=("$out")
    fi
  done
fi

GO_ERR_FILE="$PROBE_DIR/go-errors.log"
TS_FILE="$PROBE_DIR/ts.log"
TS_ERR_FILE="$PROBE_DIR/ts-errors.log"

sort_by_ts() {
  sed -E 's/^ts_ms=([0-9]+)/\1\'$'\t''&/' | sort -n -k1,1 | cut -f2-
}

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
  } | sort_by_ts
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
    } | grep -E "(^| )step=$STEP( |\$)" | sort_by_ts
    ;;
  --go)
    merge_and_sort "$GO_FILE" "${OWNER_FILES[@]:-}" "$GO_ERR_FILE"
    ;;
  --debug)

    {
      read_file "$GO_FILE"
      for f in "${OWNER_FILES[@]:-}"; do [[ -n "$f" ]] && read_file "$f"; done
    } | grep -E "(^| )kind=breadcrumb( |\$)" | grep -E "(^| )debug=true( |\$)" | sort_by_ts
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
