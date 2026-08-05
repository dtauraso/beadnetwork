#!/usr/bin/env bash
# check-doc-drift.sh — thin wrapper so scripts/audit-doc-drift.mjs (broken
# file/path references in tracked docs AND in .go/.ts/.tsx/.sh/.mjs/.yml/
# .yaml/.gitignore/Makefile source comments)
# runs in the discovered tools/check-*.sh guard loop (scripts/stop-checks.sh
# globs that dir only). Without this wrapper the audit script existed but
# nothing ever invoked it — see docs/drift-checklist.md item 1 ("Can the model
# skip a required step/tool and still answer?").
#
# PLACEMENT: none | thin wrapper invoking scripts/audit-doc-drift.mjs so it runs in the guard loop
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
node "$REPO_ROOT/scripts/audit-doc-drift.mjs"
