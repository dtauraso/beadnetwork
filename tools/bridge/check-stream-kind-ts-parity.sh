#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: src/Buffer/streamframe/stream_fds.go,src/runCommand.ts,src/runner/stream-fds.ts,src/runner/spawn-layout.ts,src/runner/stream-demux.ts | a new StreamKind must gain a WIREFOLD_STREAM_FDS env entry (runner/spawn-layout.ts builds the string, runCommand.ts's spawn env assigns it) and its own handle<Kind>Fd reader (runner/stream-demux.ts) in the ext host

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

TS_SRC="src"

if [ ! -d "$TS_SRC" ]; then
  echo "check-stream-kind-ts-parity: MISCONFIGURED — ext-host source dir not found: $TS_SRC"
  exit 1
fi

GO_KIND_FILES=$(grep -rlE '^const StreamKind[A-Za-z]+ = "' --include='*.go' . \
  --exclude-dir=node_modules --exclude-dir=out --exclude-dir=.git 2>/dev/null || true)

if [ -z "$GO_KIND_FILES" ]; then
  echo "check-stream-kind-ts-parity: MISCONFIGURED — no file declares 'const StreamKind<X> = \"…\"'."
  echo "The declaration form or its home moved; this guard is scanning nothing rather than"
  echo "passing vacuously. Repoint its pattern."
  exit 1
fi

KINDS=$(grep -hoE '^const StreamKind[A-Za-z]+ = "[a-zA-Z]+"' $GO_KIND_FILES \
  | grep -oE '"[a-zA-Z]+"' | tr -d '"' | sort -u || true)

if [ -z "$KINDS" ]; then
  echo "check-stream-kind-ts-parity: MISCONFIGURED — found $GO_KIND_FILES but extracted no kind values."
  exit 1
fi

ENV_FILES=$(grep -rl 'WIREFOLD_STREAM_FDS:' --include='*.ts' "$TS_SRC" 2>/dev/null || true)
if [ -z "$ENV_FILES" ]; then
  echo "check-stream-kind-ts-parity: MISCONFIGURED — nothing under $TS_SRC assigns"
  echo "WIREFOLD_STREAM_FDS: in a spawn env. The ext host's fd-allocation site moved;"
  echo "repoint this guard."
  exit 1
fi

PATTERN_FILES="$TS_SRC"

HITS=0
for k in $KINDS; do

  envpat="${k}:\\\$\\{[A-Za-z0-9_.]*[Ff][Dd]\\}"
  if ! grep -rhoE "$envpat" --include='*.ts' "$PATTERN_FILES" >/dev/null 2>&1; then
    echo "check-stream-kind-ts-parity: stream kind \"$k\" is declared in Go but the ext host"
    echo "  never spells it into WIREFOLD_STREAM_FDS (looked for \`$k:\${…Fd}\` under: $PATTERN_FILES)."
    echo "  Go will find no base fd for it and every emitter of that kind will silently write"
    echo "  nowhere. Allocate a base fd and push the entry."
    HITS=$((HITS + 1))
  fi

  cap="$(printf '%s' "$k" | awk '{print toupper(substr($0,1,1)) substr($0,2)}')"
  if ! grep -rqE "^[[:space:]]*(private |public |protected |async )*handle${cap}Fd[[:space:]]*\(" --include='*.ts' "$TS_SRC" 2>/dev/null; then
    echo "check-stream-kind-ts-parity: stream kind \"$k\" is declared in Go but no"
    echo "  handle${cap}Fd( reader METHOD IS DEFINED anywhere under $TS_SRC (a call site alone"
    echo "  does not count — see this check's own comment)."
    echo "  The pipe would be allocated and never read — frames pile up in the pipe buffer and"
    echo "  the entity class never appears in the editor. Add the per-kind handler."
    HITS=$((HITS + 1))
  fi
done

if [ "$HITS" -ne 0 ]; then
  echo ""
  echo "check-stream-kind-ts-parity: $HITS divergence(s) found"
  exit 1
fi

exit 0
