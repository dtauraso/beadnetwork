#!/usr/bin/env bash

# PLACEMENT: tools/**/*.sh | a guard that names a path which no longer exists passes while checking nothing; this is what makes that loud.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

python3 - <<'PY'
import pathlib
import re
import sys

SCAN_DIRS = ["tools", "scripts", ".githooks"]

TOPS = tuple(sorted(p.name + "/" for p in pathlib.Path(".").iterdir()
                    if p.is_dir() and not p.name.startswith(".") or p.name == ".githooks"))

PATHISH = re.compile(r"(?<![A-Za-z0-9_./-])((?:%s)[A-Za-z0-9_./*{}-]*)" %
                     "|".join(re.escape(t) for t in TOPS))

files = []
for d in SCAN_DIRS:
    root = pathlib.Path(d)
    if root.is_dir():
        files += [f for f in root.rglob("*.sh") if "node_modules" not in str(f)]
    if root.is_dir():
        files += [f for f in root.rglob("*.py") if "node_modules" not in str(f)]

if len(files) < 20:
    print(f"check-guard-paths-exist: MISCONFIGURED — found only {len(files)} script(s) to scan; "
          f"refusing vacuous pass", file=sys.stderr)
    sys.exit(1)

DECL = pathlib.Path("tools/repo-hygiene/paths-may-be-absent.tsv")
if not DECL.exists():
    print(f"check-guard-paths-exist: MISCONFIGURED — {DECL} not found; every declared "
          f"absence would silently become a failure", file=sys.stderr)
    sys.exit(1)

declared = {}
for lineno, line in enumerate(DECL.read_text(encoding="utf-8").splitlines(), 1):
    if not line.strip():
        continue
    parts = line.split("\t")
    if len(parts) != 3 or not parts[2].strip():
        print(f"{DECL}:{lineno}: needs script<TAB>path<TAB>reason — an absence without a "
              f"stated reason is an allowlist entry, not a decision", file=sys.stderr)
        sys.exit(1)
    declared.setdefault(parts[0], set()).add(parts[1])

for script in declared:
    if not pathlib.Path(script).exists():
        print(f"{DECL}: names {script}, which does not exist — the declaration outlived "
              f"the script it excused", file=sys.stderr)
        sys.exit(1)

missing = {}
checked = 0
for f in sorted(files):
    text = f.read_text(encoding="utf-8", errors="replace")
    allowed = declared.get(str(f), set())
    for m in PATHISH.finditer(text):
        raw = m.group(1).rstrip("/.,:;\"')")
        if not raw or raw.endswith("-") or "{" in raw or "}" in raw:
            continue
        if raw in allowed:
            continue
        checked += 1
        if "*" in raw:
            if not list(pathlib.Path(".").glob(raw)):
                missing.setdefault(str(f), set()).add(raw)
            continue
        if not pathlib.Path(raw).exists():
            missing.setdefault(str(f), set()).add(raw)

if not checked:
    print("check-guard-paths-exist: MISCONFIGURED — parsed 0 repo paths out of the scripts; "
          "format changed, this would check nothing", file=sys.stderr)
    sys.exit(1)

if missing:
    print("GUARD NAMES A PATH THAT DOES NOT EXIST — it is checking nothing there:")
    for f in sorted(missing):
        print(f"  {f}")
        for p in sorted(missing[f]):
            print(f"      {p}")
    print()
    print("  A guard that greps a moved or renamed path still exits 0. It reports success")
    print("  while covering nothing, which is worse than having no guard at all. Point it")
    print("  at where the thing lives now, or delete it if the thing is gone.")
    sys.exit(1)

print(f"check-guard-paths-exist: clean ({checked} path references across {len(files)} scripts)")
PY
