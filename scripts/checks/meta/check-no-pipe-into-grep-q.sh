#!/usr/bin/env bash

# PLACEMENT: scripts/checks,scripts/lib,src | a pipe into `grep -q` under `set -o pipefail` reports SIGPIPE as a real failure

set -uo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

hits=""
while IFS= read -r f; do
  grep -q 'pipefail' "$f" || continue
  h=$(grep -nE '\|[[:space:]]*grep[[:space:]]+-[a-zA-Z]*q' "$f" || true)
  [ -n "$h" ] && hits+="$f"$'\n'"$h"$'\n'
done < <(git ls-files '*.sh')

if [ -n "$hits" ]; then
  cat >&2 <<'WHY'
check-no-pipe-into-grep-q: FAIL

A pipeline whose right side is `grep -q` is a coin flip under `set -o pipefail`.
`grep -q` exits the moment it MATCHES, which closes the pipe while the left side
is still writing; the writer takes SIGPIPE and exits 141, pipefail promotes that
to the pipeline's status, and `if ! ...` reads the whole thing as "not found".
The check then reports a violation for text that IS present. It fires only when
the writer loses the race, so it is rare when run alone and common under the
parallel `xargs -P` guard runner - which is what made two guards look flaky and
cost a session's debugging.

Measured on the real input: 3% of greps returned 141 under load, and with ~21
greps per run that is a false failure every other run.

Use a here-string instead - no pipe, no SIGPIPE, same test:

    if ! grep -qF "$needle" <<< "$haystack"; then

Offending lines:
WHY
  printf '%s' "$hits" >&2
  exit 1
fi

echo "check-no-pipe-into-grep-q: clean (no pipefail script pipes into grep -q)"
