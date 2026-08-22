#!/usr/bin/env bash

# PLACEMENT: none | universal TS lint, not a placement decision (react-hooks rules are errors)
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

npx --no-install eslint .
