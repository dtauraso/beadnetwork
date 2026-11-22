#!/usr/bin/env bash

# PLACEMENT: Categories/Node/**/*.go | every refuseStructuralEdit(...) call site must be followed by

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

report="$(python3 - <<'PY'
import os, re

root = "nodes"
call_pat = re.compile(r'\b[Rr]efuseStructuralEdit\(')
def_pat = re.compile(r'func\s+\([a-zA-Z_]+ \*\w+\)\s+[Rr]efuseStructuralEdit\(')
next_pat = re.compile(r'\b[Ee]mitViewFrame\(')

hits = []
checked = 0

for dp, _dn, fns in os.walk(root):
    for fn in fns:
        if not fn.endswith(".go") or fn.endswith("_test.go"):
            continue
        p = os.path.join(dp, fn)
        with open(p, encoding="utf-8", errors="replace") as fh:
            lines = fh.readlines()
        stripped = [l.split("//", 1)[0] for l in lines]
        for i, code in enumerate(stripped):
            if not call_pat.search(code):
                continue
            if def_pat.search(code):
                continue  # skip the definition line itself
            checked += 1
            j = i + 1
            found = False
            while j < len(stripped):
                s = stripped[j].strip()
                if s == "":
                    j += 1
                    continue
                found = bool(next_pat.search(s))
                break
            if not found:
                hits.append(f"{p}:{i+1}: {lines[i].strip()}")

out = []
if hits:
    out.append("MISSING_EMIT")
    out += hits
out.append(f"CHECKED={checked}")
print("\n".join(out))
PY
)"

checked_line="$(grep '^CHECKED=' <<< "$report" || true)"
checked_count="${checked_line#CHECKED=}"
report="$(grep -v '^CHECKED=' <<< "$report" || true)"

echo "check-refusal-emits-frame: checked $checked_count refuseStructuralEdit( call site(s)"

if [[ -z "$report" ]]; then
  exit 0
fi

fail=0
while IFS= read -r line; do
  case "$line" in
    MISSING_EMIT)
      echo "REFUSAL WITHOUT EMIT: a refuseStructuralEdit( call is not immediately followed by an"
      echo "emitViewFrame( call — the refusal counter bumps but the webview is never told:"
      fail=1; continue ;;
    *)
      printf '  %s\n' "$line" ;;
  esac
done <<< "$report"

exit $fail
