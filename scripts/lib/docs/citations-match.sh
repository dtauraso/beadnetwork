#!/usr/bin/env bash

check_citations_match() {
  local HITS=0
  local path line doc quoted blob needle
  while IFS=$'\t' read -r path line doc quoted; do
    case "$doc" in
      CLAUDE) blob="$TMP/claude.txt" ;;
      MODEL)  blob="$TMP/model.txt" ;;
      *) continue ;;
    esac

    needle=${quoted//[*\`]/}
    needle=${needle//$'\t'/ }
    needle=${needle//$'\n'/ }
    while [[ $needle == *"  "* ]]; do needle=${needle//  / }; done

    if ! grep -qiF -- "$needle" "$blob"; then
      if [[ $HITS -eq 0 ]]; then
        echo "doc-citations: a citation quotes text that is NOT in the doc it cites:"
        echo ""
      fi
      echo "  $path:$line"
      echo "      cites $doc.md \"$quoted\""
      echo "      but that text does not appear in $doc.md"
      HITS=$((HITS + 1))
    fi
  done < "$TMP/cites.txt"

  if [[ $HITS -eq 0 ]]; then
    echo "doc-citations: clean ($(wc -l < "$TMP/cites.txt" | tr -d ' ') citations checked)"
    return 0
  fi

  echo ""
  echo "doc-citations: $HITS bad citation(s)"
  echo ""
  echo "  Either quote the doc verbatim, or drop the citation and state the point directly."
  echo "  A paraphrase presented as a citation is how a retired rule gets re-imposed as live"
  echo "  doctrine (see this script's header). If the doc changed, the citer must change too."
  return 1
}
