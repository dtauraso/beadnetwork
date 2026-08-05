#!/usr/bin/env bash
# check-memory-hygiene.sh — enforce that every memory/ entry is a well-formed, indexed
# memory, not raw agent monologue. Run from repo root: bash tools/check-memory-hygiene.sh
#
# WHY THIS EXISTS (drift-checklist item #7 — "memory poisoning"): CLAUDE.md's drift
# checklist asks "can the agent's own monologue become persistent memory/?" A repo guard
# can't watch the agent, but it CAN enforce the static shape every real memory has, so a
# malformed / typeless / unindexed blob (the shape monologue-dumped-as-memory takes) fails
# the build instead of silently persisting. Each memory/*.md must have: YAML frontmatter,
# a name, a description, a valid type (user|feedback|project|reference — accepted either as
# `metadata:\n  type:` or a top-level `type:`), a non-empty body, and an entry in
# memory/MEMORY.md (the index CLAUDE.md says is loaded each session).
#
# Exit 0 clean, exit 1 with a report — auto-discovered by scripts/stop-checks.sh via the
# tools/check-*.sh glob.
#
# PLACEMENT: memory/*.md | needs YAML frontmatter with name/description/valid type, non-empty body, and an entry in memory/MEMORY.md
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

MEM_DIR="memory"
INDEX="$MEM_DIR/MEMORY.md"
if [[ ! -d "$MEM_DIR" || ! -f "$INDEX" ]]; then
  echo "check-memory-hygiene: MISCONFIGURED — $MEM_DIR/ or $INDEX missing (moved?); refusing vacuous pass" >&2
  exit 1
fi

# Allowed memory types (CLAUDE.md memory section).
allowed_type() { case "$1" in user|feedback|project|reference) return 0;; *) return 1;; esac; }

files=()
while IFS= read -r f; do files+=("$f"); done < <(find "$MEM_DIR" -maxdepth 1 -name '*.md' ! -name 'MEMORY.md' | sort)

if [[ ${#files[@]} -eq 0 ]]; then
  echo "check-memory-hygiene: MISCONFIGURED — no memory/*.md files found; refusing vacuous pass" >&2
  exit 1
fi

# All five checks in ONE python3 pass over ALL files, instead of ~8 process spawns
# (basename/head/awk/grep×3/tr) PER file (~67 files ≈ 500 processes). Each check below is
# a line-for-line port of the shell/awk/grep/sed it replaces — same regexes, same fence
# semantics, same "first match wins" for type — so a file that failed before fails the
# same way now. One unreadable file is caught and reported per-file rather than aborting
# the batch, mirroring the old per-file `[[ -f ]]`/redirection guards.
python3 -c "
import re, sys

files = sys.argv[1:-1]
index_path = sys.argv[-1]
with open(index_path, 'r', errors='replace') as fh:
    index_content = fh.read()

fence_re = re.compile(r'^---[ \t]*\$')
name_re = re.compile(r'^name:[ \t]*\S')
desc_re = re.compile(r'^description:[ \t]*\S')
type_re = re.compile(r'^[ \t]*type:[ \t]*(.*)\$')
allowed = {'user', 'feedback', 'project', 'reference'}

fail = 0
for path in files:
    base = path.rsplit('/', 1)[-1]
    try:
        with open(path, 'r', errors='replace') as fh:
            lines = fh.read().splitlines()
    except Exception as e:
        print(f'MEMORY HYGIENE: {base} could not be read: {e}')
        fail = 1
        continue

    # 1) frontmatter fence on line 1
    if not lines or not fence_re.match(lines[0]):
        print(f\"MEMORY HYGIENE: {base} has no YAML frontmatter (first line is not '---') — raw content is not a memory.\")
        fail = 1
        continue

    # Extract the frontmatter block (between the first two --- fences), same as the old
    # 'NR==1{next} /^---\$/{exit} {print}' awk: lines after line 1, up to (excluding) the
    # next fence line, stopping at the first one found.
    fence_idx = None
    for i in range(1, len(lines)):
        if fence_re.match(lines[i]):
            fence_idx = i
            break
    fm_lines = lines[1:fence_idx] if fence_idx is not None else lines[1:]
    body_lines = lines[fence_idx + 1:] if fence_idx is not None else []

    # 2) name + description present and non-empty
    if not any(name_re.match(l) for l in fm_lines):
        print(f\"MEMORY HYGIENE: {base} frontmatter is missing a non-empty 'name:'.\")
        fail = 1
    if not any(desc_re.match(l) for l in fm_lines):
        print(f\"MEMORY HYGIENE: {base} frontmatter is missing a non-empty 'description:'.\")
        fail = 1

    # 3) a valid type (nested under metadata: or top-level — leading whitespace allowed on
    # purpose, so 'metadata:\n  type: x' and a top-level 'type: x' both match), value in
    # the allowed set. First match wins, same as the old 'head -1'.
    tval = ''
    for l in fm_lines:
        m = type_re.match(l)
        if m:
            tval = re.sub(r'\s+', '', m.group(1))
            break
    if not tval:
        print(f'MEMORY HYGIENE: {base} has no \'type:\' in frontmatter (need one of user|feedback|project|reference).')
        fail = 1
    elif tval not in allowed:
        print(f\"MEMORY HYGIENE: {base} has type '{tval}' — not one of user|feedback|project|reference.\")
        fail = 1

    # 4) non-empty body after the closing fence
    if not any(l.strip() for l in body_lines):
        print(f'MEMORY HYGIENE: {base} has empty body after frontmatter — a memory needs content.')
        fail = 1

    # 5) indexed in MEMORY.md (by filename)
    if base not in index_content:
        print(f'MEMORY HYGIENE: {base} is not referenced in {index_path} — unindexed memory is invisible to the session.')
        fail = 1

sys.exit(fail)
" "${files[@]}" "$INDEX"
exit $?
