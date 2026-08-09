#!/usr/bin/env bash
set -euo pipefail

# check-stream-kind-ts-parity.sh — a stream KIND declared in Go must exist on the TS side.
#
# PLACEMENT: Buffer/stream_fds.go,tools/topology-vscode/src/runCommand.ts,tools/topology-vscode/src/runner/stream-fds.ts,tools/topology-vscode/src/runner/stream-demux.ts | a new StreamKind must gain a WIREFOLD_STREAM_FDS env entry (runCommand.ts's spawn env) and its own handle<Kind>Fd reader (runner/stream-demux.ts) in the ext host
#
# THE BUG THIS EXISTS FOR. The "one inherited stdio pipe per emitting goroutine" transport
# is agreed BY POSITION and BY NAME, with no runtime negotiation (Buffer/stream_fds.go's
# header): the ext host computes a base fd per stream KIND, spells the kind name into
# WIREFOLD_STREAM_FDS, and attaches one reader per fd. Go then looks its own kind up in
# that env var and, when the entry is ABSENT, silently no-ops — every mover of that kind
# keeps a nil stream and emits nothing (that silent-skip path is what
# check-stream-fd-mismatch-reported.sh forces main.go to have a VOICE for).
#
# So a SIXTH stream kind added in Go, with the ext host never taught to allocate a base fd
# for it or to read its pipe, produces a build that compiles, tests that pass, and a whole
# emitting-goroutine class that writes to nothing. Today the only thing standing between
# that and the tree is prose. This guard is the enforcement.
#
# WHAT IT ASSERTS, for every `const StreamKind<X> = "<v>"` declared on the Go side:
#   1. the ext host spells "<v>" into the WIREFOLD_STREAM_FDS env string it builds — matched
#      as a template-literal entry `<v>:${…Fd}` / `<v>:${…FD}`, i.e. the interpolated
#      expression must NAME an fd. (That trailing-Fd requirement is what keeps unrelated
#      same-shaped cache keys in the same file — `edge:${row}`, `interior:${row}` — from
#      satisfying it.)
#   2. a per-kind reader method `handle<X>Fd(` exists somewhere under the ext-host src tree,
#      where <X> is the kind value with its first letter capitalised (view→handleViewFd,
#      interior→handleInteriorFd, drive→handleDriveFd).
#
# WHAT IT DELIBERATELY DOES NOT ASSERT, and why — read this before trusting it as coverage:
#   - No per-kind fd CONSTANT is checked. Only view and edge have module-level base
#     constants (VIEW_FD, EDGE_BASE_FD) in runner/stream-fds.ts; node/interior/drive bases
#     are computed PER SPAWN because they depend on edgeCount, so there is no per-kind
#     constant to look for and a check for one would be a lie for three of five kinds.
#   - No per-kind TABLE in webview/snapshot-buffer.ts is checked. The kind→table mapping is
#     genuinely NOT one-per-kind: view has no row table at all (singleton), and drive
#     deliberately shares interior's table (a drive-slot frame IS an interior-shaped frame,
#     last-writer-wins — see Buffer/stream_fds.go's StreamKindDrive). Asserting a table per
#     kind would false-fail on the tree as it stands.
#   - No per-kind FRAME TAG in src/schema/frame-tags.ts is checked, for the same reason:
#     drive reuses BUF_BLOCK_TAG_INTERIOR_STREAM.
#   - It does not check the fd ARITHMETIC agrees (base + row), nor that the reader decodes
#     the right frame shape. Those are behaviour, not vocabulary.
#   The honest summary: this is a NAME-LEVEL parity check between Go's stream-kind set and
#   the ext host's env-spelling + per-kind reader. It catches "a kind exists in Go that the
#   host has never heard of", which is the failure mode above. It does not certify a kind
#   that passes is wired CORRECTLY.
#
# Files are LOCATED BY SCANNING (memory/feedback_guards_hardcoding_single_file_break_on_split.md
# — this repo just split runCommand.ts and moved the fd constants out from under a
# hardcoded path). A scan that finds nothing reports MISCONFIGURED rather than passing.
#
# Exit 0 clean, exit 1 with a named report.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

TS_SRC="tools/topology-vscode/src"

if [ ! -d "$TS_SRC" ]; then
  echo "check-stream-kind-ts-parity: MISCONFIGURED — ext-host source dir not found: $TS_SRC"
  exit 1
fi

# --- Go side: discover the declaring file, then the kinds ---------------------
# `|| true` on every extractor assignment: under `set -e` a non-matching grep inside a
# VAR=$(...) assignment aborts the script with exit 1 and NO output, which would make the
# MISCONFIGURED branches below unreachable (a sibling guard hit exactly this).
GO_KIND_FILES=$(grep -rlE '^const StreamKind[A-Za-z]+ = "' --include='*.go' . \
  --exclude-dir=node_modules --exclude-dir=out --exclude-dir=.git 2>/dev/null || true)

if [ -z "$GO_KIND_FILES" ]; then
  echo "check-stream-kind-ts-parity: MISCONFIGURED — no file declares 'const StreamKind<X> = \"…\"'."
  echo "The declaration form or its home moved; this guard is scanning nothing rather than"
  echo "passing vacuously. Repoint its pattern."
  exit 1
fi

# kind VALUES (the wire names), e.g. view edge node interior drive.
KINDS=$(grep -hoE '^const StreamKind[A-Za-z]+ = "[a-zA-Z]+"' $GO_KIND_FILES \
  | grep -oE '"[a-zA-Z]+"' | tr -d '"' | sort -u || true)

if [ -z "$KINDS" ]; then
  echo "check-stream-kind-ts-parity: MISCONFIGURED — found $GO_KIND_FILES but extracted no kind values."
  exit 1
fi

# --- TS side: discover the file that builds WIREFOLD_STREAM_FDS ---------------
# The ASSIGNMENT site (`WIREFOLD_STREAM_FDS:` in the spawn env object), not every file that
# merely names the env var in a comment.
ENV_FILES=$(grep -rl 'WIREFOLD_STREAM_FDS:' --include='*.ts' "$TS_SRC" 2>/dev/null || true)
if [ -z "$ENV_FILES" ]; then
  echo "check-stream-kind-ts-parity: MISCONFIGURED — nothing under $TS_SRC assigns"
  echo "WIREFOLD_STREAM_FDS: in a spawn env. The ext host's fd-allocation site moved;"
  echo "repoint this guard."
  exit 1
fi

HITS=0
for k in $KINDS; do
  # 1. env spelling: `<kind>:${…Fd}` / `${…FD}` in the file that builds the env var.
  envpat="${k}:\\\$\\{[A-Za-z0-9_.]*[Ff][Dd]\\}"
  if ! grep -hoE "$envpat" $ENV_FILES >/dev/null 2>&1; then
    echo "check-stream-kind-ts-parity: stream kind \"$k\" is declared in Go but the ext host"
    echo "  never spells it into WIREFOLD_STREAM_FDS (looked for \`$k:\${…Fd}\` in: $ENV_FILES)."
    echo "  Go will find no base fd for it and every emitter of that kind will silently write"
    echo "  nowhere. Allocate a base fd and push the entry."
    HITS=$((HITS + 1))
  fi

  # 2. per-kind reader: handle<Kind>Fd(
  cap="$(printf '%s' "$k" | awk '{print toupper(substr($0,1,1)) substr($0,2)}')"
  if ! grep -rq "handle${cap}Fd(" --include='*.ts' "$TS_SRC" 2>/dev/null; then
    echo "check-stream-kind-ts-parity: stream kind \"$k\" is declared in Go but no"
    echo "  handle${cap}Fd( reader exists anywhere under $TS_SRC."
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
