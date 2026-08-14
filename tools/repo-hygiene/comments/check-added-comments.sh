#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: none | repo-wide: no comment line added since the base ref survives in a hand-edited file

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$REPO_ROOT"

report="$(WIREFOLD_COMMENT_BASE="${WIREFOLD_COMMENT_BASE:-}" python3 - <<'PY'
import os, re, subprocess, sys

def run(*a):
    return subprocess.run(a, capture_output=True, text=True).stdout

base = os.environ.get("WIREFOLD_COMMENT_BASE") or ""
if not base:
    for cand in ("origin/main", "main"):
        if run("git", "rev-parse", "--verify", "-q", cand).strip():
            base = cand
            break
if not base:
    print("SCANNED 0 file(s), no base ref")
    sys.exit(0)

LINE = {".go": "//", ".ts": "//", ".tsx": "//", ".js": "//", ".jsx": "//",
        ".sh": "#", ".py": "#", ".css": None, ".md": None}

# Comments a tool READS. Removing one changes behaviour, so none of them is
# narration however it was added: build pragmas, lint and formatter directives,
# generated-file markers, the PLACEMENT header tools/placement-brief.sh serves,
# and the ALL_CAPS fences the parity guards grep between.
KEEP = re.compile(
    r"go:|^\+build|nolint|shellcheck|eslint|@ts-|prettier|istanbul|coding[:=]|"
    r"Code generated|DO NOT EDIT|@generated|PLACEMENT:|^[A-Z][A-Z0-9_]{3,}$")

def added_lines(path):
    """New-file line numbers added since base, read out of the diff's own hunk
    headers rather than by matching text — the same line can appear twice."""
    d = run("git", "diff", "-U0", base, "--", path)
    out, n = set(), 0
    for ln in d.split("\n"):
        h = re.match(r"^@@ -\S+ \+(\d+)(?:,(\d+))? @@", ln)
        if h:
            n = int(h.group(1))
            continue
        if ln.startswith("+") and not ln.startswith("+++"):
            out.add(n)
            n += 1
        elif not ln.startswith("-") and not ln.startswith("\\"):
            n += 1
    return out

def strip(path, marker, added):
    try:
        src = open(path, encoding="utf-8").read().split("\n")
    except OSError:
        return []
    cut, killed, in_block = set(), [], False
    for i, raw in enumerate(src, start=1):
        s = raw.strip()
        is_c = False
        if in_block:
            is_c = True
            if "*/" in s:
                in_block = False
        elif marker == "//" and s.startswith("/*"):
            is_c, in_block = True, "*/" not in s
        elif marker and s.startswith(marker):
            is_c = True
        if not is_c or i not in added or KEEP.search(s.lstrip("/*# ").strip()):
            continue
        cut.add(i)
        killed.append(f"{path}:{i}: {s[:72]}")
    if not cut:
        return []
    kept, blank = [], False
    for i, raw in enumerate(src, start=1):
        if i in cut:
            continue
        if not raw.strip():
            if blank:
                continue
            blank = True
        else:
            blank = False
        kept.append(raw)
    tmp = path + ".strip.tmp"
    with open(tmp, "w", encoding="utf-8") as f:
        f.write("\n".join(kept))
    os.replace(tmp, path)
    return killed

files = [p for p in run("git", "diff", "--name-only", base).split("\n") if p]
removed, scanned = [], 0
for p in files:
    ext = os.path.splitext(p)[1]
    marker = LINE.get(ext)
    if not marker or not os.path.isfile(p) or "node_modules" in p or "/out/" in p:
        continue
    head = "".join(open(p, encoding="utf-8", errors="replace").readlines()[:3])
    if re.search(r"DO NOT EDIT|Code generated|@generated", head, re.I):
        continue
    scanned += 1
    removed += strip(p, marker, added_lines(p))

print(f"SCANNED {scanned} changed file(s) against {base}")
if removed:
    print("REMOVED")
    print("\n".join(removed))
PY
)"

summary="$(head -n1 <<< "$report")"
rest="$(tail -n +2 <<< "$report")"
echo "$summary"
[[ -z "$rest" ]] && exit 0

fail=0
while IFS= read -r line; do
  case "$line" in
    REMOVED)
      echo "ADDED COMMENTS REMOVED: these comment lines were added since the base ref and have"
      echo "been deleted from the working tree. Re-read the diff before committing — if one of"
      echo "them was load-bearing, put it back deliberately and it will not be touched again"
      echo "once it is part of the base:"
      fail=1; continue ;;
    *) printf '  %s\n' "$line" ;;
  esac
done <<< "$rest"

exit $fail
