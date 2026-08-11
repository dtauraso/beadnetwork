#!/usr/bin/env bash
#
# PLACEMENT: nodes/**/*.go | every refuseStructuralEdit(...) call site must be followed by
# an emitViewFrame(...) call — the write-then-emit split moved the VIEW frame emit OUT of
# refuseStructuralEdit and onto its callers.
# check-refusal-emits-frame.sh — forbid a refuseStructuralEdit(...) call that is not
# immediately followed by an emitViewFrame(...) call. Run from repo root:
# bash tools/network/structure/check-refusal-emits-frame.sh
#
# WHY THIS EXISTS: docs/planning/movedispatch-decomposition.md, the write-then-emit split, moved
# emitViewFrame OUT of refuseStructuralEdit (nodes/Wiring/scene_structure.go) and onto each of
# its 12 CreateNode/DeleteNode call sites by hand. refuseStructuralEdit used to emit its own
# frame — impossible to forget. Now a call site that forgets the follow-up emit bumps the
# ui.editRefused counter and returns silently: nothing is written, the run does not end, and
# the webview is never told. That reads EXACTLY like a refused edit that never happened, i.e.
# a broken build, from the outside — deleting one such emit was confirmed to produce ZERO
# `go test ./...` failures before this guard and its sibling test existed.
#
# RULE: for every call to refuseStructuralEdit( in non-test Go under nodes/, the next
# non-comment, non-blank line in the same function must be a call to emitViewFrame(. This
# skips refuseStructuralEdit own definition line (func (md *MoveDispatch) refuseStructuralEdit).
#
# Scope: production (non-test) Go under nodes/. Comments are stripped from the scanned lines
# before matching, so prose about refuseStructuralEdit does not trip it.
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/*/check-*.sh / tools/*/*/check-*.sh glob.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

report="$(python3 - <<'PY'
import os, re

root = "nodes"
call_pat = re.compile(r'\brefuseStructuralEdit\(')
def_pat = re.compile(r'func\s+\([a-zA-Z_]+ \*\w+\)\s+refuseStructuralEdit\(')
next_pat = re.compile(r'\bemitViewFrame\(')

hits = []
checked = 0

for dp, _dn, fns in os.walk(root):
    for fn in fns:
        if not fn.endswith(".go") or fn.endswith("_test.go"):
            continue
        p = os.path.join(dp, fn)
        with open(p, encoding="utf-8", errors="replace") as fh:
            lines = fh.readlines()
        # Strip line comments (prose exempt) for matching, keep originals for reporting.
        stripped = [l.split("//", 1)[0] for l in lines]
        for i, code in enumerate(stripped):
            if not call_pat.search(code):
                continue
            if def_pat.search(code):
                continue  # skip the definition line itself
            checked += 1
            # find next non-blank, non-comment-only line
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
