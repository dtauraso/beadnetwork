#!/usr/bin/env bash

# Sourced by tools/network/trace/check-breadcrumb-not-gated.sh. Shared machinery both
# per-identifier audits use: find the files mentioning an identifier, then narrow to the
# lines where it actually GUARDS code (if/&&/||), not just comments it.

# find_ident_files IDENT — prints matching non-test .go files under nodes/Buffer/Trace,
# one per line (nothing if none found).
find_ident_files() {
  local ident="$1"
  grep -rl "$ident" --include="*.go" nodes Buffer Trace 2>/dev/null | grep -v '_test\.go$' || true
}

# find_guarding_hits IDENT FILE... — prints only the "grep -Hn" hit lines whose code
# actually guards on IDENT, skipping comment-only mentions.
find_guarding_hits() {
  local ident="$1"; shift
  local files=("$@")
  [ ${#files[@]} -eq 0 ] && return 0

  local all_hits=()
  while IFS= read -r line; do
    [ -n "$line" ] && all_hits+=("$line")
  done < <(grep -Hn "$ident" "${files[@]}" 2>/dev/null || true)

  local h content trimmed
  for h in "${all_hits[@]}"; do
    content="${h#*:}"
    content="${content#*:}"
    trimmed="$(echo "$content" | sed -e 's/^[[:space:]]*//')"
    case "$trimmed" in
      //*) continue ;;
    esac
    case "$content" in
      *"if "*"$ident"*|*"$ident"*"&&"*|*"&&"*"$ident"*|*"$ident"*"||"*|*"||"*"$ident"*)
        printf '%s\n' "$h" ;;
    esac
  done
}
