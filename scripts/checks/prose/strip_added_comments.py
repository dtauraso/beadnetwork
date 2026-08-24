import os, re, subprocess, sys

def run(*a):
    return subprocess.run(a, capture_output=True, text=True).stdout

base = os.environ.get("BEADNETWORK_COMMENT_BASE") or ""
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

KEEP = re.compile(
    r"^!|go:|frametag:|^\+build|nolint|shellcheck|eslint|@ts-|prettier|istanbul|coding[:=]|"
    r"Code generated|DO NOT EDIT|@generated|PLACEMENT:|^[A-Z][A-Z0-9_]{3,}$")

FENCE_START = re.compile(r"^[A-Z][A-Z0-9_]*_START$")
FENCE_END = re.compile(r"^[A-Z][A-Z0-9_]*_END$")

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
    fenced = False
    for i, raw in enumerate(src, start=1):
        s = raw.strip()
        body = s.lstrip("/*# ").strip()
        if FENCE_START.match(body):
            fenced = True
        if fenced:
            if FENCE_END.match(body):
                fenced = False
            continue
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
    mode = os.stat(path).st_mode
    tmp = path + ".strip.tmp"
    with open(tmp, "w", encoding="utf-8") as f:
        f.write("\n".join(kept))
    os.chmod(tmp, mode)
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
