#!/usr/bin/env bash
# audit-mutation-crap.sh — SLOW-LANE, PERIODIC test-strength audit. NOT part of
# scripts/stop-checks.sh and must never be added to it: this recompiles the module
# and re-runs the whole test suite per mutation site, taking minutes to tens of
# minutes depending on scope. Run it by hand, occasionally, not per commit.
#
# WHY THIS EXISTS: docs/testing-shape.md argues cross-goroutine correctness is
# guaranteed BY CONSTRUCTION (ownership + message-passing, no locks), and a prior
# audit pass deleted ~27 tests on that argument (docs/testing-shape.md "History").
# Nothing in this repo produces EVIDENCE for or against that argument.
# tools/check-test-integrity.sh (if present) only detects tests being WEAKENED over
# time — it says nothing about whether they had strength to begin with. This script
# runs two tools that together do:
#
#   crap4go   — CC^2 x (1-cov)^3 + CC per function. Says WHERE coverage was lost:
#               complex, undertested functions are the riskiest to change blind.
#   mutate4go — mutation testing. Says whether the SURVIVING tests would actually
#               notice a broken line, one covered line at a time.
#
# crap4go ranks; mutate4go verifies the top of that ranking, one file at a time
# (mutate4go recompiles per mutation site, so it does not scale to "run on nodes/").
#
# WORKFLOW (see this script's usage output, and .claude/skills/audit-mutation-crap/SKILL.md
# for the full narrative including the manifest trap below):
#   1. audit-mutation-crap.sh crap [path-fragment...]   — rank functions, worst first
#   2. audit-mutation-crap.sh mutate-scan <file.go>      — cheap structural count, SAFE,
#      writes nothing to the source file
#   3. audit-mutation-crap.sh mutate <file.go>           — full run. WARNING: on success
#      this embeds a manifest footer comment INTO THE SOURCE FILE (see TRAP below).
#
# TRAP — mutate4go writes into your source, not just target/coverage/:
#   Any non---scan mutate4go run (including --update-manifest) rewrites the target
#   .go file in place, appending a `// mutate4go-manifest-begin ... end` JSON footer
#   comment recording the last-tested date and a hash per function
#   (~/Downloads/unclebob-repos/mutate4go/internal/manifest/manifest.go: Embed() is
#   called by cmd/mutate4go's non-scan path and writes the file with os.WriteFile).
#   That footer WILL collide with three guards in this repo's stop-checks suite:
#     - tools/check-gofmt.sh              — the rewritten file may need `gofmt -w`
#     - tools/check-no-untracked-source.sh — mutate4go also leaves a `<file>.mutate4go.bak`
#                                            backup NEXT TO the source; if a worker crashes
#                                            mid-run the .bak can survive and is untracked
#     - tools/check-comment-vocab.sh       — any retired-vocabulary token the manifest
#                                            JSON happens to echo (function names, etc.)
#                                            would trip the banned-token scan
#   None of this is a bug in mutate4go — the manifest is what makes its differential
#   mode fast on a second run. It is a bug WAITING TO HAPPEN if a full run's result is
#   ever committed by accident. Safe practice:
#     - Use `--scan` for anything you don't intend to commit results of — it reads
#       structure only, never writes.
#     - After a real (non-scan) run you intend to act on, `git diff <file>` before
#       committing: keep the manifest ONLY if you deliberately want differential mode
#       going forward, otherwise `git checkout -- <file>` to discard it and rely on a
#       fresh full scan next time.
#     - Always `git status --porcelain` afterward and delete any `*.mutate4go.bak`
#       stragglers before committing.
#
# COVERAGE ARTIFACT: both tools write target/coverage/coverage.out, matching this
# repo's `*.out` .gitignore entry (confirmed 2026-07: an untracked
# target/coverage/coverage.out already existed in the tree from a prior ad hoc run —
# evidence this needs documenting, not itself a problem since it's ignored). Verify
# with `git status --porcelain -uall` after any run; nothing under target/ should
# ever need staging.
#
# SETUP: nothing. Both tools are fetched as Go MODULES at the pinned versions below and
# cached by the Go module cache, so this needs network only on the first run per version.
#
# They used to be built from clones at ~/Downloads/unclebob-repos/. That tied a checked-in
# script to one machine's Downloads folder — it breaks silently the day that folder is
# reorganized, and the failure ("source not found") points at the tool rather than at the
# real cause. A module path is the medium's normal answer; pinning tooling to a downloads
# directory is the weird choice (CLAUDE.md "Medium vs. substance").
#
# Versions are PINNED, not @latest: @latest re-resolves over the network on every run and
# would silently change what this audit measures between two runs, which is the one thing a
# measurement tool must not do. Bump deliberately.
#
# To work from a local clone instead (offline, or to test an unreleased change), set
# CRAP4GO_SRC / MUTATE4GO_SRC to its directory — an existing directory takes precedence
# over the module path.

set -euo pipefail

CRAP4GO_MOD="${CRAP4GO_MOD:-github.com/unclebob/crap4go/cmd/crap4go@v0.0.0-20260521190544-bee16dbdadb4}"
MUTATE4GO_MOD="${MUTATE4GO_MOD:-github.com/unclebob/mutate4go/cmd/mutate4go@v0.0.0-20260523153956-9016c7adafc1}"

# Empty by default — set to a clone directory to override the module path above.
CRAP4GO_SRC="${CRAP4GO_SRC:-}"
MUTATE4GO_SRC="${MUTATE4GO_SRC:-}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

usage() {
  cat >&2 <<'EOF'
Usage:
  audit-mutation-crap.sh crap [path-fragment...]
      Build and run crap4go over this repo. With no fragment, scans everything
      under go.mod (slow: full `go test ./... -coverprofile=...`). Pass a
      fragment ("nodes") to scope both which functions are reported AND which
      test packages run (crap4go still needs coverage for the whole tree it
      analyzes unless you also pass --test-command yourself).

  audit-mutation-crap.sh mutate-scan <file.go>
      Structural-only mutation site count for one file. SAFE: writes nothing.

  audit-mutation-crap.sh mutate <file.go> [extra mutate4go args...]
      Full mutation run for one file: applies each covered mutation, reruns
      tests, reports killed/survived/uncovered. WRITES a manifest footer into
      the source file on success — see the TRAP comment at the top of this
      script before running this.

Recommended sequence (see .claude/skills/audit-mutation-crap/SKILL.md):
  1. audit-mutation-crap.sh crap nodes
  2. Take the top few functions from the CRAP table (worst first).
  3. audit-mutation-crap.sh mutate-scan nodes/Wiring/<worst file>.go
  4. audit-mutation-crap.sh mutate nodes/Wiring/<worst file>.go
  5. Address survivors/uncovered, repeat step 4 on the SAME file until clean,
     THEN move to the next file. Do not run mutate4go across many files in one
     pass — it recompiles per mutation site.
EOF
  exit 1
}

# Resolve a tool to a runnable binary. A local clone wins if one was pointed at; otherwise
# the pinned module is installed into a temp GOBIN.
#
# `go install <mod>@<ver>` rather than `go run`: go run rebuilds and re-links on every
# invocation, and `mutate` calls the binary once per mutation site. Installing once and
# exec'ing the result keeps that cost out of the inner loop. Both forms are module-aware and
# ignore THIS repo's go.mod, so neither can perturb our own dependencies — which is also why
# the old `go build ./cmd/...` had to cd into the clone ("outside main module").
build_tool() {
  local src="$1" mod="$2" name="$3" bin_var="$4"
  local bin
  if [ -n "$src" ]; then
    if [ ! -d "$src" ]; then
      echo "audit-mutation-crap: $name source override is set but is not a directory: $src" >&2
      exit 1
    fi
    bin="$(mktemp -d)/$name"
    (cd "$src" && go build -o "$bin" "./cmd/$name")
  else
    local gobin
    gobin="$(mktemp -d)"
    if ! GOBIN="$gobin" go install "$mod" 2>&1; then
      echo "audit-mutation-crap: could not install $name from $mod" >&2
      echo "  First run needs network. Offline? Point ${name^^}_SRC at a local clone." >&2
      exit 1
    fi
    bin="$gobin/$name"
  fi
  printf -v "$bin_var" '%s' "$bin"
}

cmd="${1:-}"
case "$cmd" in
  crap)
    shift
    build_tool "$CRAP4GO_SRC" "$CRAP4GO_MOD" crap4go CRAP4GO_BIN
    exec "$CRAP4GO_BIN" "$@"
    ;;
  mutate-scan)
    shift
    [ $# -ge 1 ] || usage
    build_tool "$MUTATE4GO_SRC" "$MUTATE4GO_MOD" mutate4go MUTATE4GO_BIN
    exec "$MUTATE4GO_BIN" "$1" --scan
    ;;
  mutate)
    shift
    [ $# -ge 1 ] || usage
    build_tool "$MUTATE4GO_SRC" "$MUTATE4GO_MOD" mutate4go MUTATE4GO_BIN
    echo "audit-mutation-crap: full run WRITES a manifest footer into $1 on success — see this script's header comment." >&2
    exec "$MUTATE4GO_BIN" "$@"
    ;;
  *)
    usage
    ;;
esac
