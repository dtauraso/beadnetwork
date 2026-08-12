#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: none | universal prose hygiene: any file quoting CLAUDE.md/MODEL.md must quote it verbatim

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

source "$SCRIPT_DIR/lib/normalize-docs.sh"
source "$SCRIPT_DIR/lib/collect-citing-files.sh"
source "$SCRIPT_DIR/lib/normalize-candidates.sh"
source "$SCRIPT_DIR/lib/extract-citations.sh"
source "$SCRIPT_DIR/lib/check-citations-match.sh"

require_docs_present

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

write_normalized_docs
collect_citing_files
write_normalized_candidates
extract_citations

check_citations_match
