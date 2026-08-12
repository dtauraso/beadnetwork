#!/usr/bin/env bash

# PLACEMENT: *_test.go | a bare `<recv>.Start(ctx)` call must capture and Wait() its returned WaitGroup before cleanup
set -uo pipefail
cd "$(dirname "$0")/../../.."

hits=$(git ls-files '*_test.go' -z \
  | xargs -0 grep -nE '^[[:space:]]*[A-Za-z_][A-Za-z0-9_]*\.Start\(ctx\)[[:space:]]*$' \
  2>/dev/null || true)

if [ -n "$hits" ]; then
  echo "check-start-waitgroup-used: test(s) discard MoveDispatch.Start's WaitGroup, so"
  echo "nothing waits for the goroutines before cleanup runs:"
  echo "$hits" | sed 's/^/  /'
  echo
  echo "Fix — capture the group and wait, registered AFTER t.TempDir() so t.Cleanup LIFO"
  echo "runs it before the dir RemoveAll:"
  echo "    wg := md.Start(ctx)"
  echo "    t.Cleanup(func() { cancel(); wg.Wait() })"
  exit 1
fi
exit 0
