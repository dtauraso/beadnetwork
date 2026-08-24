#!/usr/bin/env bash

# PLACEMENT: .claude/rules/ | a rule whose paths: frontmatter matches nothing never loads, and silence is what that looks like.

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

python3 - <<'PY'
import glob
import pathlib
import re
import sys

RULES = sorted(pathlib.Path(".claude/rules").glob("*.md"))

if len(RULES) < 3:
    print(f"check-rule-paths-exist: MISCONFIGURED — found only {len(RULES)} rule(s); "
          f"refusing vacuous pass", file=sys.stderr)
    sys.exit(1)

ENTRY = re.compile(r"^\s*-\s*(.+?)\s*$")

empty = {}
no_frontmatter = []
checked = 0

for rule in RULES:
    text = rule.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        no_frontmatter.append(str(rule))
        continue
    end = text.find("\n---", 4)
    if end == -1:
        no_frontmatter.append(str(rule))
        continue
    front = text[4:end]

    in_paths = False
    patterns = []
    for line in front.splitlines():
        if re.match(r"^paths:\s*$", line):
            in_paths = True
            continue
        if in_paths:
            if re.match(r"^\S", line):
                in_paths = False
                continue
            m = ENTRY.match(line)
            if m:
                patterns.append(m.group(1).strip().strip('"').strip("'"))

    if not patterns:
        no_frontmatter.append(str(rule))
        continue

    for pat in patterns:
        checked += 1
        if not glob.glob(pat, recursive=True):
            empty.setdefault(str(rule), []).append(pat)

if not checked:
    print("check-rule-paths-exist: MISCONFIGURED — parsed 0 path patterns out of the rule "
          "frontmatter; format changed, this would check nothing", file=sys.stderr)
    sys.exit(1)

if no_frontmatter:
    print("RULE HAS NO paths: FRONTMATTER — it can never load on demand:")
    for r in no_frontmatter:
        print(f"  {r}")
    print()
    print("  A rule in .claude/rules/ loads when an edited file matches its paths:.")
    print("  With none, it is dead weight nothing will ever read. Give it paths, or")
    print("  move its content to CLAUDE.md if it is meant to always apply.")
    sys.exit(1)

if empty:
    print("RULE PATH MATCHES NOTHING — the rule never loads:")
    for r in sorted(empty):
        print(f"  {r}")
        for p in empty[r]:
            print(f"      {p}")
    print()
    print("  These rules load by paths: match. A pattern pointing at a moved or deleted")
    print("  path matches no file, so the rule silently stops loading — and a rule that")
    print("  never loads looks exactly like a rule that is being followed. Point it at")
    print("  where the code lives now, or delete the rule if its subject is gone.")
    sys.exit(1)

print(f"check-rule-paths-exist: clean ({checked} path patterns across {len(RULES)} rules)")
PY
