#!/usr/bin/env bash

set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

bold() { printf '\033[1m%s\033[0m\n' "$1"; }

bold "current branch"
git rev-parse --abbrev-ref HEAD
echo

bold "open work (task/* branches — description = the item)"
found=0
for ref in $(git for-each-ref --format='%(refname:short)' refs/heads/'task/*'); do
  found=1
  desc=$(git config "branch.$ref.description" || true)
  printf '  \033[36m%s\033[0m\n' "$ref"
  if [ -n "$desc" ]; then
    printf '%s\n' "$desc" | fold -s -w 76 | sed 's/^/      /'
  else
    printf '      (no description — set with: git config branch.%s.description "...")\n' "$ref"
  fi
  echo
done
[ "$found" = 0 ] && echo "  (no task/* branches — clean)"
echo

bold "recently merged to main (last 8)"
git log --oneline --merges -8 main 2>/dev/null || git log --oneline -8 main
echo

bold "next steps"
echo "  - start a change: scripts/new-task.sh <short-kebab-name> \"one-line description\""
echo "    (makes the branch and checks it out — plain branches in this one checkout)"
echo "  - read memory/MEMORY.md (durable rules + project state)"
echo "  - read MODEL.md before any Go-network / pump change"
echo "  - verify recipe: see CLAUDE.md Workflow — bash scripts/stop-checks.sh, clean == EMPTY stdout"
echo "    (it ALWAYS exits 0 by design; \$? is not the signal)"
echo
bold "diagnostics"
echo "  - scripts/reload-gap.sh — is 'Developer: Reload Window' slow? prints the"
echo "    extension-host respawn gap per reload (~1.8s healthy, >4s regressed)."
