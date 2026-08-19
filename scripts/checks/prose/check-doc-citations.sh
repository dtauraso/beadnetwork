#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: none | universal prose hygiene: any file quoting CLAUDE.md/MODEL.md must quote it verbatim

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

source "$REPO_ROOT/scripts/lib/docs/normalize-docs.sh"
source "$REPO_ROOT/scripts/lib/docs/collect-citing-files.sh"
source "$REPO_ROOT/scripts/lib/docs/normalize-candidates.sh"
source "$REPO_ROOT/scripts/lib/docs/extract-citations.sh"
source "$REPO_ROOT/scripts/lib/docs/citations-match.sh"

require_docs_present

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

write_normalized_docs
collect_citing_files
write_normalized_candidates
extract_citations

check_citations_match
