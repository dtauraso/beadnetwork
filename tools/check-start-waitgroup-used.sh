#!/usr/bin/env bash
# check-start-waitgroup-used.sh — a test must not discard MoveDispatch.Start's WaitGroup.
#
# Start(ctx) returns a *sync.WaitGroup covering every goroutine it launches. A test that
# calls it as a BARE STATEMENT never waits for those goroutines: `defer cancel()` only
# SIGNALS shutdown, and t.TempDir's RemoveAll (registered via t.Cleanup, so it runs AFTER
# the deferred cancel) then races goroutines still closing their stream fds. That produces
# an intermittent "TempDir RemoveAll cleanup: bad file descriptor", only under the parallel
# suite, passing on every isolated rerun.
#
# The fix is a happens-before edge, not a longer sleep — capture the group and wait, with
# the cleanup registered AFTER t.TempDir() so LIFO ordering runs it first:
#
#     wg := md.Start(ctx)
#     t.Cleanup(func() { cancel(); wg.Wait() })
#
# See docs/testing-shape.md ("Signal-without-wait teardown").
set -uo pipefail
cd "$(dirname "$0")/.."

# Bare-statement call: a line whose entire content is `<recv>.Start(ctx)` with no
# assignment. An assigned call (`wg := md.Start(ctx)`) contains `=` and is not matched.
hits=$(git ls-files '*_test.go' -z \
  | xargs -0 grep -nE '^[[:space:]]*[A-Za-z_][A-Za-z0-9_]*\.Start\(ctx\)[[:space:]]*$' \
  2>/dev/null || true)

if [ -n "$hits" ]; then
  echo "check-start-waitgroup-used: test(s) discard MoveDispatch.Start's WaitGroup, so"
  echo "nothing waits for the goroutines before cleanup runs (see docs/testing-shape.md):"
  echo "$hits" | sed 's/^/  /'
  echo
  echo "Fix — capture the group and wait, registered AFTER t.TempDir() so t.Cleanup LIFO"
  echo "runs it before the dir RemoveAll:"
  echo "    wg := md.Start(ctx)"
  echo "    t.Cleanup(func() { cancel(); wg.Wait() })"
  exit 1
fi
exit 0
