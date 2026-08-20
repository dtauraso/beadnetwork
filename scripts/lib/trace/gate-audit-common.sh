#!/usr/bin/env bash

find_ident_files() {
  local ident="$1"
  grep -rl "$ident" --include="*.go" src/Node src/NodeKinds src/Ring/Bead src/schema/buffer-layout src/schema/buffer-layout 2>/dev/null | grep -v '_test\.go$' || true
}

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
