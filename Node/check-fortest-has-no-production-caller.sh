#!/usr/bin/env bash

# PLACEMENT: Node/**/*.go | a *ForTest symbol must have no production (non-_test.go) caller

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

report="$(python3 - <<'PY'
import os, re

root = "Node"
def_pat = re.compile(r'\bfunc\s+(\w*ForTest)\b')
ref_pat = re.compile(r'\b(\w*ForTest)\b')

symbols = set()
files_scanned = 0
hits = []

for dp, _dn, fns in os.walk(root):
    for fn in fns:
        if not fn.endswith(".go"):
            continue
        p = os.path.join(dp, fn)
        with open(p, encoding="utf-8", errors="replace") as fh:
            for line in fh:
                code = line.split("//", 1)[0]
                for m in def_pat.finditer(code):
                    symbols.add(m.group(1))

for dp, _dn, fns in os.walk(root):
    for fn in fns:
        if not fn.endswith(".go") or fn.endswith("_test.go"):
            continue
        p = os.path.join(dp, fn)
        files_scanned += 1
        with open(p, encoding="utf-8", errors="replace") as fh:
            for i, line in enumerate(fh, 1):
                code = line.split("//", 1)[0]           # strip line comment (doc prose exempt)
                if not code.strip():
                    continue
                is_def = def_pat.search(code) is not None
                for m in ref_pat.finditer(code):
                    name = m.group(1)
                    if name not in symbols:
                        continue
                    if is_def and def_pat.search(code).group(1) == name:
                        continue  # the definition line itself, not a call site
                    hits.append(f"{p}:{i}: {line.strip()}")

if files_scanned == 0:
    raise SystemExit(
        f"check-fortest-has-no-production-caller: MISCONFIGURED — walked {root!r} and found\n"
        "  no .go files, so every ForTest symbol would read as unused. This guard reported\n"
        "  a clean scan of zero files for a whole day after nodes/ became Node/; a bare\n"
        "  directory name is invisible to check-guard-paths-exist.sh. Repoint root."
    )

out = []
out.append(f"SCANNED {files_scanned} Node/**/*.go production files, {len(symbols)} *ForTest symbol(s): {', '.join(sorted(symbols)) if symbols else '(none found)'}")
if hits:
    out.append("HITS")
    out += hits
print("\n".join(out))
PY
)"

summary="$(head -n1 <<< "$report")"
rest="$(tail -n +2 <<< "$report")"
echo "$summary"

if [[ -z "$rest" ]]; then
  exit 0
fi

fail=0
while IFS= read -r line; do
  case "$line" in
    HITS)
      echo "FORTEST SYMBOL CALLED FROM PRODUCTION CODE: a *ForTest constructor is a test-only"
      echo "escape hatch. Give the caller a"
      echo "real, non-test entry point instead of reusing the test hatch:"
      fail=1; continue ;;
    *)
      printf '  %s\n' "$line" ;;
  esac
done <<< "$rest"

exit $fail
