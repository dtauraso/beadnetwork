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

fail=0
for f in "${files[@]}"; do
  base="$(basename "$f")"
  # 1) frontmatter fence on line 1
  if [[ "$(head -1 "$f")" != "---" ]]; then
    echo "MEMORY HYGIENE: $base has no YAML frontmatter (first line is not '---') — raw content is not a memory."
    fail=1; continue
  fi
  # Extract the frontmatter block (between the first two --- fences).
  fm="$(awk 'NR==1{next} /^---[[:space:]]*$/{exit} {print}' "$f")"
  # 2) name + description present and non-empty
  for key in name description; do
    if ! printf '%s\n' "$fm" | grep -qE "^${key}:[[:space:]]*\S"; then
      echo "MEMORY HYGIENE: $base frontmatter is missing a non-empty '${key}:'."
      fail=1
    fi
  done
  # 3) a valid type (nested under metadata: or top-level), value in the allowed set
  tval="$(printf '%s\n' "$fm" | grep -E "^[[:space:]]*type:[[:space:]]*" | head -1 | sed -E 's/^[[:space:]]*type:[[:space:]]*//' | tr -d '[:space:]')"
  if [[ -z "$tval" ]]; then
    echo "MEMORY HYGIENE: $base has no 'type:' in frontmatter (need one of user|feedback|project|reference)."
    fail=1
  elif ! allowed_type "$tval"; then
    echo "MEMORY HYGIENE: $base has type '$tval' — not one of user|feedback|project|reference."
    fail=1
  fi
  # 4) non-empty body after the closing fence
  body="$(awk 'f{print} /^---[[:space:]]*$/{c++} c==2{f=1}' "$f" | grep -c .)"
  if [[ "$body" -eq 0 ]]; then
    echo "MEMORY HYGIENE: $base has empty body after frontmatter — a memory needs content."
    fail=1
  fi
  # 5) indexed in MEMORY.md (by filename)
  if ! grep -qF "$base" "$INDEX"; then
    echo "MEMORY HYGIENE: $base is not referenced in $INDEX — unindexed memory is invisible to the session."
    fail=1
  fi
done

exit $fail
